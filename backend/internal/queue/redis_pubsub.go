package queue

import (
    "context"
    "encoding/json"
    redisClient "cinema-ticket-booking-system/internal/redis"
)

type BookingEvent struct {
    UserID     string `json:"user_id"`
    SeatNumber string `json:"seat_number"`
    Event      string `json:"event"`
}

func PublishBookingEvent(userID string, seatNumber string, event string) error {
    payload := BookingEvent{
        UserID:     userID,
        SeatNumber: seatNumber,
        Event:      event,
    }

    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    return redisClient.Client.Publish(
        context.Background(),
        "booking_events",
        string(data),
    ).Err()
}