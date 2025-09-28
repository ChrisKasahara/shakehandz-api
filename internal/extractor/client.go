package extractor

import (
	"context"
	"fmt"
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/shared/apierror"
	"shakehandz-api/internal/shared/auth/oauth"
	"shakehandz-api/internal/shared/llm/gemini"
	gmsg "shakehandz-api/internal/shared/message/gmail"
	"shakehandz-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/gmail/v1"
)

func RefreshExtractorTokenForBackground(authContext auth.AuthContext) (*gemini.Client, *gmail.Service, error) {
	// 独立したcontextを作成（gin.Contextではない）
	ctx := context.Background()

	cli, err := gemini.NewGeminiClientWithRefresh(ctx, GeminiModel, authContext.Token.RefreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	gmail_svc, err := gmsg.NewGmailClientWithRefresh(ctx, authContext.Token.RefreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return cli, gmail_svc, nil
}

func RefreshExtractorToken(c *gin.Context) (*gemini.Client, *gmail.Service, error) {
	verified, err := oauth.IsUserVerified(c)
	ctx := c.Request.Context()

	if err != nil {
		response.SendError(c, apierror.Common.Unauthorized, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return nil, nil, err
	}

	cli, _ := gemini.NewGeminiClientWithRefresh(ctx, "models/gemini-2.5-flash", verified.Token.RefreshToken)

	gmail_svc, err := gmsg.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
	if err != nil {
		response.SendError(c, apierror.Gmail.CreateClientFailed, response.ErrorDetail{
			Detail:   "failed to create gmail service",
			Resource: "gmail",
		})
		return nil, nil, err
	}

	// トークンのリフレッシュ処理を実装
	return cli, gmail_svc, nil
}
