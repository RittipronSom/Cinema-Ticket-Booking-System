package queue

import (
    "context"
    "fmt"
    redisClient "cinema-ticket-booking-system/internal/redis"
)

func StartSubscriber() {
    pubsub := redisClient.Client.Subscribe(
        context.Background(),
        "booking_events",
    )

    fmt.Println("Queue subscriber started...")

    go func() {
        for msg := range pubsub.Channel() {
            fmt.Println("📨 Booking event received:", msg.Payload)
            // mock notification ตรงนี้
        }
    }()
}