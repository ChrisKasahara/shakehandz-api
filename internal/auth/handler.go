package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// UserHandlerはユーザー関連のHTTPリクエストを処理します
type UserHandler struct {
	service *AuthService
}

// NewUserHandlerは新しいUserHandlerを生成します
func NewUserHandler(service *AuthService) *UserHandler {
	return &UserHandler{service: service}
}

// UpsertUserRequestはユーザー作成・更新リクエストのボディを表します
type UpsertUserRequest struct {
	GoogleID     string `json:"googleId" binding:"required"`
	RefreshToken string `json:"refreshToken" binding:"required"`
	Scope        string `json:"scope" binding:"required"`
	Email        string `json:"email" binding:"required"`
	Name         string `json:"name"`
	Image        string `json:"image"`
}

// UpsertUserはユーザー情報をデータベースに作成または更新します
// Next.jsサーバーからのみ呼び出されることを想定しています
func UpsertUserHandler(s *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 内部APIトークンを検証
		internalToken := c.GetHeader("X-App-Auth")
		if internalToken != os.Getenv("BACKEND_INTERNAL_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			return
		}

		// 2. リクエストボディを検証・バインド
		var req UpsertUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
			return
		}

		// 3. サービスレイヤーを呼び出してビジネスロジックを実行
		if err := s.UpsertUserWithToken(req); err != nil {
			// サービスレイヤーで発生したエラーに応じて適切なステータスコードを返す
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

}
