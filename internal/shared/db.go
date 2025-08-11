package config

import (
	"fmt"
	"log"
	"os"
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/humanresource"
	"shakehandz-api/internal/project"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "todo"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}

	if err := db.AutoMigrate(&project.Project{}, &humanresource.HumanResource{}, &auth.User{}, &auth.OAuthToken{}); err != nil {
		log.Fatal("マイグレーション失敗:", err)
	}

	return db
}
