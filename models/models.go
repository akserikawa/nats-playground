package models

import "time"

type Task struct {
	ID        int       `json:"id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskResult struct {
	TaskID     int           `json:"task_id"`
	ConsumerID int           `json:"consumer_id"`
	Status     string        `json:"status"`
	Duration   time.Duration `json:"duration_ms"`
	FinishedAt time.Time     `json:"finished_at"`
}

type WSEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ConsumerStats struct {
	ConsumerID    int       `json:"consumer_id"`
	TotalHandled  int       `json:"total_handled"`
	AvgDurationMs float64   `json:"avg_duration_ms"`
	LastActive    time.Time `json:"last_active"`
}

type DashboardState struct {
	Consumers     map[int]*ConsumerStats `json:"consumers"`
	RecentResults []TaskResult           `json:"recent_results"`
	TotalPublished int                   `json:"total_published"`
}
