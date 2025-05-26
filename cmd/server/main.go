package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

func main() {
	hub := websocket.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := hub.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrade WebSocket:", err)
			return
		}
		hub.Register <- conn

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					hub.Unregister <- conn
					break
				}
			}
		}()
	})

	go consumeKafka(hub)

	log.Println("WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func consumeKafka(hub *websocket.Hub) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "notifications",
		GroupID: "websocket-notifier",
	})

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

		log.Printf("Notification received: %s\n", notif.Message)
		hub.Broadcast <- msg.Value
	}
}
