package models

import (
	"math"
	"sort"
	"strings"
)

// 新增：领域分类器
type DomainClassifier struct {
	// 技术栈到领域的映射
	TechStackDomains map[string][]string
	// 项目特征到领域的映射
	ProjectFeatureDomains map[string][]string
	// 行业分类
	IndustryDomains map[string][]string
}

// 新增：多维度领域判断
func (dc *DomainClassifier) ClassifyDomain(developer *Developer) []string {
	domains := make(map[string]float64)

	// 基于技术栈判断
	techDomains := dc.classifyByTechStack(developer.Skills)
	for domain, weight := range techDomains {
		domains[domain] += weight * 0.4
	}

	// 基于项目特征判断
	projectDomains := dc.classifyByProjects(developer.Repositories)
	for domain, weight := range projectDomains {
		domains[domain] += weight * 0.3
	}

	// 基于行业判断
	industryDomains := dc.classifyByIndustry(developer.Repositories)
	for domain, weight := range industryDomains {
		domains[domain] += weight * 0.3
	}

	return getTopDomains(domains, 3) // 返回前3个最可能的领域
}

// classifyByTechStack 基于技术栈判断领域
func (dc *DomainClassifier) classifyByTechStack(skills []string) map[string]float64 {
	domains := make(map[string]float64)

	// 技术栈权重映射
	techWeights := map[string]map[string]float64{
		"backend": {
			"Go": 0.9, "Java": 0.8, "Python": 0.7, "C++": 0.8,
			"Node.js": 0.7, "Ruby": 0.7, "PHP": 0.7,
		},
		"frontend": {
			"JavaScript": 0.9, "TypeScript": 0.9, "React": 0.8,
			"Vue": 0.8, "Angular": 0.8, "HTML": 0.6, "CSS": 0.6,
		},
		"mobile": {
			"Swift": 0.9, "Kotlin": 0.9, "Java": 0.7,
			"Flutter": 0.8, "React Native": 0.8,
		},
		// ... 其他领域
	}

	// 计算每个领域的得分
	for domain, weights := range techWeights {
		var score float64
		matchCount := 0
		for _, skill := range skills {
			if weight, exists := weights[skill]; exists {
				score += weight
				matchCount++
			}
		}
		if matchCount > 0 {
			domains[domain] = score / float64(matchCount)
		}
	}

	return domains
}

// classifyByProjects 基于项目特征判断领域
func (dc *DomainClassifier) classifyByProjects(repos []string) map[string]float64 {
	domains := make(map[string]float64)

	// 项目特征关键词
	projectFeatures := map[string][]string{
		"backend":  {"server", "api", "database", "microservice"},
		"frontend": {"ui", "interface", "component", "web"},
		"mobile":   {"android", "ios", "mobile", "app"},
		// ... 其他领域
	}

	// 分析项目名称和描述
	for domain, features := range projectFeatures {
		var score float64
		for _, repo := range repos {
			for _, feature := range features {
				if strings.Contains(strings.ToLower(repo), feature) {
					score += 0.5
				}
			}
		}
		if score > 0 {
			domains[domain] = score
		}
	}

	return domains
}

// getTopDomains 获取得分最高的领域
func getTopDomains(domains map[string]float64, limit int) []string {
	type domainScore struct {
		domain string
		score  float64
	}

	// 转换为切片以便排序
	scores := make([]domainScore, 0, len(domains))
	for d, s := range domains {
		scores = append(scores, domainScore{d, s})
	}

	// 按得分降序排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 获取前N个领域
	result := make([]string, 0, limit)
	for i := 0; i < limit && i < len(scores); i++ {
		result = append(result, scores[i].domain)
	}

	return result
}

// DomainPrediction 领域预测结果
type DomainPrediction struct {
	Domains    []string
	Confidence float64
}

// PredictDomains 预测开发者的领域
func (dc *DomainClassifier) PredictDomains(developer *Developer) *DomainPrediction {
	// 获取各个维度的领域得分
	techScores := dc.classifyByTechStack(developer.Skills)
	projectScores := dc.classifyByProjects(developer.Repositories)

	// 合并得分
	combinedScores := combineScores(techScores, projectScores)

	// 获取top领域
	domains := getTopDomains(combinedScores, 3)

	// 计算置信度
	confidence := calculateDomainConfidence(combinedScores, domains)

	return &DomainPrediction{
		Domains:    domains,
		Confidence: confidence,
	}
}

// combineScores 合并不同维度的得分
func combineScores(techScores, projectScores map[string]float64) map[string]float64 {
	combined := make(map[string]float64)

	// 合并技术栈得分
	for domain, score := range techScores {
		combined[domain] += score * 0.6 // 技术栈权重更高
	}

	// 合并项目特征得分
	for domain, score := range projectScores {
		combined[domain] += score * 0.4
	}

	return combined
}

// calculateDomainConfidence 计算领域预测的置信度
func calculateDomainConfidence(scores map[string]float64, topDomains []string) float64 {
	if len(topDomains) == 0 {
		return 0
	}

	// 获取最高分
	var maxScore float64
	for _, domain := range topDomains {
		if scores[domain] > maxScore {
			maxScore = scores[domain]
		}
	}

	// 基础置信度
	baseConfidence := 0.3

	// 根据最高分调整置信度
	scoreConfidence := maxScore / 10.0 // 假设最高可能分数为10

	// 根据领域数量调整置信度
	domainConfidence := float64(len(scores)) / 10.0

	totalConfidence := baseConfidence + scoreConfidence + domainConfidence
	return math.Min(totalConfidence, 1.0) * 100
}

// classifyByIndustry 基于行业特征判断领域
func (dc *DomainClassifier) classifyByIndustry(repos []string) map[string]float64 {
	domains := make(map[string]float64)

	// 行业关键词映射
	industryKeywords := map[string][]string{
		"fintech":   {"payment", "banking", "finance", "blockchain", "crypto"},
		"ecommerce": {"shop", "store", "retail", "marketplace"},
		"gaming":    {"game", "unity", "unreal", "gaming"},
		"ai":        {"machine-learning", "deep-learning", "neural", "ai"},
		// ... 添加更多行业
	}

	for _, repo := range repos {
		for industry, keywords := range industryKeywords {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(repo), keyword) {
					domains[industry] += 0.5
				}
			}
		}
	}

	return domains
}
