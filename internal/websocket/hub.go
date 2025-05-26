package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Upgrader   websocket.Upgrader
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewHub() *Hub {
	var up = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	return &Hub{
		Upgrader:   up,
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.mu.Lock()
			h.Clients[conn] = true
			h.mu.Unlock()
			log.Println("New client registered")

		case conn := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
				log.Println("Client unregistered")
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for conn := range h.Clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("error message sent:", err)
					conn.Close()
					delete(h.Clients, conn)
				}
			}
			h.mu.Unlock()
		}
	}
}
