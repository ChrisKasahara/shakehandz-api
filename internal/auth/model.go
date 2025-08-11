package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// users
type User struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey"`
	GoogleID string    `gorm:"type:varchar(255);uniqueIndex;not null"` // Googleのsubを冗長保持するならここでもOK
	Email    string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Name     string    `gorm:"type:varchar(255)"`
	Picture  string    `gorm:"type:text"`
}

// oauth_tokens
type OAuthToken struct {
	ID     uint      `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"type:char(36);not null;index:idx_user_provider,unique"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID;references:ID"`

	Provider string  `gorm:"size:32;not null;index:idx_user_provider,unique;index:uq_provider_sub,unique"` // "google"
	Sub      string  `gorm:"size:191;not null;index:uq_provider_sub,unique"`                               // Googleのsub（=不変の一意ID）
	Scope    *string `gorm:"type:text"`

	// 暗号化済みのrefresh_token。バイナリ格納（AES-GCMのnonce+ct）を推奨
	RefreshToken []byte `gorm:"type:varbinary(1024);not null"`

	// access_tokenの有効期限を保存したいならnullableで
	Expiry    *time.Time
	TokenType string `gorm:"size:32;default:Bearer"`

	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
