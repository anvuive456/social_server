package requests

import "time"

// InitiateCallRequest represents a request to initiate a call
type InitiateCallRequest struct {
	CalleeID uint   `json:"callee_id" binding:"required" validate:"min=1"`
	Type     string `json:"type" binding:"required,oneof=video audio"`
}

// CallResponseRequest represents a response to a call request
type CallResponseRequest struct {
	Response string `json:"response" binding:"required,oneof=accept decline"`
}

// JoinCallRequest represents a request to join a call
type JoinCallRequest struct {
	RoomID string `json:"room_id" binding:"required"`
}

// UpdateCallRequest represents a request to update call settings
type UpdateCallRequest struct {
	IsMuted bool `json:"is_muted,omitempty"`
}

// CallHistoryQueryRequest represents query parameters for call history
type CallHistoryQueryRequest struct {
	Before    string `form:"before,omitempty"`
	After     string `form:"after,omitempty"`
	Limit     int    `form:"limit" binding:"min=1,max=100"`
	Status    string `form:"status" binding:"omitempty,oneof=pending ringing ongoing ended declined missed"`
	Type      string `form:"type" binding:"omitempty,oneof=video audio"`
	StartDate string `form:"start_date,omitempty"` // Format: 2006-01-02
	EndDate   string `form:"end_date,omitempty"`   // Format: 2006-01-02
}

// WebRTCSignalingRequest represents WebRTC signaling data
type WebRTCSignalingRequest struct {
	Type      string      `json:"type" binding:"required,oneof=offer answer ice_candidate"`
	SDP       string      `json:"sdp,omitempty"`
	Candidate string      `json:"candidate,omitempty"`
	CallID    uint        `json:"call_id" binding:"required"`
	Data      interface{} `json:"data,omitempty"`
}

// GroupCallRequest represents a request to create a group call
type GroupCallRequest struct {
	ParticipantIDs []uint `json:"participant_ids" binding:"required,min=1,max=10"`
	Type           string `json:"type" binding:"required,oneof=video audio"`
	Title          string `json:"title,omitempty" validate:"max=100"`
}

// InviteToCallRequest represents a request to invite someone to an existing call
type InviteToCallRequest struct {
	UserID uint `json:"user_id" binding:"required,min=1"`
}

// CallSettingsRequest represents call settings update
type CallSettingsRequest struct {
	VideoEnabled bool `json:"video_enabled,omitempty"`
	AudioEnabled bool `json:"audio_enabled,omitempty"`
	IsMuted      bool `json:"is_muted,omitempty"`
}

// StartRecordingRequest represents a request to start call recording
type StartRecordingRequest struct {
	Quality string `json:"quality,omitempty" binding:"omitempty,oneof=low medium high"`
}

// ParsedCallHistoryQuery represents parsed query parameters
type ParsedCallHistoryQuery struct {
	Cursor    string
	Limit     int
	Status    string
	Type      string
	StartDate *time.Time
	EndDate   *time.Time
}
