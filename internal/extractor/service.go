package extractor

import (
	"fmt"
	"net/http"
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/shared/auth/oauth"
	gmsg "shakehandz-api/internal/shared/message/gmail"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 　Gmailメッセージ取得→Gemini解析→（将来）DB保存
type Service struct {
	Fetcher gmsg.MessageIF
	DB      *gorm.DB
	rdb     *redis.Client
}

func NewExtractorService(f gmsg.MessageIF, db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{Fetcher: f, DB: db, rdb: rdb}
}

func (s *Service) Run(c *gin.Context) error {
	user, err := auth.GetUser(c)
	if err != nil {
		fmt.Println("ユーザー情報の取得に失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return err
	}

	verified, err := oauth.IsUserVerified(c)
	if err != nil {
		fmt.Println("ユーザーの認証に失敗しました")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return err
	}

	var batchExecution ExtractorBatchExecution

	result := s.DB.Where("user_id = ? AND extractor_type = ?", user.ID, TypeHumanResource).
		Order("execution_date desc").
		First(&batchExecution)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve batch execution record"})
		return result.Error
	}

	fmt.Printf("RefreshExtractorToken called for user: %s\n", user.ID)

	// 更新不要の場合は何もしない
	// クライアント更新済みの旨を返却
	if time.Since(batchExecution.ExecutionDate) <= MessageTTL && batchExecution.Status != StatusFailed {
		fmt.Println("クライアントの更新は不要です")
		c.JSON(http.StatusOK, gin.H{"message": "現在バッチが進行中"})
		return nil
	}

	// ・本日の最後の処理が失敗している
	var isFailed bool = batchExecution.Status == StatusFailed
	// ・最後の処理から24時間以上経過している
	var isExpired bool = time.Since(batchExecution.ExecutionDate) > MessageTTL

	var newBatchExecution ExtractorBatchExecution
	// クライアントを更新して返却
	if isExpired || isFailed {
		// ユーザ認証
		cli, gmail_svc, err := RefreshExtractorTokenForBackground(verified)
		if err != nil {
			fmt.Println("クライアントの更新に失敗しました")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh extractor token"})
			return err
		}

		UpdateClientData(user.ID, gmail_svc, cli)

		if isExpired {
			fmt.Println("有効期限切れのためClientを更新します")
		} else {
			fmt.Println("失敗ステータスのためClientを更新します")

		}

		// 新しいバッチレコードを現在のバッチとして設定
		newBatchExecution = ExtractorBatchExecution{
			UserID:        user.ID,
			ExtractorType: TypeHumanResource,
			TriggerFrom:   TriggerFront,
			Status:        StatusInProgress,
			ExecutionDate: time.Now(),
		}

		// 新しいバッチレコードを作成
		r := s.DB.Create(&newBatchExecution)

		if r.Error != nil {
			fmt.Printf("バッチレコード作成エラー: %v\n", r.Error)
			return r.Error
		}

	}

	// バッチ処理を開始
	// クライアントを更新した場合のみ、新しいバッチレコードを渡す
	err = RunHumanResourceExtractionBatch(s, user, MessageTTL, &newBatchExecution)

	if err != nil {
		fmt.Println("バッチ処理の開始に失敗しました")

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start batch processing"})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Extractor token refreshed"})

	return err

}
