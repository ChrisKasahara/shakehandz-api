// Gmail APIクライアント初期化
package gmail

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// NewService はGmail APIクライアントを初期化して返す
func NewService(ctx context.Context, token *oauth2.Token) (*gmail.Service, error) {
	return gmail.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
}
