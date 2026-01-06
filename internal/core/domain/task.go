package domain

import (
	"encoding/json"
	"time"
)

type Task struct {
	ID        string          `json:"id"`
	UserID    uint            `json:"user_id"`
	BookID    uint            `json:"book_id"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	Progress  int             `json:"progress"`
	Meta      json.RawMessage `json:"meta"`
	Error     string          `json:"error"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Prompt struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Content  string `json:"content"`
	IsSystem bool   `json:"is_system"`
}