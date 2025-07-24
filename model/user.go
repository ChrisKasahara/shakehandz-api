package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey"`
	GoogleID string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Email    string    `gorm:"type:varchar(255);not null"`
	Name     string
	Picture  string
}
