package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)

func HandleWebSocket(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	clients[conn] = true

	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			delete(clients, conn)
			conn.Close()
			break
		}
	}
}

func BroadcastSeatUpdate() {

	for client := range clients {

		err := client.WriteJSON(gin.H{
			"message": "seat_updated",
		})

		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}