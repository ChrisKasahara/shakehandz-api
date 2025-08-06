package gemini

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Geminiサービス呼び出し用ハンドラ
func NewStructureWithGeminiHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Bearer トークン抽出
		var accessToken string
		if _, err := fmt.Sscanf(c.GetHeader("Authorization"), "Bearer %s", &accessToken); err != nil || accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
			return
		}

		msgs, err := svc.Run(c, accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}
