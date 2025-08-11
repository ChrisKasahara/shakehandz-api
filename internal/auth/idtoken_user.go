package auth

import (
	"context"
	"os"

	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

func UserFromIDToken(ctx context.Context, db *gorm.DB, idToken string) (*User, error) {
	aud := os.Getenv("GOOGLE_CLIENT_ID")
	payload, err := idtoken.Validate(ctx, idToken, aud)
	if err != nil {
		return nil, err
	}
	sub, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)

	var user User
	if err := db.Where("google_id = ?", sub).First(&user).Error; err != nil {
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			return nil, err
		}
	}
	return &user, nil
}
