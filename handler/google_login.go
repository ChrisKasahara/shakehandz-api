package handler

import (
	"context"
	"net/http"
	"os"
	"strings"

	"shakehandz-api/model"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type GoogleLoginHandler struct {
	DB *gorm.DB
}

func NewGoogleLoginHandler(db *gorm.DB) *GoogleLoginHandler {
	return &GoogleLoginHandler{DB: db}
}

func (h *GoogleLoginHandler) GoogleLogin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
		return
	}
	idToken := strings.TrimPrefix(authHeader, "Bearer ")

	payload, err := idtoken.Validate(context.Background(), idToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid id_token"})
		return
	}

	sub, _ := payload.Claims["sub"].(string)
	name, _ := payload.Claims["name"].(string)
	email, _ := payload.Claims["email"].(string)
	picture, _ := payload.Claims["picture"].(string)

	user := model.User{
		GoogleID: sub,
		Email:    email,
		Name:     name,
		Picture:  picture,
	}

	// Upsert: GoogleIDで検索し、なければ新規作成、あれば更新
	var existing model.User
	result := h.DB.Where("google_id = ?", sub).First(&existing)
	if result.Error == nil {
		// 更新
		existing.Email = email
		existing.Name = name
		existing.Picture = picture
		h.DB.Save(&existing)
	} else {
		// 新規
		h.DB.Create(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"google_id": sub,
		"email":     email,
		"name":      name,
		"picture":   picture,
	})
}
