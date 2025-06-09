package postgres

import (
	"time"

	"gorm.io/gorm"
)

// User roles constants
const (
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
	RoleSuperUser = "superuser"
)

// Permission levels
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionAdmin  = "admin"
)

type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Email        string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Role         string     `gorm:"size:20;default:user;not null" json:"role"`
	Permissions  string     `gorm:"type:text" json:"permissions,omitempty"` // JSON array of permissions
	IsOnline     bool       `gorm:"default:false" json:"is_online"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	IsVerified   bool       `gorm:"default:false" json:"is_verified"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsBanned     bool       `gorm:"default:false" json:"is_banned"`
	BannedUntil  *time.Time `json:"banned_until,omitempty"`
	BanReason    string     `gorm:"type:text" json:"ban_reason,omitempty"`

	// Relationships
	UserFriends            []UserFriend    `gorm:"foreignKey:UserID" json:"user_friends"`
	FriendOf               []UserFriend    `gorm:"foreignKey:FriendID" json:"friend_of"`
	SentFriendRequests     []FriendRequest `gorm:"foreignKey:SenderID" json:"sent_friend_requests"`
	ReceivedFriendRequests []FriendRequest `gorm:"foreignKey:ReceiverID" json:"received_friend_requests"`
	BlockedUsers           []User          `gorm:"many2many:user_blocks;" json:"blocked_users"`
	Profile                *Profile        `gorm:"foreignKey:UserID" json:"profile,omitempty"`

	// Embedded settings
	Settings UserSettings `gorm:"embedded;embeddedPrefix:settings_" json:"settings"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserSettings struct {
	// Privacy settings
	PrivacyProfileVisibility   string `gorm:"size:20;default:public" json:"privacy_profile_visibility"`
	PrivacyShowOnlineStatus    bool   `gorm:"default:true" json:"privacy_show_online_status"`
	PrivacyAllowFriendRequests bool   `gorm:"default:true" json:"privacy_allow_friend_requests"`

	// Notification settings
	NotificationsEmail          bool `gorm:"default:true" json:"notifications_email"`
	NotificationsPush           bool `gorm:"default:true" json:"notifications_push"`
	NotificationsFriendRequests bool `gorm:"default:true" json:"notifications_friend_requests"`
	NotificationsMessages       bool `gorm:"default:true" json:"notifications_messages"`
	NotificationsPosts          bool `gorm:"default:true" json:"notifications_posts"`
}

// Table names
func (User) TableName() string {
	return "users"
}

// Helper methods for role checking
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperUser
}

func (u *User) IsModerator() bool {
	return u.Role == RoleModerator || u.IsAdmin()
}

func (u *User) IsSuperUser() bool {
	return u.Role == RoleSuperUser
}

func (u *User) HasRole(role string) bool {
	switch role {
	case RoleSuperUser:
		return u.IsSuperUser()
	case RoleAdmin:
		return u.IsAdmin()
	case RoleModerator:
		return u.IsModerator()
	case RoleUser:
		return true // All users have user role
	default:
		return false
	}
}

func (u *User) CanAccess(requiredRole string) bool {
	return u.HasRole(requiredRole)
}

func (u *User) IsAccountActive() bool {
	if !u.IsActive || u.IsBanned {
		return false
	}

	if u.BannedUntil != nil && u.BannedUntil.After(time.Now()) {
		return false
	}

	return true
}
