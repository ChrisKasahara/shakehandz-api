package gmail

import (
	"context"
	msg "shakehandz-api/internal/shared/message"

	"google.golang.org/api/gmail/v1"
)

type GmailMsgFetcher struct{}

func NewGmailMsgFetcher() MessageIF {
	return &GmailMsgFetcher{}
}

func (fetcher *GmailMsgFetcher) FetchMsg(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*msg.Message, error) {
	if max <= 0 {
		max = 10
	}
	list, err := fetcher.FetchMsgIds(svc, query, max)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return []*msg.Message{}, nil
	}
	return fetcher.FetchMsgDetails(ctx, svc, list)
}

// ページング版（svc使い回し）
func (fetcher *GmailMsgFetcher) FetchMsgWithPaging(ctx context.Context, svc *gmail.Service, query string, pageSize int64, pageToken string) ([]*msg.Message, string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	list, nextPageToken, err := fetcher.FetchMsgIdsWithPaging(svc, query, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	if len(list) == 0 {
		return []*msg.Message{}, nextPageToken, nil
	}
	msgs, err := fetcher.FetchMsgDetails(ctx, svc, list)
	return msgs, nextPageToken, err
}
