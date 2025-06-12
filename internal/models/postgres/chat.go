package postgres

import (
	"social_server/internal/models/constants"
	"time"

	"gorm.io/gorm"
)

type ChatRoomType = constants.ChatRoomType
type ParticipantRole = constants.ParticipantRole
type MessageType = constants.MessageType
type MessageStatus = constants.MessageStatus

const (
	ChatRoomTypePrivate = constants.ChatRoomTypePrivate
	ChatRoomTypeGroup   = constants.ChatRoomTypeGroup
)

const (
	ParticipantRoleMember = constants.ParticipantRoleMember
	ParticipantRoleAdmin  = constants.ParticipantRoleAdmin
	ParticipantRoleOwner  = constants.ParticipantRoleOwner
)

const (
	MessageTypeText     = constants.MessageTypeText
	MessageTypeImage    = constants.MessageTypeImage
	MessageTypeVideo    = constants.MessageTypeVideo
	MessageTypeAudio    = constants.MessageTypeAudio
	MessageTypeFile     = constants.MessageTypeFile
	MessageTypeSystem   = constants.MessageTypeSystem
	MessageTypeLocation = constants.MessageTypeLocation
)

const (
	MessageStatusSent      = constants.MessageStatusSent
	MessageStatusDelivered = constants.MessageStatusDelivered
	MessageStatusRead      = constants.MessageStatusRead
	MessageStatusFailed    = constants.MessageStatusFailed
)

type ChatRoom struct {
	ID           uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string       `gorm:"size:100" json:"name"`
	Description  string       `gorm:"type:text" json:"description"`
	Type         ChatRoomType `gorm:"size:20;not null" json:"type"`
	Avatar       string       `gorm:"size:500" json:"avatar"`
	CreatedBy    uint         `gorm:"not null;index" json:"created_by"`
	IsArchived   bool         `gorm:"default:false" json:"is_archived"`
	LastActivity *time.Time   `gorm:"index" json:"last_activity"`

	// Embedded settings
	Settings ChatRoomSettings `gorm:"embedded;embeddedPrefix:settings_" json:"settings"`

	// Relationships
	Creator      User          `gorm:"foreignKey:CreatedBy" json:"creator"`
	Participants []Participant `gorm:"foreignKey:ChatRoomID" json:"participants"`
	Messages     []Message     `gorm:"foreignKey:ChatRoomID" json:"messages"`

	// Timestamps
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"-"`
}

type ChatRoomSettings struct {
	AllowFileSharing    bool `gorm:"default:true" json:"allow_file_sharing"`
	AllowImageSharing   bool `gorm:"default:true" json:"allow_image_sharing"`
	AllowVideoSharing   bool `gorm:"default:true" json:"allow_video_sharing"`
	OnlyAdminsCanPost   bool `gorm:"default:false" json:"only_admins_can_post"`
	OnlyAdminsCanInvite bool `gorm:"default:false" json:"only_admins_can_invite"`
	MessageEncryption   bool `gorm:"default:false" json:"message_encryption"`
}

type Participant struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatRoomID  uint            `gorm:"not null;index" json:"chat_room_id"`
	UserID      uint            `gorm:"not null;index" json:"user_id"`
	Role        ParticipantRole `gorm:"size:20;default:member" json:"role"`
	JoinedAt    time.Time       `json:"joined_at"`
	LastReadAt  time.Time       `json:"last_read_at"`
	IsMuted     bool            `gorm:"default:false" json:"is_muted"`
	IsBlocked   bool            `gorm:"default:false" json:"is_blocked"`
	Nickname    string          `gorm:"size:100" json:"nickname"`
	Permissions string          `gorm:"type:text" json:"permissions"` // JSON array as string

	// Relationships
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	User     User     `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Message struct {
	ID               uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatRoomID       uint        `gorm:"not null;index" json:"chat_room_id"`
	SenderID         uint        `gorm:"not null;index" json:"sender_id"`
	Type             MessageType `gorm:"size:20;not null" json:"type"`
	Content          string      `gorm:"type:text" json:"content"`
	EncryptedContent string      `gorm:"type:text" json:"encrypted_content,omitempty"`
	ReplyToID        *uint       `gorm:"index" json:"reply_to_id,omitempty"`
	ForwardedFromID  *uint       `gorm:"index" json:"forwarded_from_id,omitempty"`
	EditedAt         *time.Time  `json:"edited_at,omitempty"`
	DeliveryStatus   string      `gorm:"type:text" json:"delivery_status"` // JSON as string
	Mentions         string      `gorm:"type:text" json:"mentions"`        // JSON array as string
	Tags             string      `gorm:"type:text" json:"tags"`            // JSON array as string

	// Embedded media and location
	Media    *MessageMedia    `gorm:"embedded;embeddedPrefix:media_" json:"media,omitempty"`
	Location *MessageLocation `gorm:"embedded;embeddedPrefix:location_" json:"location,omitempty"`

	// Relationships
	ChatRoom      ChatRoom          `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	Sender        User              `gorm:"foreignKey:SenderID" json:"sender"`
	ReplyTo       *Message          `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	ForwardedFrom *Message          `gorm:"foreignKey:ForwardedFromID" json:"forwarded_from,omitempty"`
	ReadBy        []MessageRead     `gorm:"foreignKey:MessageID" json:"read_by"`
	Reactions     []MessageReaction `gorm:"foreignKey:MessageID" json:"reactions"`

	// Timestamps
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type MessageMedia struct {
	Type      string `gorm:"size:50" json:"type"`
	URL       string `gorm:"size:500" json:"url"`
	Filename  string `gorm:"size:255" json:"filename"`
	Size      int64  `json:"size"`
	MimeType  string `gorm:"size:100" json:"mime_type"`
	Thumbnail string `gorm:"size:500" json:"thumbnail,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

type MessageLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `gorm:"size:500" json:"address,omitempty"`
	PlaceName string  `gorm:"size:255" json:"place_name,omitempty"`
}

type MessageRead struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ReadAt    time.Time `json:"read_at"`

	// Relationships
	Message Message `gorm:"foreignKey:MessageID" json:"message"`
	User    User    `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

type MessageReaction struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Emoji     string    `gorm:"size:50;not null" json:"emoji"`
	ReactedAt time.Time `json:"reacted_at"`

	// Relationships
	Message Message `gorm:"foreignKey:MessageID" json:"message"`
	User    User    `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

type TypingIndicator struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatRoomID uint      `gorm:"not null;index" json:"chat_room_id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	IsTyping   bool      `gorm:"default:false" json:"is_typing"`
	LastTyped  time.Time `json:"last_typed"`

	// Relationships
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	User     User     `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OnlineStatus struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	IsOnline   bool      `gorm:"default:false" json:"is_online"`
	LastSeen   time.Time `json:"last_seen"`
	Status     string    `gorm:"size:20;default:offline" json:"status"`
	DeviceInfo string    `gorm:"size:255" json:"device_info"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatInvite struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatRoomID uint      `gorm:"not null;index" json:"chat_room_id"`
	InviterID  uint      `gorm:"not null;index" json:"inviter_id"`
	InviteeID  uint      `gorm:"not null;index" json:"invitee_id"`
	Status     string    `gorm:"size:20;default:pending" json:"status"`
	Message    string    `gorm:"type:text" json:"message"`
	ExpiresAt  time.Time `gorm:"index" json:"expires_at"`

	// Relationships
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	Inviter  User     `gorm:"foreignKey:InviterID" json:"inviter"`
	Invitee  User     `gorm:"foreignKey:InviteeID" json:"invitee"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ChatNotification struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	ChatRoomID uint       `gorm:"not null;index" json:"chat_room_id"`
	MessageID  *uint      `gorm:"index" json:"message_id,omitempty"`
	Type       string     `gorm:"size:50;not null" json:"type"`
	Title      string     `gorm:"size:255" json:"title"`
	Content    string     `gorm:"type:text" json:"content"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	Data       string     `gorm:"type:text" json:"data,omitempty"` // JSON as string
	ReadAt     *time.Time `json:"read_at,omitempty"`

	// Relationships
	User     User     `gorm:"foreignKey:UserID" json:"user"`
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	Message  *Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Table names
func (ChatRoom) TableName() string {
	return "chat_rooms"
}

func (Participant) TableName() string {
	return "participants"
}

func (Message) TableName() string {
	return "messages"
}

func (MessageRead) TableName() string {
	return "message_reads"
}

func (MessageReaction) TableName() string {
	return "message_reactions"
}

func (TypingIndicator) TableName() string {
	return "typing_indicators"
}

func (OnlineStatus) TableName() string {
	return "online_statuses"
}

func (ChatInvite) TableName() string {
	return "chat_invites"
}

func (ChatNotification) TableName() string {
	return "chat_notifications"
}
