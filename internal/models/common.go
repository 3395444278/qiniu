package models

import "time"

type ContributionLevel struct {
	Commits       int     // 提交数
	PRs           int     // PR 数
	Reviews       int     // 代码评审数
	Issues        int     // Issue 数
	CoreDev       bool    // 是否是核心开发者
	CommitQuality float64 // 提交质量
	PRQuality     float64 // PR 质量
	Impact        float64 // 贡献影响力
}

// DeveloperMetrics 开发者评估指标
type DeveloperMetrics struct {
	Contributions ContributionsMetrics
	Projects      ProjectsMetrics
	Influence     InfluenceMetrics
	Activity      ActivityMetrics
	Expertise     ExpertiseMetrics
}

// ContributionsMetrics 基础指标结构体
type ContributionsMetrics struct {
	CommitCount int     // 提交数量
	PRCount     int     // PR数量
	ReviewCount int     // 代码审查数量
	IssueCount  int     // Issue数量
	Quality     float64 // 代码质量分数
}

// ProjectsMetrics 项目指标结构体
type ProjectsMetrics struct {
	TotalCount   int     // 项目总数
	StarCount    int     // 获得的 star 数
	ForkCount    int     // 获得的 fork 数
	WatchCount   int     // 观察者数量
	CoreProjects int     // 核心项目数量
	Quality      float64 // 项目质量分数
}

// InfluenceMetrics 影响力指标结构体
type InfluenceMetrics struct {
	Followers   int     // 关注者数量
	Following   int     // 关注数量
	Reach       float64 // 影响力范围
	Recognition float64 // 行业认可度
}

// ActivityMetrics 活跃度指标结构体
type ActivityMetrics struct {
	LastActive  time.Time // 最后活跃时间
	Frequency   float64   // 活动频率
	Consistency float64   // 持续性
	Growth      float64   // 增长趋势
}

// ExpertiseMetrics 专业度指标结构体
type ExpertiseMetrics struct {
	Languages   []string // 编程语言
	Domains     []string // 技术领域
	Specialties []string // 专长领域
	Depth       float64  // 技术深度
}
