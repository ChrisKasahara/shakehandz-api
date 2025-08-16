package shared_auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	mw "shakehandz-api/internal/middleware"
)

func IsUserVerified(c *gin.Context) (mw.AuthContext, error) {
	userValue, exists := c.Get("auth")

	// 2. 存在しなかった場合（ありえないはずだが念のため）
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in context"})
		return mw.AuthContext{}, errors.New("user not found in context")
	}

	var verified mw.AuthContext
	// 3. 型アサーションで元の auth.User 型に戻す
	verified, ok := userValue.(mw.AuthContext)
	if !ok {
		// 型アサーションに失敗した場合
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return mw.AuthContext{}, errors.New("invalid user type in context")
	}

	return verified, nil
}
