package model

import "time"

type Chapter struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	BookID     uint      `json:"bookId" gorm:"index;not null"`
	Index      int       `json:"index" gorm:"not null"`
	Title      string    `json:"title" gorm:"size:255;not null"`
	ChapterMD5 string    `json:"chapterMd5" gorm:"size:32;not null"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

type ChapterContent struct {
	ChapterMD5 string    `json:"chapterMd5" gorm:"primaryKey;size:32"`
	Content    string    `json:"content" gorm:"type:longtext;not null"`
	WordsCount int       `json:"wordsCount" gorm:"not null"`
	TokenCount int       `json:"tokenCount" gorm:"not null"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
