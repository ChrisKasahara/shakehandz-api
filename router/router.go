package router

import (
	"shakehandz-api/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	projectHandler := handler.NewProjectHandler(db)
	humanHandler := handler.NewHumanResourcesHandler(db)

	projects := r.Group("/projects")
	{
		projects.GET("", projectHandler.GetProjects)
		projects.GET("/:id", projectHandler.GetProject)
	}

	humanResources := r.Group("/human_resources")
	{
		humanResources.GET("", humanHandler.GetHumanResources)
		humanResources.GET("/:id", humanHandler.GetHumanResource)
	}

	return r
}
