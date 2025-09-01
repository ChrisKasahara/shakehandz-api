package middleware

import (
	"net/http"
	"os"

	"shakehandz-api/internal/auth" // User, OauthTokenモデルのパス

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthMiddlewareは、内部認証とユーザー検索を行うGinミドルウェアです
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 内部APIトークンを検証
		internalToken := c.GetHeader("X-App-Auth")
		if internalToken != os.Getenv("BACKEND_INTERNAL_TOKEN") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			return
		}

		// ヘッダーから 'X-Google-ID' (sub) を取得
		googleID := c.GetHeader("X-Google-ID")
		if googleID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing google id"})
			return
		}

		// OAuthTokenを起点にユーザーを検索
		var token auth.OauthToken
		if err := db.Where("provider = ? AND sub = ?", "google", googleID).First(&token).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "token not found for the user"})
			return
		}

		if token.UserID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user id not associated with the token"})
			return
		}

		var user auth.User
		if err := db.First(&user, token.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		verifiedUser := auth.AuthContext{
			User:  user,
			Token: token,
		}

		// ユーザー情報をコンテキストに保存して、次のハンドラーに渡す
		c.Set("auth", verifiedUser)

		// 次の処理（ミドルウェアまたはハンドラー）に進む
		c.Next()
	}
}
