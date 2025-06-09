package postgres

import (
	"time"

	"gorm.io/gorm"
)

type FriendRequestStatus string

const (
	FriendRequestPending  FriendRequestStatus = "pending"
	FriendRequestAccepted FriendRequestStatus = "accepted"
	FriendRequestRejected FriendRequestStatus = "rejected"
)

type FriendRequest struct {
	ID         uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID   uint                `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint                `gorm:"not null;index" json:"receiver_id"`
	Status     FriendRequestStatus `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected
	Message    string              `gorm:"type:text" json:"message"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	DeletedAt  gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relationships
	Sender   *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver *User `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
