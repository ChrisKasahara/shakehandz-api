package message

import (
	"context"
	"log"
	msg "shakehandz-api/internal/shared/message"
	gmsg "shakehandz-api/internal/shared/message/gmail"

	"google.golang.org/api/gmail/v1"
)

type MessageService struct {
	Fetcher gmsg.MessageIF
}

func NewMessageService(f gmsg.MessageIF) *MessageService {
	return &MessageService{Fetcher: f}
}

// Gmailメッセージ取得
func (s *MessageService) Run(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*msg.Message, error) {

	log.Printf("Gmailメッセージを取得中: query=%s, max=%d", query, max)
	idMsgs, err := s.Fetcher.FetchMsg(ctx, svc, query, max)
	if err != nil {
		return nil, err
	}

	return idMsgs, nil
}

func (s *MessageService) RunGetSingleMessage(ctx context.Context, svc *gmail.Service, messageID string) (*msg.Message, error) {
	return s.Fetcher.GetSingleMessage(ctx, svc, messageID)
}
