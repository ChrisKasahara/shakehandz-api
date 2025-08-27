package cache

import "time"

// Redisに保存するステータスの構造体
type RedisCache struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
