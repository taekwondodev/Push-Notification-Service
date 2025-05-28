package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/db"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

var hub *websocket.Hub
var repo *db.NotificationRepository

func main() {
	hub := websocket.NewHub()
	repo, err := db.NewNotificationRepository()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer repo.Close()

	go consumeKafka(hub, repo)

	http.HandleFunc("/ws", handleWebSocket)

	http.HandleFunc("/notifications", getNotificationsHandler)

	log.Println("WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func consumeKafka(hub *websocket.Hub, db *db.NotificationRepository) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "notifications",
		GroupID: "websocket-notifier",
	})
	defer reader.Close()

	log.Println("Kafka consumer started...")
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("error Kafka:", err)
			continue
		}

		var notif models.Notification
		if err := json.Unmarshal(msg.Value, &notif); err != nil {
			log.Println("error parsing notification:", err)
			continue
		}

		notif.CreatedAt = time.Now().Unix()

		if err := db.SaveNotification(context.Background(), &notif); err != nil {
			log.Println("error saving notification to DB:", err)
		}

		hub.SendToUser(notif.Receiver, notif)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "need user", http.StatusBadRequest)
		return
	}

	conn, err := hub.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	hub.Register(user, conn)

	log.Println("User connected via WebSocket:", user)
}

func getNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "need user", http.StatusBadRequest)
		return
	}

	notifications, err := repo.GetNotifications(r.Context(), user)
	if err != nil {
		http.Error(w, "failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
