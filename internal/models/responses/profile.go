package responses

import (
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type ProfileDetailResponse struct {
	ID            uint          `json:"id"`
	UserID        uint          `json:"user_id"`
	FirstName     string        `json:"first_name"`
	LastName      string        `json:"last_name"`
	DisplayName   string        `json:"display_name"`
	Avatar        string        `json:"avatar"`
	AvatarHash    string        `json:"avatar_hash"`
	WallImage     string        `json:"wall_image"`
	WallImageHash string        `json:"wall_image_hash"`
	Bio           string        `json:"bio"`
	DateOfBirth   *time.Time    `json:"date_of_birth,omitempty"`
	Phone         string        `json:"phone,omitempty"`
	User          UserBasicInfo `json:"user"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ProfilesResponse is no longer needed for 1-1 relationship
// Use postgres.Profile directly for single profile responses

type UserBasicInfo struct {
	ID         uint       `json:"id"`
	Email      string     `json:"email"`
	IsOnline   bool       `json:"is_online"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	IsVerified bool       `json:"is_verified"`
	IsActive   bool       `json:"is_active"`
}

// UserProfilesResponse is no longer needed for 1-1 relationship
// Use ProfileDetailResponse directly for single user profile

type ProfileSearchResponse struct {
	Profiles   []ProfileResponse `json:"profiles"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`
}
