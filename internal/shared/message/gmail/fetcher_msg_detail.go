package message_gmail

import (
	"context"
	"log"
	"sort"
	"sync"

	msg "shakehandz-api/internal/shared/message"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/gmail/v1"
)

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
