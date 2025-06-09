package postgres

import (
	"social_server/internal/models/constants"
	"time"

	"gorm.io/gorm"
)

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

type Post struct {
	ID             uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID       uint        `gorm:"not null;index" json:"author_id"`
	Type           PostType    `gorm:"size:20;not null" json:"type"`
	Content        string      `gorm:"type:text" json:"content"`
	Privacy        PostPrivacy `gorm:"size:20;default:public" json:"privacy"`
	Location       string      `gorm:"size:255" json:"location"`
	Tags           string      `gorm:"type:text" json:"tags"`            // JSON array as string
	MentionedUsers string      `gorm:"type:text" json:"mentioned_users"` // JSON array as string
	IsEdited       bool        `gorm:"default:false" json:"is_edited"`
	EditedAt       *time.Time  `json:"edited_at,omitempty"`
	IsArchived     bool        `gorm:"default:false" json:"is_archived"`
	LikeCount      int         `gorm:"default:0" json:"like_count"`
	CommentCount   int         `gorm:"default:0" json:"comment_count"`
	ShareCount     int         `gorm:"default:0" json:"share_count"`
	ViewCount      int         `gorm:"default:0" json:"view_count"`

	// Relationships
	Author   User        `gorm:"foreignKey:AuthorID" json:"author"`
	Media    []PostMedia `gorm:"foreignKey:PostID" json:"media"`
	Likes    []Like      `gorm:"foreignKey:PostID" json:"likes"`
	Comments []Comment   `gorm:"foreignKey:PostID" json:"comments"`
	Shares   []Share     `gorm:"foreignKey:PostID" json:"shares"`
	Views    []PostView  `gorm:"foreignKey:PostID" json:"views"`

	// Timestamps
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PostMedia struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint   `gorm:"not null;index" json:"post_id"`
	Type      string `gorm:"size:50;not null" json:"type"`
	URL       string `gorm:"size:500;not null" json:"url"`
	Filename  string `gorm:"size:255" json:"filename"`
	Size      int64  `json:"size"`
	MimeType  string `gorm:"size:100" json:"mime_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	Thumbnail string `gorm:"size:500" json:"thumbnail,omitempty"`
	AltText   string `gorm:"size:255" json:"alt_text,omitempty"`
	Order     int    `gorm:"default:0" json:"order"`
	BlurHash  string `gorm:"size:255" json:"blur_hash,omitempty"`

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"post"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Like struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID uint   `gorm:"not null;index" json:"post_id"`
	UserID uint   `gorm:"not null;index" json:"user_id"`
	Type   string `gorm:"size:20;default:like" json:"type"` // like, love, haha, wow, sad, angry

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"post"`
	User User `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Comment struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID         uint       `gorm:"not null;index" json:"post_id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	ParentID       *uint      `gorm:"index" json:"parent_id,omitempty"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	Media          string     `gorm:"type:text" json:"media"`           // JSON array as string
	MentionedUsers string     `gorm:"type:text" json:"mentioned_users"` // JSON array as string
	IsEdited       bool       `gorm:"default:false" json:"is_edited"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
	LikeCount      int        `gorm:"default:0" json:"like_count"`
	ReplyCount     int        `gorm:"default:0" json:"reply_count"`

	// Relationships
	Post    Post          `gorm:"foreignKey:PostID" json:"post"`
	User    User          `gorm:"foreignKey:UserID" json:"user"`
	Parent  *Comment      `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment     `gorm:"foreignKey:ParentID" json:"replies"`
	Likes   []CommentLike `gorm:"foreignKey:CommentID" json:"likes"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CommentLike struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID uint   `gorm:"not null;index" json:"comment_id"`
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Type      string `gorm:"size:20;default:like" json:"type"`

	// Relationships
	Comment Comment `gorm:"foreignKey:CommentID" json:"comment"`
	User    User    `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Share struct {
	ID      uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID  uint        `gorm:"not null;index" json:"post_id"`
	UserID  uint        `gorm:"not null;index" json:"user_id"`
	Content string      `gorm:"type:text" json:"content"` // Optional comment when sharing
	Privacy PostPrivacy `gorm:"size:20;default:public" json:"privacy"`

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"post"`
	User User `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PostView struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID       uint   `gorm:"not null;index" json:"post_id"`
	UserID       *uint  `gorm:"index" json:"user_id,omitempty"` // NULL for anonymous views
	IPAddress    string `gorm:"size:45" json:"ip_address"`
	UserAgent    string `gorm:"size:1000" json:"user_agent"`
	ViewDuration int    `json:"view_duration"` // in seconds

	// Relationships
	Post Post  `gorm:"foreignKey:PostID" json:"post"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

type PostReport struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID      uint       `gorm:"not null;index" json:"post_id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	Reason      string     `gorm:"size:100;not null" json:"reason"`
	Details     string     `gorm:"type:text" json:"details"`
	Status      string     `gorm:"size:20;default:pending" json:"status"` // pending, reviewed, resolved, dismissed
	ReviewedBy  *uint      `gorm:"index" json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes string     `gorm:"type:text" json:"review_notes"`

	// Relationships
	Post           Post  `gorm:"foreignKey:PostID" json:"post"`
	User           User  `gorm:"foreignKey:UserID" json:"user"`
	ReviewedByUser *User `gorm:"foreignKey:ReviewedBy" json:"reviewed_by_user,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type SavedPost struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID   uint   `gorm:"not null;index" json:"post_id"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	Category string `gorm:"size:100" json:"category"` // Optional category for organization
	Notes    string `gorm:"type:text" json:"notes"`   // Optional personal notes

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"post"`
	User User `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PostTag struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Slug        string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Color       string `gorm:"size:7" json:"color"` // Hex color code
	UseCount    int    `gorm:"default:0" json:"use_count"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Table names
func (Post) TableName() string {
	return "posts"
}

func (PostMedia) TableName() string {
	return "post_media"
}

func (Like) TableName() string {
	return "likes"
}

func (Comment) TableName() string {
	return "comments"
}

func (CommentLike) TableName() string {
	return "comment_likes"
}

func (Share) TableName() string {
	return "shares"
}

func (PostView) TableName() string {
	return "post_views"
}

func (PostReport) TableName() string {
	return "post_reports"
}

func (SavedPost) TableName() string {
	return "saved_posts"
}

func (PostTag) TableName() string {
	return "post_tags"
}
