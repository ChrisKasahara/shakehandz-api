package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	DB *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

// GET /projects
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var projects []Project
	if err := h.DB.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得失敗"})
		return
	}
	c.JSON(http.StatusOK, projects)
}

// GET /projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	var project Project
	if err := h.DB.First(&project, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "見つかりません"})
		return
	}
	c.JSON(http.StatusOK, project)
}
