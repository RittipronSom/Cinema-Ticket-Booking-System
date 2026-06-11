package handler

import (
	"context"
	"net/http"
	"time"

	"cinema-ticket-booking-system/internal/database"
	"cinema-ticket-booking-system/internal/models"
	redisClient "cinema-ticket-booking-system/internal/redis"

	"github.com/gin-gonic/gin"
)

var seats = []models.Seat{
	{
		ID:         1,
		SeatNumber: "A1",
		Status:     "AVAILABLE",
	},
	{
		ID:         2,
		SeatNumber: "A2",
		Status:     "BOOKED",
	},
	{
		ID:         3,
		SeatNumber: "A3",
		Status:     "LOCKED",
	},
	{
		ID:         4,
		SeatNumber: "A4",
		Status:     "AVAILABLE",
	},
}

func GetSeats(c *gin.Context) {

	for i, seat := range seats {

		lockKey := "seat:" + seat.SeatNumber

		exists, err := redisClient.Client.Exists(
			redisClient.Ctx,
			lockKey,
		).Result()

		if err != nil {
			continue
		}

		if exists == 1 {
			seats[i].Status = "LOCKED"
		} else {
			if seats[i].Status != "BOOKED" {
				seats[i].Status = "AVAILABLE"
			}
		}
	}

	c.JSON(http.StatusOK, seats)
}

func LockSeat(c *gin.Context) {
	var request struct {
		SeatNumber string `json:"seat_number"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	lockKey := "seat:" + request.SeatNumber

	success, err := redisClient.Client.SetNX(
		redisClient.Ctx,
		lockKey,
		"locked",
		10*time.Second,
	).Result()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Redis error",
		})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Seat already locked",
		})
		return
	}

	for i, seat := range seats {
		if seat.SeatNumber == request.SeatNumber {
			seats[i].Status = "LOCKED"
		}
	}
	BroadcastSeatUpdate()
	c.JSON(http.StatusOK, gin.H{
		"message": "Seat locked successfully",
	})
}
func ConfirmBooking(c *gin.Context) {

	var request struct {
		UserID     string `json:"user_id"`
		SeatNumber string `json:"seat_number"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	for i, seat := range seats {

		if seat.SeatNumber == request.SeatNumber {

			lockKey := "seat:" + request.SeatNumber

			exists, err := redisClient.Client.Exists(
				redisClient.Ctx,
				lockKey,
			).Result()

			if err != nil {

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Redis error",
				})

				return
			}

			if exists == 0 {

				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Seat is not locked",
				})

				return
			}

			seats[i].Status = "BOOKED"
			redisClient.Client.Del(
				redisClient.Ctx,
				lockKey,
			)
		}
	}

	booking := models.Booking{
		UserID:     request.UserID,
		SeatNumber: request.SeatNumber,
		Status:     "BOOKED",
	}

	collection := database.DB.Collection("bookings")

	_, err := collection.InsertOne(
		context.Background(),
		booking,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save booking",
		})
		return
	}

	BroadcastSeatUpdate()

	c.JSON(http.StatusOK, gin.H{
		"message": "จองที่นั่งสำเร็จ",
	})
}
