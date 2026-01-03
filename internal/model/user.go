package model

import "time"

// User 用户模型
type User struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex"`
	Role         string    `json:"role" gorm:"default:operator"` // admin, operator, viewer
	Status       string    `json:"status" gorm:"default:active"` // active, disabled
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
