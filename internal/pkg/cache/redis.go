package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"qinniu/internal/models"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return RedisClient.Ping(context.Background()).Err()
}

func CacheUser(username string, data interface{}, expiration time.Duration) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, "user:"+username, jsonData, expiration).Err()
}

func GetCachedUser(username string) ([]byte, error) {
	ctx := context.Background()
	return RedisClient.Get(ctx, "user:"+username).Bytes()
}

// CacheDeveloper 缓存开发者数据
func CacheDeveloper(developer interface{}) error {
	key := fmt.Sprintf("developer:%v", developer.(*models.Developer).Username)
	return setCache(key, developer, 24*time.Hour)
}

// GetCachedDeveloper 获取��存的开发者数据
func GetCachedDeveloper(username string) (interface{}, error) {
	key := fmt.Sprintf("developer:%s", username)
	data, err := RedisClient.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var developer interface{}
	if err := json.Unmarshal(data, &developer); err != nil {
		return nil, err
	}
	return developer, nil
}

// CacheRepositories 缓存仓库列表
func CacheRepositories(username string, repos []string) error {
	ctx := context.Background()
	key := "repos:" + username

	// 将仓库列表序列化为JSON
	data, err := json.Marshal(repos)
	if err != nil {
		return err
	}

	// 缓存数据，设置12小时过期
	return RedisClient.Set(ctx, key, data, 12*time.Hour).Err()
}

// GetCachedRepositories 获取缓存的仓库列表
func GetCachedRepositories(username string) ([]string, error) {
	ctx := context.Background()
	key := "repos:" + username

	// 从缓存获取数据
	data, err := RedisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	}
	if err != nil {
		return nil, err
	}

	// 反序列化数据
	var repos []string
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// ClearCache 除用户相关的所有缓存
func ClearCache(username string) error {
	ctx := context.Background()
	keys := []string{
		"developer:" + username,
		"repos:" + username,
	}

	for _, key := range keys {
		if err := RedisClient.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	return nil
}

// 添加缓存预热功能
func WarmupCache(usernames []string, fetchFunc func(string) (interface{}, error)) error {
	// 并发预热
	concurrency := 5
	sem := make(chan struct{}, concurrency)
	errChan := make(chan error, len(usernames))

	for _, username := range usernames {
		sem <- struct{}{} // 获取信号量
		go func(user string) {
			defer func() { <-sem }() // 释放信号量

			// 检查缓存是否已存在且未过期
			if exists, _ := RedisClient.Exists(context.Background(), "developer:"+user).Result(); exists == 1 {
				return
			}

			// 使用传入的函数获取数据
			if data, err := fetchFunc(user); err == nil {
				if err := CacheDeveloper(data); err != nil {
					errChan <- err
				}
			}
		}(username)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < concurrency; i++ {
		sem <- struct{}{}
	}

	// 检查错误
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// setCache 设置缓存
func setCache(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, jsonData, expiration).Err()
}
