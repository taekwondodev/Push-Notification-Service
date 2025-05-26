package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/models"
)

func main() {
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "notifications",
		Balancer: &kafka.LeastBytes{},
	}

	notification := models.Notification{
		From:    "Alice",
		To:      "Bob",
		Message: "Alice sent you a message!",
	}

	data, _ := json.Marshal(notification)

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(notification.To),
			Value: data,
		},
	)
	if err != nil {
		log.Fatal("error message sent:", err)
	}

	log.Println("Notification sent!")
	_ = writer.Close()
}
