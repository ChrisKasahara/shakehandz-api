package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// users
type User struct {
	ID    uuid.UUID `gorm:"type:char(36);primaryKey"`
	Email string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Name  string    `gorm:"type:varchar(255)"`
	Image string    `gorm:"type:text"`
}

// oauth_tokens
type OauthToken struct {
	ID     uint      `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"type:char(36);not null;index:idx_user_provider,unique"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID;references:ID"`

	Provider string  `gorm:"size:32;not null;index:idx_user_provider,unique;index:uq_provider_sub,unique"`
	Sub      string  `gorm:"size:191;not null;index:uq_provider_sub,unique"`
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

type AuthContext struct {
	User  User
	Token OauthToken
}
