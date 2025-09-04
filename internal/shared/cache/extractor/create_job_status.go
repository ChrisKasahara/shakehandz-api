package cache_extractor

import (
	"shakehandz-api/internal/shared/cache"
	"time"
)

// Redisの初回登録時に使用するJobStatusのコンストラクタ
func CreateNewJobStatus(message string) *JobStatus {
	return &JobStatus{
		RedisCache: cache.RedisCache{
			ID:        "status",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Status:  "idle",
		Message: message,
	}
}

func (js *JobStatus) UpdateJobStatus(status, message string) {
	js.Status = status
	js.Message = message
	js.UpdatedAt = time.Now()
}

func (js *JobStatus) StartJob(message string) {
	if js.Status == "idle" {
		js.Status = "pending"
		js.UpdatedAt = time.Now()
		js.StartedAt = time.Now()
		js.Message = message
	}
}
