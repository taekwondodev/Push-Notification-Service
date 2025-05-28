package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/taekwondodev/push-notification-service/internal/models"
)

type Hub struct {
	Upgrader websocket.Upgrader
	clients  map[string]*websocket.Conn
	mu       sync.Mutex
}

func NewHub() *Hub {
	var up = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	return &Hub{
		Upgrader: up,
		clients:  make(map[string]*websocket.Conn),
	}
}

func (h *Hub) Register(user string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[user] = conn
	log.Printf("[%s] connected\n", user)
}

func (h *Hub) SendToUser(user string, message models.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.clients[user]; ok {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("error sending message to user %s: %v", user, err)
		}
	}
}
