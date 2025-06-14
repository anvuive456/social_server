package requests

import "social_server/internal/models/constants"

type ChatRoomType = constants.ChatRoomType
type MessageType = constants.MessageType

const (
	ChatRoomTypePrivate = constants.ChatRoomTypePrivate
	ChatRoomTypeGroup   = constants.ChatRoomTypeGroup
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

type CreateChatRoomRequest struct {
	Name         string       `form:"name,omitempty"`
	Description  string       `form:"description,omitempty"`
	Type         ChatRoomType `form:"type" binding:"required"`
	Avatar       string       `form:"avatar,omitempty"`
	Participants []uint       `form:"participants,omitempty"`
}

type GetChatRoomsRequest struct {
	Limit   int    `form:"limit,omitempty"`
	Archive bool   `form:"archive,omitempty"`
	Before  string `form:"before,omitempty"`
	After   string `form:"after,omitempty"`
}

type DeleteChatRoomRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type UpdateChatRoomRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

type SendMessageRequest struct {
	Type      MessageType                 `json:"type" binding:"required"`
	Content   string                      `json:"content"`
	ReplyToID *uint                       `json:"reply_to_id,omitempty"`
	Mentions  []uint                      `json:"mentions,omitempty"`
	Media     *SendMessageMediaRequest    `json:"media,omitempty"`
	Location  *SendMessageLocationRequest `json:"location,omitempty"`
}

type SendMessageMediaRequest struct {
	Type      string `json:"type" binding:"required"`
	URL       string `json:"url" binding:"required"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

type SendMessageLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   string  `json:"address,omitempty"`
	PlaceName string  `json:"place_name,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type InviteToChatRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
	Message string `json:"message,omitempty"`
}

type JoinChatRoomRequest struct {
	InviteCode string `json:"invite_code,omitempty"`
}

type UpdateParticipantRequest struct {
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Nickname    string   `json:"nickname,omitempty"`
}

type MessageReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

type TypingIndicatorRequest struct {
	IsTyping bool `json:"is_typing"`
}

type MarkMessagesReadRequest struct {
	MessageIDs []uint `json:"message_ids" binding:"required"`
}

type ChatRoomSettingsRequest struct {
	AllowFileSharing    *bool `json:"allow_file_sharing,omitempty"`
	AllowImageSharing   *bool `json:"allow_image_sharing,omitempty"`
	AllowVideoSharing   *bool `json:"allow_video_sharing,omitempty"`
	OnlyAdminsCanPost   *bool `json:"only_admins_can_post,omitempty"`
	OnlyAdminsCanInvite *bool `json:"only_admins_can_invite,omitempty"`
	MessageEncryption   *bool `json:"message_encryption,omitempty"`
}

type SyncChatRoomsRequest struct {
	LastID *uint `form:"last_id"`
}
