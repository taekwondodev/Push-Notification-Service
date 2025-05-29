package controller

import (
	"net/http"

	"github.com/gorilla/websocket"
	ws "github.com/taekwondodev/push-notification-service/internal/websocket"
)

type WebSocketController struct {
    hub    *ws.Hub
}

func NewWebSocketController(hub *ws.Hub) *WebSocketController {
    return &WebSocketController{
        hub:    hub,
    }
}

func (h *WebSocketController) HandleConnection(w http.ResponseWriter, r *http.Request) {
    user := r.URL.Query().Get("user")
    if user == "" {
        http.Error(w, "user parameter required", http.StatusBadRequest)
        return
    }

    conn, err := h.hub.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    h.hub.Register(user, conn)

    go h.handleConnectionLifecycle(user, conn)
}

func (h *WebSocketController) handleConnectionLifecycle(user string, conn *websocket.Conn) {
    defer h.hub.Unregister(user)
    
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }
}