package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) (User, error) {
	authValue, exists := c.Get("auth")

	if !exists {
		return User{}, errors.New("unauthorized")
	}

	verifiedUser, ok := authValue.(AuthContext)
	if !ok {
		return User{}, errors.New("unauthorized")
	}

	return verifiedUser.User, nil
}
