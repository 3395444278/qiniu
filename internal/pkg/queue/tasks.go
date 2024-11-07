package queue

import "time"

type EvaluationTask struct {
	Username     string    `json:"username"`
	ProfileURL   string    `json:"profile_url"`
	BlogURL      string    `json:"blog_url"`
	Description  string    `json:"description"`
	Repositories []string  `json:"repositories"`
	CreatedAt    time.Time `json:"created_at"`
}

type EvaluationResult struct {
	Username      string            `json:"username"`
	Specialties   []string          `json:"specialties"`
	Experience    map[string]string `json:"experience"`
	AIEvaluation  string            `json:"ai_evaluation"`
	LastEvaluated time.Time         `json:"last_evaluated"`
}
