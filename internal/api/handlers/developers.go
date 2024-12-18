package handlers

import (
	"context"
	"fmt"
	"net/http"
	"qinniu/internal/models"

	"strconv"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 添加领域技能映射
var domainSkills = map[string][]string{
	"backend": {
		"Go", "Java", "Python", "Ruby", "PHP", "C++", "C#",
		"Node.js", "Rust", "Scala", "Kotlin", "Spring",
		"Django", "Laravel", "Express",
	},
	"frontend": {
		"JavaScript", "TypeScript", "React", "Vue", "Angular",
		"HTML", "CSS", "Svelte", "Next.js", "Nuxt.js",
		"Webpack", "Sass", "Less", "TailwindCSS",
	},
	"mobile": {
		"Swift", "Kotlin", "Java", "Objective-C", "Flutter",
		"React Native", "Android", "iOS", "Xamarin", "Dart",
	},
	"ai": {
		"Python", "TensorFlow", "PyTorch", "Jupyter Notebook",
		"R", "Scikit-learn", "Pandas", "NumPy", "CUDA", "OpenCV",
	},
	"devops": {
		"Docker", "Kubernetes", "Jenkins", "Ansible", "Terraform",
		"Shell", "AWS", "Azure", "GCP", "GitLab", "CircleCI",
		"Prometheus", "Grafana",
	},
	"database": {
		"SQL", "MongoDB", "Redis", "PostgreSQL", "MySQL",
		"Oracle", "Cassandra", "Elasticsearch",
	},
	"security": {
		"Python", "C", "Assembly", "Shell", "Ruby", "Go",
		"Metasploit", "Wireshark", "Burp Suite",
	},
	"blockchain": {
		"Solidity", "Go", "JavaScript", "Rust", "C++",
		"Web3.js", "Ethereum", "Smart Contracts",
	},
	"gamedev": {
		"C++", "C#", "Unity", "Unreal Engine", "JavaScript",
		"OpenGL", "DirectX", "Vulkan", "SDL", "SFML",
	},
	"embedded": {
		"C", "C++", "Assembly", "Arduino", "Raspberry Pi",
		"RTOS", "ARM", "IoT",
	},
	"systems": {
		"C", "C++", "Rust", "Go", "Assembly", "Linux",
		"Windows", "Kernel", "Drivers",
	},
}

// SearchDevelopers 搜索开发者
func SearchDevelopers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	// 构建查询条件
	conditions := []bson.M{}

	// 1. 基本信息模糊搜索
	if keyword := c.Query("keyword"); keyword != "" {
		// 关键词可以匹配用户名、姓名、邮箱或位置
		keywordQuery := bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": keyword, "$options": "i"}},
				{"name": bson.M{"$regex": keyword, "$options": "i"}},
				{"email": bson.M{"$regex": keyword, "$options": "i"}},
				{"location": bson.M{"$regex": keyword, "$options": "i"}},
			},
		}
		conditions = append(conditions, keywordQuery)
	}

	// 添加姓名模糊查询
	if name := c.Query("name"); name != "" {
		nameQuery := bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": name, "$options": "i"}},
				{"username": bson.M{"$regex": name, "$options": "i"}},
			},
		}
		conditions = append(conditions, nameQuery)
	}

	// 2. 按领域搜索
	if domain := c.Query("domain"); domain != "" {
		if skills, exists := domainSkills[strings.ToLower(domain)]; exists {
			conditions = append(conditions, bson.M{"skills": bson.M{"$in": skills}})
		}
	}

	// 3. 按国家筛选（支持多个国家）
	if nations := strings.Split(c.Query("nations"), ","); len(nations) > 0 {
		nationQueries := make([]bson.M, 0)
		for _, nation := range nations {
			if nation = strings.TrimSpace(nation); nation != "" {
				nationQueries = append(nationQueries, bson.M{
					"nation":            strings.ToUpper(nation),
					"nation_confidence": bson.M{"$gte": 60},
				})
			}
		}
		if len(nationQueries) > 0 {
			conditions = append(conditions, bson.M{"$or": nationQueries})
		}
	}

	// 4. 按技能筛选（支持多个技能）
	if skills := c.QueryArray("skills"); len(skills) > 0 {
		conditions = append(conditions, bson.M{"skills": bson.M{"$all": skills}})
	}

	// 5. 按活跃度筛选
	if minActivity := c.Query("min_activity"); minActivity != "" {
		if activityDays, err := strconv.Atoi(minActivity); err == nil {
			conditions = append(conditions, bson.M{
				"last_updated": bson.M{
					"$gte": time.Now().AddDate(0, 0, -activityDays),
				},
			})
		}
	}

	// 6. 按贡献度筛选
	if minCommits := c.Query("min_commits"); minCommits != "" {
		if commits, err := strconv.Atoi(minCommits); err == nil {
			conditions = append(conditions, bson.M{"commit_count": bson.M{"$gte": commits}})
		}
	}

	// 7. 按影响力筛选
	if minStars := c.Query("min_stars"); minStars != "" {
		if stars, err := strconv.Atoi(minStars); err == nil {
			conditions = append(conditions, bson.M{"star_count": bson.M{"$gte": stars}})
		}
	}

	// 8. 按 TalentRank 筛选
	if minRank := c.Query("min_rank"); minRank != "" {
		if rankFloat, err := strconv.ParseFloat(minRank, 64); err == nil {
			conditions = append(conditions, bson.M{"talent_rank": bson.M{"$gte": rankFloat}})
		}
	}

	// 9. 按更新时间范围筛选
	if updatedAfter := c.Query("updated_after"); updatedAfter != "" {
		if t, err := time.Parse(time.RFC3339, updatedAfter); err == nil {
			conditions = append(conditions, bson.M{"updated_at": bson.M{"$gte": t}})
		}
	}

	// 添加仓库星星数范围筛选
	if repoStars := c.Query("repo_stars"); repoStars != "" {
		stars, err := strconv.Atoi(repoStars)
		if err == nil {
			conditions = append(conditions, bson.M{
				"repo_stars": bson.M{"$elemMatch": bson.M{"$gte": stars}},
			})
		}
	}

	// 添加特定仓库名称搜索
	if repoName := c.Query("repo_name"); repoName != "" {
		conditions = append(conditions, bson.M{
			"repository_urls": bson.M{"$exists": true},
			"$where":          fmt.Sprintf("Object.keys(this.repository_urls).some(k => k.toLowerCase().includes('%s'.toLowerCase()))", repoName),
		})
	}

	// 组合查询条件
	query := bson.M{}
	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	// 排序选项
	sortField := c.DefaultQuery("sort_by", "talent_rank")
	sortOrder := -1                    // 默认降序
	if c.Query("sort_asc") == "true" { // 修改这里，只检查 "true"
		sortOrder = 1
	}

	// 构建聚合管道
	pipeline := []bson.M{
		{"$match": query},
		{"$group": bson.M{ // 添加去重
			"_id": "$username",
			"doc": bson.M{"$first": "$$ROOT"},
		}},
		{"$replaceRoot": bson.M{"newRoot": "$doc"}},
		{"$sort": bson.M{sortField: sortOrder}},
		{"$skip": (page - 1) * pageSize},
		{"$limit": pageSize},
		// 修改投影，只包含需要的字段
		{"$project": bson.M{
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
			"profile_url":       1,
			"repository_urls":   1,
			"repo_stars":        1,
			"data_validation":   1,
			"update_frequency":  1,
			// 不包含 tech_evaluation 字段，而不是显式排除
		}},
	}

	// 执行聚合查询
	developers, err := models.AggregateSearch(pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取总数
	total, err := models.CountDevelopers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"developers": developers,
	})
}

// GetDeveloper 获取单个开发者
func GetDeveloper(c *gin.Context) {
	id := c.Param("id")
	developer, err := models.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "开发者不存在"})
		return
	}

	c.JSON(http.StatusOK, developer)
}

// CreateDeveloper 创建开发者
func CreateDeveloper(c *gin.Context) {
	var developer models.Developer
	if err := c.ShouldBindJSON(&developer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := developer.Create(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, developer)
}

// UpdateDeveloper 更新开发者
func UpdateDeveloper(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var developer models.Developer
	if err := c.ShouldBindJSON(&developer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	developer.ID = objectID
	if err := developer.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, developer)
}

// DeleteDeveloper 删除开发者
func DeleteDeveloper(c *gin.Context) {
	id := c.Param("id")
	developer, err := models.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "开发者不存在"})
		return
	}

	if err := developer.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "开发者已删除"})
}

// GetAllNations 获取所有国家
func GetAllNations(c *gin.Context) {
	// 使用聚合管道获取去重的国家列表
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"nation": bson.M{
					"$exists": true,
					"$ne":     "",
				},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$nation",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"nation": "$_id",
				"count":  1,
				"_id":    0,
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	collection := models.GetCollection()
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	type NationInfo struct {
		Nation string `json:"nation"`
		Count  int    `json:"count"`
	}

	var nations []NationInfo
	if err := cursor.All(context.Background(), &nations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nations,
	})
}
