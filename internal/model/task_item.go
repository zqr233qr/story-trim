package model

import "time"

// TaskItem 任务章节处理记录。
type TaskItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    string    `json:"task_id" gorm:"index;size:36;not null"`
	ChapterID uint      `json:"chapter_id" gorm:"index;not null"`
	PromptID  uint      `json:"prompt_id" gorm:"index;not null"`
	Status    string    `json:"status" gorm:"size:20;not null"`
	Error     string    `json:"error" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
