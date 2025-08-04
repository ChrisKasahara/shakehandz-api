package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	msg "shakehandz-api/internal/shared/message"
)

// Service: Gmailメッセージ取得→Gemini解析→（将来）DB保存
type Service struct {
	Fetcher msg.MessageFetcher
	Gemini  *Client
}

func NewService(f msg.MessageFetcher, g *Client) *Service {
	return &Service{Fetcher: f, Gemini: g}
}

func (s *Service) Run(ctx context.Context, token string) ([]*msg.Message, error) {
	msgs, err := s.Fetcher.FetchMsg(ctx, token, "has:attachment", 10)
	if err != nil {
		return msgs, err
	}

	// JSON形式で構造を表示（console.log風）
	jsonBytes, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		fmt.Printf("JSON変換エラー: %v\n", err)
		return msgs, err
	}
	fmt.Println(string(jsonBytes))

	// TODO: Gemini 解析ロジックをここに実装
	// 例: chat := s.Gemini.Model.StartChat() ...

	return msgs, err
}
