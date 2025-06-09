package responses

import (
	"social_server/internal/models/postgres"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type FriendRequestsResponse struct {
	Requests   []postgres.FriendRequest `json:"requests"`
	TotalCount uint                     `json:"total_count"`
	NextCursor *paginator.Cursor        `json:"next_cursor"`
}

type FriendStatsResponse struct {
	TotalReceived uint `json:"total_received"`
	TotalSent     uint `json:"total_sent"`
}

type FriendsResponse struct {
	Friends    []postgres.UserFriend `json:"friends"`
	NextCursor *paginator.Cursor     `json:"next_cursor,omitempty"`
	TotalCount uint                  `json:"total_count"`
}
