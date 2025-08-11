package auth

import (
	"errors"

	"gorm.io/gorm"
)

func FindGoogleRefreshTokenEncByUserID(db *gorm.DB, userID string) ([]byte, error) {
	var rec OAuthToken
	err := db.Where("user_id = ? AND provider = ?", userID, "google").First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rec.RefreshToken, nil
}
