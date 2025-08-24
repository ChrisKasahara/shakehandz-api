package cache_extractor

import (
	"shakehandz-api/internal/shared/cache"
	"time"
)

type JobStatus struct {
	cache.RedisCache
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	StartedAt time.Time `json:"started_at,omitempty"`
}
