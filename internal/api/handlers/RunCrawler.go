package handlers

import (
	"fmt"
	"net/http"
	"qinniu/internal/models"
	"strings"

	"sync"

	githubcrawler "qinniu/internal/crawler"

	"github.com/gin-gonic/gin"
)

// RunCrawlerRequest 定义HTTP请求的JSON结构
type RunCrawlerRequest struct {
	Usernames   string `json:"usernames" binding:"required"`   // 逗号分隔的GitHub用户名
	Concurrency int    `json:"concurrency" binding:"required"` // 并发数
}

type CrawlResult struct {
	Username string `json:"username"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// RunCrawlerHandler 处理 /run-crawler 路由的请求
func RunCrawlerHandler(c *gin.Context) {

	var req RunCrawlerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Concurrency <= 0 || req.Concurrency > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Concurrency must be between 1 and 6"})
		return
	}

	userList := strings.Split(req.Usernames, ",")
	for i, username := range userList {
		userList[i] = strings.TrimSpace(username)
	}

	crawlerInstance := githubcrawler.NewGitHubCrawler()
	userChan := make(chan string)
	var wg sync.WaitGroup
	var results []CrawlResult
	var mu sync.Mutex

	// 启动工作协程
	for i := 0; i < req.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for username := range userChan {
				developer, err := crawlerInstance.GetUserData(username)
				if err != nil {
					mu.Lock()
					results = append(results, CrawlResult{
						Username: username,
						Success:  false,
						Error:    err.Error(),
					})
					mu.Unlock()
					continue
				}
				PrintResult(developer)
			}
		}()
	}

	// 发送用户名到通道
	go func() {
		for _, username := range userList {
			if username != "" {
				userChan <- username
			}
		}
		close(userChan)
	}()

	// 等待所有工作完成
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"status":  "Crawling completed",
		"results": results,
	})
}
func PrintResult(developer *models.Developer) {
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

	// 添加更多的调试信息
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
