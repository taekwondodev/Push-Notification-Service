package db

import (
	"context"
	"log"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *NotificationRepository) SaveNotification(ctx context.Context, notif *models.Notification) error {
	_, err := r.collection.InsertOne(ctx, notif)
	return err
}

func (r *NotificationRepository) GetNotifications(ctx context.Context, user string) ([]models.Notification, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"receiver": user})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"read": true}},
	)
	if err != nil {
		return err
	}
	return nil
}
