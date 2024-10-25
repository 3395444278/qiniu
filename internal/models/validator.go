package models

import (
	"time"
)

type Validator struct {
	MinStars         int // 改为 int
	MinContributions int // 改为 int
	MinFollowers     int // 改为 int
	MaxInactiveTime  time.Duration
}

func NewValidator() *Validator {
	return &Validator{
		MinStars:         5,
		MinContributions: 10,
		MinFollowers:     3,
		MaxInactiveTime:  365 * 24 * time.Hour, // 1年
	}
}

func (v *Validator) ValidateDeveloper(dev *Developer) bool {
	// 检查基本数据
	if dev.Username == "" {
		return false
	}

	// 检查活跃度
	if time.Since(dev.LastActive) > v.MaxInactiveTime {
		return false
	}

	// 检查最低要求
	if dev.StarCount < v.MinStars ||
		dev.CommitCount < v.MinContributions {
		return false
	}

	return true
}
