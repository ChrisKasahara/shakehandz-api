package message_gmail

import (
	"context"

	"google.golang.org/api/gmail/v1"
)

func (fetcher *GmailMsgFetcher) FetchMsgIds(ctx context.Context, svc *gmail.Service, query string, max int64) ([]*gmail.Message, error) {
	if max <= 0 {
		max = 10
	}
	list, err := svc.Users.Messages.List("me").MaxResults(max).Q(query).Do()
	if err != nil {
		return nil, err
	}
	if len(list.Messages) == 0 {
		return []*gmail.Message{}, nil
	}
	return list.Messages, nil
}

// ページング版（svc使い回し）
func (fetcher *GmailMsgFetcher) FetchMsgIdsWithPaging(ctx context.Context, svc *gmail.Service, query string, pageSize int64, pageToken string) ([]*gmail.Message, string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	call := svc.Users.Messages.List("me").MaxResults(pageSize).Q(query)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	list, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	if len(list.Messages) == 0 {
		return []*gmail.Message{}, list.NextPageToken, nil
	}

	return list.Messages, list.NextPageToken, nil
}
