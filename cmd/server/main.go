// main.go: Gin, Geminiクライアント初期化＋ルーティング
package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"shakehandz-api/internal/gmail"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("環境変数 GEMINI_API_KEY が設定されていません")
	}

	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Geminiクライアント初期化失敗: %v", err)
	}
	defer genaiClient.Close()

	r := gin.Default()

	// Gmail API
	r.GET("/api/gmail/sync", gmail.GmailMessagesHandler)
	// r.POST("/api/gmail/process", gmail.NewProcessHandler(...)) // handler_proc.go 実装後に追加

	// 他ドメインのルーティングもここに追加

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
