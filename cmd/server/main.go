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
	router := http.NewServeMux()

	go consumeKafka(hub, repo)

	router.HandleFunc("GET /ws", handleWebSocket)
	router.HandleFunc("GET /notifications", getNotificationsHandler)
	router.HandleFunc("POST /notifications", postNotificationHandler)
	router.HandleFunc("POST /notifications/{id}/read", markNotificationAsRead)

	log.Println("WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
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

func postNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var n models.Notification
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	n.CreatedAt = time.Now().Unix()
	n.Read = false

	msg, _ := json.Marshal(n)
	writer := kafka.Writer{
		Addr:     kafka.TCP("kafka:9092"),
		Topic:    "notifications",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	err := writer.WriteMessages(r.Context(), kafka.Message{Value: msg})
	if err != nil {
		http.Error(w, "Kafka error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func markNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := repo.MarkAsRead(r.Context(), id); err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
