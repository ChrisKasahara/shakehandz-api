package humanresource

import (
	"net/http"

	"shakehandz-api/internal/shared/apierror"
	"shakehandz-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HumanResourcesHandler struct {
	DB *gorm.DB
}

func NewHumanResourcesHandler(db *gorm.DB) *HumanResourcesHandler {
	return &HumanResourcesHandler{DB: db}
}

func (h *HumanResourcesHandler) GetHumanResources(c *gin.Context) {
	var humans []HumanResource

	if err := h.DB.Order("created_at DESC").Limit(100).Find(&humans).Error; err != nil {
		response.SendError(c, apierror.Common.JSONParseFailed, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	response.SendSuccess(c, http.StatusOK, humans)
}

func (h *HumanResourcesHandler) GetHumanResourceByID(c *gin.Context) {
	id := c.Param("id")
	var human HumanResource
	if err := h.DB.First(&human, "id = ?", id).Error; err != nil {
		response.SendError(c, apierror.Common.JSONParseFailed, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}
	response.SendSuccess(c, http.StatusOK, human)
}
