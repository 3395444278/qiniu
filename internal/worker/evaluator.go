package worker

import (
	"context"
	"log"
	"qinniu/internal/models"
	"qinniu/internal/pkg/ai"
	"qinniu/internal/pkg/queue"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Evaluator struct {
	aiClient *ai.Client
	queue    queue.Queue
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewEvaluator(aiClient *ai.Client, queue queue.Queue) *Evaluator {
	return &Evaluator{
		aiClient: aiClient,
		queue:    queue,
		quit:     make(chan struct{}),
	}
}

func (e *Evaluator) Start() error {
	log.Println("Starting evaluator worker...")
	e.queue.Subscribe(e.ProcessEvaluationTask)
	return nil
}
func (e *Evaluator) Stop() error {
	log.Println("Stopping evaluator worker...")
	close(e.quit)
	e.wg.Wait()
	log.Println("Evaluator worker stopped.")
	return nil
}
func (e *Evaluator) ProcessEvaluationTask(task *queue.EvaluationTask) error {
	log.Printf("Processing evaluation task for user: %s", task.Username)

	// 1. 获取用户信息
	developer, err := models.FindByUsername(task.Username)
	if err != nil {
		log.Printf("Error finding developer %s: %v", task.Username, err)
		return err
	}

	// 2. 收集更广泛的评估信息
	info := map[string]interface{}{
		// 基本信息
		"username": developer.Username,
		"name":     developer.Name,
		"bio":      task.Description,
		"location": developer.Location,
		"email":    developer.Email,

		// 链接信息
		"blog":        task.BlogURL,
		"profile_url": task.ProfileURL,
		"repos":       developer.Repositories,
		"repo_urls":   developer.RepositoryURLs,

		// 技术信息
		"skills":    developer.Skills,
		"languages": developer.Skills, // 所有使用的编程语言

		// 统计信息
		"stars":      developer.StarCount,
		"commits":    developer.CommitCount,
		"forks":      developer.ForkCount,
		"repo_stars": developer.RepoStars,

		// 活跃度信息
		"last_active": developer.LastUpdated,
		"created_at":  developer.CreatedAt,
		"updated_at":  developer.UpdatedAt,
	}

	// 3. 使用 AI 进行评估
	evaluation, err := e.aiClient.EvaluateDeveloper(context.Background(), info)
	if err != nil {
		log.Printf("Error evaluating developer %s: %v", task.Username, err)
		return err
	}

	// 4. 更新开发者信息
	developer.TechEvaluation = models.TechEvaluation{
		BlogURL:         task.BlogURL,
		PersonalSiteURL: task.ProfileURL,
		Biography:       task.Description,
		Specialties:     evaluation.Specialties,
		Experience:      evaluation.Experience,
		AIEvaluation:    evaluation.AIEvaluation,
		LastEvaluated:   time.Now(),
	}

	// 5. 直接更新数据库中的 tech_evaluation 字段
	collection := models.GetCollection()
	filter := bson.M{"username": task.Username}
	update := bson.M{
		"$set": bson.M{
			"tech_evaluation": developer.TechEvaluation,
			"updated_at":      time.Now(),
		},
	}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating tech evaluation for %s: %v", task.Username, err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Printf("Warning: No documents updated for %s", task.Username)
	} else {
		log.Printf("Successfully updated tech evaluation for %s", task.Username)
	}

	return nil
}
