package model

import (
	"time"
)

// Script 脚本模型
type Script struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"uniqueIndex;not null"`
	Description string            `json:"description" gorm:"type:text"`
	Language    string            `json:"language" gorm:"default:bash"` // bash, python, powershell
	Content     string            `json:"content" gorm:"type:text;not null"`
	Parameters  []ScriptParameter `json:"parameters" gorm:"serializer:json"`
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// ScriptParameter 脚本参数
type ScriptParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ScriptLanguage 脚本语言常量
const (
	ScriptLanguageBash       = "bash"
	ScriptLanguagePython     = "python"
	ScriptLanguagePowerShell = "powershell"
)
