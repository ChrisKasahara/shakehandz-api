package router

import (
	"shakehandz-api/handler"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, genaiClient *genai.GenerativeModel) *gin.Engine {
	r := gin.Default()

	// CORS: allow http://localhost:3000
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	projectHandler := handler.NewProjectHandler(db)
	humanHandler := handler.NewHumanResourcesHandler(db)
	geminiHandler := handler.NewGeminiHandler(genaiClient)

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

	// Gmail API
	r.GET("/gmail/messages", handler.GmailMessagesHandler)

	// Google ID Token Login
	googleLoginHandler := handler.NewGoogleLoginHandler(db)
	r.POST("/api/auth/google-login", googleLoginHandler.GoogleLogin)

	r.POST("/api/gemini/convert", geminiHandler.Convert)

	return r
}
