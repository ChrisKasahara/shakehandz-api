// main.go: Gin, Geminiクライアント初期化＋ルーティング
package main

import (
	"log"

	"github.com/joho/godotenv"

	"shakehandz-api/internal/router"
)

func main() {
	// ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました")
	}

	// rdb, err := cache.NewRedisClient(ctx)
	// if err != nil {
	// 	log.Fatalf("Redisクライアントの初期化に失敗しました: %v", err)
	// }

	// r := router.SetupRouter(rdb)
	r := router.SetupRouter(nil)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバー起動失敗: %v", err)
	}
}
