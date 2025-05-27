package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/db"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

func main() {
	hub := websocket.NewHub()
	repo, err := db.NewNotificationRepository()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer repo.Close()

	go consumeKafka(hub, repo)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "need user", http.StatusBadRequest)
			return
		}

		conn, err := hub.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error WebSocket:", err)
			return
		}
		hub.Register(user, conn)

		go func() {
			defer hub.Unregister(user, conn)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	})

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

		if err := db.SaveNotification(&notif); err != nil {
			log.Println("error saving notification to DB:", err)
		}

		payload, _ := json.Marshal(notif)
		hub.SendToUser(notif.To, payload)
	}
}
