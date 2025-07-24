package models

import (
	"time"

	"gorm.io/gorm"
)

type Artwork struct {
	ID                uint          `json:"id" gorm:"primaryKey"`
	UserID            uint          `json:"user_id" gorm:"not null;index"`
	Title             string        `json:"title" gorm:"not null"`
	Description       string        `json:"description"`
	Filename          string        `json:"filename" gorm:"not null"`
	CloudflareImageID string        `json:"cloudflare_image_id" gorm:"not null"`
	ImageURL          string        `json:"image_url" gorm:"-"` // Computed field, not stored
	ImageVariants     ImageVariants `json:"image_variants" gorm:"embedded"`
	FileSize          int64         `json:"file_size" gorm:"not null"`
	Width             int           `json:"width"`
	Height            int           `json:"height"`
	Format            string        `json:"format" gorm:"not null"` // JPEG, PNG, PSD, etc.
	MimeType          string        `json:"mime_type" gorm:"not null"`
	Software          string        `json:"software"`
	Tags              string        `json:"tags"` // JSON array or comma-separated

	// Copyright and protection
	CopyrightRegistered bool `json:"copyright_registered" gorm:"default:true"`
	AIProtectionEnabled bool `json:"ai_protection_enabled" gorm:"default:false"`

	// Metadata
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User       User        `json:"user" gorm:"foreignKey:UserID"`
	Portfolios []Portfolio `json:"portfolios" gorm:"foreignKey:ArtworkID"`
}

type ImageVariants struct {
	Thumbnail string `json:"thumbnail"` // 200px variant
	Medium    string `json:"medium"`    // 800px variant
	Large     string `json:"large"`     // 1200px variant
	Original  string `json:"original"`  // Full size variant
}
