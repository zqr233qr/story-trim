package model

import "time"

type Chapter struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	BookID     uint      `json:"book_id" gorm:"index:idx_bookid_index,unique;not null"`
	Index      int       `json:"index" gorm:"index:idx_bookid_index,unique;not null"`
	Title      string    `json:"title" gorm:"size:255;not null"`
	ChapterMD5 string    `json:"chapter_md5" gorm:"size:32;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type ChapterContent struct {
	ChapterMD5 string    `json:"chapter_md5" gorm:"primaryKey;size:32"`
	Content    string    `json:"content" gorm:"type:longtext;not null"`
	WordsCount int       `json:"words_count" gorm:"not null"`
	TokenCount int       `json:"token_count" gorm:"not null"`
	CreatedAt  time.Time `json:"create_aAt" gorm:"autoCreateTime"`
}
