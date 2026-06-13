package handler

import (
	"context"
	"cinema-ticket-booking-system/internal/database"
	"cinema-ticket-booking-system/internal/models"
)

func InitSeatsFromDB() {
	collection := database.DB.Collection("bookings")
	cursor, err := collection.Find(
		context.Background(),
		map[string]interface{}{"status": "BOOKED"},
	)
	if err != nil {
		return
	}

	var bookings []models.Booking
	cursor.All(context.Background(), &bookings)

	for _, b := range bookings {
		for i, seat := range seats {
			if seat.SeatNumber == b.SeatNumber {
				seats[i].Status = "BOOKED"
			}
		}
	}
}