package gmail

import (
	"net/http"
	"strings"

	"shakehandz-api/internal/auth"
	gmsg "shakehandz-api/internal/shared/message/gmail"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gmail同期API: id_tokenからユーザー解決→DBのrefreshを使ってFetch
func NewGmailHandler(svc *GmailMsgService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1) Authorization: Bearer <id_token>
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing id_token"})
			return
		}
		idToken := strings.TrimPrefix(authz, "Bearer ")

		// 2) id_token → User
		user, err := auth.UserFromIDToken(ctx, db, idToken)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 3) DBから暗号化refresh_token（[]byte）取得
		enc, err := auth.FindGoogleRefreshTokenEncByUserID(db, user.ID.String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if len(enc) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "no_refresh_token", "reauthorize": true})
			return
		}

		refresh_gmail_svc, err := gmsg.NewServiceWithRefresh(ctx, enc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gmail service"})
			return
		}

		// Runでメッセージ取得
		msgs, err := svc.Run(ctx, refresh_gmail_svc, c.Query("query"), 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}
