package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *AuthService
}

func NewUserHandler(service *AuthService) *UserHandler {
	return &UserHandler{service: service}
}

type UpsertUserRequest struct {
	GoogleID     string `json:"googleId" binding:"required"`
	RefreshToken string `json:"refreshToken" binding:"required"`
	Scope        string `json:"scope" binding:"required"`
	Email        string `json:"email" binding:"required"`
	Name         string `json:"name"`
	Image        string `json:"image"`
}

func UpsertUserHandler(s *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 内部APIトークン検証
		internalToken := c.GetHeader("X-App-Auth")
		if internalToken != os.Getenv("BACKEND_INTERNAL_TOKEN") {

			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			return
		}

		var req UpsertUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
			return
		}

		if err := s.UpsertUserWithToken(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

}
