package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"qinniu/internal/api"
	"qinniu/internal/pkg/ai"
	"qinniu/internal/pkg/queue"
	"qinniu/internal/worker"
	"syscall"
	"time"

	"qinniu/internal/pkg/initconfig"

	"github.com/gin-gonic/gin"
)

func main() {
	initconfig.Init()

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	api.SetupRoutes(r)

	// 初始化 AI 客户端
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("AI_API_KEY environment variable not set")
	}

	log.Printf("Starting evaluator service with API key: %s...", apiKey[:10])

	aiClient := ai.NewClient(apiKey)
	queueClient := queue.NewQueue()
	evaluator := worker.NewEvaluator(aiClient, queueClient)

	// 启动评估服务在一个新的协程中
	go func() {
		log.Println("Starting evaluator worker...")
		if err := evaluator.Start(); err != nil {
			log.Fatalf("Failed to start evaluator: %v", err)
		}
		log.Println("Evaluator service is running...")
	}()

	// 获取服务器端口
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 启动服务器在一个新的协程中
	go func() {
		log.Printf("Starting Gin server on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("无法启动服务器: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down evaluator and server...")

	// 优雅关闭评估服务
	if err := evaluator.Stop(); err != nil {
		log.Printf("Error shutting down evaluator: %v", err)
	}

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Application has been shut down.")
}
