package cache_extractor

import (
	"context"
	"encoding/json"
	"fmt"

	// このパッケージで必要であれば残す
	"github.com/redis/go-redis/v9"
)

func FetchJobStatus(ctx context.Context, rdb *redis.Client, id string) (*JobStatus, error) {
	// 1. キーを生成
	key := fmt.Sprintf("job:status:%s", id)

	// 2. RedisからJSON文字列を取得
	statusJSON, err := rdb.Get(ctx, key).Result()
	if err != nil {
		// これにより、呼び出し元は「キーが存在しない」ことを検知できる
		return nil, err
	}

	// 3. 取得したJSON文字列を、GoのJobStatusオブジェクトに変換（デシリアライズ）
	var status JobStatus
	if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
		// JSONのパースに失敗した場合、エラーに文脈を追加して返す
		return nil, fmt.Errorf("failed to unmarshal job status JSON (id: %s): %w", id, err)
	}

	// 4. 成功した場合、復元したオブジェクトのポインタを返す
	return &status, nil
}
