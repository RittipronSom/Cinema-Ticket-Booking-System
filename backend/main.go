package main

import (
	"net/http"

	"cinema-ticket-booking-system/internal/database"
	"cinema-ticket-booking-system/internal/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	
	err := database.ConnectMongo()

	if err != nil {
		panic(err)
	}
	r := gin.Default()	

	// เปิดใช้งาน CORS
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Cinema Booking Backend Running",
		})
	})

	r.GET("/seats", handler.GetSeats)
	r.POST("/lock-seat", handler.LockSeat)
	r.GET("/ws", handler.HandleWebSocket)
	r.POST("/confirm-booking", handler.ConfirmBooking)
	r.GET("/bookings", handler.GetBookings)

	r.Run(":8080")
}
