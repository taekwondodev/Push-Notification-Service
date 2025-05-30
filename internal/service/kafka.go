package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/config"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

type KafkaServiceInterface interface {
	PublishNotification(ctx context.Context, notification *models.Notification) error
	StartConsumer(ctx context.Context) error
}

type KafkaService struct {
	config   *config.KafkaConfig
	hub      *websocket.Hub
	notifSvc *NotificationService
}

func NewKafkaService(cfg *config.KafkaConfig, hub *websocket.Hub, notifSvc *NotificationService) *KafkaService {
	return &KafkaService{
		config:   cfg,
		hub:      hub,
		notifSvc: notifSvc,
	}
}

func (k *KafkaService) PublishNotification(ctx context.Context, notification *models.Notification) error {
	writer := kafka.Writer{
		Addr:     kafka.TCP(k.config.Brokers[0]),
		Topic:    k.config.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	notification.CreatedAt = time.Now().Unix()
	notification.Read = false

	msg, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(notification.Receiver),
		Value: msg,
	})
	if err != nil {
		return err
	}

	return nil
}

func (k *KafkaService) StartConsumer(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.config.Brokers,
		Topic:       k.config.Topic,
		GroupID:     k.config.GroupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	log.Printf("Kafka consumer started for topic: %s", k.config.Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer shutting down...")
			return ctx.Err()
		default:
			if err := k.processMessage(ctx, reader); err != nil {
				continue
			}
		}
	}
}

func (k *KafkaService) processMessage(ctx context.Context, reader *kafka.Reader) error {
	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		return err
	}

	var notif models.Notification
	if err := json.Unmarshal(msg.Value, &notif); err != nil {
		return err
	}

	if err := k.notifSvc.CreateNotification(ctx, &notif); err != nil {
		return err
	}

	k.hub.SendToUser(notif.Receiver, notif)
	log.Printf("Notification processed for user: %s", notif.Receiver)
	return nil
}
