package postgres

import (
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	FirstName     string     `gorm:"size:50" json:"first_name"`
	LastName      string     `gorm:"size:50" json:"last_name"`
	DisplayName   string     `gorm:"size:100" json:"display_name"`
	Avatar        string     `gorm:"size:500" json:"avatar"`
	AvatarHash    string     `gorm:"size:500" json:"avatar_hash"`
	WallImage     string     `gorm:"size:500" json:"wall_image"`
	WallImageHash string     `gorm:"size:500" json:"wall_image_hash"`
	Bio           string     `gorm:"type:text" json:"bio"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	Phone         string     `gorm:"size:20" json:"phone,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Table name
func (Profile) TableName() string {
	return "profiles"
}
