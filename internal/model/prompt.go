package model

type Prompt struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	Name             string  `json:"name" gorm:"size:50;not null"`
	Description      string  `json:"description" gorm:"size:255"`
	PromptContent    string  `json:"-" gorm:"type:text"`
	TargetRatioMin   float64 `json:"-" gorm:"not null"`
	TargetRatioMax   float64 `json:"-" gorm:"not null"`
	BoundaryRatioMin float64 `json:"-" gorm:"not null"`
	BoundaryRatioMax float64 `json:"-" gorm:"not null"`
	IsSystem         bool    `json:"-" gorm:"not null;default:false"`
	IsDefault        bool    `json:"is_default" gorm:"not null;default:false"`
}
