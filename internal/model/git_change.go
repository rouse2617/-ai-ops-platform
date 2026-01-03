package model

import "time"

// GitCommit represents a git commit with change details
type GitCommit struct {
	Hash          string    `json:"hash"`
	Author        string    `json:"author"`
	Email         string    `json:"email"`
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
	ChangedFiles  []string  `json:"changed_files"`
	Repository    string    `json:"repository"`
	Branch        string    `json:"branch"`
}

// CodeChangeCorrelation represents the correlation between alert and code change
type CodeChangeCorrelation struct {
	AlertID       string      `json:"alert_id"`
	AlertName     string      `json:"alert_name"`
	Service       string      `json:"service"`
	Commits       []GitCommit `json:"commits"`
	CorrelationScore float64  `json:"correlation_score"`
	SuspiciousFiles  []SuspiciousFile `json:"suspicious_files"`
	TimeWindow    TimeWindow  `json:"time_window"`
}

// SuspiciousFile represents a file change that might cause the alert
type SuspiciousFile struct {
	FilePath      string  `json:"file_path"`
	CommitHash    string  `json:"commit_hash"`
	ChangeType    string  `json:"change_type"` // added, modified, deleted
	Score         float64 `json:"score"`
	Reason        string  `json:"reason"`
	DiffURL       string  `json:"diff_url"`
}

// TimeWindow represents the time range for correlation
type TimeWindow struct {
	AlertTime   time.Time `json:"alert_time"`
	SearchStart time.Time `json:"search_start"`
	SearchEnd   time.Time `json:"search_end"`
}

// ServiceCodeMapping represents the mapping between service and code paths
type ServiceCodeMapping struct {
	ServiceName   string   `json:"service_name"`
	Repositories  []string `json:"repositories"`
	CodePaths     []string `json:"code_paths"`
	Keywords      []string `json:"keywords"`
}
