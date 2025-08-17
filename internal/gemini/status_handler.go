package gemini

import (
	"fmt"
	"log"
	"net/http"
	cache_ai "shakehandz-api/internal/shared/cache/ai"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// ハンドラ名を役割に合わせて変更
func GetStructureStatusHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redisのキーを生成
		key := fmt.Sprintf("job:status:%s", "status")

		// Redisからステータス(JSON文字列)を取得
		statusJSON, err := rdb.Get(c.Request.Context(), key).Result()
		if err == redis.Nil {
			// ステータスが存在しない場合は、初期状態を設定
			progressStatus := cache_ai.JobStatus{
				JobID:   "status",
				Status:  "idle",
				Message: "あなたの指示を待っています。",
			}
			if err := cache_ai.UpdateStatusInRedis(c.Request.Context(), rdb, progressStatus); err != nil {
				log.Printf("ERROR: Failed to update redis: %v", err)
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
			return
		}

		// 取得したJSON文字列をそのままクライアントに返す
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, statusJSON)
	}
}
