package domain

import (
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"index" json:"user_id"`
	User   User `json:"-"`

	Title         string `gorm:"size:255" json:"title"`
	Author        string `gorm:"size:255" json:"author"`
	TotalChapters int    `json:"total_chapters"`
	Status        string `json:"status"` // e.g., "processing", "completed"
	
	Chapters []Chapter `gorm:"foreignKey:BookID" json:"chapters,omitempty"`
}
