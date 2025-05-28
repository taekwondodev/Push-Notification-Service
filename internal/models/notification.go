package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitzero"`
	Sender    string             `json:"sender" bson:"sender"`
	Receiver  string             `json:"receiver" bson:"receiver"`
	Message   string             `json:"message" bson:"message"`
	Read      bool               `json:"read" bson:"read,omitzero"`
	CreatedAt int64              `json:"createdAt" bson:"createdAt,omitzero"`
}
