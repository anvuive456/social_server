package requests

import "social_server/internal/models/constants"

type PostType = constants.PostType
type PostPrivacy = constants.PostPrivacy

const (
	PostTypeText  = constants.PostTypeText
	PostTypeImage = constants.PostTypeImage
	PostTypeVideo = constants.PostTypeVideo
	PostTypeLink  = constants.PostTypeLink
	PostTypeAudio = constants.PostTypeAudio
)

const (
	PostPrivacyPublic  = constants.PostPrivacyPublic
	PostPrivacyFriends = constants.PostPrivacyFriends
	PostPrivacyPrivate = constants.PostPrivacyPrivate
)

type GetPostsRequest struct {
	UserID    uint64 `json:"user_id,omitempty"`
	Limit     int    `json:"limit" binding:"required"`
	Sort      string `json:"sort,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
	After     string `json:"after,omitempty"`
	Before    string `json:"before,omitempty"`
}

type CreatePostRequest struct {
	Type     PostType           `json:"type" binding:"required"`
	Content  string             `json:"content"`
	Privacy  PostPrivacy        `json:"privacy"`
	Location string             `json:"location,omitempty"`
	Tags     []string           `json:"tags,omitempty"`
	Media    []PostMediaRequest `json:"media,omitempty"`
}

type UpdatePostRequest struct {
	Content  string      `json:"content,omitempty"`
	Privacy  PostPrivacy `json:"privacy,omitempty"`
	Location string      `json:"location,omitempty"`
	Tags     []string    `json:"tags,omitempty"`
}

type PostMediaRequest struct {
	Type      string `json:"type" binding:"required"`
	URL       string `json:"url" binding:"required"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
	AltText   string `json:"alt_text,omitempty"`
	Order     int    `json:"order"`
}

type CreateCommentRequest struct {
	Content        string `json:"content" binding:"required"`
	ParentID       *uint  `json:"parent_id,omitempty"`
	MentionedUsers []uint `json:"mentioned_users,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type SharePostRequest struct {
	Content string      `json:"content,omitempty"`
	Privacy PostPrivacy `json:"privacy"`
}
