// Package response は、APIのJSONレスポンスに関する共通の型定義とヘルパー関数を提供します。
package response

import (
	"net/http"
	"shakehandz-api/internal/shared/apierror"

	"github.com/gin-gonic/gin"
)

//================================================================
// 成功レスポンス
//================================================================

type SuccessResponse[T any] struct {
	Status int `json:"status"`
	Data   T   `json:"data"`
}

// SendSuccess は成功レスポンスをクライアントに送信します。
func SendSuccess[T any](c *gin.Context, status int, data T) {
	c.JSON(status, SuccessResponse[T]{
		// 渡されたstatusを使うように修正
		Status: http.StatusOK,
		Data:   data,
	})
}

//================================================================
// エラーレスポンス (apierrorと連携)
//================================================================

// ErrorResponse はエラー時のレスポンス構造体です。
type ErrorResponse struct {
	Message string        `json:"message"`
	Code    string        `json:"code"`
	Errors  []ErrorDetail `json:"errors"`
}

// ErrorDetail は個々のエラーの詳細です。
type ErrorDetail struct {
	Detail   string `json:"detail"`
	Resource string `json:"resource,omitempty"`
	Field    string `json:"field,omitempty"`
}

func SendError(c *gin.Context, code apierror.Code, details ...ErrorDetail) {
	errInfo := apierror.GetInfo(code)

	finalDetails := []ErrorDetail{}
	if len(details) > 0 {
		finalDetails = details
	}

	c.AbortWithStatusJSON(errInfo.HTTPStatus, ErrorResponse{
		Message: errInfo.Message,
		Code:    string(code),
		Errors:  finalDetails,
	})
}
