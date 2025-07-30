package main

import (
	"context"
	"log"
	"os"
	"shakehandz-api/config"
	"shakehandz-api/router"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました")
	}

	// .envファイルからAPIキーを取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("環境変数 GEMINI_API_KEY が設定されていません")
	}

	db := config.InitDB()

	// Gemini client を正しい方法で初期化
	ctx := context.Background()

	// NewClientの第2引数にAPIキーを渡す
	genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Geminiクライアント初期化失敗: %v", err)
	}
	// 使い終わったらクライアントを閉じる
	defer genaiClient.Close()

	// 正しいモデル名を指定
	model := genaiClient.GenerativeModel("gemini-1.5-flash")

	// 依存注入: DB と GeminiClient
	r := router.SetupRouter(db, model)

	// サーバー起動
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
