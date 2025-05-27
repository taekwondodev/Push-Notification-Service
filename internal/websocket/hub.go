package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Upgrader websocket.Upgrader
	clients  map[string][]*websocket.Conn
	mu       sync.Mutex
}

func NewHub() *Hub {
	var up = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	return &Hub{
		Upgrader: up,
		clients:  make(map[string][]*websocket.Conn),
	}
}

func (h *Hub) Register(user string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[user] = append(h.clients[user], conn)
	log.Printf("[%s] connected (%d conn)\n", user, len(h.clients[user]))
}

func (h *Hub) Unregister(user string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.clients[user]
	for i, c := range conns {
		if c == conn {
			h.clients[user] = append(conns[:i], conns[i+1:]...)
			c.Close()
			break
		}
	}
	if len(h.clients[user]) == 0 {
		delete(h.clients, user)
		log.Printf("[%s] disconnected\n", user)
	}
}

func (h *Hub) SendToUser(user string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.clients[user]
	if !ok {
		log.Printf("User [%s] not connected, notifica persa\n", user)
		return
	}

	for _, conn := range conns {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("error WebSocket:", err)
			conn.Close()
		}
	}
}
