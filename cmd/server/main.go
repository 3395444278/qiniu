package main

import (
	"log"
	"os"
	"qinniu/internal/api"
	"qinniu/internal/pkg/cache"
	"qinniu/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// 加载环境变量
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Printf("警告: 未能加载 .env 文件: %v", err)
	}
}

func main() {
	// 初始化数据库连接
	if err := database.InitMongoDB(); err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	// 初始化Redis连接
	if err := cache.InitRedis(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Continuing without Redis cache...")
	} else {
		log.Println("Successfully connected to Redis")
	}

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	api.SetupRoutes(r)

	// 获取服务器端口
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("无法启动服务器: %v", err)
	}
}
