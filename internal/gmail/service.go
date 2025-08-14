package gmail

import (
	"context"
	msg "shakehandz-api/internal/shared/message"
	gmsg "shakehandz-api/internal/shared/message/gmail"

	"google.golang.org/api/gmail/v1"
)

type GmailMsgService struct {
	Fetcher gmsg.MessageIF
}

func NewGmailMsgService(f gmsg.MessageIF) *GmailMsgService {
	return &GmailMsgService{Fetcher: f}
}

// Gmailメッセージ取得
func (s *GmailMsgService) Run(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*msg.Message, error) {

	idMsgs, err := s.Fetcher.FetchMsgIds(ctx, svc, query, max)
	if err != nil {
		return nil, err
	}

	return idMsgs, nil
}
