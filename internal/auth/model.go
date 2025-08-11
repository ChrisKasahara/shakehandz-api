package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey"`
	GoogleID string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Email    string    `gorm:"type:varchar(255);not null"`
	Name     string
	Picture  string
}

type OAuthToken struct {
	ID       uint      `gorm:"primaryKey"`
	UserID   uuid.UUID `gorm:"type:char(36);index"`
	Provider string    `gorm:"size:32;not null"` // "google"
	Sub      string    `gorm:"size:191;not null;uniqueIndex:uq_provider_sub"`
	Scope    *string

	AccessToken  string    `gorm:"type:text"`
	RefreshToken []byte    // 暗号化して格納
	Expiry       time.Time `gorm:"not null"`
	TokenType    string    `gorm:"size:32;default:Bearer"`

	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
