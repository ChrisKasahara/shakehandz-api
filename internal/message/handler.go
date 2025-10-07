package message

import (
	"fmt"
	"net/http"

	"shakehandz-api/internal/shared/apierror"
	oauth "shakehandz-api/internal/shared/auth/oauth"
	"shakehandz-api/internal/shared/response"

	"shakehandz-api/internal/shared/message/gmail"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Gmail同期API: id_tokenからユーザー解決→DBのrefreshを使ってFetch
// 新しいBFFアーキテクチャに合わせて認証方法を変更しています
func MessageHandler(svc *MessageService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		verified, err := oauth.IsUserVerified(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		gmail_svc, err := gmail.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create gmail service"})
			return
		}

		msgs, err := svc.Run(ctx, gmail_svc, c.Query("query"), 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func MessageHandlerWithID(svc *MessageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		fmt.Printf("MessageHandlerWithID called with ID: %s\n", c.Param("id"))

		verified, err := oauth.IsUserVerified(c)
		if err != nil {
			response.SendError(c, apierror.Common.Unauthorized, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "mail",
			})

			return
		}

		gmail_svc, err := gmail.NewGmailClientWithRefresh(ctx, verified.Token.RefreshToken)
		if err != nil {
			response.SendError(c, apierror.Gmail.CreateClientFailed, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "mail",
			})
			return
		}

		messageID := c.Param("id")
		if messageID == "" {
			response.SendError(c, apierror.Common.BadRequest, response.ErrorDetail{
				Detail:   "message ID is required",
				Resource: "mail",
			})
			return
		}

		message, err := svc.RunGetSingleMessage(ctx, gmail_svc, messageID)
		if err != nil {
			response.SendError(c, apierror.Gmail.MessageFetchFailed, response.ErrorDetail{
				Detail:   err.Error(),
				Resource: "mail",
			})
			return
		}

		if message == nil {
			response.SendError(c, apierror.Gmail.MessageNotFound, response.ErrorDetail{
				Detail:   "message not found",
				Resource: "mail",
			})
			return
		}

		response.SendSuccess(c, http.StatusOK, message)
	}
}
