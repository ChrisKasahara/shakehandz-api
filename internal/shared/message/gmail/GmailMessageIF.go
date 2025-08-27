package gmail

import (
	"context"
	m "shakehandz-api/internal/shared/message"

	"google.golang.org/api/gmail/v1"
)

type MsgIDFetcherIF interface {
	FetchMsgIds(svc *gmail.Service, query string, max int64) ([]*gmail.Message, error)
	FetchMsgIdsWithPaging(svc *gmail.Service, query string, pageSize int64, pageToken string) ([]*gmail.Message, string, error)
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
