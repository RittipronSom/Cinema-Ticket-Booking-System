package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminSecret := c.GetHeader("X-Admin-Secret")

		if adminSecret != os.Getenv("ADMIN_SECRET") || adminSecret == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}