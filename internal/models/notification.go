package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID      primitive.ObjectID `bson:"_id,omitzero" json:"id"`
	From    string             `bson:"from" json:"from"`
	To      string             `bson:"to" json:"to"`
	Message string             `bson:"message" json:"message"`
}
