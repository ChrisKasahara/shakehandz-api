package message

import (
	"net/http"

	sa "shakehandz-api/internal/shared/auth"
	gmsg "shakehandz-api/internal/shared/message/gmail"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gmail同期API: id_tokenからユーザー解決→DBのrefreshを使ってFetch
// 新しいBFFアーキテクチャに合わせて認証方法を変更しています
func GmailHandler(svc *GmailMsgService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		verified, err := sa.IsUserVerified(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		gmail_svc, err := gmsg.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
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
