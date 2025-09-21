package humanresource

import (
	"net/http"

	"shakehandz-api/internal/auth"
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

func (h *HumanResourcesHandler) GetHumanResourceByID(c *gin.Context) {
	id := c.Param("id")
	user, err := auth.GetUser(c)

	if err != nil {
		response.SendError(c, apierror.Common.Unauthorized, response.ErrorDetail{
			Detail:   "unauthorized",
			Resource: "human resource",
		})
		return
	}

	var humanResources HumanResource
	if err := h.DB.Where("created_by_id = ?", user.ID).First(&humanResources, "id = ?", id).Error; err != nil {
		response.SendError(c, apierror.Common.JSONParseFailed, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	response.SendSuccess(c, http.StatusOK, humanResources)
}

func (h *HumanResourcesHandler) GetHumanResourcesWithFilter(c *gin.Context) {
	var humansResource []HumanResource
	var total int64

	user, err := auth.GetUser(c)

	if err != nil {
		response.SendError(c, apierror.Common.Unauthorized, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	// フィルターパラメータを解析
	filter, err := h.parseFilterParams(c)
	if err != nil {
		response.SendError(c, apierror.Common.BadRequest, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	// デフォルト値を設定
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	// ベースクエリを構築
	query := h.DB.Where("created_by_id = ?", user.ID)

	// 動的フィルターを適用
	query = h.applyFilters(query, filter)

	// 総数を取得（ページング用）
	if err := query.Model(&HumanResource{}).Count(&total).Error; err != nil {
		response.SendError(c, apierror.Common.DatabaseError, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	// ページングを適用
	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	query = query.Order("created_at DESC")

	if err := query.Find(&humansResource).Error; err != nil {
		response.SendError(c, apierror.Common.DatabaseError, response.ErrorDetail{
			Detail:   err.Error(),
			Resource: "human resource",
		})
		return
	}

	totalPages := (total + int64(filter.Limit) - 1) / int64(filter.Limit)

	responseData := HumanResourceResponse{
		Pagination: PaginationInfo{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
		AppliedFilters:     filter,
		HumanResourcesData: humansResource,
	}

	response.SendSuccess(c, http.StatusOK, responseData)
}
