package responses

import (
	"social_server/internal/models/postgres"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type PostResponse struct {
	Posts      []postgres.Post   `json:"items"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type CommentResponse struct {
	Comments   []postgres.Comment `json:"items"`
	NextCursor *paginator.Cursor  `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type LikeResponse struct {
	Likes      []postgres.Like   `json:"items"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type ShareResponse struct {
	Shares     []postgres.Share  `json:"items"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type PostStats struct {
	TotalPosts    int64 `json:"total_posts"`
	TotalLikes    int64 `json:"total_likes"`
	TotalComments int64 `json:"total_comments"`
	TotalShares   int64 `json:"total_shares"`
	TotalViews    int64 `json:"total_views"`
}

type FeedPost struct {
	postgres.Post
	IsLiked    bool   `json:"is_liked"`
	IsShared   bool   `json:"is_shared"`
	IsSaved    bool   `json:"is_saved"`
	MyReaction string `json:"my_reaction,omitempty"`
}

type FeedResponse struct {
	Posts      []FeedPost        `json:"posts"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`
}

type PostDetailResponse struct {
	Post       postgres.Post      `json:"post"`
	Comments   []postgres.Comment `json:"comments"`
	IsLiked    bool               `json:"is_liked"`
	IsShared   bool               `json:"is_shared"`
	IsSaved    bool               `json:"is_saved"`
	MyReaction string             `json:"my_reaction,omitempty"`
}

type PopularPostsResponse struct {
	Posts   []FeedPost `json:"posts"`
	Period  string     `json:"period"`
	Total   int        `json:"total"`
	HasMore bool       `json:"has_more"`
}

type TrendingTagsResponse struct {
	Tags []postgres.PostTag `json:"tags"`
}

type PostSearchResponse struct {
	Posts    []FeedPost `json:"posts"`
	Total    int        `json:"total"`
	Query    string     `json:"query"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	HasMore  bool       `json:"has_more"`
}

type PostMediaResponse struct {
	Media []postgres.PostMedia `json:"media"`
	Total int                  `json:"total"`
}

type UserPostsResponse struct {
	Posts      []FeedPost        `json:"posts"`
	User       UserProfile       `json:"user"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Stats PostStats `json:"stats"`
}
