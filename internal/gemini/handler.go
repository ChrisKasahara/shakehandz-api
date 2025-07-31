package gemini

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageFetcherをDI
func NewExtractDataFromMessageWithGemini(svc *GmailMessageFetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Authorization ヘッダーから Bearer トークン取得
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

		n, err := svc.Fetch(ctx, accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process mails"})
			return
		}

		// TODO: ここで取得したメールを解析キューに積む処理を実装予定
		//       解析ロジック・Gemini呼び出しは未実装

		c.JSON(http.StatusAccepted, gin.H{
			"queued": true,
			"count":  n,
		})
	}
}
