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

// ChapterContent 表示章节内容的元信息。
type ChapterContent struct {
	ChapterMD5 string    `json:"chapter_md5" gorm:"primaryKey;size:32"`
	ObjectKey  string    `json:"object_key" gorm:"size:255;not null"`
	Size       int64     `json:"size" gorm:"not null"`
	Content    string    `json:"content" gorm:"-"`
	WordsCount int       `json:"words_count" gorm:"not null"`
	CreatedAt  time.Time `json:"create_aAt" gorm:"autoCreateTime"`
}
