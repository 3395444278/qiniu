package main

import (
	"log"
	"os"
	"os/signal"
	"qinniu/internal/pkg/ai"
	"qinniu/internal/pkg/cache"
	"qinniu/internal/pkg/database"
	"qinniu/internal/pkg/queue"
	"qinniu/internal/worker"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

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

	// 初始化 AI 客户端
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("AI_API_KEY environment variable not set")
	}

	log.Printf("Starting evaluator service with API key: %s...", apiKey[:10])

	aiClient := ai.NewClient(apiKey)
	queueClient := queue.NewQueue()
	evaluator := worker.NewEvaluator(aiClient, queueClient)

	// 启动工作器
	log.Println("Starting evaluator worker...")
	if err := evaluator.Start(); err != nil {
		log.Fatalf("Failed to start evaluator: %v", err)
	}

	log.Println("Evaluator service is running...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down evaluator...")
}
