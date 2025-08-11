package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RequireInternalCall() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("BACKEND_INTERNAL_TOKEN")
		authHeader := c.GetHeader("X-App-Auth")
		if secret == "" || authHeader != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
