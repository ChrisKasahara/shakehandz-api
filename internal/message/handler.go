package message

import (
	"net/http"

	oauth "shakehandz-api/internal/shared/auth/oauth"

	"shakehandz-api/internal/shared/message/gmail"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gmail同期API: id_tokenからユーザー解決→DBのrefreshを使ってFetch
// 新しいBFFアーキテクチャに合わせて認証方法を変更しています
func MessageHandler(svc *MessageService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		verified, err := oauth.IsUserVerified(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		gmail_svc, err := gmail.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gmail service"})
			return
		}

		msgs, err := svc.Run(ctx, gmail_svc, c.Query("query"), 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}
