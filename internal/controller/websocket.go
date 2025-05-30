package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/middleware"
	ws "github.com/taekwondodev/push-notification-service/internal/websocket"
)

type WebSocketController struct {
	hub *ws.Hub
}

func NewWebSocketController(hub *ws.Hub) *WebSocketController {
	return &WebSocketController{
		hub: hub,
	}
}

func (h *WebSocketController) HandleConnection(w http.ResponseWriter, r *http.Request) error {
	username, err := middleware.GetUsernameFromContext(r.Context())
	if err != nil {
		return customerrors.ErrBadRequest
	}

	conn, err := h.hub.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	h.hub.Register(username, conn)

	go h.handleConnectionLifecycle(username, conn)
	return nil
}

func (h *WebSocketController) handleConnectionLifecycle(user string, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		h.hub.Unregister(user)
	}()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", user, err)
			}
			return
		}

		h.handleMessage(user, message)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

func (h *WebSocketController) handleMessage(user string, message []byte) {
	log.Printf("Received message from user %s: %s", user, string(message))
}
