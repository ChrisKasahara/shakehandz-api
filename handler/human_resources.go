package handler

import (
	"net/http"
	"shakehandz-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HumanResourcesHandler struct {
	DB *gorm.DB
}

func NewHumanResourcesHandler(db *gorm.DB) *HumanResourcesHandler {
	return &HumanResourcesHandler{DB: db}
}

// GET /human_resources
func (h *HumanResourcesHandler) GetHumanResources(c *gin.Context) {
	var humans []model.HumanResource
	if err := h.DB.Find(&humans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得失敗"})
		return
	}
	c.JSON(http.StatusOK, humans)
}

// GET /human_resources/:id
func (h *HumanResourcesHandler) GetHumanResource(c *gin.Context) {
	id := c.Param("id")
	var human model.HumanResource
	if err := h.DB.First(&human, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "見つかりません"})
		return
	}
	c.JSON(http.StatusOK, human)
}
