package domain

import (
	"time"
)

// SharedEncyclopedia 公共百科池 (基于书籍指纹共享)
type SharedEncyclopedia struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	BookFingerprint string    `gorm:"index:idx_book_range;size:32" json:"book_fingerprint"`
	RangeEnd       int       `gorm:"index:idx_book_range" json:"range_end"` // 例如 50 表示覆盖 1-50 章
	Content        string    `gorm:"type:text" json:"content"`
	Version        string    `gorm:"size:20" json:"version"`
	CreatedAt      time.Time `json:"created_at"`
}
