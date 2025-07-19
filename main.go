package main

import (
	"log"
	"shakehandz-api/config"
	"shakehandz-api/handler"
	"shakehandz-api/router"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました")
	}

	db := config.InitDB()

	// DI
	handler.NewHumanResourcesHandler(db)
	handler.NewProjectHandler(db)

	r := router.SetupRouter(db)

	r.Run(":8080")
}
