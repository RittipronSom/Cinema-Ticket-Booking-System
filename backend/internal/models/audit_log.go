package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type AuditLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Event       string             `bson:"event" json:"event"`
	SeatNumber  string             `bson:"seat_number" json:"seat_number"`
	Description string             `bson:"description" json:"description"`
}