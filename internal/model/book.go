package model

import "time"

type Book struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"index;not null"`
	BookMD5       string    `json:"book_md5" gorm:"size:32;index"`
	Title         string    `json:"title" gorm:"size:255;not null"`
	TotalChapters int       `json:"total_chapters" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}
