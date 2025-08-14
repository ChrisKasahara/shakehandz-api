package message_gmail

import (
	"context"
	m "shakehandz-api/internal/shared/message"

	"google.golang.org/api/gmail/v1"
)

type MsgIDFetcherIF interface {
	FetchMsgIds(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*m.Message, error)
	FetchMsgIdsWithPaging(ctx context.Context, svc *gmail.Service, query string, pageSize int64, pageToken string) ([]*m.Message, string, error)
}

type MsgDetailFetcherIF interface {
	FetchMsgDetails(ctx context.Context, srv *gmail.Service, messages []*gmail.Message) ([]*m.Message, error)
}

type MessageFetcherIF interface {
	FetchMsg(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*m.Message, error)
	FetchMsgWithPaging(ctx context.Context, svc *gmail.Service, query string, pageSize int64, pageToken string) ([]*m.Message, string, error)
}

type MessageIF interface {
	MsgIDFetcherIF
	MsgDetailFetcherIF
	MessageFetcherIF
}
