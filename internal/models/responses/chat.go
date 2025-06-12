package responses

import (
	"social_server/internal/models/postgres"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

// type ChatRoomMember struct {
// 	UserID      uint                     `json:"user_id"`
// 	Username    string                   `json:"username"`
// 	DisplayName string                   `json:"display_name"`
// 	Avatar      string                   `json:"avatar"`
// 	Role        postgres.ParticipantRole `json:"role"`
// 	IsOnline    bool                     `json:"is_online"`
// 	LastSeen    time.Time                `json:"last_seen"`
// 	JoinedAt    time.Time                `json:"joined_at"`
// }

type ChatRoomSummary struct {
	ID               uint                  `json:"id"`
	Name             string                `json:"name"`
	Type             postgres.ChatRoomType `json:"type"`
	Avatar           string                `json:"avatar"`
	ParticipantCount int                   `json:"participant_count"`
	LastMessage      *postgres.Message     `json:"last_message"`
	LastActivity     time.Time             `json:"last_activity"`
	UnreadCount      int                   `json:"unread_count"`
	IsMuted          bool                  `json:"is_muted"`
	CreatedAt        time.Time             `json:"created_at"`
}
type ChatRoomsResponse struct {
	Conversations []ChatRoomSummary `json:"rooms"`
	NextCursor    *paginator.Cursor `json:"next_cursor,omitempty"`
}

type MessageSearchResult struct {
	Message    *postgres.Message  `json:"message"`
	ChatRoom   *postgres.ChatRoom `json:"chat_room"`
	Highlights []string           `json:"highlights"`
	Score      float64            `json:"score"`
}

type ChatRoomResponse struct {
	Rooms      []ChatRoomSummary `json:"rooms"`
	NextCursor *paginator.Cursor `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type MessageResponse struct {
	Messages   []postgres.Message `json:"messages"`
	NextCursor *paginator.Cursor  `json:"next_cursor,omitempty"`

	Total int `json:"total"`
}

type NotificationResponse struct {
	Notifications []postgres.ChatNotification `json:"notifications"`
	NextCursor    *paginator.Cursor           `json:"next_cursor,omitempty"`
	HasMore       bool                        `json:"has_more"`
	Total         int                         `json:"total"`
}

type ChatRoomDetailResponse struct {
	Room         postgres.ChatRoom `json:"room"`
	Participants []postgres.User   `json:"participants"`
	MyRole       string            `json:"my_role"`
	UnreadCount  int               `json:"unread_count"`
	IsMuted      bool              `json:"is_muted"`
}

type MessageHistoryResponse struct {
	Messages    []postgres.Message `json:"messages"`
	ChatRoom    ChatRoomSummary    `json:"chat_room"`
	NextCursor  *paginator.Cursor  `json:"next_cursor,omitempty"`
	HasMore     bool               `json:"has_more"`
	UnreadCount int                `json:"unread_count"`
}

type OnlineParticipantsResponse struct {
	Participants []postgres.User `json:"participants"`
	Total        int             `json:"total"`
}

type ChatStatisticsResponse struct {
	TotalRooms         int `json:"total_rooms"`
	TotalMessages      int `json:"total_messages"`
	UnreadMessages     int `json:"unread_messages"`
	ActiveParticipants int `json:"active_participants"`
}

type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type WebSocketEventType string

const (
	WSEventNewMessage       WebSocketEventType = "new_message"
	WSEventMessageRead      WebSocketEventType = "message_read"
	WSEventMessageDelivered WebSocketEventType = "message_delivered"
	WSEventTypingStart      WebSocketEventType = "typing_start"
	WSEventTypingStop       WebSocketEventType = "typing_stop"
	WSEventUserOnline       WebSocketEventType = "user_online"
	WSEventUserOffline      WebSocketEventType = "user_offline"
	WSEventRoomUpdate       WebSocketEventType = "room_update"
	WSEventParticipantJoin  WebSocketEventType = "participant_join"
	WSEventParticipantLeave WebSocketEventType = "participant_leave"
	WSEventMessageReaction  WebSocketEventType = "message_reaction"
)

type TypingResponse struct {
	RoomID    uint      `json:"room_id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

type MessageDeliveryResponse struct {
	MessageID   uint      `json:"message_id"`
	DeliveredTo []uint    `json:"delivered_to"`
	ReadBy      []uint    `json:"read_by"`
	DeliveredAt time.Time `json:"delivered_at"`
}
