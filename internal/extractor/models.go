package extractor

import (
	"time"

	"github.com/google/uuid"
)

type ExtractorBatchExecution struct {
	ID     uint      `gorm:"primaryKey"`
	UserID uuid.UUID `gorm:"type:char(36);not null;index:idx_user_provider"`

	ExtractorType string    `gorm:"type:varchar(20);not null;index"`          // human_resource, project
	TriggerFrom   string    `gorm:"type:varchar(20);not null;default:'auto'"` // front, auto
	ExecutionDate time.Time `gorm:"type:datetime(3);not null;index;default:CURRENT_TIMESTAMP(3)"`
	Status        string    `gorm:"type:varchar(20);not null;index"` // pending, in_progress, completed, failed
}

const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusNoData     = "no_data"
	StatusFailed     = "failed"
	StatusExpired    = "expired"
)

const (
	TriggerFront = "front"
	TriggerAuto  = "auto"
)

const (
	TypeHumanResource = "human_resource"
	TypeProject       = "project"
)
