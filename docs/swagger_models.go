package docs

import (
	"time"
)

// ErrorResponse represents a standard error response
// @Description Standard error response format
type ErrorResponse struct {
	Error   string      `json:"error" example:"invalid_request"`
	Message string      `json:"message" example:"The request is invalid"`
	Code    int         `json:"code,omitempty" example:"400"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
// @Description Standard success response format
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationMetadata represents pagination information
// @Description Pagination metadata for list responses
type PaginationMetadata struct {
	Page       int  `json:"page" example:"1"`
	Limit      int  `json:"limit" example:"20"`
	Total      int  `json:"total" example:"100"`
	TotalPages int  `json:"total_pages" example:"5"`
	HasMore    bool `json:"has_more" example:"true"`
}

// ChatRoomListResponse represents paginated chat room list
// @Description Response for getting user's chat rooms
type ChatRoomListResponse struct {
	Rooms      []ChatRoomSummary `json:"rooms"`
	NextCursor *ChatRoomCursor   `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more" example:"true"`
}

// ChatRoomSummary represents a summary view of chat room
// @Description Summary information for chat room in listings
type ChatRoomSummary struct {
	ID               string          `json:"id" example:"507f1f77bcf86cd799439011"`
	Name             string          `json:"name" example:"Project Team"`
	Type             string          `json:"type" example:"group"`
	Avatar           string          `json:"avatar" example:"/uploads/avatars/room.png"`
	ParticipantCount int             `json:"participant_count" example:"5"`
	LastMessage      *MessageSummary `json:"last_message"`
	LastActivity     time.Time       `json:"last_activity"`
	UnreadCount      int             `json:"unread_count" example:"3"`
	IsMuted          bool            `json:"is_muted" example:"false"`
	CreatedAt        time.Time       `json:"created_at"`
}

// MessageSummary represents a summary view of message
// @Description Summary of last message in room listing
type MessageSummary struct {
	ID        string    `json:"id" example:"507f1f77bcf86cd799439020"`
	Content   string    `json:"content" example:"Hello everyone!"`
	Type      string    `json:"type" example:"text"`
	SenderID  string    `json:"sender_id" example:"507f1f77bcf86cd799439012"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRoomCursor represents pagination cursor for chat rooms
// @Description Cursor for chat room pagination
type ChatRoomCursor struct {
	ID           string    `json:"id" example:"507f1f77bcf86cd799439011"`
	LastActivity time.Time `json:"last_activity"`
}

// MessageListResponse represents paginated message list
// @Description Response for getting room messages
type MessageListResponse struct {
	Messages   []MessageDetail `json:"messages"`
	NextCursor *MessageCursor  `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more" example:"true"`
}

// MessageDetail represents detailed message information
// @Description Detailed message with sender information
type MessageDetail struct {
	ID         string            `json:"id" example:"507f1f77bcf86cd799439020"`
	ChatRoomID string            `json:"chat_room_id" example:"507f1f77bcf86cd799439011"`
	SenderID   string            `json:"sender_id" example:"507f1f77bcf86cd799439012"`
	Type       string            `json:"type" example:"text"`
	Content    string            `json:"content" example:"Hello everyone!"`
	Media      *MessageMedia     `json:"media,omitempty"`
	ReplyToID  *string           `json:"reply_to_id,omitempty" example:"507f1f77bcf86cd799439019"`
	EditedAt   *time.Time        `json:"edited_at,omitempty"`
	DeletedAt  *time.Time        `json:"deleted_at,omitempty"`
	ReadBy     []MessageRead     `json:"read_by"`
	Reactions  []MessageReaction `json:"reactions"`
	Mentions   []string          `json:"mentions"`
	Tags       []string          `json:"tags" example:"urgent,meeting"`
	SenderInfo *UserSummary      `json:"sender_info"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// MessageCursor represents pagination cursor for messages
// @Description Cursor for message pagination
type MessageCursor struct {
	ID        string    `json:"id" example:"507f1f77bcf86cd799439020"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageMedia represents media attachment
// @Description Media attachment information
type MessageMedia struct {
	Type      string `json:"type" example:"image"`
	URL       string `json:"url" example:"/uploads/images/photo.jpg"`
	Filename  string `json:"filename" example:"photo.jpg"`
	Size      int64  `json:"size" example:"1024000"`
	MimeType  string `json:"mime_type" example:"image/jpeg"`
	Thumbnail string `json:"thumbnail,omitempty" example:"/uploads/thumbnails/photo_thumb.jpg"`
	Duration  int    `json:"duration,omitempty" example:"0"`
	Width     int    `json:"width,omitempty" example:"1920"`
	Height    int    `json:"height,omitempty" example:"1080"`
}

// MessageRead represents read receipt
// @Description Message read receipt information
type MessageRead struct {
	UserID string    `json:"user_id" example:"507f1f77bcf86cd799439013"`
	ReadAt time.Time `json:"read_at"`
}

// MessageReaction represents message reaction
// @Description Message reaction information
type MessageReaction struct {
	UserID    string    `json:"user_id" example:"507f1f77bcf86cd799439013"`
	Emoji     string    `json:"emoji" example:"👍"`
	ReactedAt time.Time `json:"reacted_at"`
}

// UserSummary represents user summary information
// @Description Summary user information for messages and participants
type UserSummary struct {
	ID          string `json:"id" example:"507f1f77bcf86cd799439012"`
	Username    string `json:"username" example:"john_doe"`
	DisplayName string `json:"display_name" example:"John Doe"`
	Avatar      string `json:"avatar" example:"/uploads/avatars/john.png"`
}

// ParticipantDetail represents detailed participant information
// @Description Detailed participant information with status
type ParticipantDetail struct {
	UserID      string    `json:"user_id" example:"507f1f77bcf86cd799439012"`
	Username    string    `json:"username" example:"john_doe"`
	DisplayName string    `json:"display_name" example:"John Doe"`
	Avatar      string    `json:"avatar" example:"/uploads/avatars/john.png"`
	Role        string    `json:"role" example:"admin"`
	IsOnline    bool      `json:"is_online" example:"true"`
	LastSeen    time.Time `json:"last_seen"`
	JoinedAt    time.Time `json:"joined_at"`
	IsMuted     bool      `json:"is_muted" example:"false"`
	IsBlocked   bool      `json:"is_blocked" example:"false"`
}

// CreateRoomRequest represents room creation request
// @Description Request body for creating a chat room
type CreateRoomRequest struct {
	Name           string           `json:"name" example:"Project Team"`
	Description    string           `json:"description" example:"Discussion about the new project"`
	Type           string           `json:"type" binding:"required" example:"group"`
	Avatar         string           `json:"avatar" example:"/uploads/avatars/room.png"`
	ParticipantIDs []string         `json:"participant_ids"`
	Settings       ChatRoomSettings `json:"settings"`
}

// ChatRoomSettings represents room settings
// @Description Settings and permissions for a chat room
type ChatRoomSettings struct {
	AllowFileSharing    bool `json:"allow_file_sharing" example:"true"`
	AllowImageSharing   bool `json:"allow_image_sharing" example:"true"`
	AllowVideoSharing   bool `json:"allow_video_sharing" example:"true"`
	OnlyAdminsCanPost   bool `json:"only_admins_can_post" example:"false"`
	OnlyAdminsCanInvite bool `json:"only_admins_can_invite" example:"false"`
	MessageEncryption   bool `json:"message_encryption" example:"false"`
}

// CreateRoomResponse represents room creation response
// @Description Response for successful room creation
type CreateRoomResponse struct {
	ID             string           `json:"id" example:"507f1f77bcf86cd799439011"`
	Name           string           `json:"name" example:"Project Team"`
	Description    string           `json:"description" example:"Discussion about the new project"`
	Type           string           `json:"type" example:"group"`
	Avatar         string           `json:"avatar" example:"/uploads/avatars/room.png"`
	CreatedBy      string           `json:"created_by" example:"507f1f77bcf86cd799439012"`
	ParticipantIDs []string         `json:"participant_ids"`
	Settings       ChatRoomSettings `json:"settings"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// SendMessageRequest represents message sending request
// @Description Request body for sending a message
type SendMessageRequest struct {
	Content   string           `json:"content" example:"Hello everyone!"`
	Type      string           `json:"type" example:"text"`
	Media     *MessageMedia    `json:"media,omitempty"`
	ReplyToID *string          `json:"reply_to_id,omitempty" example:"507f1f77bcf86cd799439019"`
	Mentions  []string         `json:"mentions,omitempty"`
	Tags      []string         `json:"tags,omitempty" example:"urgent,meeting"`
	Location  *MessageLocation `json:"location,omitempty"`
}

// MessageLocation represents location information
// @Description Location information for location messages
type MessageLocation struct {
	Latitude  float64 `json:"latitude" example:"37.7749"`
	Longitude float64 `json:"longitude" example:"-122.4194"`
	Address   string  `json:"address,omitempty" example:"San Francisco, CA"`
	PlaceName string  `json:"place_name,omitempty" example:"Union Square"`
}

// SendMessageResponse represents message sending response
// @Description Response for successful message sending
type SendMessageResponse struct {
	ID         string        `json:"id" example:"507f1f77bcf86cd799439020"`
	ChatRoomID string        `json:"chat_room_id" example:"507f1f77bcf86cd799439011"`
	SenderID   string        `json:"sender_id" example:"507f1f77bcf86cd799439012"`
	Type       string        `json:"type" example:"text"`
	Content    string        `json:"content" example:"Hello everyone!"`
	Media      *MessageMedia `json:"media,omitempty"`
	ReplyToID  *string       `json:"reply_to_id,omitempty"`
	Mentions   []string      `json:"mentions,omitempty"`
	Tags       []string      `json:"tags,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// AddReactionRequest represents reaction addition request
// @Description Request body for adding a reaction to a message
type AddReactionRequest struct {
	Emoji string `json:"emoji" binding:"required" example:"👍"`
}

// AddParticipantRequest represents participant addition request
// @Description Request body for adding a participant to a room
type AddParticipantRequest struct {
	UserID string `json:"user_id" binding:"required" example:"507f1f77bcf86cd799439016"`
}

// RoomSearchResponse represents room search results
// @Description Response for room search
type RoomSearchResponse struct {
	Rooms      []ChatRoomSummary `json:"rooms"`
	NextCursor *ChatRoomCursor   `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more" example:"true"`
	Query      string            `json:"query" example:"project"`
	TotalFound int               `json:"total_found" example:"5"`
}

// WebSocketEventData represents WebSocket event data
// @Description Data structure for WebSocket events
type WebSocketEventData struct {
	Type      string      `json:"type" example:"new_message"`
	RoomID    string      `json:"room_id,omitempty" example:"507f1f77bcf86cd799439011"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// TypingEventData represents typing event data
// @Description Data for typing indicator events
type TypingEventData struct {
	UserID   string `json:"user_id" example:"507f1f77bcf86cd799439012"`
	Username string `json:"username" example:"john_doe"`
	IsTyping bool   `json:"is_typing" example:"true"`
	RoomID   string `json:"room_id" example:"507f1f77bcf86cd799439011"`
}

// OnlineStatusData represents online status event data
// @Description Data for user online/offline events
type OnlineStatusData struct {
	UserID    string    `json:"user_id" example:"507f1f77bcf86cd799439012"`
	IsOnline  bool      `json:"is_online" example:"true"`
	LastSeen  time.Time `json:"last_seen"`
	Timestamp time.Time `json:"timestamp"`
}

// MessageOperationResponse represents response for message operations
// @Description Response for message operations like delete, mark as read
type MessageOperationResponse struct {
	Success   bool   `json:"success" example:"true"`
	Message   string `json:"message" example:"Operation completed successfully"`
	MessageID string `json:"message_id,omitempty" example:"507f1f77bcf86cd799439020"`
}

// ParticipantOperationResponse represents response for participant operations
// @Description Response for participant operations like add, remove
type ParticipantOperationResponse struct {
	Success       bool   `json:"success" example:"true"`
	Message       string `json:"message" example:"Participant operation completed"`
	RoomID        string `json:"room_id,omitempty" example:"507f1f77bcf86cd799439011"`
	ParticipantID string `json:"participant_id,omitempty" example:"507f1f77bcf86cd799439016"`
}

// ValidationError represents validation error details
// @Description Detailed validation error information
type ValidationError struct {
	Field   string `json:"field" example:"content"`
	Tag     string `json:"tag" example:"required"`
	Message string `json:"message" example:"Content is required"`
	Value   string `json:"value,omitempty" example:""`
}

// ValidationErrorResponse represents validation error response
// @Description Response for validation errors
type ValidationErrorResponse struct {
	Error   string            `json:"error" example:"validation_failed"`
	Message string            `json:"message" example:"Request validation failed"`
	Details []ValidationError `json:"details"`
}

// RateLimitError represents rate limit error
// @Description Rate limit exceeded error information
type RateLimitError struct {
	Error      string `json:"error" example:"rate_limit_exceeded"`
	Message    string `json:"message" example:"Too many requests"`
	RetryAfter int    `json:"retry_after" example:"60"`
	Limit      int    `json:"limit" example:"100"`
	Window     int    `json:"window" example:"3600"`
}

// UnauthorizedError represents authentication error
// @Description Authentication or authorization error
type UnauthorizedError struct {
	Error   string `json:"error" example:"unauthorized"`
	Message string `json:"message" example:"Authentication required"`
	Code    int    `json:"code" example:"401"`
}

// NotFoundError represents resource not found error
// @Description Resource not found error
type NotFoundError struct {
	Error    string `json:"error" example:"not_found"`
	Message  string `json:"message" example:"Resource not found"`
	Resource string `json:"resource,omitempty" example:"chat_room"`
	ID       string `json:"id,omitempty" example:"507f1f77bcf86cd799439011"`
}

// ForbiddenError represents permission denied error
// @Description Permission denied error
type ForbiddenError struct {
	Error    string `json:"error" example:"forbidden"`
	Message  string `json:"message" example:"Permission denied"`
	Action   string `json:"action,omitempty" example:"send_message"`
	Resource string `json:"resource,omitempty" example:"chat_room"`
}

// InternalServerError represents server error
// @Description Internal server error
type InternalServerError struct {
	Error     string    `json:"error" example:"internal_server_error"`
	Message   string    `json:"message" example:"An internal error occurred"`
	RequestID string    `json:"request_id,omitempty" example:"req_123456789"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheckResponse represents health check response
// @Description Health check response
type HealthCheckResponse struct {
	Status    string    `json:"status" example:"ok"`
	Service   string    `json:"service" example:"social_server"`
	Version   string    `json:"version" example:"2.0.0"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime,omitempty" example:"2h30m15s"`
}

// APIStats represents API statistics
// @Description API usage statistics
type APIStats struct {
	TotalRequests       int64     `json:"total_requests" example:"12345"`
	ActiveConnections   int       `json:"active_connections" example:"25"`
	AverageResponseTime float64   `json:"average_response_time" example:"125.5"`
	ErrorRate           float64   `json:"error_rate" example:"0.05"`
	LastReset           time.Time `json:"last_reset"`
}
