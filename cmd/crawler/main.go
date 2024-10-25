package main

import (
	"flag"
	"fmt"
	"log"
	"qinniu/internal/crawler"
	"qinniu/internal/models"
	"qinniu/internal/pkg/cache"
	"qinniu/internal/pkg/database"
	"strings"
	"sync"
	"time"

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

	// 支持多用户名输入
	usernames := flag.String("users", "", "GitHub usernames to analyze (comma-separated)")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent crawlers")
	flag.Parse()

	if *usernames == "" {
		log.Fatal("Please provide GitHub usernames using -users flag")
	}

	// 创建爬虫实例
	crawlerInstance := crawler.NewGitHubCrawler()

	// 创建工作池
	userChan := make(chan string)
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for username := range userChan {
				developer, err := processUserWithRetry(crawlerInstance, username)
				if err != nil {
					log.Printf("Error processing user %s: %v\n", username, err)
					continue
				}

				// 打印结果
				printResult(developer)

				// 保存到数据库
				if err := developer.Create(); err != nil {
					log.Printf("Error saving user %s: %v\n", username, err)
				}
			}
		}()
	}

	// 发送用户名到通道
	userList := strings.Split(*usernames, ",")
	for _, username := range userList {
		username = strings.TrimSpace(username)
		if username != "" {
			userChan <- username
		}
	}
	close(userChan)

	// 等待所有工作完成
	wg.Wait()
}

// 添加重试机制
func processUserWithRetry(crawler *crawler.GitHubCrawler, username string) (*models.Developer, error) {
	fmt.Printf("\nProcessing user: %s\n", username) // 添加处理提示

	var developer *models.Developer
	var err error

	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		developer, err = crawler.GetUserData(username)
		if err == nil {
			return developer, nil
		}

		log.Printf("Attempt %d failed: %v\n", i+1, err) // 添加错误日志

		if strings.Contains(err.Error(), "rate limit") {
			time.Sleep(time.Minute)
			continue
		}

		time.Sleep(retryDelay)
		retryDelay *= 2
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

func printResult(developer *models.Developer) {
	if developer == nil {
		return
	}

	fmt.Println("\n====================================")
	fmt.Printf("Username: %s\n", developer.Username)
	if developer.Name != "" {
		fmt.Printf("Name: %s\n", developer.Name)
	}
	fmt.Printf("Location: %s\n", developer.Location)
	if developer.Nation != "" {
		fmt.Printf("Nation: %s (置信度: %.2f%%)\n", developer.Nation, developer.NationConfidence)
	}
	fmt.Printf("TalentRank: %.2f\n", developer.TalentRank)
	fmt.Printf("Confidence: %.2f%%\n", developer.Confidence)

	// 修改这里，添加更多的调试信息
	fmt.Printf("Avatar URL: %s\n", developer.Avatar)
	if developer.Avatar == "" {
		fmt.Println("Warning: Avatar URL is empty!")
	}

	if len(developer.Skills) > 0 {
		fmt.Printf("Skills: %v\n", developer.Skills)
	}
	fmt.Printf("Number of Repositories: %d\n", len(developer.Repositories))

	// 添加数据库操作的调试信息
	fmt.Printf("Database ID: %s\n", developer.ID.Hex())
	fmt.Printf("Created At: %v\n", developer.CreatedAt)
	fmt.Printf("Updated At: %v\n", developer.UpdatedAt)

	fmt.Println("Saved to database successfully!")
	fmt.Println("====================================\n")
}
