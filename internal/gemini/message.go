package gemini

import (
	"fmt"
	"net/http"
	"strings"

	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/humanresource"
	msg "shakehandz-api/internal/shared/message"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Service) fetchUnprocessedMessages(c *gin.Context, token string, target int, db *gorm.DB) ([]*msg.Message, error) {
	const pageSize = 50
	const maxPages = 10

	var candidates []*msg.Message
	seenIDs := make(map[string]bool)
	pageToken := ""
	pageCount := 0

	ctx := c.Request.Context()

	for len(candidates) < target && pageCount < maxPages {
		pageCount++
		fmt.Printf("ページ %d/%d を処理中...\n", pageCount, maxPages)

		// 1) Authorization: Bearer <id_token>
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing id_token"})
			return nil, fmt.Errorf("missing id_token in Authorization header")
		}
		idToken := strings.TrimPrefix(authz, "Bearer ")

		// 2) id_token → User
		user, err := auth.UserFromIDToken(ctx, db, idToken)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid id_token"})
			return nil, fmt.Errorf("invalid id_token: %w", err)
		}

		// 3) DBから暗号化refresh_token（[]byte）取得
		enc, err := auth.FindGoogleRefreshTokenEncByUserID(db, user.ID.String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return nil, fmt.Errorf("db error: %w", err)
		}
		if len(enc) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "no_refresh_token", "reauthorize": true})
			return nil, fmt.Errorf("no_refresh_token: %w", err)
		}

		// ページング対応でメッセージを取得
		msgs, nextPageToken, err := s.Fetcher.FetchMsgWithPaging(ctx, enc, "has:attachment", pageSize, pageToken)
		if err != nil {
			return nil, fmt.Errorf("Gmail API 呼び出し失敗: %w", err)
		}

		if len(msgs) == 0 {
			fmt.Println("これ以上のメッセージはありません")
			break
		}

		// メッセージIDを抽出して重複除外
		var messageIDs []string
		for _, msg := range msgs {
			if !seenIDs[msg.Id] {
				messageIDs = append(messageIDs, msg.Id)
				seenIDs[msg.Id] = true
			}
		}

		if len(messageIDs) == 0 {
			fmt.Println("新しいメッセージIDがありません")
			pageToken = nextPageToken
			if pageToken == "" {
				break
			}
			continue
		}

		// DBで既存チェック（MessageIDを使用）
		var existingIDs []string
		err = s.DB.Model(&humanresource.HumanResource{}).
			Where("message_id IN (?)", messageIDs).
			Pluck("message_id", &existingIDs).Error
		if err != nil {
			return nil, fmt.Errorf("DB照会失敗: %w", err)
		}

		// 既存IDをマップに変換
		existingIDMap := make(map[string]bool)
		for _, id := range existingIDs {
			existingIDMap[id] = true
		}

		fmt.Printf("取得メッセージ数: %d, DB既存件数: %d\n", len(messageIDs), len(existingIDs))

		// 未処理メッセージのみを candidates に追加
		for _, msg := range msgs {
			if !existingIDMap[msg.Id] && len(candidates) < target {
				candidates = append(candidates, msg)
			}
		}

		fmt.Printf("未処理候補累計: %d件\n", len(candidates))

		// 次ページがない場合は終了
		pageToken = nextPageToken
		if pageToken == "" {
			fmt.Println("全ページを処理完了")
			break
		}
	}

	if pageCount >= maxPages {
		fmt.Printf("最大ページ数 %d に達しました\n", maxPages)
	}

	// 目標件数まで切り詰め
	if len(candidates) > target {
		candidates = candidates[:target]
	}

	fmt.Printf("最終的な未処理メッセージ件数: %d件\n", len(candidates))
	return candidates, nil
}
