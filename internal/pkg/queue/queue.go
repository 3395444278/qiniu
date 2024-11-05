package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	QueueName = "developer_evaluation"
)

type Queue interface {
	Publish(task *EvaluationTask) error
	Subscribe(handler func(*EvaluationTask) error)
}

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

func NewQueue() Queue {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	log.Printf("Connecting to Redis at %s", redisAddr)

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		DB:       0,
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Error connecting to Redis: %v", err)
	} else {
		log.Println("Successfully connected to Redis")
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}
}

func (q *RedisQueue) Publish(task *EvaluationTask) error {
	log.Printf("Publishing evaluation task for user: %s", task.Username)

	data, err := json.Marshal(task)
	if err != nil {
		log.Printf("Error marshaling task: %v", err)
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	err = q.client.LPush(q.ctx, QueueName, data).Err()
	if err != nil {
		log.Printf("Error publishing task to Redis: %v", err)
		return err
	}

	log.Printf("Successfully published task for user: %s to Redis queue", task.Username)
	return nil
}

func (q *RedisQueue) Subscribe(handler func(*EvaluationTask) error) {
	log.Printf("Starting to listen for tasks on queue: %s", QueueName)

	for {
		// 使用 BRPOP 阻塞等待任务
		result, err := q.client.BRPop(q.ctx, 0, QueueName).Result()
		if err != nil {
			log.Printf("Error getting task from queue: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Received task from queue")

		var task EvaluationTask
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("Error unmarshaling task: %v", err)
			continue
		}

		log.Printf("Processing task for user: %s", task.Username)

		// 同步处理任务，确保错误被正确处理
		if err := handler(&task); err != nil {
			log.Printf("Error handling task for %s: %v", task.Username, err)
			// 失败重试
			time.Sleep(time.Second * 5)
			if err := q.Publish(&task); err != nil {
				log.Printf("Error re-queuing task: %v", err)
			}
		} else {
			log.Printf("Successfully processed task for user: %s", task.Username)
		}
	}
}
