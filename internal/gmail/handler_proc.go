// Gmail メール処理キュー登録用ハンドラ
package gmail

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Fetcher interface {
	FetchMessageIDList(ctx context.Context, token *oauth2.Token, query string, max int) ([]interface{}, error)
}

// Fetcher インターフェースは fetcher.go で定義されている想定
func NewProcessHandler(f Fetcher) gin.HandlerFunc {
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

		token := &oauth2.Token{AccessToken: accessToken}

		// メール取得（has:attachment, 最大5件）
		mails, err := f.FetchMessageIDList(ctx, token, "has:attachment", 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mails"})
			return
		}
		n := len(mails)

		// TODO: ここで取得したメールを解析キューに積む処理を実装予定
		//       解析ロジック・Gemini呼び出しは未実装

		c.JSON(http.StatusAccepted, gin.H{
			"queued": true,
			"count":  n,
		})
	}
}
