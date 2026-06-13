package main

import (
	"cinema-ticket-booking-system/internal/database"
	"cinema-ticket-booking-system/internal/handler"
	"cinema-ticket-booking-system/internal/middleware"
	"cinema-ticket-booking-system/internal/queue"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	err := database.ConnectMongo()
	if err != nil {
		panic(err)
	}

	err = middleware.InitFirebase()
	if err != nil {
		panic(err)
	}

	handler.InitSeatsFromDB()
	handler.StartLockExpiryWatcher()
	go queue.StartSubscriber()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization","X-Admin-Secret"},
		AllowCredentials: true,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Cinema Booking Backend Running",
		})	
	})

	r.GET("/seats", handler.GetSeats)
	r.GET("/ws", handler.HandleWebSocket)

	// Routes ที่ต้อง Login
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.POST("/lock-seat", handler.LockSeat)
		authGroup.POST("/confirm-booking", handler.ConfirmBooking)
	}
	authGroup.POST("/unlock-seat", handler.UnlockSeat)

	// Admin Routes
	r.GET("/bookings", middleware.AdminOnly(), handler.GetBookings)

	r.Run(":8080")
}
