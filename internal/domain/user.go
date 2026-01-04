package domain

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleGuest Role = "guest"
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"uniqueIndex;size:50" json:"username"`
	Password string `json:"-"` // 密码哈希，不返回给前端
	Role     Role   `gorm:"default:'user'" json:"role"`
}
