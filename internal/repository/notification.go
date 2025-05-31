package repository

import (
	"context"
	"log"
	"time"

	"github.com/taekwondodev/push-notification-service/internal/customerrors"
	"github.com/taekwondodev/push-notification-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository interface {
	Save(ctx context.Context, notification *models.Notification) error
	FindByReceiver(ctx context.Context, receiver string, unreadOnly bool) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	Close() error
}

type mongoNotificationRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoNotificationRepository(uri, database string) NotificationRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("failed to connect to database", "error", err)
	}

	collection := client.Database(database).Collection("notifications")

	repo := &mongoNotificationRepository{
		client:     client,
		collection: collection,
	}

	if err := repo.createIndexes(ctx); err != nil {
		log.Printf("Warning: failed to create indexes: %v", err)
	}

	return repo
}

func (r *mongoNotificationRepository) Save(ctx context.Context, notification *models.Notification) error {
	_, err := r.collection.InsertOne(ctx, notification)
	return err
}

func (r *mongoNotificationRepository) FindByReceiver(ctx context.Context, receiver string, unreadOnly bool) ([]models.Notification, error) {
	mongoFilter := r.buildMongoFilter(receiver, unreadOnly)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
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

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"read": true}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return customerrors.ErrNotificationNotFound
	}

	return nil
}

func (r *mongoNotificationRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

func (r *mongoNotificationRepository) buildMongoFilter(receiver string, unread bool) bson.M {
	mongoFilter := bson.M{}
	mongoFilter["receiver"] = receiver

	if unread {
		mongoFilter["$or"] = []bson.M{
			{"read": false},
			{"read": bson.M{"$exists": false}},
		}
	}

	return mongoFilter
}

func (r *mongoNotificationRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "receiver", Value: 1}, {Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "sender", Value: 1}, {Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "receiver", Value: 1}, {Key: "read", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
