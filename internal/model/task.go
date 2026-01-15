package model

import "time"

type Task struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	BookID    uint      `json:"book_id" gorm:"index;not null"`
	PromptID  uint      `json:"prompt_id" gorm:"not null;default:0"`
	Type      string    `json:"type" gorm:"size:20;not null"`
	Status    string    `json:"status" gorm:"size:20;not null"`
	Progress  int       `json:"progress" gorm:"not null;default:0"`
	TakeTime  float64   `json:"take_time" gorm:"not null;default:0"`
	Error     string    `json:"error" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
