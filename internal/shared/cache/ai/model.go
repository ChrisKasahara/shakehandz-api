package cache_ai

import "time"

// Redisに保存するステータスの構造体
type JobStatus struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
}
