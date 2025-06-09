package responses

import (
	"social_server/internal/models/postgres"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

// CallResponse represents basic call information
type CallResponse struct {
	ID        uint      `json:"id"`
	CallerID  uint      `json:"caller_id"`
	CalleeID  *uint     `json:"callee_id,omitempty"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	RoomID    string    `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CallDetailResponse represents detailed call information
type CallDetailResponse struct {
	ID           uint                       `json:"id"`
	CallerID     uint                       `json:"caller_id"`
	CalleeID     *uint                      `json:"callee_id,omitempty"`
	Type         string                     `json:"type"`
	Status       string                     `json:"status"`
	Duration     int                        `json:"duration"`
	StartedAt    *time.Time                 `json:"started_at,omitempty"`
	EndedAt      *time.Time                 `json:"ended_at,omitempty"`
	IsGroupCall  bool                       `json:"is_group_call"`
	RoomID       string                     `json:"room_id"`
	Participants []postgres.CallParticipant `json:"participants"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// CallHistoryResponse represents paginated call history
type CallHistoriesResponse struct {
	Calls      []postgres.Call  `json:"calls"`
	NextCursor paginator.Cursor `json:"next_cursor,omitempty"`
}

// ActiveCallsResponse represents user's active calls
type ActiveCallsResponse struct {
	Calls []CallResponse `json:"calls"`
	Total int            `json:"total"`
}

// CallStatsResponse represents call statistics
type CallStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// CallParticipantResponse represents call participant information
type CallParticipantResponse struct {
	ID       uint       `json:"id"`
	UserID   uint       `json:"user_id"`
	Username string     `json:"username"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
	IsActive bool       `json:"is_active"`
	IsMuted  bool       `json:"is_muted"`
	PeerID   string     `json:"peer_id,omitempty"`
}

// GroupCallResponse represents group call information
type GroupCallResponse struct {
	ID           uint                      `json:"id"`
	CreatorID    uint                      `json:"creator_id"`
	Title        string                    `json:"title,omitempty"`
	Type         string                    `json:"type"`
	Status       string                    `json:"status"`
	RoomID       string                    `json:"room_id"`
	Participants []CallParticipantResponse `json:"participants"`
	MaxUsers     int                       `json:"max_users"`
	CreatedAt    time.Time                 `json:"created_at"`
}

// WebRTCConfigResponse represents WebRTC configuration
type WebRTCConfigResponse struct {
	ICEServers []ICEServerResponse `json:"ice_servers"`
}

// ICEServerResponse represents ICE server configuration
type ICEServerResponse struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// CallInvitationResponse represents call invitation information
type CallInvitationResponse struct {
	CallID    uint              `json:"call_id"`
	RoomID    string            `json:"room_id"`
	Caller    UserBasicResponse `json:"caller"`
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

// UserBasicResponse represents basic user information for calls
type UserBasicResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ConnectionStatsResponse represents WebSocket connection statistics
type ConnectionStatsResponse struct {
	TotalConnections int       `json:"total_connections"`
	ActiveRooms      int       `json:"active_rooms"`
	ConnectedUsers   int       `json:"connected_users"`
	Timestamp        time.Time `json:"timestamp"`
}

// CallQualityResponse represents call quality metrics
type CallQualityResponse struct {
	CallID          uint      `json:"call_id"`
	AudioQuality    float64   `json:"audio_quality"`
	VideoQuality    float64   `json:"video_quality"`
	ConnectionType  string    `json:"connection_type"`
	Latency         int       `json:"latency_ms"`
	PacketLoss      float64   `json:"packet_loss_percent"`
	Bandwidth       int       `json:"bandwidth_kbps"`
	LastMeasurement time.Time `json:"last_measurement"`
}

// CallRecordingResponse represents call recording information
type CallRecordingResponse struct {
	ID        uint      `json:"id"`
	CallID    uint      `json:"call_id"`
	URL       string    `json:"url"`
	Duration  int       `json:"duration"`
	FileSize  int64     `json:"file_size"`
	Quality   string    `json:"quality"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RoomInfoResponse represents room information
type RoomInfoResponse struct {
	ID           string                    `json:"id"`
	CallID       uint                      `json:"call_id"`
	Type         string                    `json:"type"`
	Status       string                    `json:"status"`
	Participants []CallParticipantResponse `json:"participants"`
	MaxUsers     int                       `json:"max_users"`
	CreatedAt    time.Time                 `json:"created_at"`
}

// CallEventResponse represents call events for real-time updates
type CallEventResponse struct {
	Type      string      `json:"type"`
	CallID    uint        `json:"call_id"`
	RoomID    string      `json:"room_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type CallStatisticsResponse struct {
	TotalCalls        int `json:"total_calls"`
	TotalDuration     int `json:"total_duration"`
	MissedCalls       int `json:"missed_calls"`
	CompletedCalls    int `json:"completed_calls"`
	AverageCallLength int `json:"average_call_length"`
}

type ActiveCallResponse struct {
	Call         CallResponse              `json:"call"`
	Participants []CallParticipantResponse `json:"participants"`
	MyStatus     string                    `json:"my_status"`
	CanJoin      bool                      `json:"can_join"`
}

type CallInviteResponse struct {
	CallID    uint        `json:"call_id"`
	Caller    UserProfile `json:"caller"`
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id"`
	InvitedAt time.Time   `json:"invited_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

type WebRTCOfferResponse struct {
	CallID    uint      `json:"call_id"`
	OfferSDP  string    `json:"offer_sdp"`
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
}

type WebRTCAnswerResponse struct {
	CallID    uint      `json:"call_id"`
	AnswerSDP string    `json:"answer_sdp"`
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ICECandidateResponse struct {
	CallID    uint      `json:"call_id"`
	Candidate string    `json:"candidate"`
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
}
