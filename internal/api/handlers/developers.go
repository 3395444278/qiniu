package handlers

import (
	"log"
	"net/http"
	"qinniu/internal/models"

	"strconv"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// GetDevelopers 获取开发者列表
func GetDevelopers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	// 使用聚合管道来处理数据，添加去重逻辑
	pipeline := []bson.M{
		// 首先按更新时间排序，确保获取最新数据
		{"$sort": bson.M{"updated_at": -1}},
		// 按用户名去重，保留最新的记录
		{"$group": bson.M{
			"_id": "$username",
			"doc": bson.M{"$first": "$$ROOT"},
		}},
		// 恢复文档结构
		{"$replaceRoot": bson.M{"newRoot": "$doc"}},
		// 按 talent_rank 排序
		{"$sort": bson.M{"talent_rank": -1}},
		// 分页
		{"$skip": (page - 1) * pageSize},
		{"$limit": pageSize},
		// 投影
		{"$project": bson.M{
			"_id":               1,
			"username":          1,
			"name":              1,
			"email":             1,
			"location":          1,
			"nation":            1,
			"nation_confidence": 1,
			"talent_rank":       1,
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
			"data_validation": bson.M{
				"is_valid":       "$data_validation.is_valid",
				"confidence":     "$data_validation.confidence",
				"last_validated": "$data_validation.last_validated",
				"issues":         bson.M{"$ifNull": []interface{}{"$data_validation.issues", primitive.A{}}},
			},
			"update_frequency": 1,
		}},
	}

	developers, err := models.AggregateSearch(pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加调试日志
	for _, dev := range developers {
		log.Printf("Debug - Developer %s has avatar URL: %s", dev.Username, dev.Avatar)
	}

	c.JSON(http.StatusOK, developers)
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

// SearchDevelopers 搜索开发者
func SearchDevelopers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	// 构建查询条件
	conditions := []bson.M{}

	// 1. 按领域搜索
	if domain := c.Query("domain"); domain != "" {
		if skills, exists := domainSkills[strings.ToLower(domain)]; exists {
			conditions = append(conditions, bson.M{"skills": bson.M{"$in": skills}})
		}
	}

	// 2. 按国家筛选
	if nation := c.Query("nation"); nation != "" {
		nationQuery := bson.M{
			"nation":            strings.ToUpper(nation),
			"nation_confidence": bson.M{"$gte": 60},
		}
		conditions = append(conditions, nationQuery)
	}

	// 3. 按技能筛选
	if skill := c.Query("skills"); skill != "" {
		skillQuery := bson.M{"skills": bson.M{"$in": []string{skill}}}
		conditions = append(conditions, skillQuery)
	}

	// 4. 按最低 TalentRank 选
	if minRank := c.Query("min_rank"); minRank != "" {
		if rankFloat, err := strconv.ParseFloat(minRank, 64); err == nil {
			conditions = append(conditions, bson.M{"talent_rank": bson.M{"$gte": rankFloat}})
		}
	}

	// 组合所有条件
	query := bson.M{}
	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	// 5. 使用聚合管道进行查询和去重
	pipeline := []bson.M{
		{"$match": query},
		{"$sort": bson.M{"talent_rank": -1}},
		{"$group": bson.M{
			"_id": "$username",
			"doc": bson.M{"$first": "$$ROOT"},
		}},
		{"$replaceRoot": bson.M{"newRoot": "$doc"}},
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
			"talent_rank":       1,
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
			"avatar":            1, // 确保包含 avatar 字段
			"data_validation": bson.M{
				"is_valid":       "$data_validation.is_valid",
				"confidence":     "$data_validation.confidence",
				"last_validated": "$data_validation.last_validated",
				"issues":         bson.M{"$ifNull": []interface{}{"$data_validation.issues", primitive.A{}}},
			},
			"update_frequency": 1,
		}},
	}

	// 执行聚合查询
	developers, err := models.AggregateSearch(pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加调试日志
	for _, dev := range developers {
		log.Printf("Debug - Developer %s has avatar URL: %s", dev.Username, dev.Avatar)
	}

	c.JSON(http.StatusOK, developers)
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

// AdvancedSearch 高级搜索功能
func AdvancedSearch(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	// 构建高级搜索查询
	query := buildAdvancedQuery(
		c.Query("domains"),           // 领域
		c.Query("skills"),            // 技能
		c.Query("nations"),           // 国家
		c.Query("talent_rank_range"), // TalentRank范围
		c.Query("activity_period"),   // 活跃时间段
		c.Query("project_type"),      // 项目类型
		c.Query("contribution_type"), // 贡献类型
	)

	// 排序
	sortField := c.DefaultQuery("sort", "talent_rank")
	sortOrder := c.DefaultQuery("order", "desc")

	var sortValue int
	if sortOrder == "asc" {
		sortValue = 1
	} else {
		sortValue = -1
	}

	// 创建排序选项
	sortOpts := bson.D{{Key: sortField, Value: sortValue}}

	// 执行搜索
	developers, err := models.SearchWithOptions(query, page, pageSize, options.Find().
		SetSort(sortOpts))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data":      developers,
		"page":      page,
		"page_size": pageSize,
		"query":     query, // 返回查询件，方便调试
	})
}

// 新增：趋势分析
func AnalyzeTrends(c *gin.Context) {
	// 分析开发者在不同时期的表现
	trends := analyzeDeveloperTrends(
		c.Param("id"),
		c.Query("period"),
		c.Query("metrics"),
	)

	c.JSON(http.StatusOK, trends)
}

// buildAdvancedQuery 构建高级搜索查询
func buildAdvancedQuery(domains, skills, nations, rankRange, activityPeriod, projectType, contributionType string) bson.M {
	query := bson.M{}

	// 处理领域
	if domains != "" {
		domainList := strings.Split(domains, ",")
		query["domains"] = bson.M{"$in": domainList}
	}

	// 处理技能
	if skills != "" {
		skillList := strings.Split(skills, ",")
		query["skills"] = bson.M{"$all": skillList}
	}

	// 处理国家
	if nations != "" {
		nationList := strings.Split(nations, ",")
		query["nation"] = bson.M{"$in": nationList}
	}

	// 处理TalentRank范围
	if rankRange != "" {
		ranges := strings.Split(rankRange, "-")
		if len(ranges) == 2 {
			min, err1 := strconv.ParseFloat(ranges[0], 64)
			max, err2 := strconv.ParseFloat(ranges[1], 64)
			if err1 == nil && err2 == nil {
				query["talent_rank"] = bson.M{
					"$gte": min,
					"$lte": max,
				}
			}
		}
	}

	// 处理活跃时间段
	if activityPeriod != "" {
		switch activityPeriod {
		case "1m":
			query["last_active"] = bson.M{"$gte": time.Now().AddDate(0, -1, 0)}
		case "6m":
			query["last_active"] = bson.M{"$gte": time.Now().AddDate(0, -6, 0)}
		case "1y":
			query["last_active"] = bson.M{"$gte": time.Now().AddDate(-1, 0, 0)}
		}
	}

	// 处理项目类型
	if projectType != "" {
		query["project_types"] = projectType
	}

	// 处理贡献类型
	if contributionType != "" {
		query["contribution_types"] = contributionType
	}

	return query
}

// analyzeDeveloperTrends 分析开发者趋势
func analyzeDeveloperTrends(id, period, metrics string) map[string]interface{} {
	// 解析时间段
	var startTime time.Time
	switch period {
	case "1m":
		startTime = time.Now().AddDate(0, -1, 0)
	case "6m":
		startTime = time.Now().AddDate(0, -6, 0)
	case "1y":
		startTime = time.Now().AddDate(-1, 0, 0)
	default:
		startTime = time.Now().AddDate(0, -1, 0)
	}

	// 解析指标
	metricsList := strings.Split(metrics, ",")
	trends := make(map[string]interface{})

	// 获取开发者数据
	developer, err := models.FindByID(id)
	if err != nil {
		return trends
	}

	// 分析各项指标的趋势
	for _, metric := range metricsList {
		switch metric {
		case "talent_rank":
			trends["talent_rank"] = analyzeTalentRankTrend(developer, startTime)
		case "contributions":
			trends["contributions"] = analyzeContributionsTrend(developer, startTime)
		case "activity":
			trends["activity"] = analyzeActivityTrend(developer, startTime)
		}
	}

	return trends
}

// 分析TalentRank趋势
func analyzeTalentRankTrend(developer *models.Developer, startTime time.Time) []float64 {
	// TODO: 实现TalentRank趋势分析
	return []float64{}
}

// 分析贡献趋势
func analyzeContributionsTrend(developer *models.Developer, startTime time.Time) []int {
	// TODO: 实现贡献趋势分析
	return []int{}
}

// 分析活跃度趋势
func analyzeActivityTrend(developer *models.Developer, startTime time.Time) []float64 {
	// TODO: 实现活跃度趋势分析
	return []float64{}
}
