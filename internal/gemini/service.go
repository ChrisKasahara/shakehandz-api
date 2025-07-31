// Gmailサービス層: fetcherの結果をGemini解析・DB保存
package gemini

import (
	"context"
	msg "shakehandz-api/internal/shared/message"
)

// GmailMessageFetcherはGmailのメッセージをGeminiで解析するためのインターフェース
type GmailMessageFetcher struct {
	MessageFetcher msg.MessageFetcher
}

func NewMessageFetcherWithGemini(msgF msg.MessageFetcher) *GmailMessageFetcher {
	return &GmailMessageFetcher{MessageFetcher: msgF}
}

// Fetch: メール取得→（解析・保存は未実装）
func (s *GmailMessageFetcher) Fetch(ctx context.Context, token string) (int, error) {
	msgs, err := s.MessageFetcher.FetchMsg(ctx, token, "has:attachment", 5)
	if err != nil {
		return 0, err
	}
	// TODO: Gemini解析・DB保存（未実装）
	return len(msgs), nil
}
