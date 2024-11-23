package initconfig

import (
	"log"

	"qinniu/internal/pkg/cache"
	"qinniu/internal/pkg/database"

	"github.com/joho/godotenv"
)

// Init 初始化项目所需的环境变量、数据库和 Redis 连接
func Init() {
	// 加载环境变量
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Printf("警告: 未能加载 .env 文件: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitMongoDB(); err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	// 初始化 Redis 连接
	if err := cache.InitRedis(); err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
		log.Println("继续运行，但不使用 Redis 缓存...")
	} else {
		log.Println("成功连接到 Redis")
	}
}
