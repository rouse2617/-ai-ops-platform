package model

import "time"

type Task struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	SessionID string     `json:"sessionId" gorm:"index"`
	Type      string     `json:"type"`
	Status    string     `json:"status" gorm:"index"`
	Input     string     `json:"input" gorm:"type:text"`
	Output    string     `json:"output,omitempty" gorm:"type:text"`
	Error     string     `json:"error,omitempty" gorm:"type:text"`
	HostID    string     `json:"hostId,omitempty" gorm:"index"`
	HostName  string     `json:"hostName,omitempty"`
	ToolName  string     `json:"toolName,omitempty"`
	StartTime time.Time  `json:"startTime" gorm:"index"`
	EndTime   *time.Time `json:"endTime,omitempty"`
	Duration  int64      `json:"duration,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
