package gmail

import (
	"context"
	"errors"

	"shakehandz-api/internal/shared/auth/oauth"
	"shakehandz-api/internal/shared/crypto"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// 保存済みの暗号化refresh_tokenからgmail.Serviceを生成（自動リフレッシュ）
func NewGmailClientWithRefresh(ctx context.Context, encRefresh []byte) (*gmail.Service, error) {
	if len(encRefresh) == 0 {
		return nil, errors.New("empty refresh token")
	}
	rt, err := crypto.DecryptFromBytes(encRefresh)
	if err != nil {
		return nil, err
	}
	cfg := oauth.OAuth2ConfigFromEnv()
	tok := &oauth2.Token{RefreshToken: rt}

	baseTS := cfg.TokenSource(ctx, tok)
	ts := oauth2.ReuseTokenSource(nil, baseTS)
	if ts == nil {
		return nil, errors.New("failed to create token source")
	}

	return gmail.NewService(ctx, option.WithTokenSource(ts))
}
