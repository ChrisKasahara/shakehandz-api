// internal/router/router.go: CORS対応（localhost:3000許可）＋DB依存のルーティング
package router

import (
	"shakehandz-api/internal/auth"
	g "shakehandz-api/internal/gmail"
	"shakehandz-api/internal/humanresource"
	"shakehandz-api/internal/project"
	config "shakehandz-api/internal/shared"
	gmsg "shakehandz-api/internal/shared/message/gmail"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS: localhost:3000のみ許可
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-App-Auth"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	db := config.InitDB()

	// Gemini Client/Service DI
	// ctx := context.Background()
	// geminiCl, _ := gemini.NewClient(ctx, os.Getenv("GEMINI_API_KEY"), "models/gemini-2.5-flash")
	// geminiService := gemini.NewService(fetcher, geminiCl, db)

	// gmailService := gmail.NewService(fetcher)

	gmailMsgFetcherSvc := g.NewGmailMsgService(gmsg.NewGmailMsgFetcher())

	r.GET("/api/gmail/messages", g.NewGmailHandler(gmailMsgFetcherSvc, db))
	// r.POST("/api/gemini/structure-resources", gemini.NewStructureWithGeminiHandler(geminiService))
	// gemini/gmailの他ルーティングもfetcherを使う場合は同様に修正

	// HumanResource
	hrHandler := humanresource.NewHumanResourcesHandler(db)
	r.GET("/api/humanresources", hrHandler.GetHumanResources)
	r.GET("/api/humanresources/:id", hrHandler.GetHumanResource)

	// Project
	projectHandler := project.NewProjectHandler(db)
	r.GET("/api/projects", projectHandler.GetProjects)
	r.GET("/api/projects/:id", projectHandler.GetProject)

	// Auth
	authHandler := auth.NewGoogleLoginHandler(db)
	r.POST("/api/auth/google-login", authHandler.GoogleLogin)
	consentHandler := auth.NewGoogleConsentHandler(db)
	r.POST("/api/auth/google-consent", auth.RequireInternalCall(), consentHandler.SaveRefreshToken)

	return r
}
