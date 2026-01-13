package model

import "time"

type Book struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"userId" gorm:"index;not null"`
	BookMD5       string    `json:"bookMd5" gorm:"size:32;index"`
	Title         string    `json:"title" gorm:"size:255;not null"`
	TotalChapters int       `json:"totalChapters" gorm:"not null"`
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
