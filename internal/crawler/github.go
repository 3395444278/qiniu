package crawler

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"os"
	"qinniu/internal/models"
	"qinniu/internal/pkg/cache"
	"qinniu/internal/pkg/queue"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"qinniu/internal/pkg/ai"

	"github.com/google/go-github/v45/github"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
)

type GitHubCrawler struct {
	client   *github.Client
	ctx      context.Context
	aiClient *ai.Client
}

func NewGitHubCrawler() *GitHubCrawler {
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("未发现 GITHUB_TOKEN 环境变量")
	}

	// 添加调试信息
	log.Printf("Using GitHub token: %s...", token[:10]) // 只打印token的前10个字符

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 验证token
	_, _, err := client.Users.Get(ctx, "") // 获取当前用户信息来验证token
	if err != nil {
		log.Fatalf("GitHub token 验证失败: %v", err)
	}

	log.Println("GitHub token 验证成功")

	// 初始化 AI 客户端
	aiApiKey := os.Getenv("AI_API_KEY")
	if aiApiKey == "" {
		log.Printf("Warning: DEEPSEEK_API_KEY not found, AI prediction will be disabled")
		return &GitHubCrawler{
			client: client,
			ctx:    ctx,
		}
	}

	aiClient := ai.NewClient(aiApiKey)
	log.Printf("AI client initialized successfully")

	return &GitHubCrawler{
		client:   client,
		ctx:      ctx,
		aiClient: aiClient,
	}
}

// CleanUsername 清理和验证 GitHub 用户名
func CleanUsername(rawUsername string) (string, error) {
	// 移除特殊字符和额外的描述
	username := strings.Split(rawUsername, " ")[0] // 取第一部分
	username = strings.Split(username, "·")[0]     // 移除 · 后面的内容
	username = strings.TrimSpace(username)         // 移除空格

	// GitHub 用户名规则验证
	// 只允许字母、数字和单个连字符（不能以连字符开始或结束）
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`)

	if !validUsername.MatchString(username) {
		return "", fmt.Errorf("invalid GitHub username format: %s", username)
	}

	// 检查连续的连字符
	if strings.Contains(username, "--") {
		return "", fmt.Errorf("invalid GitHub username format (contains consecutive hyphens): %s", username)
	}

	// 检查长度
	if len(username) > 39 {
		return "", fmt.Errorf("username too long (max 39 characters): %s", username)
	}

	return username, nil
}

// GetUserData 获取用基本信息
func (gc *GitHubCrawler) GetUserData(username string) (*models.Developer, error) {
	// 1. 首先检查数据库中是否已存在该用户
	existingDev, err := models.FindByUsername(username)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 如果存在且不需要更新，直返回
	if existingDev != nil && !existingDev.ShouldUpdate() {
		return existingDev, nil
	}

	// 如果Redis客户端可用，尝试从缓存获取
	if cache.RedisClient != nil {
		if cached, err := cache.GetCachedDeveloper(username); err == nil && cached != nil {
			if developer, ok := cached.(*models.Developer); ok {
				return developer, nil
			}
		}
	}

	// 加超时控制
	ctx, cancel := context.WithTimeout(gc.ctx, 30*time.Second)
	defer cancel()

	// 并发获取用户信息和仓库信息
	var user *github.User
	var repos []*github.Repository
	var userErr, repoErr error

	wg := sync.WaitGroup{}
	wg.Add(2)

	// 获取用户信息
	go func() {
		defer wg.Done()
		user, _, userErr = gc.client.Users.Get(ctx, username)
	}()

	// 获取仓库信息
	go func() {
		defer wg.Done()
		repos, repoErr = gc.GetUserRepositories(username)
	}()

	wg.Wait()

	if userErr != nil {
		return nil, userErr
	}
	if repoErr != nil {
		return nil, repoErr
	}

	// 获取用户头像 URL - 只在这里获取一次
	avatarURL := user.GetAvatarURL()
	if avatarURL == "" && user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	// 快速处理基本信息
	name := getPtrValue(user.Name)
	if name == "" {
		name = getPtrValue(user.Login)
	}

	// 删除这里的重复获取
	// avatarURL = user.GetAvatarURL()  // 删除这行
	if avatarURL == "" {
		log.Printf("Warning - No avatar URL found for user: %s", username)
	} else {
		log.Printf("Debug - Got avatar URL for user %s: %s", username, avatarURL)
	}

	// 获取仓库的语言信息
	var skills []string
	skillMap := make(map[string]struct{})

	for _, repo := range repos {
		// 获取主语言
		if repo.Language != nil && *repo.Language != "" {
			skillMap[*repo.Language] = struct{}{}
		}

		// 获取所有使用的语言
		languages, _, err := gc.client.Repositories.ListLanguages(gc.ctx, username, getPtrValue(repo.Name))
		if err != nil {
			continue
		}

		for lang := range languages {
			skillMap[lang] = struct{}{}
		}
	}

	// 转换为切片
	for skill := range skillMap {
		skills = append(skills, skill)
	}

	sort.Strings(skills)

	// 计算总 star 数和 fork 数
	var totalStars int
	var totalForks int
	var contributions int
	for _, repo := range repos {
		// 跳过 fork 的仓库
		if repo.GetFork() {
			log.Printf("跳过 fork 的仓库: %s", repo.GetName())
			continue
		}

		// 添加详细的日志
		log.Printf("处理原创仓库: %s", repo.GetName())
		log.Printf("Stars: %d, Forks: %d, Size: %d, Fork: %v",
			repo.GetStargazersCount(),
			repo.GetForksCount(),
			repo.GetSize(),
			repo.GetFork())

		totalStars += repo.GetStargazersCount()
		totalForks += repo.GetForksCount()
		contributions += repo.GetSize()
	}

	// 添加汇总日志
	log.Printf("最终统计结果 - 总 Stars: %d, 总 Forks: %d, 总贡献: %d",
		totalStars, totalForks, contributions)

	// 创建新的开发者记录
	developer := &models.Developer{
		Username:     getPtrValue(user.Login),
		Name:         name,
		Email:        getPtrValue(user.Email),
		Location:     getPtrValue(user.Location),
		Avatar:       avatarURL,
		ProfileURL:   user.GetHTMLURL(),
		UpdatedAt:    time.Now(),
		LastUpdated:  time.Now(),
		Skills:       skills,
		Repositories: repoNames(repos),
		StarCount:    totalStars,
		CommitCount:  contributions,
		ForkCount:    totalForks,
	}
	// 添加调试日志，确认 developer 对象中的 Avatar 字段
	log.Printf("Debug - Developer object created with Avatar URL: %s", developer.Avatar)

	// 假设有函数计算项目重要性和贡献度
	projectImportance := calculateProjectImportance(repos)
	contributionLevel := gc.calculateContributionLevel(username, repos)

	// 创建 DeveloperMetrics 对象
	developerMetrics := &models.DeveloperMetrics{}

	// 设置贡献指标
	developerMetrics.Contributions.CommitCount = contributions
	developerMetrics.Contributions.Quality = 0.8 // 默认质量分数

	// 设置项目指标
	developerMetrics.Projects.StarCount = totalStars
	developerMetrics.Projects.ForkCount = totalForks
	developerMetrics.Projects.TotalCount = len(repos)
	developerMetrics.Projects.Quality = projectImportance

	// 设置影响力指标
	developerMetrics.Influence.Followers = getPtrValue(user.Followers)
	developerMetrics.Influence.Recognition = contributionLevel

	// 设置活跃度指
	developerMetrics.Activity.LastActive = time.Now()
	developerMetrics.Activity.Frequency = calculateActivityFrequency(contributions)
	developerMetrics.Activity.Consistency = 0.8 // 默认持续性分数
	developerMetrics.Activity.Growth = calculateGrowthTrend(contributions)

	// 设置专业度指标（可以从其地方获取）
	developerMetrics.Expertise.Languages = skills
	developerMetrics.Expertise.Depth = 0.8 // 可以根据实际情况计算

	// 调用 calculateTalentRank
	developer.TalentRank = gc.calculateTalentRank(developerMetrics)

	// 处理 Nation 信息
	var nationConfidence float64
	nation := extractNation(developer.Location)
	if nation == "" {
		log.Printf("尝试使用 AI 预测国家...")
		aiNation, aiConfidence := gc.predictNationWithAI(user, repos)
		if aiNation != "" {
			nation = aiNation
			nationConfidence = aiConfidence
			log.Printf("AI 预测成功 - 国家: %s, 置信度: %.2f", nation, nationConfidence)
		} else {
			log.Printf("AI 预测失败")
		}
	}

	if nation == "" {
		quickPred := QuickPredictNation(username, user, repos)
		if quickPred != nil && quickPred.Confidence >= 40 {
			nation = quickPred.Nation
			nationConfidence = quickPred.Confidence
		}
	}

	// 如果所有预测方法仍然未能确定国家，调用 AI 客户端
	if nation == "" {
		aiNation, aiConfidence := gc.predictNationWithAI(user, repos)
		if aiNation != "" {
			nation = aiNation
			nationConfidence = aiConfidence
		}
	}

	developer.Nation = nation
	developer.NationConfidence = nationConfidence

	// 计算置信（使用新的方法或移除）
	developer.Confidence = calculateConfidence(
		contributions,
		totalStars,
		getPtrValue(user.Followers),
		developer.Location != "",
	)

	// 添加数据验证
	developer.DataValidation = models.ValidationResult{
		IsValid:       true,
		Confidence:    developer.Confidence,
		LastValidated: time.Now(),
		// 不要初始化 Issues 字段，让它保持为 nil
	}

	// 设置更新频率（根据活跃度调整）
	if developer.CommitCount > 1000 {
		developer.UpdateFrequency = 24 * time.Hour // 活跃用户每天更新
	} else {
		developer.UpdateFrequency = 7 * 24 * time.Hour // 不活跃用户每周更新
	}

	// 保存到数据库
	if existingDev != nil {
		developer.ID = existingDev.ID
		developer.CreatedAt = existingDev.CreatedAt
		if err := developer.Update(); err != nil {
			return nil, fmt.Errorf("更新用户失败: %v", err)
		}
	} else {
		if err := developer.Create(); err != nil {
			return nil, fmt.Errorf("创建用户失败: %v", err)
		}
	}

	// 创建并发送评估任务
	evaluationTask := &queue.EvaluationTask{
		Username:     developer.Username,
		ProfileURL:   developer.ProfileURL,
		BlogURL:      user.GetBlog(),
		Description:  user.GetBio(),
		Repositories: developer.Repositories,
		CreatedAt:    time.Now(),
	}

	// 发送评估任务到队列
	queueClient := queue.NewQueue()
	if err := queueClient.Publish(evaluationTask); err != nil {
		log.Printf("Warning: Failed to publish evaluation task for %s: %v", username, err)
	} else {
		log.Printf("Successfully published evaluation task for %s", username)
	}

	// 验证保存后的数据
	savedDev, err := models.FindByUsername(developer.Username)
	if err == nil && savedDev != nil {
		log.Printf("Debug - Verified saved data for %s, tech_evaluation: %+v",
			savedDev.Username, savedDev.TechEvaluation)
	}

	// 如Redis客户端可用，保存到缓存
	if cache.RedisClient != nil {
		if err := cache.CacheDeveloper(developer); err != nil {
			log.Printf("Warning: Failed to cache developer data: %v", err)
		}
	}

	// 在保存到数据库之前再次确认
	if developer.Avatar == "" {
		log.Printf("Warning - Developer %s has no avatar URL before saving to database", developer.Username)
	}

	return developer, nil
}

// predictNation 通过其他信息预测用户的国家
func (gc *GitHubCrawler) predictNation(user *github.User, repos []*github.Repository) string {
	// 1. 分析提交时间分布
	commitTimes := gc.analyzeCommitTimes(repos)
	if timezone := predictTimezone(commitTimes); timezone != "" {
		return timezoneToCountry(timezone)
	}

	// 2. 分析码注释语言
	if lang := gc.analyzeCodeComments(repos); lang != "" {
		return gc.languageToCountry(lang)
	}

	// 3. 分析用户名特征
	if country := gc.analyzeUsername(getPtrValue(user.Login)); country != "" {
		return country
	}

	// 4. 分析仓库名称和描述
	if country := gc.analyzeRepoInfo(repos); country != "" {
		return country
	}

	return ""
}

// analyzeCommitTimes 分析提交时间分布
func (gc *GitHubCrawler) analyzeCommitTimes(repos []*github.Repository) []time.Time {
	var commitTimes []time.Time

	for _, repo := range repos {
		commits, _, err := gc.client.Repositories.ListCommits(gc.ctx, getPtrValue(repo.Owner.Login), getPtrValue(repo.Name), nil)
		if err != nil {
			continue
		}

		for _, commit := range commits {
			if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Author.Date != nil {
				commitTimes = append(commitTimes, *commit.Commit.Author.Date)
			}
		}
	}

	return commitTimes
}

// predictTimezone 据时间分布预测时区
func predictTimezone(commitTimes []time.Time) string {
	if len(commitTimes) == 0 {
		return ""
	}

	// 统计提交时间的小时分布
	hourCount := make(map[int]int)
	for _, t := range commitTimes {
		hour := t.Hour()
		hourCount[hour]++
	}

	// 找出最活跃的时间段（6小时）
	maxCount := 0
	peakStartHour := 0

	for hour := 0; hour < 24; hour++ {
		count := 0
		for i := 0; i < 6; i++ {
			h := (hour + i) % 24
			count += hourCount[h]
		}
		if count > maxCount {
			maxCount = count
			peakStartHour = hour
		}
	}

	// 根据活跃时间段推测时区
	switch {
	case peakStartHour >= 8 && peakStartHour <= 10:
		return "Asia/Shanghai"
	case peakStartHour >= 6 && peakStartHour <= 8:
		return "America/New_York"
		// 可以添加更多时区判断
	}

	return ""
}

// timezoneToCountry 将时区映到可能的国家
func timezoneToCountry(timezone string) string {
	if country, ok := timezoneCountryMap[timezone]; ok {
		return country
	}
	return ""
}

// analyzeCodeComments 分析代码注释中的语言
func (gc *GitHubCrawler) analyzeCodeComments(repos []*github.Repository) string {
	commentPatterns := map[string]*regexp.Regexp{
		"chinese":  regexp.MustCompile(`[\p{Han}]`),
		"english":  regexp.MustCompile(`^[a-zA-Z\s.,!?]+$`),
		"japanese": regexp.MustCompile(`[\p{Hiragana}\p{Katakana}]`),
		"korean":   regexp.MustCompile(`[\p{Hangul}]`),
	}

	langCount := make(map[string]int)

	for _, repo := range repos {
		// 获取仓库内容
		_, contents, _, err := gc.client.Repositories.GetContents(gc.ctx, getPtrValue(repo.Owner.Login), getPtrValue(repo.Name), "", nil)
		if err != nil {
			continue
		}

		for _, content := range contents {
			if content.Type != nil && *content.Type == "file" {
				// 分析文件内容中的注释
				fileContent, err := getFileContent(content)
				if err != nil {
					continue
				}

				for lang, pattern := range commentPatterns {
					if matches := pattern.FindAllString(fileContent, -1); len(matches) > 0 {
						langCount[lang] += len(matches)
					}
				}
			}
		}
	}

	// 返回最常见的语言
	maxCount := 0
	dominantLang := ""
	for lang, count := range langCount {
		if count > maxCount {
			maxCount = count
			dominantLang = lang
		}
	}

	return dominantLang
}

// languageToCountry 将语言映射到可能的家
func (gc *GitHubCrawler) languageToCountry(language string) string {
	if country, ok := languageCountryMap[language]; ok {
		return country
	}
	return ""
}

// analyzeUsername 分析用户名特征
func (gc *GitHubCrawler) analyzeUsername(username string) string {
	// 汉字匹配
	chinesePattern := regexp.MustCompile(`[\p{Han}]`)
	if chinesePattern.MatchString(username) {
		return "CN"
	}

	// 中文拼音特征
	pinyinPattern := regexp.MustCompile(`^[a-z]+(zhang|wang|li|zhao|liu|chen|yang|huang|zhou|wu|sun|zhu|ma|lin|xu|luo|guo|he|gao|zheng)$`)
	if pinyinPattern.MatchString(strings.ToLower(username)) {
		return "CN"
	}

	// 日文假名和汉字特征
	japanesePattern := regexp.MustCompile(`([\p{Hiragana}\p{Katakana}]|kawa|naka|taka|hiro|yama|sato|suzuki|tana|ito|watanabe|yamamoto|nakamura|kobayashi|saito)`)
	if japanesePattern.MatchString(strings.ToLower(username)) {
		return "JP"
	}

	// 韩文和韩名称特征
	koreanPattern := regexp.MustCompile(`([\p{Hangul}]|kim|lee|park|choi|jung|kang|cho|yoon|jang|lim|han|oh|seo|shin)`)
	if koreanPattern.MatchString(strings.ToLower(username)) {
		return "KR"
	}

	return ""
}

// analyzeRepoInfo 分析仓库信息
func (gc *GitHubCrawler) analyzeRepoInfo(repos []*github.Repository) string {
	langCount := make(map[string]int)

	for _, repo := range repos {
		// 分析仓库描述
		if repo.Description != nil {
			desc := strings.ToLower(*repo.Description)

			// 检测中文
			if strings.Contains(desc, "中国") || strings.Contains(desc, "中文") {
				langCount["chinese"]++
			}

			// 检测日语
			if strings.Contains(desc, "日本") || strings.Contains(desc, "にほん") {
				langCount["japanese"]++
			}

			// 检测韩语
			if strings.Contains(desc, "한국") || strings.Contains(desc, "조선") {
				langCount["korean"]++
			}
		}

		// 分析仓库主要语言
		if repo.Language != nil {
			switch *repo.Language {
			case "Chinese":
				langCount["chinese"]++
			case "Japanese":
				langCount["japanese"]++
			case "Korean":
				langCount["korean"]++
			}
		}
	}

	// 返回最见的语言对应的国家
	maxCount := 0
	dominantLang := ""
	for lang, count := range langCount {
		if count > maxCount {
			maxCount = count
			dominantLang = lang
		}
	}

	return languageCountryMap[dominantLang]
}

// 修改 GetUserRepositories 方法，使用并发处理
func (gc *GitHubCrawler) GetUserRepositories(username string) ([]*github.Repository, error) {
	ctx, cancel := context.WithTimeout(gc.ctx, 20*time.Second)
	defer cancel()

	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Type:        "all",
		Sort:        "updated",
		Direction:   "desc",
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := gc.client.Repositories.List(ctx, username, opts)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// 使用工作池并发获取详细信息
	results := make([]*github.Repository, len(allRepos))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数

	for i, repo := range allRepos {
		wg.Add(1)
		go func(i int, repo *github.Repository) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			fullRepo, _, err := gc.client.Repositories.Get(ctx, username, repo.GetName())
			if err != nil {
				log.Printf("Warning: Failed to get full repository info for %s: %v", repo.GetName(), err)
				results[i] = repo // 如果获取失败，使用原始数据
				return
			}
			results[i] = fullRepo
		}(i, repo)
	}

	wg.Wait()
	return results, nil
}

// calculateTalentRank 计算开发者的 TalentRank
func (gc *GitHubCrawler) calculateTalentRank(metrics *models.DeveloperMetrics) float64 {
	// 调用 models.CalculateTalentRank 进行计算
	return models.CalculateTalentRank(metrics)
}

// 计算活动频率
func calculateActivityFrequency(contributions int) float64 {
	// 简单的活动频率计算
	return math.Min(float64(contributions)/1000.0, 1.0)
}

// 计算增长趋势
func calculateGrowthTrend(contributions int) float64 {
	// 简单的增长趋势计算
	return math.Min(float64(contributions)/500.0, 1.0)
}

// 辅助函数
func getPtrValue[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// extractSkills 提取技能
func (gc *GitHubCrawler) extractSkills(repos []*github.Repository) []string {
	skillMap := make(map[string]struct{})

	for _, repo := range repos {
		if repo.Language != nil && *repo.Language != "" {
			skillMap[strings.ToLower(*repo.Language)] = struct{}{}
		}
		languages, _, err := gc.client.Repositories.ListLanguages(gc.ctx, getPtrValue(repo.Owner.Login), getPtrValue(repo.Name))
		if err != nil {
			continue
		}
		for lang := range languages {
			skillMap[strings.ToLower(lang)] = struct{}{}
		}
	}

	skills := make([]string, 0, len(skillMap))
	for skill := range skillMap {
		if skill != "" {
			skills = append(skills, skill)
		}
	}
	sort.Strings(skills)
	return skills
}

func repoNames(repos []*github.Repository) []string {
	// 使用 map 去重
	nameMap := make(map[string]struct{})
	for _, repo := range repos {
		nameMap[getPtrValue(repo.Name)] = struct{}{}
	}

	// 转换回切片
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}

	// 排序以保持稳定顺序
	sort.Strings(names)
	return names
}

// 使用 models 包中的函数
func extractNation(location string) string {
	return models.ExtractNation(location)
}

// 新增：获取仓库总 star 数
func getTotalStars(repos []*github.Repository) int {
	var total int
	for _, repo := range repos {
		total += getPtrValue(repo.StargazersCount)
	}
	return total
}

// 新增：从URL中提取位置信息
func extractLocationFromURL(url string) string {
	// TODO: 实现从个人网站提取位置信息的逻辑
	return ""
}

// 新增：从仓库中提取位置信息
func (gc *GitHubCrawler) extractLocationFromRepos(repos []*github.Repository) string {
	for _, repo := range repos {
		// 检查README内容
		if readme := gc.getRepoReadme(repo); readme != "" {
			// 查找位置相关的关键词
			locationKeywords := []string{
				"Location:", "Based in", "Living in", "From", "所在地:", "位置:", "常驻:", "来自:",
				"作地点:", "工作城市:", "所在城市:", "所在省份:", "所在国家:",
			}

			// 将README内容按分割
			lines := strings.Split(readme, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// 跳过空行和明显不是位置信息的行
				if line == "" || strings.Contains(line, "import") || strings.Contains(line, "export") {
					continue
				}

				for _, keyword := range locationKeywords {
					if strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
						// 提取关键词后的
						text := line[strings.Index(line, keyword)+len(keyword):]
						// 清文本
						text = strings.TrimSpace(text)
						text = strings.Trim(text, ":")
						text = strings.TrimSpace(text)

						// 验证提取的文本是否看起来像位置信息
						if isValidLocation(text) {
							return text
						}
					}
				}
			}
		}

		// 检查仓库描述
		if desc := getPtrValue(repo.Description); desc != "" {
			// 查找位置相关的关键词
			for _, keyword := range []string{"based in", "located in", "from"} {
				if idx := strings.Index(strings.ToLower(desc), keyword); idx != -1 {
					// 提取关键词后的文本
					text := desc[idx+len(keyword):]
					if end := strings.IndexAny(text, ".,;"); end != -1 {
						text = strings.TrimSpace(text[:end])
						if isValidLocation(text) {
							return text
						}
					}
					text = strings.TrimSpace(text)
					if isValidLocation(text) {
						return text
					}
				}
			}
		}
	}
	return ""
}

// isValidLocation 验证文本是看起来像位置信息
func isValidLocation(text string) bool {
	// 如果文本太长，可能不是位置信息
	if len(text) > 100 {
		return false
	}

	// 检查是否包含明显不是位置的关键词
	invalidKeywords := []string{
		"book", "ebook", "tutorial", "guide", "manual", "documentation",
		"scratch", "project", "repository", "code", "software", "app",
		"import", "export", "function", "class", "const", "var", "let",
		"return", "component", "from", "require", "module",
	}

	textLower := strings.ToLower(text)
	for _, keyword := range invalidKeywords {
		if strings.Contains(textLower, keyword) {
			return false
		}
	}

	// 检查是否包含常见的位置相关词汇
	locationKeywords := []string{
		"city", "province", "state", "country", "region", "district",
		"城市", "省份", "国家", "地区", "区域",
	}

	for _, keyword := range locationKeywords {
		if strings.Contains(textLower, keyword) {
			return true
		}
	}

	// 检查是否匹配任何已知的城市国家名称
	if extractNation(text) != "" {
		return true
	}

	// 如果文本很短且不包含特殊字符，可能是有效的位置
	if len(text) < 30 && !strings.ContainsAny(text, "{}[]()<>=/\\") {
		// 额外检查文本应该主要包含字、空格和逗号
		validChars := 0
		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsSpace(r) || r == ',' || r == '.' {
				validChars++
			}
		}
		return float64(validChars)/float64(len(text)) > 0.8
	}

	return false
}

// extractLocationFromBio 从Bio中提取位置信息
func extractLocationFromBio(bio string) string {
	// 位置相关的关键词
	locationKeywords := []string{
		"Location:", "Based in", "Living in", "From", "所在地:", "位置:", "常驻:", "来自:",
		"工作地点:", "工作城市:", "所在城市:", "所在省份:", "所在国家:",
	}

	// 将bio按行分割
	lines := strings.Split(bio, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, keyword := range locationKeywords {
			if strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
				// 提取关键词后的文本
				text := line[strings.Index(line, keyword)+len(keyword):]
				// 清理文本
				text = strings.TrimSpace(text)
				text = strings.Trim(text, ":")
				text = strings.TrimSpace(text)
				if text != "" && isValidLocation(text) {
					return text
				}
			}
		}
	}

	// 如果没有找到关键词，尝试直接匹配城市或国家名
	words := strings.Fields(bio)
	for _, word := range words {
		// 检查是否是城市名或国家名
		if isValidLocation(word) && extractNation(word) != "" {
			return word
		}
	}

	return ""
}

// getRepoReadme 获取仓库的README内容
func (gc *GitHubCrawler) getRepoReadme(repo *github.Repository) string {
	// 获取仓库的所有者和名称
	owner := getPtrValue(repo.Owner.Login)
	repoName := getPtrValue(repo.Name)

	// 尝试获取 README 内容
	readme, _, err := gc.client.Repositories.GetReadme(gc.ctx, owner, repoName, nil)
	if err != nil {
		return ""
	}

	// 解码 README 内容
	content, err := readme.GetContent()
	if err != nil {
		return ""
	}

	return content
}

// 添加一个函数用于缓存预热
func (gc *GitHubCrawler) WarmupCache(usernames []string) error {
	return cache.WarmupCache(usernames, func(username string) (interface{}, error) {
		return gc.GetUserData(username)
	})
}

// 新增：速预测函数
func QuickPredictNation(username string, user *github.User, repos []*github.Repository) *models.PredictionResult {
	points := make(map[string]float64)
	factors := make([]string, 0)

	// 1. 检查邮箱域名
	if email := getPtrValue(user.Email); email != "" {
		switch {
		case strings.HasSuffix(email, ".cn"):
			points["CN"] += 2.0
			factors = append(factors, "邮箱域名(.cn)")
		case strings.HasSuffix(email, ".jp"):
			points["JP"] += 2.0
			factors = append(factors, "邮箱名(.jp)")
		case strings.Contains(email, "foxmail.com"):
			points["CN"] += 1.5
			factors = append(factors, "邮箱服务(foxmail)")
		case strings.Contains(email, "qq.com"):
			points["CN"] += 1.5
			factors = append(factors, "箱服务(qq)")
		case strings.Contains(email, "163.com"):
			points["CN"] += 1.5
			factors = append(factors, "邮箱服务(163)")
		}
	}

	// 2. 检查用户名和显示名称
	usernameLower := strings.ToLower(username)
	nameLower := strings.ToLower(getPtrValue(user.Name))

	// 检查中文字符
	if containsChinese(nameLower) {
		points["CN"] += 2.0
		factors = append(factors, "中文名称")
	}

	// 检查特定关键词
	for keyword, country := range map[string]string{
		"china": "CN",
		"cn":    "CN",
		"jp":    "JP",
		"kr":    "KR",
		"sg":    "SG",
	} {
		if strings.Contains(usernameLower, keyword) || strings.Contains(nameLower, keyword) {
			points[country] += 1.5
			factors = append(factors, "用户名/显示名称关词")
		}
	}

	// 3. 查仓库描述README
	chineseCount := 0
	for _, repo := range repos {
		desc := strings.ToLower(getPtrValue(repo.Description))

		// 检查中文内容
		if containsChinese(desc) {
			chineseCount++
		}

		// 检查关键词
		if strings.Contains(desc, "中国") || strings.Contains(desc, "china") {
			points["CN"] += 1.0
			factors = append(factors, "仓库描述包含地理关键词")
		}
	}

	// 如果超过30%的仓库描述包含中文
	if float64(chineseCount)/float64(len(repos)) > 0.3 {
		points["CN"] += 2.0
		factors = append(factors, "大量中文仓���描述")
	}

	// 4. 检查公司信息
	if company := getPtrValue(user.Company); company != "" {
		companyLower := strings.ToLower(company)
		if containsChinese(company) ||
			strings.Contains(companyLower, "china") ||
			strings.Contains(companyLower, "beijing") ||
			strings.Contains(companyLower, "shanghai") {
			points["CN"] += 1.5
			factors = append(factors, "公司信息")
		}
	}

	// 找出得分最高的国家
	var maxScore float64
	var predictedNation string
	for country, score := range points {
		if score > maxScore {
			maxScore = score
			predictedNation = country
		}
	}

	// 计算置信度
	confidence := calculateQuickConfidence(maxScore, len(factors))

	return &models.PredictionResult{
		Nation:     predictedNation,
		Confidence: confidence,
		Factors:    factors,
	}
}

// containsChinese 检测字符串是否包含中文字符
func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// containsJapanese 检测字符串是否包含日字
func containsJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana) {
			return true
		}
	}
	return false
}

// calculateQuickConfidence 计算快速预测的置信度
func calculateQuickConfidence(maxScore float64, factorCount int) float64 {
	// 基础置信度
	baseConfidence := 0.3

	// 根据得分增加置信度
	scoreConfidence := math.Min(maxScore/5.0, 0.4)

	// 根据影响因素数量增加置信度
	factorConfidence := math.Min(float64(factorCount)/3.0, 0.3)

	totalConfidence := baseConfidence + scoreConfidence + factorConfidence
	return math.Min(totalConfidence, 1.0) * 100
}

// analyzeBasicInfo 基于基本信息的快速分析
func analyzeBasicInfo(username string, repos []*github.Repository) string {
	// 1. 检查用户名
	usernameLower := strings.ToLower(username)
	switch {
	case strings.Contains(usernameLower, "cn") || strings.Contains(usernameLower, "china"):
		return "CN"
	case strings.Contains(usernameLower, "jp") || strings.Contains(usernameLower, "japan"):
		return "JP"
	case strings.Contains(usernameLower, "kr") || strings.Contains(usernameLower, "korea"):
		return "KR"
	}

	// 2. 快速检查仓库和描述
	for _, repo := range repos {
		desc := strings.ToLower(getPtrValue(repo.Description))
		switch {
		case strings.Contains(desc, "中") || strings.Contains(desc, "china"):
			return "CN"
		case strings.Contains(desc, "日本") || strings.Contains(desc, "japan"):
			return "JP"
		case strings.Contains(desc, "한국") || strings.Contains(desc, "korea"):
			return "KR"
		}
	}

	return ""
}

// 计算项目重要性，基于仓库的 star 数、fork 数等
func calculateProjectImportance(repos []*github.Repository) float64 {
	if len(repos) == 0 {
		return 0.0
	}

	var totalScore float64
	var validRepos int

	for _, repo := range repos {
		// 跳过 fork 的仓库
		if repo.GetFork() {
			continue
		}

		stars := repo.GetStargazersCount()
		forks := repo.GetForksCount()
		size := repo.GetSize()

		// 计算单个仓库的得分
		repoScore := float64(0)

		// 1. Star 权重计算（使用对数计算，避免单个高 star 项目过度影响）
		if stars > 0 {
			repoScore += math.Log10(float64(stars)) * 2
		}

		// 2. Fork 权重计算
		if forks > 0 {
			repoScore += math.Log10(float64(forks)) * 1.5
		}

		// 3. 项目大小权重（考虑代码量，但影响较小）
		if size > 0 {
			repoScore += math.Log10(float64(size)) * 0.3
		}

		// 4. 活跃度权重
		if !repo.GetArchived() && time.Since(repo.GetUpdatedAt().Time) < 180*24*time.Hour { // 半年内有更新
			repoScore *= 1.2
		}

		totalScore += repoScore
		validRepos++
	}

	if validRepos == 0 {
		return 0.0
	}

	// 计算平均分并归一化到 0-1 范围
	avgScore := totalScore / float64(validRepos)
	normalizedScore := math.Min(avgScore/10.0, 1.0)

	return normalizedScore
}

// 计算开发者的贡献度，基 commit 数或其他贡献指标
func (gc *GitHubCrawler) calculateContributionLevel(username string, repos []*github.Repository) float64 {
	var totalScore float64
	var validRepos int

	for _, repo := range repos {
		// 跳过 fork 的仓库
		if repo.GetFork() {
			continue
		}

		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()

		// 1. 获取用户在该仓库的提交数
		userCommits := gc.getUserCommitsInRepo(username, owner, repoName)
		if userCommits == 0 {
			continue
		}

		// 2. 计算提交得分（使用对数计算）
		commitScore := math.Log10(float64(userCommits)) * 2

		// 3. 如果是仓库所有者，获得额外加分
		if owner == username {
			commitScore *= 1.5
		}

		// 4. 根据仓库质量调整得分
		repoQuality := float64(repo.GetStargazersCount()) / 100.0
		if repoQuality > 1.0 {
			repoQuality = 1.0 + math.Log10(repoQuality) // 对高质量项目进行对数加成
		}
		commitScore *= (1.0 + repoQuality)

		totalScore += commitScore
		validRepos++
	}

	if validRepos == 0 {
		return 0.0
	}

	// 归一化到 0-1 范围
	return math.Min(totalScore/float64(validRepos*10), 1.0)
}

// 实现获取仓库的总提交数
func (gc *GitHubCrawler) getRepoTotalCommits(owner, repoName string) int {
	commits, _, err := gc.client.Repositories.ListCommits(gc.ctx, owner, repoName, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil || len(commits) == 0 {
		return 0
	}

	// 获取最后一个提交的 SHA
	lastCommitSHA := commits[0].GetSHA()

	// 获取提交次数
	comparison, _, err := gc.client.Repositories.CompareCommits(gc.ctx, owner, repoName, "HEAD~0", lastCommitSHA, nil)
	if err != nil {
		return 0
	}

	totalCommits := comparison.GetTotalCommits() + 1 // 加上初始提交
	return totalCommits
}

// 实现获取用户在仓库中的提交数
func (gc *GitHubCrawler) getUserCommitsInRepo(username, owner, repoName string) int {
	opts := &github.CommitsListOptions{
		Author: username,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	totalCommits := 0
	for {
		commits, resp, err := gc.client.Repositories.ListCommits(gc.ctx, owner, repoName, opts)
		if err != nil {
			break
		}

		totalCommits += len(commits)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return totalCommits
}

// 新增：计置信度的函数
func calculateConfidence(contributions, stars, followers int, hasLocation bool) float64 {
	// 基础置信度
	baseConfidence := 0.5

	// 根据贡献调整置信度
	contributionConfidence := math.Min(float64(contributions)/1000.0, 0.3)

	// 据 star 数调整置信度
	starConfidence := math.Min(float64(stars)/10000.0, 0.1)

	// 根据关注者数调整置信度
	followerConfidence := math.Min(float64(followers)/1000.0, 0.1)

	// 位置信息提供额外置信度
	locationConfidence := 0.0
	if hasLocation {
		locationConfidence = 0.1
	}

	// 算总置信度
	totalConfidence := baseConfidence +
		contributionConfidence +
		starConfidence +
		followerConfidence +
		locationConfidence

	// 确保置信度在 0-100 之间
	return math.Min(totalConfidence*100, 100)
}

// 辅助函数：获取文件内容
func getFileContent(content *github.RepositoryContent) (string, error) {
	if content.Content != nil {
		decoded, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}
	return "", fmt.Errorf("no content available")
}

// timezoneCountryMap 时区到国家代码的映射
var timezoneCountryMap = map[string]string{
	// 亚洲
	"Asia/Shanghai":    "CN",
	"Asia/Hong_Kong":   "CN",
	"Asia/Chongqing":   "CN",
	"Asia/Harbin":      "CN",
	"Asia/Urumqi":      "CN",
	"Asia/Tokyo":       "JP",
	"Asia/Seoul":       "KR",
	"Asia/Singapore":   "SG",
	"Asia/Taipei":      "TW",
	"Asia/Bangkok":     "TH",
	"Asia/Jakarta":     "ID",
	"Asia/Manila":      "PH",
	"Asia/Kolkata":     "IN",
	"Asia/Ho_Chi_Minh": "VN",

	// 北美
	"America/New_York":    "US",
	"America/Los_Angeles": "US",
	"America/Chicago":     "US",
	"America/Denver":      "US",
	"America/Phoenix":     "US",
	"America/Toronto":     "CA",
	"America/Vancouver":   "CA",
	"America/Montreal":    "CA",

	// 欧洲
	"Europe/London":    "GB",
	"Europe/Paris":     "FR",
	"Europe/Berlin":    "DE",
	"Europe/Madrid":    "ES",
	"Europe/Rome":      "IT",
	"Europe/Amsterdam": "NL",
	"Europe/Stockholm": "SE",
	"Europe/Oslo":      "NO",
	"Europe/Moscow":    "RU",
}

// languageCountryMap 语言到国家代码的映射
var languageCountryMap = map[string]string{
	// 亚洲语言
	"chinese":    "CN",
	"mandarin":   "CN",
	"cantonese":  "CN",
	"japanese":   "JP",
	"korean":     "KR",
	"vietnamese": "VN",
	"thai":       "TH",
	"indonesian": "ID",
	"malay":      "MY",
	"hindi":      "IN",

	// 欧洲语言
	"german":     "DE",
	"french":     "FR",
	"italian":    "IT",
	"spanish":    "ES",
	"portuguese": "PT",
	"russian":    "RU",
	"dutch":      "NL",
	"swedish":    "SE",
	"norwegian":  "NO",
	"danish":     "DK",

	// 英语变体
	"english":    "US", // 默认美式英语
	"british":    "GB",
	"australian": "AU",
	"canadian":   "CA",
}

// predictNationWithAI 使用 AI 客户端预测国家
func (gc *GitHubCrawler) predictNationWithAI(user *github.User, repos []*github.Repository) (string, float64) {
	info := map[string]interface{}{
		"username":    getPtrValue(user.Login),
		"name":        getPtrValue(user.Name),
		"email":       getPtrValue(user.Email),
		"location":    getPtrValue(user.Location),
		"profile_url": user.GetHTMLURL(),
		"skills":      gc.extractSkills(repos),
		"languages":   gc.extractLanguages(repos),
		"repos":       repoNames(repos),
		"commits":     gc.getTotalCommits(repos),
		"stars":       getTotalStars(repos),
		"forks":       getTotalForks(repos),
		"last_active": getLastActiveTime(repos).Format("2006-01-02"),
	}

	evaluation, err := gc.aiClient.EvaluateDeveloper(gc.ctx, info)
	if err != nil {
		log.Printf("AI 评估失败: %v", err)
		return "", 0
	}

	if evaluation.Nation == "" {
		log.Printf("AI 未能预测国家")
		return "", 0
	}

	log.Printf("AI 预测结果 - 国家: %s, 置信度: %.2f", evaluation.Nation, evaluation.Confidence)
	return evaluation.Nation, evaluation.Confidence
}

// extractLanguages 提取所有使用的编程语言
func (gc *GitHubCrawler) extractLanguages(repos []*github.Repository) []string {
	languageSet := make(map[string]struct{})

	for _, repo := range repos {
		if repo.Language != nil && *repo.Language != "" {
			languageSet[strings.ToLower(*repo.Language)] = struct{}{}
		}
		languages, _, err := gc.client.Repositories.ListLanguages(gc.ctx, getPtrValue(repo.Owner.Login), getPtrValue(repo.Name))
		if err != nil {
			continue
		}
		for lang := range languages {
			languageSet[strings.ToLower(lang)] = struct{}{}
		}
	}

	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	sort.Strings(languages)
	return languages
}

// 提取仓库 URLs
func repoURLs(repos []*github.Repository) []string {
	urls := make([]string, 0, len(repos))
	for _, repo := range repos {
		urls = append(urls, repo.GetHTMLURL())
	}
	return urls
}

// getTotalCommits 计算所有仓库的提交总数
func (gc *GitHubCrawler) getTotalCommits(repos []*github.Repository) int {
	total := 0
	for _, repo := range repos {
		total += gc.getUserCommitsInRepo(
			getPtrValue(repo.Owner.Login),
			getPtrValue(repo.Owner.Login),
			getPtrValue(repo.Name),
		)
	}
	return total
}

// getTotalForks 计算所有仓库的 Fork 总数
func getTotalForks(repos []*github.Repository) int {
	total := 0
	for _, repo := range repos {
		total += repo.GetForksCount()
	}
	return total
}

// getRepoStars 计算所有仓库的 Stars 总数
func getRepoStars(repos []*github.Repository) int {
	total := 0
	for _, repo := range repos {
		total += repo.GetStargazersCount()
	}
	return total
}

// getLastActiveTime 获取用户最近的活跃时间
func getLastActiveTime(repos []*github.Repository) time.Time {
	var lastActive time.Time
	for _, repo := range repos {
		updatedAt := repo.GetUpdatedAt().Time
		if updatedAt.After(lastActive) {
			lastActive = updatedAt
		}
	}
	return lastActive
}

// getCreationTime 获取用户账户的创建时间
func getCreationTime(user *github.User) time.Time {
	return user.GetCreatedAt().Time
}

// getUpdateTime 获取用户账户的最后更新时间
func getUpdateTime(user *github.User) time.Time {
	return user.GetUpdatedAt().Time
}
