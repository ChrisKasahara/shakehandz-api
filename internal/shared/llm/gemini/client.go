package gemini

import (
	"context"
	"errors"
	"shakehandz-api/internal/shared/auth/oauth"
	"shakehandz-api/internal/shared/crypto"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type Client struct {
	Model *genai.GenerativeModel
}

func NewGeminiClientWithRefresh(ctx context.Context, model string, encRefresh []byte) (*Client, error) {
	if len(encRefresh) == 0 {
		return nil, errors.New("empty refresh token")
	}

	// 1. リフレッシュトークンを復号
	rt, err := crypto.DecryptFromBytes(encRefresh)
	if err != nil {
		return nil, err
	}

	// 2. OAuth2 ConfigとTokenSourceを生成 (Gmailの時と同じ)
	cfg := oauth.OAuth2ConfigFromEnv()
	tok := &oauth2.Token{RefreshToken: rt}
	baseTS := cfg.TokenSource(ctx, tok)
	ts := oauth2.ReuseTokenSource(nil, baseTS)
	if ts == nil {
		return nil, errors.New("failed to create token source")
	}

	// 3. ★★★ WithAPIKeyの代わりにWithTokenSourceを使う ★★★
	cli, err := genai.NewClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	return &Client{Model: cli.GenerativeModel(model)}, nil
}
