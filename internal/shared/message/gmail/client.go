package message_gmail

import (
	"context"
	"errors"

	"shakehandz-api/internal/shared/crypto"
	googleutil "shakehandz-api/internal/shared/google"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// 旧: NewService(ctx, accessToken) は1時間で失効 → 今後は未使用に
// func NewService(ctx context.Context, token string) (*gmail.Service, error) { ... }

// 保存済みの暗号化refresh_tokenからgmail.Serviceを生成（自動リフレッシュ）
func NewServiceWithRefresh(ctx context.Context, encRefresh []byte) (*gmail.Service, error) {
	if len(encRefresh) == 0 {
		return nil, errors.New("empty refresh token")
	}
	rt, err := crypto.DecryptFromBytes(encRefresh)
	if err != nil {
		return nil, err
	}
	cfg := googleutil.OAuth2ConfigFromEnv()
	tok := &oauth2.Token{RefreshToken: rt}
	ts := cfg.TokenSource(ctx, tok) // ← 自動でaccess_token発行＆更新
	return gmail.NewService(ctx, option.WithTokenSource(ts))
}
