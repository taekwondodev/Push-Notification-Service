package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	ws "github.com/taekwondodev/push-notification-service/internal/websocket"
)

type WebSocketController struct {
	hub *ws.Hub
}

const (
	readTimeout  = 60 * time.Second
	pingInterval = 30 * time.Second
)

func NewWebSocketController(hub *ws.Hub) *WebSocketController {
	return &WebSocketController{
		hub: hub,
	}
}

func (h *WebSocketController) HandleConnection(w http.ResponseWriter, r *http.Request) error {
	username := r.URL.Query().Get("username")
	if username == "" {
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
	defer h.hub.Unregister(user)

	h.setupConnectionTimeouts(conn)

	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	done := make(chan struct{})
	go h.readMessageLoop(user, conn, done)

	h.pingLoop(conn, done, pingTicker)
}

func (h *WebSocketController) setupConnectionTimeouts(conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(readTimeout))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})
}

func (h *WebSocketController) readMessageLoop(user string, conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.handleReadError(user, err)
			return
		}

		h.handleMessage(user, message)
		conn.SetReadDeadline(time.Now().Add(readTimeout))
	}
}

func (h *WebSocketController) handleMessage(user string, message []byte) {
	log.Printf("Received message from user %s: %s", user, string(message))
}

func (h *WebSocketController) handleReadError(user string, err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("WebSocket error for user %s: %v", user, err)
	}
}

func (h *WebSocketController) pingLoop(conn *websocket.Conn, done chan struct{}, pingTicker *time.Ticker) {
	for {
		select {
		case <-done:
			return
		case <-pingTicker.C:
			if err := h.sendPing(conn); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketController) sendPing(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.PingMessage, nil)
}
