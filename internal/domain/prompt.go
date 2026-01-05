package domain

import (
	"time"

	"gorm.io/gorm"
)

type Prompt struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name      string `gorm:"size:50" json:"name"`
	Version   string `gorm:"size:20" json:"version"`
	Content   string `gorm:"type:text" json:"content"`
	IsSystem  bool   `gorm:"index" json:"is_system"`
}
