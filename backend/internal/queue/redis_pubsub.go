package queue

import (
	"context"
	"fmt"
	"time"

	redisClient "cinema-ticket-booking-system/internal/redis"
)

const BookingChannel = "booking_events"

func PublishBookingEvent(message string) {

	err := redisClient.Client.Publish(
		context.Background(),
		BookingChannel,
		message,
	).Err()

	if err != nil {
		fmt.Println("Publish Error:", err)
		return
	}

	fmt.Println("Event Published:", message)
}

func StartBookingSubscriber() {

	pubsub := redisClient.Client.Subscribe(
		context.Background(),
		BookingChannel,
	)

	channel := pubsub.Channel()

	fmt.Println("Booking Subscriber Started...")

	for msg := range channel {

		fmt.Println("Processing Booking Event:", msg.Payload)

		time.Sleep(5 * time.Second)

		fmt.Println("Notification Sent:", msg.Payload)
	}
}
