// internal/router/router.go: CORS対応（localhost:3000許可）＋DB依存のルーティング
package router

import (
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/extractor"
	"shakehandz-api/internal/humanresource"
	message "shakehandz-api/internal/message"
	"shakehandz-api/internal/middleware"
	"shakehandz-api/internal/project"
	config "shakehandz-api/internal/shared"
	"shakehandz-api/internal/shared/message/gmail"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-App-Auth"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	db := config.InitDB()

	// DI
	gmailMsgFetcher := gmail.NewGmailMsgFetcher()
	messageSvc := message.NewMessageService(gmailMsgFetcher)
	geminiService := extractor.NewGeminiService(gmailMsgFetcher, db, rdb)
	authService := auth.NewAuthService(db)
	hrHandler := humanresource.NewHumanResourcesHandler(db)
	projectHandler := project.NewProjectHandler(db)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(db))
	{
		// メール
		protected.GET("/message/gmail", message.MessageHandler(messageSvc, db))

		// AI
		protected.POST("/structure/humanresource", extractor.StructureWithGeminiHandler(geminiService))
		protected.GET("/structure/status", extractor.GetStructureStatusHandler(rdb))

		// 要員管理
		protected.GET("/humanresource", hrHandler.GetHumanResources)
		protected.GET("/humanresource/:id", hrHandler.GetHumanResourceByID)

		// 案件管理
		protected.GET("/projects", projectHandler.GetProjects)
		protected.GET("/projects/:id", projectHandler.GetProject)

	}

	r.POST("/api/auth/upsert", auth.UpsertUserHandler(authService))

	// Gemini Client/Service DI
	// geminiService := gemini.NewService(fetcher, geminiCl, db)

	// gmailService := gmail.NewService(fetcher)

	// r.POST("/api/gemini/structure-resources", gemini.NewStructureWithGeminiHandler(geminiService))
	// gemini/gmailの他ルーティングもfetcherを使う場合は同様に修正

	// HumanResource

	// Project

	return r
}
