// Gmailサービス層: fetcherの結果をGemini解析・DB保存
package gmail

import (
	"context"
	shm "shakehandz-api/internal/shared/mail"
)

// ProcessService: fetcherをDI
type ProcessService struct {
	Fetcher shm.Fetcher
}

func NewProcessService(f shm.Fetcher) *ProcessService {
	return &ProcessService{Fetcher: f}
}

// Run: メール取得→（解析・保存は未実装）
func (s *ProcessService) Run(ctx context.Context, token string) (int, error) {
	msgs, err := s.Fetcher.Fetch(ctx, token, "has:attachment", 5)
	if err != nil {
		return 0, err
	}
	// TODO: Gemini解析・DB保存（未実装）
	return len(msgs), nil
}
