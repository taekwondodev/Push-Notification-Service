package db

import (
	"context"
	"log"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewNotificationRepository() (*NotificationRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database("notificationsdb").Collection("notifications")

	log.Println("MongoDB connected")

	return &NotificationRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (r *NotificationRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Disconnect(ctx)
}

func (r *NotificationRepository) SaveNotification(notif *models.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, notif)
	return err
}
