package main

import (
	"context"
	"log"
	"shakehandz-api/config"
	"shakehandz-api/router"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました")
	}

	db := config.InitDB()

	// Gemini client を初期化
	genaiClient, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		log.Fatalf("Geminiクライアント初期化失敗: %v", err)
	}

	// 依存注入: DB と GeminiClient
	r := router.SetupRouter(db, genaiClient)

	// サーバー起動
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
