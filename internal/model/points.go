package model

import "time"

// UserPoints 用户积分余额表。
type UserPoints struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	Balance   int       `json:"balance" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PointsLedger 积分流水记录表。
type PointsLedger struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	Change       int       `json:"change" gorm:"not null"`
	BalanceAfter int       `json:"balance_after" gorm:"not null"`
	Type         string    `json:"type" gorm:"size:20;not null"`
	Reason       string    `json:"reason" gorm:"size:50;not null"`
	RefType      string    `json:"ref_type" gorm:"size:20"`
	RefID        string    `json:"ref_id" gorm:"size:64"`
	Extra        string    `json:"extra" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
