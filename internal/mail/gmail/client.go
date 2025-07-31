package gmailfetcher

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// NewService はGmail APIクライアントを初期化して返す
func NewService(ctx context.Context, token string) (*gmail.Service, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return gmail.NewService(ctx, option.WithTokenSource(ts))
}
