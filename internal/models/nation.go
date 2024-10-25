package models

import (
	"context"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// NationPredictor 国家预测器
type NationPredictor struct {
	client *github.Client
	ctx    context.Context
}

// PredictionResult 预测结果
type PredictionResult struct {
	Nation     string   // 预测的国家代码
	Confidence float64  // 预测的置信度
	Factors    []string // 影响预测的因素
}

// 预测方法
func PredictNation(username string, user *github.User, repos []*github.Repository) *PredictionResult {
	factors := make([]string, 0)
	points := make(map[string]float64)

	// 1. 分析活动时间分布
	if timeZone := analyzeActivityTime(username); timeZone != "" {
		country := timeZoneToCountry(timeZone)
		points[country] += 3.0
		factors = append(factors, "活动时间分布")
	}

	// 2. 分析用户的关注者和被关注者
	if followers, following, err := analyzeConnections(username); err == nil {
		for country, score := range followers {
			points[country] += score * 2.0
		}
		for country, score := range following {
			points[country] += score * 1.5
		}
		factors = append(factors, "社交网络分析")
	}

	// 3. 分析代码注释和提交信息的语言
	if langPoints := analyzeLanguageUsage(repos); len(langPoints) > 0 {
		for country, score := range langPoints {
			points[country] += score
		}
		factors = append(factors, "语言使用分析")
	}

	// 4. 分析仓库信息
	if repoPoints := analyzeRepositories(repos); len(repoPoints) > 0 {
		for country, score := range repoPoints {
			points[country] += score
		}
		factors = append(factors, "仓库信息分析")
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
	confidence := calculateConfidence(maxScore, len(factors))

	return &PredictionResult{
		Nation:     predictedNation,
		Confidence: confidence,
		Factors:    factors,
	}
}

// analyzeActivityTime 分析用户活动时间
func analyzeActivityTime(username string) string {
	// TODO: 实现活动时间分析
	// 1. 获取用户最近的活动记录
	// 2. 统计活动时间分布
	// 3. 根据时间分布推测时区
	return ""
}

// analyzeConnections 分析用户的社交网络
func analyzeConnections(username string) (map[string]float64, map[string]float64, error) {
	followers := make(map[string]float64)
	following := make(map[string]float64)

	// TODO: 实现社交网络分析
	// 1. 获取用户的关注者列表
	// 2. 获取用户关注的人的列表
	// 3. 分析这些用户的地理位置分布

	return followers, following, nil
}

// analyzeLanguageUsage 分析代码和注释中的语言使用
func analyzeLanguageUsage(repos []*github.Repository) map[string]float64 {
	points := make(map[string]float64)

	for _, repo := range repos {
		// 分析仓库描述
		desc := strings.ToLower(getPtrValue(repo.Description))

		// 检测中文使用
		if ContainsChinese(desc) {
			points["CN"] += 1.0
		}

		// 检测日语使用
		if ContainsJapanese(desc) {
			points["JP"] += 1.0
		}

		// 检测韩语使用
		if ContainsKorean(desc) {
			points["KR"] += 1.0
		}

		// TODO: 分析代码注释
		// 1. 克隆仓库
		// 2. 分析代码文件中的注释
		// 3. 检测注释中使用的语言
	}

	return points
}

// analyzeRepositories 分析仓库信息
func analyzeRepositories(repos []*github.Repository) map[string]float64 {
	points := make(map[string]float64)

	// 检查地理相关关键词
	keywords := map[string]map[string]struct{}{
		"CN": {
			"china": {}, "chinese": {}, "中国": {},
			"beijing": {}, "shanghai": {},
			"golang中文": {}, "中文文档": {}, "汉化": {},
			"微信": {}, "支付宝": {}, "腾讯": {}, "阿里": {},
		},
		"US": {
			"usa": {}, "united states": {}, "america": {},
			"california": {}, "new york": {}, "silicon valley": {},
		},
		// ... 其他国家的关键词
	}

	for _, repo := range repos {
		// 分析仓库名称和描述
		name := strings.ToLower(getPtrValue(repo.Name))
		desc := strings.ToLower(getPtrValue(repo.Description))
		readme := getRepoReadme(repo)

		// 使用关键词检查
		for country, words := range keywords {
			for word := range words {
				if strings.Contains(name, word) || strings.Contains(desc, word) {
					points[country] += 1.0
				}
			}
		}

		// 检查语言使用
		if ContainsChinese(desc) || ContainsChinese(readme) {
			points["CN"] += 2.0
		}
		if ContainsJapanese(desc) || ContainsJapanese(readme) {
			points["JP"] += 2.0
		}
		if ContainsKorean(desc) || ContainsKorean(readme) {
			points["KR"] += 2.0
		}

		// 检查域名和链接
		domains := extractDomains(desc + " " + readme)
		for domain := range domains {
			switch {
			case strings.HasSuffix(domain, ".cn"):
				points["CN"] += 1.5
			case strings.HasSuffix(domain, ".jp"):
				points["JP"] += 1.5
			case strings.HasSuffix(domain, ".kr"):
				points["KR"] += 1.5
				// ... 其他国家的域名
			}
		}

		// 分析提交信息的语言
		commitLanguage := analyzeCommitMessages(repo)
		if commitLanguage != "" {
			points[commitLanguage] += 1.0
		}
	}

	return points
}

// getRepoReadme 获取仓库的README内容
func getRepoReadme(repo *github.Repository) string {
	// 创建一个 GitHub 客户端
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return ""
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 获取仓库的所有者和名称
	owner := getPtrValue(repo.Owner.Login)
	repoName := getPtrValue(repo.Name)

	// 尝试获取 README 内容
	readme, _, err := client.Repositories.GetReadme(ctx, owner, repoName, nil)
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

// extractDomains 从文本中提取域名
func extractDomains(text string) map[string]struct{} {
	domains := make(map[string]struct{})
	// 使用正则表达式匹配域名
	re := regexp.MustCompile(`(?i)[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.(?:[a-z]{2,}\.)*[a-z]{2,}`)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		domains[match] = struct{}{}
	}
	return domains
}

// analyzeCommitMessages 分析提交信息的语言
func analyzeCommitMessages(repo *github.Repository) string {
	// 创建一个 GitHub 客户端
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return ""
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 获取仓库的所有者和名称
	owner := getPtrValue(repo.Owner.Login)
	repoName := getPtrValue(repo.Name)

	// 获取最近的提交记录
	commits, _, err := client.Repositories.ListCommits(ctx, owner, repoName, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // 获取最近100条提交
		},
	})
	if err != nil {
		return ""
	}

	// 统计不同语言的使用频率
	langPoints := make(map[string]int)
	totalCommits := 0

	for _, commit := range commits {
		message := getPtrValue(commit.Commit.Message)
		if message == "" {
			continue
		}

		totalCommits++

		// 检测中文
		if ContainsChinese(message) {
			langPoints["CN"]++
			continue
		}

		// 检测日语
		if ContainsJapanese(message) {
			langPoints["JP"]++
			continue
		}

		// 检测韩语
		if ContainsKorean(message) {
			langPoints["KR"]++
			continue
		}

		// 检测其他语言特征
		switch {
		case ContainsRussian(message):
			langPoints["RU"]++
		case ContainsArabic(message):
			langPoints["AR"]++
		case ContainsHindi(message):
			langPoints["IN"]++
		case ContainsThai(message):
			langPoints["TH"]++
		case ContainsVietnamese(message):
			langPoints["VN"]++
		}
	}

	// 如果总提交数太少，返回空
	if totalCommits < 5 {
		return ""
	}

	// 找出使用最多的语言
	var maxCount int
	var dominantLang string
	for lang, count := range langPoints {
		// 只有当某种语言使用频率超过20%时才考虑
		if count > maxCount && float64(count)/float64(totalCommits) > 0.2 {
			maxCount = count
			dominantLang = lang
		}
	}

	return dominantLang
}

// 添加更多语言检测函数
func ContainsRussian(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Cyrillic) {
			return true
		}
	}
	return false
}

func ContainsArabic(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Arabic) {
			return true
		}
	}
	return false
}

func ContainsHindi(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Devanagari) {
			return true
		}
	}
	return false
}

func ContainsThai(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Thai) {
			return true
		}
	}
	return false
}

func ContainsVietnamese(s string) bool {
	// 检测越南语特有的字符
	vietnameseChars := []rune{'ă', 'â', 'đ', 'ê', 'ô', 'ơ', 'ư', 'ạ', 'ả', 'ấ', 'ầ', 'ẩ', 'ẫ', 'ậ'}
	for _, vChar := range vietnameseChars {
		if strings.ContainsRune(s, vChar) {
			return true
		}
	}
	return false
}

// calculateConfidence 计算预测置信度
func calculateConfidence(maxScore float64, factorCount int) float64 {
	// 基础置信度
	baseConfidence := 0.3

	// 根据得分增加置信度
	scoreConfidence := math.Min(maxScore/10.0, 0.4)

	// 根据影响因素数量增加置信度
	factorConfidence := math.Min(float64(factorCount)/5.0, 0.3)

	totalConfidence := baseConfidence + scoreConfidence + factorConfidence
	return math.Min(totalConfidence, 1.0) * 100
}

// 辅助函数

func ContainsKorean(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hangul) {
			return true
		}
	}
	return false
}

// timeZoneToCountry 将时区映射到国家
func timeZoneToCountry(timezone string) string {
	// 时区到国家的映射
	timezoneMap := map[string]string{
		"Asia/Shanghai":    "CN",
		"Asia/Beijing":     "CN",
		"Asia/Tokyo":       "JP",
		"Asia/Seoul":       "KR",
		"America/New_York": "US",
		"Europe/London":    "GB",
		// 添加更多映射...
	}
	return timezoneMap[timezone]
}

// 辅助函数：获取指针值
func getPtrValue[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// 新增：基于提交时间的位置预测
func predictLocationFromCommitTimes(commits []time.Time) string {
	// 分析提交时间分布，推测时区
	timeZoneDistribution := analyzeTimeZones(commits)
	return mostLikelyLocation(timeZoneDistribution)
}

// 新增：基于社交网络的位置预测
func predictLocationFromNetwork(followers, following []*github.User) string {
	// 分析关注者和被关注者的地理分布
	networkDistribution := analyzeNetworkLocations(followers, following)
	return mostLikelyLocation(networkDistribution)
}

// 新增：基于项目协作的位置预测
func predictLocationFromCollaboration(collaborators []*github.User) string {
	// 分析项目协作者的地理分布
	collabDistribution := analyzeCollaboratorLocations(collaborators)
	return mostLikelyLocation(collabDistribution)
}

// 添加缺失的函数
func analyzeTimeZones(commits []time.Time) map[string]int {
	zones := make(map[string]int)
	for _, t := range commits {
		hour := t.Hour()
		// 根据提交时间推测时区
		switch {
		case hour >= 9 && hour <= 18: // 工作时间
			zones["local"]++
		case hour >= 19 || hour <= 2: // 晚间
			zones["offset+"]++
		case hour >= 3 && hour <= 8: // 凌晨
			zones["offset-"]++
		}
	}
	return zones
}

// mostLikelyLocation 根据分布确定最可能的位置
func mostLikelyLocation(distribution interface{}) string {
	switch dist := distribution.(type) {
	case map[string]int:
		var maxCount int
		var maxZone string
		for zone, count := range dist {
			if count > maxCount {
				maxCount = count
				maxZone = zone
			}
		}
		return zoneToCountry(maxZone)

	case map[string]float64:
		var maxScore float64
		var maxZone string
		for zone, score := range dist {
			if score > maxScore {
				maxScore = score
				maxZone = zone
			}
		}
		return zoneToCountry(maxZone)

	default:
		return ""
	}
}

// zoneToCountry 将区域代码转换为国家代码
func zoneToCountry(zone string) string {
	switch zone {
	case "local":
		return "CN" // 假设本地时区是中国
	case "offset+":
		return "US" // 假设是美国
	case "offset-":
		return "EU" // 假设是欧洲
	default:
		return zone // 如果已经是国家代码，直接返回
	}
}

func analyzeNetworkLocations(followers, following []*github.User) map[string]float64 {
	locations := make(map[string]float64)
	// TODO: 实现网络位置分析
	return locations
}

func analyzeCollaboratorLocations(collaborators []*github.User) map[string]float64 {
	locations := make(map[string]float64)
	// TODO: 实现协作者位置分析
	return locations
}

// 将函数名首字母大写以导出
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func ContainsJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana) {
			return true
		}
	}
	return false
}

func CalculateConfidence(maxScore float64, factorCount int) float64 {
	// 基础置信度
	baseConfidence := 0.3

	// 根据得分增加置信度
	scoreConfidence := math.Min(maxScore/10.0, 0.4)

	// 根据影响因素数量增加置信度
	factorConfidence := math.Min(float64(factorCount)/5.0, 0.3)

	totalConfidence := baseConfidence + scoreConfidence + factorConfidence
	return math.Min(totalConfidence, 1.0) * 100
}

// 导出 ExtractNation 函数
func ExtractNation(location string) string {
	if location == "" {
		return ""
	}
	// 扩展国家关键词映射
	countries := map[string]string{
		// 亚洲
		"China":     "CN",
		"中国":        "CN",
		"Japan":     "JP",
		"日本":        "JP",
		"Korea":     "KR",
		"韩国":        "KR",
		"Singapore": "SG",
		"新加坡":       "SG",
		"India":     "IN",
		"Thailand":  "TH",
		"Vietnam":   "VN",
		"Malaysia":  "MY",
		"Indonesia": "ID",

		// 北美洲
		"USA":           "US",
		"United States": "US",
		"America":       "US",
		"Canada":        "CA",
		"Mexico":        "MX",

		// 欧洲
		"UK":             "GB",
		"United Kingdom": "GB",
		"England":        "GB",
		"Germany":        "DE",
		"Deutschland":    "DE",
		"France":         "FR",
		"Italy":          "IT",
		"Italia":         "IT",
		"Spain":          "ES",
		"España":         "ES",
		"Netherlands":    "NL",
		"Nederland":      "NL",
		"Sweden":         "SE",
		"Sverige":        "SE",
		"Norway":         "NO",
		"Danmark":        "DK",
		"Finland":        "FI",
		"Switzerland":    "CH",
		"Ireland":        "IE",
		"Poland":         "PL",
		"Russia":         "RU",
		"Россия":         "RU",

		// 大洋洲
		"Australia":   "AU",
		"New Zealand": "NZ",

		// 南美洲
		"Brazil":    "BR",
		"Brasil":    "BR",
		"Argentina": "AR",
		"Chile":     "CL",

		// 非洲
		"South Africa": "ZA",
		"Egypt":        "EG",
		"Nigeria":      "NG",
	}

	location = strings.ToLower(location)
	for key, value := range countries {
		if strings.Contains(strings.ToLower(location), strings.ToLower(key)) {
			return value
		}
	}

	// 如果没有直接匹配，尝试使用城市来判断
	cities := map[string]string{
		// 中国城市
		"beijing":   "CN",
		"shanghai":  "CN",
		"shenzhen":  "CN",
		"guangzhou": "CN",
		"hangzhou":  "CN",
		"chengdu":   "CN",
		"nanjing":   "CN",
		"wuhan":     "CN",
		"xian":      "CN",
		"suzhou":    "CN",

		// 日本城市
		"tokyo":    "JP",
		"osaka":    "JP",
		"kyoto":    "JP",
		"yokohama": "JP",
		"sapporo":  "JP",
		"fukuoka":  "JP",
		"nagoya":   "JP",

		// 韩国城市
		"seoul":   "KR",
		"busan":   "KR",
		"incheon": "KR",

		// 美国城市
		"new york":      "US",
		"san francisco": "US",
		"seattle":       "US",
		"boston":        "US",
		"chicago":       "US",
		"los angeles":   "US",
		"san jose":      "US",
		"austin":        "US",
		"portland":      "US",
		"washington":    "US",

		// 英国城市
		"london":     "GB",
		"manchester": "GB",
		"cambridge":  "GB",
		"oxford":     "GB",
		"edinburgh":  "GB",
		"glasgow":    "GB",
		"bristol":    "GB",
	}

	for city, country := range cities {
		if strings.Contains(location, city) {
			return country
		}
	}

	return ""
}
