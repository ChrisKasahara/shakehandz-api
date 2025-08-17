package cache_ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// JobStatusをRedisに書き込むヘルパー関数
func UpdateStatusInRedis(ctx context.Context, rdb *redis.Client, status JobStatus) error {
	key := fmt.Sprintf("job:status:%s", status.JobID)

	statusJSON, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status to JSON: %w", err)
	}

	expiration := 5 * time.Minute
	err = rdb.Set(ctx, key, statusJSON, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set status in redis: %w", err)
	}

	return nil
}
