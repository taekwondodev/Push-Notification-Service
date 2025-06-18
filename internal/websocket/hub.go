package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/taekwondodev/push-notification-service/internal/models"
)

type Hub struct {
	Upgrader   websocket.Upgrader
	clients    map[string]*websocket.Conn
	register   chan *clientConnection
	unregister chan string
	broadcast  chan *broadcastMessage
	shutdown   chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
}

type clientConnection struct {
	user string
	conn *websocket.Conn
}

type broadcastMessage struct {
	user    string
	message models.Notification
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &Hub{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients:    make(map[string]*websocket.Conn),
		register:   make(chan *clientConnection, 100),
		unregister: make(chan string, 100),
		broadcast:  make(chan *broadcastMessage, 1000),
		shutdown:   make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}

	go hub.run()
	return hub
}

func (h *Hub) run() {
	defer h.cleanup()

	for {
		select {
		case client := <-h.register:
			h.handleClientRegistration(client)

		case user := <-h.unregister:
			h.handleClientUnregistration(user)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)

		case <-h.shutdown:
			return

		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) Register(user string, conn *websocket.Conn) {
	select {
	case h.register <- &clientConnection{user: user, conn: conn}:
	case <-h.ctx.Done():
		conn.Close()
	}
}

func (h *Hub) handleClientRegistration(client *clientConnection) {
	if existingConn, exists := h.clients[client.user]; exists {
		existingConn.Close()
		log.Printf("[%s] existing connection replaced\n", client.user)
	}

	h.clients[client.user] = client.conn
	log.Printf("[%s] connected\n", client.user)
}

func (h *Hub) Unregister(user string) {
	select {
	case h.unregister <- user:
	case <-h.ctx.Done():
	}
}

func (h *Hub) handleClientUnregistration(user string) {
	if conn, exists := h.clients[user]; exists {
		conn.Close()
		delete(h.clients, user)
		log.Printf("[%s] disconnected\n", user)
	}
}

func (h *Hub) SendToUser(user string, message models.Notification) {
	select {
	case h.broadcast <- &broadcastMessage{user: user, message: message}:
	case <-h.ctx.Done():
	default:
		log.Printf("broadcast channel full, dropping message for user %s", user)
	}
}

func (h *Hub) handleBroadcast(msg *broadcastMessage) {
	conn, exists := h.clients[msg.user]
	if !exists {
		return
	}

	if err := conn.WriteJSON(msg.message); err != nil {
		log.Printf("error sending message to user %s: %v", msg.user, err)
		h.removeFailedConnection(msg.user, conn)
	}
}

func (h *Hub) Shutdown() {
	h.cancel()
	close(h.shutdown)
}

func (h *Hub) cleanup() {
	for user, conn := range h.clients {
		conn.Close()
		log.Printf("[%s] disconnected (shutdown)\n", user)
	}
	close(h.register)
	close(h.unregister)
	close(h.broadcast)
}

func (h *Hub) removeFailedConnection(user string, conn *websocket.Conn) {
	conn.Close()
	delete(h.clients, user)
}
