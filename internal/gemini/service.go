package gemini

import (
	"context"

	mail "shakehandz-api/internal/mail"

	"github.com/google/generative-ai-go/genai"
)

type Service struct {
	Fetcher mail.Fetcher
	Client  *genai.GenerativeModel
}

func NewService(f mail.Fetcher, c *genai.GenerativeModel) *Service {
	return &Service{Fetcher: f, Client: c}
}

// Run: メール取得→（Gemini解析・保存は未実装）
func (s *Service) Run(ctx context.Context, token string) (int, error) {
	msgs, err := s.Fetcher.Fetch(ctx, token, "has:attachment", 5)
	if err != nil {
		return 0, err
	}
	// TODO: Gemini解析・DB保存（未実装）
	return len(msgs), nil
}
