package responses

import (
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type ProfileResponse struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	DisplayName string     `json:"display_name"`
	Avatar      string     `json:"avatar"`
	AvatarHash  string     `json:"avatar_hash"`
	Bio         string     `json:"bio"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type UserProfile struct {
	ID          uint            `json:"id"`
	IsOnline    bool            `json:"is_online"`
	IsVerified  bool            `json:"is_verified"`
	FriendCount int             `json:"friend_count"`
	PostCount   int             `json:"post_count"`
	Profile     ProfileResponse `json:"profile,omitempty"`
}

type UserListResponse struct {
	Users      []UserProfile     `json:"users"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type FriendRequestResponse struct {
	ID         uint        `json:"id"`
	SenderID   uint        `json:"sender_id"`
	ReceiverID uint        `json:"receiver_id"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Sender     UserProfile `json:"sender"`
	Receiver   UserProfile `json:"receiver"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type FriendRequestListResponse struct {
	Requests   []FriendRequestResponse `json:"requests"`
	NextCursor *paginator.Cursor       `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
	Total      int                     `json:"total"`
}

type UserSettingsResponse struct {
	PrivacyProfileVisibility    string `json:"privacy_profile_visibility"`
	PrivacyShowOnlineStatus     bool   `json:"privacy_show_online_status"`
	PrivacyAllowFriendRequests  bool   `json:"privacy_allow_friend_requests"`
	NotificationsEmail          bool   `json:"notifications_email"`
	NotificationsPush           bool   `json:"notifications_push"`
	NotificationsFriendRequests bool   `json:"notifications_friend_requests"`
	NotificationsMessages       bool   `json:"notifications_messages"`
	NotificationsPosts          bool   `json:"notifications_posts"`
}

type UserDetailResponse struct {
	ID          uint                 `json:"id"`
	Email       string               `json:"email"`
	IsOnline    bool                 `json:"is_online"`
	LastSeen    *time.Time           `json:"last_seen,omitempty"`
	IsVerified  bool                 `json:"is_verified"`
	IsActive    bool                 `json:"is_active"`
	FriendCount int                  `json:"friend_count"`
	PostCount   int                  `json:"post_count"`
	Settings    UserSettingsResponse `json:"settings"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type UserSearchResponse struct {
	Users    []UserProfile `json:"users"`
	Total    int           `json:"total"`
	Query    string        `json:"query"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	HasMore  bool          `json:"has_more"`
}

type FriendSuggestionResponse struct {
	User          UserProfile `json:"user"`
	MutualFriends int         `json:"mutual_friends"`
	Reason        string      `json:"reason"`
	Score         float64     `json:"score"`
}

type FriendSuggestionsResponse struct {
	Suggestions []FriendSuggestionResponse `json:"suggestions"`
	Total       int                        `json:"total"`
}

type OnlineStatusResponse struct {
	UserID     uint      `json:"user_id"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen"`
	Status     string    `json:"status"`
	DeviceInfo string    `json:"device_info"`
}

type BlockedUserResponse struct {
	User      UserProfile `json:"user"`
	BlockedAt time.Time   `json:"blocked_at"`
}

type BlockedUsersResponse struct {
	Users   []BlockedUserResponse `json:"users"`
	Total   int                   `json:"total"`
	HasMore bool                  `json:"has_more"`
}

type UserSearchSimpleResponse struct {
	ID      uint            `json:"id"`
	Email   string          `json:"email"`
	Profile ProfileResponse `json:"profile"`
}

type UserSearchSimpleListResponse struct {
	Users      []UserSearchSimpleResponse `json:"users"`
	NextCursor *paginator.Cursor          `json:"next_cursor,omitempty"`
	Total      int                        `json:"total"`
}
