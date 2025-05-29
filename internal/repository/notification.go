package repository

import (
	"context"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository interface {
    Save(ctx context.Context, notification *models.Notification) error
    FindByReceiver(ctx context.Context, receiver string) ([]models.Notification, error)
    MarkAsRead(ctx context.Context, id string) error
    Close() error
}

type mongoNotificationRepository struct {
    client     *mongo.Client
    collection *mongo.Collection
}

func NewMongoNotificationRepository(uri, database string) (NotificationRepository, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOpts := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOpts)
    if err != nil {
        return nil, err
    }

    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }

    collection := client.Database(database).Collection("notifications")

    return &mongoNotificationRepository{
        client:     client,
        collection: collection,
    }, nil
}

func (r *mongoNotificationRepository) Save(ctx context.Context, notification *models.Notification) error {
    _, err := r.collection.InsertOne(ctx, notification)
    return err
}

func (r *mongoNotificationRepository) FindByReceiver(ctx context.Context, receiver string) ([]models.Notification, error) {
    cursor, err := r.collection.Find(ctx, bson.M{"receiver": receiver})
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

func (r *mongoNotificationRepository) MarkAsRead(ctx context.Context, id string) error {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return err
    }

    _, err = r.collection.UpdateOne(
        ctx,
        bson.M{"_id": objID},
        bson.M{"$set": bson.M{"read": true}},
    )
    return err
}

func (r *mongoNotificationRepository) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return r.client.Disconnect(ctx)
}