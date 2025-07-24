package models

import (
	"time"

	"gorm.io/gorm"
)

type Portfolio struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	ArtworkID     uint           `json:"artwork_id" gorm:"not null;index"`
	PortfolioName string         `json:"portfolio_name" gorm:"not null;default:'Main Portfolio'"`
	Position      int            `json:"position" gorm:"not null"` // 1-12 for ordering
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID"`
	Artwork Artwork `json:"artwork" gorm:"foreignKey:ArtworkID"`
}

// Ensure unique combination of UserID, PortfolioName, Position
// This prevents duplicate positions within the same portfolio
func (Portfolio) TableName() string {
	return "portfolios"
}
