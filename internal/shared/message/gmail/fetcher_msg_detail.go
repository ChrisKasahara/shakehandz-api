package gmail

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	msg "shakehandz-api/internal/shared/message"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/gmail/v1"
)

// 単一のメッセージIDから内容を取得
func (fetcher *GmailMsgFetcher) GetSingleMessage(ctx context.Context, srv *gmail.Service, messageID string) (*msg.Message, error) {
	// gmail.Message スライスを作成（1つだけ）
	messages := []*gmail.Message{{Id: messageID}}

	// 既存メソッドを使用
	results, err := fetcher.FetchMsgDetails(ctx, srv, messages)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("メッセージが見つかりませんでした: %s", messageID)
	}

	return results[0], nil
}

// fetchMessageDetails: メッセージ詳細の並列取得処理
func (fetcher *GmailMsgFetcher) FetchMsgDetails(ctx context.Context, srv *gmail.Service, messages []*gmail.Message) ([]*msg.Message, error) {
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	var result []*msg.Message

	// 同時に実行するリクエスト数を10に制限
	sem := semaphore.NewWeighted(10)

	for _, m := range messages {
		mid := m.Id
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
			break
		}
		g.Go(func() error {
			defer sem.Release(1)

			msg, err := srv.Users.Messages.Get("me", mid).Format("full").Do()
			if err != nil {
				return err
			}
			dto, err := ParseMessage(msg)
			if err != nil {
				return err
			}
			mu.Lock()
			result = append(result, dto)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Printf("fetcher: detail fetch error: %v", err)
		return nil, err
	}
	// 日付降順
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})
	return result, nil
}
