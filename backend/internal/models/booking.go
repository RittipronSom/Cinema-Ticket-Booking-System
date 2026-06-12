package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Booking struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"user_id" json:"user_id"`
	SeatNumber string             `bson:"seat_number" json:"seat_number"`
	Status     string             `bson:"status" json:"status"`
}
