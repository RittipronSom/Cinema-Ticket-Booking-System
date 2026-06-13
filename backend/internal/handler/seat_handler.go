package handler

import (
	"context"
	"net/http"
	"time"
	"cinema-ticket-booking-system/internal/queue"
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
		Status:     "AVAILABLE",
	},
	{
		ID:         3,
		SeatNumber: "A3",
		Status:     "AVAILABLE",
	},
	{
		ID:         4,
		SeatNumber: "A4",
		Status:     "AVAILABLE",
	},
}

func GetSeats(c *gin.Context) {
    collection := database.DB.Collection("bookings")

    // ดึงที่นั่งที่ BOOKED จาก MongoDB
    cursor, err := collection.Find(
        context.Background(),
        map[string]interface{}{"status": "BOOKED"},
    )

    bookedSeats := map[string]bool{}
    if err == nil {
        var bookings []models.Booking
        cursor.All(context.Background(), &bookings)
        for _, b := range bookings {
            bookedSeats[b.SeatNumber] = true
        }
    }

    for i, seat := range seats {
        if bookedSeats[seat.SeatNumber] {
            seats[i].Status = "BOOKED"
            continue
        }

        // เช็ค Redis lock
        lockKey := "seat:" + seat.SeatNumber
        exists, err := redisClient.Client.Exists(redisClient.Ctx, lockKey).Result()
        if err == nil && exists == 1 {
            seats[i].Status = "LOCKED"
        } else {
            seats[i].Status = "AVAILABLE"
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
	for _, seat := range seats {

		if seat.SeatNumber == request.SeatNumber &&
			seat.Status == "BOOKED" {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Seat already booked",
			})

			return
		}
	}

	lockKey := "seat:" + request.SeatNumber

	success, err := redisClient.Client.SetNX(
		redisClient.Ctx,
		lockKey,
		"locked",
		5*time.Minute, // กำหนดเวลาล็อค 5 นาที
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
		SaveAuditLog(
			"LOCK_FAILED",
			request.SeatNumber,
			"User tried locking already locked seat",
		)
		return
	}

	for i, seat := range seats {
		if seat.SeatNumber == request.SeatNumber {
			seats[i].Status = "LOCKED"
		}
	}
	SaveAuditLog(
		"SEAT_LOCKED",
		request.SeatNumber,
		"Seat locked successfully",
	)
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

	for _, seat := range seats {

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
				SaveAuditLog(
					"BOOKING_FAILED",
					request.SeatNumber,
					"User tried booking expired lock seat",
				)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "ที่นั่งนี้หมดเวลาในการจองแล้ว กรุณาลองใหม่อีกครั้ง",
				})

				return
			}

		}
	}

	collection := database.DB.Collection("bookings")

	existingBooking := collection.FindOne(
		context.Background(),
		gin.H{
			"seat_number": request.SeatNumber,
			"status":      "BOOKED",
		},
	)

	if existingBooking.Err() == nil {
		SaveAuditLog(
			"BOOKING_FAILED",
			request.SeatNumber,
			"ที่นั่งนี้ถูกจองแล้ว",
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ที่นั่งนี้ถูกจองแล้ว",
		})

		return
	}

	booking := models.Booking{
		UserID:     request.UserID,
		SeatNumber: request.SeatNumber,
		Status:     "BOOKED",
	}

	_, err := collection.InsertOne(
		context.Background(),
		booking,
	)
	queue.PublishBookingEvent(request.UserID, request.SeatNumber, "BOOKING_SUCCESS")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save booking",
		})
		return
	}
	for i, seat := range seats {

		if seat.SeatNumber == request.SeatNumber {

			seats[i].Status = "BOOKED"
		}
	}
	lockKey := "seat:" + request.SeatNumber
	redisClient.Client.Del(
		redisClient.Ctx,
		lockKey,
	)
	SaveAuditLog(
		"BOOKING_SUCCESS",
		request.SeatNumber,
		"Booking completed successfully",
	)
	
	BroadcastSeatUpdate()

	c.JSON(http.StatusOK, gin.H{
		"message": "จองที่นั่งสำเร็จ",
	})
}
func GetBookings(c *gin.Context) {

	collection := database.DB.Collection("bookings")

	cursor, err := collection.Find(
		context.Background(),
		map[string]interface{}{},
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bookings",
		})

		return
	}

	var bookings []models.Booking

	if err := cursor.All(context.Background(), &bookings); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode bookings",
		})

		return
	}

	c.JSON(http.StatusOK, bookings)
}
func StartLockExpiryWatcher() {
    go func() {
        for {
            time.Sleep(3 * time.Second) // เช็คทุก 3 วินาที

            changed := false

            for i, seat := range seats {
                if seat.Status == "LOCKED" {
                    lockKey := "seat:" + seat.SeatNumber

                    exists, err := redisClient.Client.Exists(
                        redisClient.Ctx,
                        lockKey,
                    ).Result()

                    if err != nil {
                        continue
                    }

                    // Redis key หมดอายุแล้ว → คืนสถานะ
                    if exists == 0 {
                        seats[i].Status = "AVAILABLE"
                        changed = true
                        SaveAuditLog(
                            "SEAT_RELEASED",
                            seat.SeatNumber,
                            "Lock expired, seat released",
                        )
                    }
                }
            }

            if changed {
                BroadcastSeatUpdate() // แจ้ง frontend ทันที
            }
        }
    }()
}
func UnlockSeat(c *gin.Context) {
    var request struct {
        SeatNumber string `json:"seat_number"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    lockKey := "seat:" + request.SeatNumber
    redisClient.Client.Del(redisClient.Ctx, lockKey)

    for i, seat := range seats {
        if seat.SeatNumber == request.SeatNumber && seat.Status == "LOCKED" {
            seats[i].Status = "AVAILABLE"
        }
    }

    SaveAuditLog("SEAT_UNLOCKED", request.SeatNumber, "User cancelled lock")
    BroadcastSeatUpdate()

    c.JSON(http.StatusOK, gin.H{"message": "Seat unlocked"})
}
func SaveAuditLog(
	event string,
	seatNumber string,
	description string,
) {

	collection := database.DB.Collection("audit_logs")

	log := models.AuditLog{
		Event:       event,
		SeatNumber:  seatNumber,
		Description: description,
	}

	collection.InsertOne(
		context.Background(),
		log,
	)
}
