package model

type Prompt struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	Name             string  `json:"name" gorm:"size:50;not null"`
	Description      string  `json:"description" gorm:"size:255"`
	PromptContent    string  `json:"promptContent" gorm:"type:text"`
	TargetRatioMin   float64 `json:"targetRatioMin" gorm:"not null"`
	TargetRatioMax   float64 `json:"targetRatioMax" gorm:"not null"`
	BoundaryRatioMin float64 `json:"boundaryRatioMin" gorm:"not null"`
	BoundaryRatioMax float64 `json:"boundaryRatioMax" gorm:"not null"`
	IsSystem         bool    `json:"isSystem" gorm:"not null;default:false"`
	IsDefault        bool    `json:"isDefault" gorm:"not null;default:false"`
}
