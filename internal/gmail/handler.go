package gmail

import (
	"net/http"
	"strings"

	"shakehandz-api/internal/auth"
	msg "shakehandz-api/internal/shared/message"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gmail同期API: id_tokenからユーザー解決→DBのrefreshを使ってFetch
func NewGmailHandler(f msg.MessageFetcher, db *gorm.DB) gin.HandlerFunc {
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

		// 5) 取得（クエリや件数はそのまま）
		msgs, err := f.FetchMsg(ctx, enc, "has:attachment", 100)
		if err != nil {
			// 失効や撤回が疑われる場合は再同意を促す
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "reauthorize": true})
			return
		}
		c.JSON(http.StatusOK, msgs)
	}
}
