package models

import (
	"encoding/json"
	"social_server/internal/models/postgres"
	"time"
)

// WebSocket message types for signaling
type MessageType string

const (
	MessageTypeConnected        MessageType = "connected"
	MessageTypeOffer            MessageType = "offer"
	MessageTypeAnswer           MessageType = "answer"
	MessageTypeICECandidate     MessageType = "ice_candidate"
	MessageTypeJoinRoom         MessageType = "join_room"
	MessageTypeLeaveRoom        MessageType = "leave_room"
	MessageTypeCallRequest      MessageType = "call_request"
	MessageTypeCallResponse     MessageType = "call_response"
	MessageTypeCallEnd          MessageType = "call_end"
	MessageTypeError            MessageType = "error"
	MessageTypeHeartbeat        MessageType = "heartbeat"
	MessageTypeUserJoined       MessageType = "user_joined"
	MessageTypeUserLeft         MessageType = "user_left"
	MessageTypeCallStatus       MessageType = "call_status"
	MessageTypeUserOnlineStatus MessageType = "user_online_status"

	// For chat
	MessageTypeChatCreateRoom     MessageType = "create_room"
	MessageTypeChatSendMessage    MessageType = "send_message"
	MessageTypeChatReceiveMessage MessageType = "receive_message"
)

// Main WebSocket message structure
type WSMessage struct {
	Type      MessageType     `json:"type"`
	From      uint            `json:"from,omitempty"`
	To        uint            `json:"to,omitempty"`
	RoomID    string          `json:"room_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp string          `json:"timestamp"`
}

// WebRTC Offer message
type OfferMessage struct {
	SDP    string `json:"sdp"`
	CallID uint   `json:"call_id"`
	Type   string `json:"type"` // video, audio
}

// WebRTC Answer message
type AnswerMessage struct {
	SDP    string `json:"sdp"`
	CallID uint   `json:"call_id"`
}

// ICE Candidate message
type ICECandidateMessage struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdp_m_line_index"`
	SDPMid        string `json:"sdp_mid"`
	CallID        uint   `json:"call_id"`
}

// Join room message
type JoinRoomMessage struct {
	RoomID   string `json:"room_id"`
	UserID   uint   `json:"user_id"`
	CallType string `json:"call_type"` // video, audio
}

// Leave room message
type LeaveRoomMessage struct {
	RoomID string `json:"room_id"`
	UserID uint   `json:"user_id"`
}

// Call request message
type CallRequestMessage struct {
	CallID   uint   `json:"call_id,omitempty"`
	CallerID uint   `json:"caller_id,omitempty"`
	CalleeID uint   `json:"callee_id"`
	CallType string `json:"call_type"` // video, audio
	RoomID   string `json:"room_id"`
}

// Call response message
type CallResponseMessage struct {
	CallID   uint   `json:"call_id"`
	Response string `json:"response"` // accept, decline
	RoomID   string `json:"room_id"`
}

// Call end message
type CallEndMessage struct {
	CallID uint   `json:"call_id"`
	RoomID string `json:"room_id"`
	Reason string `json:"reason,omitempty"` // hangup, timeout, error
}

// Error message
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Heartbeat message
type HeartbeatMessage struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// User joined message
type UserJoinedMessage struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoomID   string `json:"room_id"`
}

// User left message
type UserLeftMessage struct {
	UserID uint   `json:"user_id"`
	RoomID string `json:"room_id"`
	Reason string `json:"reason,omitempty"`
}

// Call status message
type CallStatusMessage struct {
	CallID uint   `json:"call_id"`
	Status string `json:"status"` // ringing, ongoing, ended
	RoomID string `json:"room_id"`
}

// WebSocket connection info
type WSConnection struct {
	UserID      uint      `json:"user_id"`
	ConnID      string    `json:"conn_id"`
	RoomID      string    `json:"room_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPing    time.Time `json:"last_ping"`
}

// Room info
type Room struct {
	ID           string    `json:"id"`
	CallID       uint      `json:"call_id"`
	Type         string    `json:"type"`   // video, audio
	Status       string    `json:"status"` // waiting, active, ended
	Participants []uint    `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	MaxUsers     int       `json:"max_users"`
}

// STUN/TURN server configuration
type ICEServerConfig struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// WebRTC configuration for client
type WebRTCConfig struct {
	ICEServers []ICEServerConfig `json:"ice_servers"`
}

// User online status message
type UserOnlineStatusMessage struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
}

// Send chat message
type SendChatMessageMessage struct {
	RoomID    uint      `json:"room_id"`
	LocalID   uint      `json:"local_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateChatRoomMessage struct {
	LocalID        uint                  `json:"local_id"`
	Type           postgres.ChatRoomType `json:"type"`
	Name           string                `json:"name"`
	Description    *string               `json:"description"`
	ParticipantIDs []uint                `json:"participant_ids"`
	CreatedAt      time.Time             `json:"created_at"`
}
