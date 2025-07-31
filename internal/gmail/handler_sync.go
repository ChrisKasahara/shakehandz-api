package gmail

import (
	"fmt"
	"net/http"

	mail "shakehandz-api/internal/mail"

	"github.com/gin-gonic/gin"
)

// Gmail同期API: Fetcherを使ってメール一覧を返す
func NewSyncHandler(f mail.Fetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		var accessToken string
		_, err := fmt.Sscanf(authHeader, "Bearer %s", &accessToken)
		if err != nil || accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			return
		}
		msgs, err := f.Fetch(ctx, accessToken, "has:attachment", 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mails"})
			return
		}
		c.JSON(http.StatusOK, msgs)
	}
}
