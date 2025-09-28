package extractor

import (
	"context"
	"fmt"

	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/humanresource"
	msg "shakehandz-api/internal/shared/message"

	"google.golang.org/api/gmail/v1"
)

func (s *Service) fetchUnprocessedMessages(ctx context.Context, user auth.User, gmail_svc *gmail.Service, target int) ([]*msg.Message, error) {
	// 解析結果を保持
	var candidates []*msg.Message
	seenIDs := make(map[string]bool)
	pageToken := ""
	pageCount := 0

	// 指定件数に達するまでページングでメッセージを取得
	for len(candidates) < target && pageCount < MaxPages {
		pageCount++
		fmt.Printf("ページ %d/%d を処理中...\n", pageCount, MaxPages)

		// ページング対応でメッセージを取得
		msgs, nextPageToken, err := s.Fetcher.FetchMsgWithPaging(ctx, gmail_svc, "has:attachment", PageSize, pageToken)
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
			Where("created_by_id = ?", user.ID).
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

				// 取得件数がに達した場合でも次ページを取得した場合処理が進行してしまうので、件数に達した場合ここで処理を終了
				if len(candidates) >= target {
					break
				}
			}
		}

		// 次ページがない場合は終了
		pageToken = nextPageToken
		if pageToken == "" {
			fmt.Println("全ページを処理完了")
			break
		}
	}

	if pageCount >= MaxPages {
		fmt.Printf("最大ページ数 %d に達しました\n", MaxPages)
	}

	// 目標件数まで切り詰め
	if len(candidates) > target {
		candidates = candidates[:target]
	}
	return candidates, nil
}
