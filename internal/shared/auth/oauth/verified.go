package oauth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"shakehandz-api/internal/auth"
)

func IsUserVerified(c *gin.Context) (auth.AuthContext, error) {
	userValue, exists := c.Get("auth")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in context"})
		return auth.AuthContext{}, errors.New("user not found in context")
	}

	var verified auth.AuthContext
	// 型アサーションで元の auth.User 型に戻す
	verified, ok := userValue.(auth.AuthContext)
	if !ok {
		// 型アサーションに失敗した場合
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return auth.AuthContext{}, errors.New("invalid user type in context")
	}

	return verified, nil

}
