package extractor

import (
	"net/http"
	"shakehandz-api/internal/shared/apierror"
	cache_extractor "shakehandz-api/internal/shared/cache/extractor"
	"shakehandz-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// ハンドラ名を役割に合わせて変更
func GetStructureStatusHandler(rc *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redisからステータスを取得
		progressStatus, err := cache_extractor.FetchJobStatus(c.Request.Context(), rc, "status")
		if err != nil {
			response.SendError(c, apierror.Common.JSONParseFailed, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "job status",
			})
		}
		if err == redis.Nil {
			// ステータスが存在しない場合は、初期状態を設定
			progressStatus = cache_extractor.CreateNewJobStatus("あなたの指示を待っています。")
			if err := cache_extractor.UpdateStatusInRedis(c.Request.Context(), rc, progressStatus); err != nil {
				response.SendError(c, apierror.Redis.UpdateDataFailed, response.ErrorDetail{
					Detail:   err.Error(),
					Resource: "job status",
				})
			}
		} else if err != nil {
			response.SendError(c, apierror.Redis.GetDataFailed, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "job status",
			})
			return
		}

		// 取得したJSON文字列をそのままクライアントに返す
		response.SendSuccess(c, http.StatusOK, progressStatus)
	}
}
