package options

import (
	"net/http"
	"shakehandz-api/internal/shared/apierror"
	"shakehandz-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OptionsHandler struct {
	DB *gorm.DB
}

func NewOptionsHandler(db *gorm.DB) *OptionsHandler {
	return &OptionsHandler{DB: db}
}

func (h *OptionsHandler) GetSkills(c *gin.Context) {
	var skills []Skills
	if err := h.DB.Order("count DESC").Limit(30).Find(&skills).Error; err != nil {
		response.SendError(c, apierror.Common.Unknown, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	response.SendSuccess(c, http.StatusOK, skills)

}
