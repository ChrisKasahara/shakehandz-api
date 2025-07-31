package gmail

import (
	"context"
	"encoding/base64"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Fetcher 実装
type fetcher struct{}

func NewFetcher() Fetcher {
	return &fetcher{}
}

// Fetch: Gmail API からメール一覧を取得
// ctx, token, query, max を受け取り、[]interface{} で返す
func (f *fetcher) FetchMessageIDList(ctx context.Context, token *oauth2.Token, query string, max int) ([]interface{}, error) {
	srv, err := gmail.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, err
	}
	msgsList, err := srv.Users.Messages.List("me").MaxResults(int64(max)).Q(query).Do()
	if err != nil {
		return nil, err
	}
	var result []interface{}
	for _, m := range msgsList.Messages {
		result = append(result, m)
	}
	return result, nil
}

func ExtractPlainText(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		return DecodeBase64URL(payload.Body.Data)
	}

	// マルチパートの場合はパーツを再帰的に探索
	for _, part := range payload.Parts {
		result := ExtractPlainText(part)
		if result != "" {
			return result
		}
	}

	return ""
}

func DecodeBase64URL(data string) string {
	decoded, err := base64.URLEncoding.DecodeString(strings.ReplaceAll(data, "-", "+"))
	if err != nil {
		return ""
	}
	return string(decoded)
}
