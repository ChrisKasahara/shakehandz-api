package gemini

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	Model *genai.GenerativeModel
}

func NewClient(ctx context.Context, apiKey, model string) (*Client, error) {
	cli, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{Model: cli.GenerativeModel(model)}, nil
}
