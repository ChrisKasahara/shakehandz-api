package extractor

import (
	"net/http"

	"shakehandz-api/internal/shared/apierror"
	"shakehandz-api/internal/shared/auth/oauth"
	"shakehandz-api/internal/shared/llm/gemini"
	"shakehandz-api/internal/shared/message/gmail"
	"shakehandz-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Geminiサービス呼び出し用ハンドラ
func StructureWithGeminiHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// ユーザ認証
		verified, err := oauth.IsUserVerified(c)
		if err != nil {
			response.SendError(c, apierror.Common.Unauthorized, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "human resource",
			})
			return
		}

		cli, _ := gemini.NewGeminiClientWithRefresh(ctx, "models/gemini-2.5-flash", verified.Token.RefreshToken)

		gmail_svc, err := gmail.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
		if err != nil {
			response.SendError(c, apierror.Gmail.CreateClientFailed, response.ErrorDetail{
				Detail:   "failed to create gmail service",
				Resource: "gmail",
			})
			return
		}

		ok, err := svc.Run(c, cli, gmail_svc)
		if err != nil {
			response.SendError(c, apierror.Extractor.Unknown, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "extractor",
			})
			return
		}

		c.JSON(http.StatusOK, ok)
	}
}
