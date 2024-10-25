package models

import "math"

// 导入公共定义
// import "qinniu/internal/models/common.go"

// 其他代码保持不变

// CalculateTalentRank 计算开发者的技术能力评分
func CalculateTalentRank(metrics *DeveloperMetrics) float64 {
	const maxScore = 100.0

	// 1. 计算贡献度分数 (25%)
	contributionScore := calculateContributionScore(metrics.Contributions)

	// 2. 计算项目影响力分数 (25%)
	projectScore := calculateProjectScore(metrics.Projects)

	// 3. 计算影响力分数 (20%)
	influenceScore := calculateInfluenceScore(metrics.Influence)

	// 4. 计算活跃度分数 (15%)
	activityScore := calculateActivityScore(metrics.Activity)

	// 5. 计算专业度分数 (15%)
	expertiseScore := calculateExpertiseScore(metrics.Expertise)

	// 计算加权总分
	totalScore := (contributionScore*0.25 +
		projectScore*0.25 +
		influenceScore*0.20 +
		activityScore*0.15 +
		expertiseScore*0.15) * maxScore

	return math.Min(totalScore, maxScore)
}

// 辅助函数实现...

// 计算贡献度分数
func calculateContributionScore(contributions ContributionsMetrics) float64 {
	// 1. 代码提交权重 (40%)
	commitScore := math.Log1p(float64(contributions.CommitCount)) / math.Log1p(10000)

	// 2. PR 数量和质量权重 (30%)
	prScore := math.Log1p(float64(contributions.PRCount)) / math.Log1p(1000) * contributions.Quality

	// 3. 代码审查权重 (20%)
	reviewScore := math.Log1p(float64(contributions.ReviewCount)) / math.Log1p(500)

	// 4. Issue 参与度 (10%)
	issueScore := math.Log1p(float64(contributions.IssueCount)) / math.Log1p(1000)

	return commitScore*0.4 + prScore*0.3 + reviewScore*0.2 + issueScore*0.1
}

// 计算项目影响力分数
func calculateProjectScore(projects ProjectsMetrics) float64 {
	// 1. Star 影响力 (35%)
	starScore := math.Log1p(float64(projects.StarCount)) / math.Log1p(100000)

	// 2. Fork 影响力 (25%)
	forkScore := math.Log1p(float64(projects.ForkCount)) / math.Log1p(10000)

	// 3. 核心项目权重 (25%)
	coreScore := 0.0
	if projects.TotalCount > 0 {
		coreScore = float64(projects.CoreProjects) / float64(projects.TotalCount)
	}

	// 4. 项目质量 (15%)
	qualityScore := projects.Quality

	return starScore*0.35 + forkScore*0.25 + coreScore*0.25 + qualityScore*0.15
}

// 计算影响力分数
func calculateInfluenceScore(influence InfluenceMetrics) float64 {
	// 1. 关注者影响力 (40%)
	followerScore := math.Log1p(float64(influence.Followers)) / math.Log1p(10000)

	// 2. 行业认可度 (35%)
	recognitionScore := influence.Recognition

	// 3. 影响力范围 (25%)
	reachScore := influence.Reach

	return followerScore*0.4 + recognitionScore*0.35 + reachScore*0.25
}

// 计算活跃度分数
func calculateActivityScore(activity ActivityMetrics) float64 {
	// 1. 活动频率 (35%)
	frequencyScore := activity.Frequency

	// 2. 持续性 (35%)
	consistencyScore := activity.Consistency

	// 3. 增长趋势 (30%)
	growthScore := activity.Growth

	return frequencyScore*0.35 + consistencyScore*0.35 + growthScore*0.3
}

// 计算专业度分数
func calculateExpertiseScore(expertise ExpertiseMetrics) float64 {
	// 1. 技术广度 (30%)
	breadthScore := math.Min(float64(len(expertise.Languages))/10.0, 1.0)

	// 2. 领域覆盖 (30%)
	domainScore := math.Min(float64(len(expertise.Domains))/5.0, 1.0)

	// 3. 技术深度 (40%)
	depthScore := expertise.Depth

	return breadthScore*0.3 + domainScore*0.3 + depthScore*0.4
}
