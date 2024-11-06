package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"qinniu/internal/pkg/database"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Developer struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username         string             `bson:"username" json:"username"`
	Name             string             `bson:"name" json:"name"`
	Email            string             `bson:"email" json:"email"`
	Location         string             `bson:"location" json:"location"`
	Nation           string             `bson:"nation" json:"nation"`
	NationConfidence float64            `bson:"nation_confidence" json:"nation_confidence"`
	TalentRank       float64            `bson:"talent_rank" json:"talent_rank"`
	Confidence       float64            `bson:"confidence" json:"confidence"`
	Skills           []string           `bson:"skills" json:"skills"`
	Repositories     []string           `bson:"repositories" json:"repositories"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	LastActive       time.Time          `bson:"last_active" json:"last_active"`
	CommitCount      int                `bson:"commit_count" json:"commit_count"`
	StarCount        int                `bson:"star_count" json:"star_count"`
	ForkCount        int                `bson:"fork_count" json:"fork_count"` // 新增
	LastUpdated      time.Time          `bson:"last_updated" json:"last_updated"`
	DataValidation   ValidationResult   `bson:"data_validation" json:"data_validation"`
	UpdateFrequency  time.Duration      `bson:"update_frequency" json:"update_frequency"`
	Avatar           string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	ProfileURL       string             `bson:"profile_url,omitempty" json:"profile_url,omitempty"`
	RepositoryURLs   map[string]string  `bson:"repository_urls,omitempty" json:"repository_urls,omitempty"`
	RepoStars        map[string]int     `bson:"repo_stars,omitempty" json:"repo_stars,omitempty"`
	TechEvaluation   TechEvaluation     `bson:"tech_evaluation,omitempty" json:"tech_evaluation,omitempty"`
	// 添加其他必要的字段
}

// 新增：数据验证结果
type ValidationResult struct {
	IsValid       bool      `bson:"is_valid" json:"is_valid"`
	Confidence    float64   `bson:"confidence" json:"confidence"`
	LastValidated time.Time `bson:"last_validated" json:"last_validated"`
	Issues        []string  `bson:"issues,omitempty" json:"issues,omitempty"` // 添加 omitempty
}

type TechEvaluation struct {
	BlogURL         string            `bson:"blog_url,omitempty" json:"blog_url,omitempty"`
	PersonalSiteURL string            `bson:"personal_site_url,omitempty" json:"personal_site_url,omitempty"`
	Biography       string            `bson:"biography,omitempty" json:"biography,omitempty"`
	Specialties     []string          `bson:"specialties,omitempty" json:"specialties,omitempty"`
	Experience      map[string]string `bson:"experience,omitempty" json:"experience,omitempty"`
	AIEvaluation    string            `bson:"ai_evaluation,omitempty" json:"ai_evaluation,omitempty"`
	LastEvaluated   time.Time         `bson:"last_evaluated,omitempty" json:"last_evaluated,omitempty"`
}

const collectionName = "developers"

// GetCollection 获取开发者集合
func GetCollection() *mongo.Collection {
	return database.DB.Collection(collectionName)
}

// init 初始化 MongoDB 选项
func init() {
	// 设置全局选项，允许截断
	registry := bson.NewRegistry()
	registry.RegisterTypeMapEntry(bson.TypeDouble, reflect.TypeOf(int32(0)))

	// 应用到现有的 MongoDB 客户端
	if database.DB != nil {
		clientOpts := options.Client().
			SetRegistry(registry)

		// 更新客户端选项
		if client, err := mongo.NewClient(clientOpts); err == nil {
			database.DB = client.Database(database.DB.Name())
		}
	}
}

// Create 创建新开发者
func (d *Developer) Create() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 添加调试日志
	log.Printf("Debug - Creating developer with Avatar URL: %s", d.Avatar)

	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()

	// 确保在插入新文档时，不设置 _id 字段
	d.ID = primitive.NilObjectID

	// 插入文档，包括 Avatar 字段
	result, err := GetCollection().InsertOne(ctx, d)
	if err != nil {
		log.Printf("Error creating developer: %v", err)
		return err
	}

	d.ID = result.InsertedID.(primitive.ObjectID)

	// 验证插入后的数据
	var inserted Developer
	err = GetCollection().FindOne(ctx, bson.M{"_id": d.ID}).Decode(&inserted)
	if err != nil {
		log.Printf("Error verifying inserted data: %v", err)
	} else {
		log.Printf("Debug - Verified Avatar URL after insert: %s", inserted.Avatar)
	}

	return nil
}

// Update 更新开发者信息
func (d *Developer) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查是否有有效的 ID
	if d.ID.IsZero() {
		return fmt.Errorf("更新失败，没有有效的 ID")
	}

	d.UpdatedAt = time.Now()

	// 用 _id 作为查询条件
	filter := bson.M{"_id": d.ID}

	// 构建更新文档，排除 _id 字段
	update := bson.M{
		"$set": bson.M{
			"username":          d.Username,
			"name":              d.Name,
			"email":             d.Email,
			"location":          d.Location,
			"nation":            d.Nation,
			"nation_confidence": d.NationConfidence,
			"talent_rank":       d.TalentRank,
			"confidence":        d.Confidence,
			"skills":            d.Skills,
			"repositories":      d.Repositories,
			"updated_at":        d.UpdatedAt,
			"last_active":       d.LastActive,
			"commit_count":      d.CommitCount,
			"star_count":        d.StarCount,
			"fork_count":        d.ForkCount, // 新增
			"last_updated":      d.LastUpdated,
			"data_validation":   d.DataValidation,
			"update_frequency":  d.UpdateFrequency,
			"avatar":            d.Avatar, // 确保包含 Avatar 字段
			"profile_url":       d.ProfileURL,
			"repository_urls":   d.RepositoryURLs,
			"repo_stars":        d.RepoStars,
			"tech_evaluation":   d.TechEvaluation,
			// 不要包含 "_id" 字段
		},
	}

	_, err := GetCollection().UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除开发者
func (d *Developer) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := GetCollection().DeleteOne(ctx, bson.M{"_id": d.ID})
	return err
}

// FindByID 通过ID查找开发者
func FindByID(id string) (*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var developer Developer
	err = GetCollection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&developer)
	if err != nil {
		return nil, err
	}

	return &developer, nil
}

// FindByUsername 通过用户名查找开发者

func FindByUsername(username string) (*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var developer Developer
	err := GetCollection().FindOne(ctx, bson.M{"username": username}).Decode(&developer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 返回 nil, nil 表示未找到
		}
		return nil, err
	}

	return &developer, nil
}

// FindAll 获取所有开发者
func FindAll(page, pageSize int64) ([]*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用聚合管道来处理数据
	pipeline := []bson.M{
		{"$sort": bson.M{"talent_rank": -1}},
		{"$skip": (page - 1) * pageSize},
		{"$limit": pageSize},
		{"$project": bson.M{
			"_id":               1,
			"username":          1,
			"name":              1,
			"email":             1,
			"location":          1,
			"nation":            1,
			"nation_confidence": 1,
			"talent_rank":       bson.M{"$toInt": "$talent_rank"},
			"confidence":        1,
			"skills":            1,
			"repositories":      1,
			"created_at":        1,
			"updated_at":        1,
			"last_active":       1,
			"commit_count":      1,
			"star_count":        1,
			"fork_count":        1,
			"last_updated":      1,
			"avatar":            1,
			"update_frequency":  1,
		}},
	}

	// 执合查询
	cursor, err := GetCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developers []*Developer
	if err = cursor.All(ctx, &developers); err != nil {
		return nil, err
	}

	// 如果没有结果，返回空数组而不是 nil
	if developers == nil {
		developers = make([]*Developer, 0)
	}

	return developers, nil
}

// Search 搜索开发者
func Search(query bson.M, page, pageSize int64) ([]*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.D{{Key: "talent_rank", Value: -1}})

	cursor, err := GetCollection().Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developers []*Developer
	if err = cursor.All(ctx, &developers); err != nil {
		return nil, err
	}

	return developers, nil
}

// BatchCreate 批量创建开者
func BatchCreate(developers []*Developer) error {
	if len(developers) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 准备批量插入的文档
	docs := make([]interface{}, len(developers))
	for i, dev := range developers {
		dev.CreatedAt = time.Now()
		dev.UpdatedAt = time.Now()
		docs[i] = dev
	}

	// 执行批量插入
	result, err := GetCollection().InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	// 更新ID
	for i, id := range result.InsertedIDs {
		developers[i].ID = id.(primitive.ObjectID)
	}

	return nil
}

// SearchWithOptions 带选项的搜索开发者
func SearchWithOptions(query bson.M, page, pageSize int64, opts ...*options.FindOptions) ([]*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 添加分页选项
	findOpts := options.MergeFindOptions(opts...)
	findOpts.SetSkip((page - 1) * pageSize)
	findOpts.SetLimit(pageSize)

	cursor, err := GetCollection().Find(ctx, query, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developers []*Developer
	if err = cursor.All(ctx, &developers); err != nil {
		return nil, err
	}

	return developers, nil
}

// 增：定期更新机制
func (d *Developer) ShouldUpdate() bool {
	return time.Since(d.LastUpdated) > d.UpdateFrequency
}

// FindTopDevelopers 获取排靠前的开发者
func FindTopDevelopers(limit int) ([]*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "talent_rank", Value: -1}})

	cursor, err := GetCollection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developers []*Developer
	if err = cursor.All(ctx, &developers); err != nil {
		return nil, err
	}

	return developers, nil
}

// AggregateSearch 执行聚合查询
func AggregateSearch(pipeline []bson.M) ([]*Developer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 在管道末尾添加类型转换，确保所有数值字段都是正确的类型
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"_id":               1,
			"username":          1,
			"name":              1,
			"email":             1,
			"location":          1,
			"nation":            1,
			"nation_confidence": bson.M{"$toDouble": "$nation_confidence"},
			"talent_rank":       bson.M{"$toDouble": "$talent_rank"},
			"confidence":        bson.M{"$toDouble": "$confidence"},
			"skills":            1,
			"repositories":      1,
			"created_at":        1,
			"updated_at":        1,
			"last_active":       1,
			"commit_count":      bson.M{"$toInt": "$commit_count"},
			"star_count":        bson.M{"$toInt": "$star_count"},
			"fork_count":        bson.M{"$toInt": "$fork_count"},
			"last_updated":      1,
			"avatar":            1,
			"update_frequency":  1,
			"profile_url":       1,
			"repository_urls":   1,
			"repo_stars":        1,
			"tech_evaluation":   1,
		},
	})

	cursor, err := GetCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developers []*Developer
	if err = cursor.All(ctx, &developers); err != nil {
		return nil, err
	}

	// 如果没有结果，返回空数组而不是 nil
	if developers == nil {
		developers = make([]*Developer, 0)
	}

	return developers, nil
}

// DeleteByUsername 删除指定用户名的所有记录
func DeleteByUsername(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := GetCollection().DeleteMany(ctx, bson.M{"username": username})
	return err
}

// CountDevelopers 统计符合条件的开发者数量
func CountDevelopers(query bson.M) (int64, error) {
	collection := GetCollection() // 使用已定义的 getCollection 函数
	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
