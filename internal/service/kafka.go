package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/taekwondodev/push-notification-service/internal/config"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/websocket"
)

type KafkaService struct {
    config  *config.KafkaConfig
    hub     *websocket.Hub
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

    err = writer.WriteMessages(ctx, kafka.Message{Value: msg})
    if err != nil {
        return err
    }

    return nil
}

func (k *KafkaService) StartConsumer(ctx context.Context) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: k.config.Brokers,
        Topic:   k.config.Topic,
        GroupID: k.config.GroupID,
    })
    defer reader.Close()
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            msg, err := reader.ReadMessage(context.Background())
            if err != nil {
                continue
            }

            var notif models.Notification
            if err := json.Unmarshal(msg.Value, &notif); err != nil {
                continue
            }

            if err := k.notifSvc.CreateNotification(context.Background(), &notif); err != nil {
                // return err
            }

            k.hub.SendToUser(notif.Receiver, notif)
        }
    }
}