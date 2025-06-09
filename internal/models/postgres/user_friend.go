package postgres

import (
	"time"

	"gorm.io/gorm"
)

// UserFriend status constants
const (
	FriendStatusActive   = "active"
	FriendStatusBlocked  = "blocked"
	FriendStatusInactive = "inactive"
)

type UserFriend struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	FriendID uint   `gorm:"not null;index" json:"friend_id"`
	Status   string `gorm:"size:20;default:active;not null" json:"status"`

	// Relationships
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Friend *User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Table name
func (UserFriend) TableName() string {
	return "user_friends"
}

// Helper methods
func (uf *UserFriend) IsActive() bool {
	return uf.Status == FriendStatusActive
}

func (uf *UserFriend) IsBlocked() bool {
	return uf.Status == FriendStatusBlocked
}

// Composite unique index to prevent duplicate friendships
func (UserFriend) TableOptions() string {
	return "ENGINE=InnoDB"
}
