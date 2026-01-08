package models

import "time"

// Schedule represents a GitLab pipeline schedule
type Schedule struct {
	ID           int        `json:"id"`
	Description  string     `json:"description"`
	Ref          string     `json:"ref"`           // Target branch
	Cron         string     `json:"cron"`          // Cron expression
	CronTimezone string     `json:"cron_timezone"` // Timezone for cron
	NextRunAt    *time.Time `json:"next_run_at"`
	Active       bool       `json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Owner        Owner      `json:"owner"`
	LastPipeline *Pipeline  `json:"last_pipeline"`
	Variables    []Variable `json:"variables"`
}

// Owner represents the schedule owner
type Owner struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// User represents the current GitLab user
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// Pipeline represents the last pipeline status
type Pipeline struct {
	ID        int        `json:"id"`
	IID       int        `json:"iid"`
	SHA       string     `json:"sha"`
	Ref       string     `json:"ref"`
	Status    string     `json:"status"` // "success", "failed", "pending", "running", "canceled"
	WebURL    string     `json:"web_url"`
	Source    string     `json:"source"` // "push", "web", "trigger", "schedule", "pipeline", "parent_pipeline", etc.
	Name      string     `json:"name"`   // Pipeline name (usually commit title)
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	User      *User      `json:"user"`
}

// PipelineJob represents a job in a pipeline
type PipelineJob struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Stage     string     `json:"stage"`
	Status    string     `json:"status"` // "success", "failed", "pending", "running", "canceled", "skipped"
	CreatedAt *time.Time `json:"created_at"`
	StartedAt *time.Time `json:"started_at"`
	Duration  float64    `json:"duration"`
	WebURL    string     `json:"web_url"`
}

// PipelineWithJobs represents a pipeline with its jobs for display
type PipelineWithJobs struct {
	Pipeline Pipeline
	Jobs     []PipelineJob
	Stages   []StageInfo
}

// StageInfo represents aggregated stage information
type StageInfo struct {
	Name   string
	Status string
}

// PipelineCreateRequest represents a request to create a pipeline
type PipelineCreateRequest struct {
	Ref       string     `json:"ref"`
	Variables []Variable `json:"variables,omitempty"`
}

// Variable represents a pipeline schedule variable
type Variable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	VariableType string `json:"variable_type"` // "env_var" or "file"
}

// Branch represents a GitLab branch
type Branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Default   bool   `json:"default"`
}

// ScheduleCreateRequest represents request to create a schedule
type ScheduleCreateRequest struct {
	Description  string     `json:"description"`
	Ref          string     `json:"ref"`
	Cron         string     `json:"cron"`
	CronTimezone string     `json:"cron_timezone,omitempty"`
	Active       bool       `json:"active"`
	Variables    []Variable `json:"-"` // Handled separately
}

// ScheduleUpdateRequest represents request to update a schedule
type ScheduleUpdateRequest struct {
	Description  *string `json:"description,omitempty"`
	Ref          *string `json:"ref,omitempty"`
	Cron         *string `json:"cron,omitempty"`
	CronTimezone *string `json:"cron_timezone,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}
