package cache_extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func UpdateStatusInRedis(ctx context.Context, rdb *redis.Client, status *JobStatus) error {

	key := fmt.Sprintf("job:status:%s", status.ID)
	// Redisからステータスを取得
	oldStatus, err := FetchJobStatus(ctx, rdb, status.ID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch redis: %v", err)
	}

	if err == redis.Nil {
		status.CreatedAt = time.Now()
	} else if err != nil {
		return fmt.Errorf("failed to get data from redis before update: %w", err)
	} else {
		status.CreatedAt = oldStatus.CreatedAt
	}

	// 3. 準備が整ったstatusオブジェクト（CreatedAtが正しく設定されている）をMarshalしてSETする
	newStatusJSON, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal job status: %w", err)
	}

	expiration := 5 * time.Minute
	if err := rdb.Set(ctx, key, newStatusJSON, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set job status in redis: %w", err)
	}

	return nil
}
