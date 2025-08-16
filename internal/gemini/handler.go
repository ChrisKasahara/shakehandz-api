package gemini

import (
	"net/http"

	sa "shakehandz-api/internal/shared/auth"
	"shakehandz-api/internal/shared/message/gmail"

	"github.com/gin-gonic/gin"
)

// Geminiサービス呼び出し用ハンドラ
func StructureWithGeminiHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		verified, err := sa.IsUserVerified(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cli, _ := NewGeminiClientWithRefresh(ctx, "models/gemini-2.5-flash", verified.Token.RefreshToken)

		gmail_svc, err := gmail.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gmail service"})
			return
		}

		ok, err := svc.Run(c, cli, gmail_svc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ok)
	}
}
