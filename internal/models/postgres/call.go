package postgres

import (
	"time"
	"social_server/internal/models/constants"
	"gorm.io/gorm"
)

type Call struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CallerID    uint           `gorm:"not null;index" json:"caller_id"`
	CalleeID    *uint          `gorm:"index" json:"callee_id,omitempty"`
	Type        string         `gorm:"size:20;not null;default:video" json:"type"` // video, audio
	Status      string         `gorm:"size:20;not null;default:pending" json:"status"` // pending, ringing, ongoing, ended, declined, missed
	Duration    int            `gorm:"default:0" json:"duration"` // in seconds
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	EndedAt     *time.Time     `json:"ended_at,omitempty"`
	IsGroupCall bool           `gorm:"default:false" json:"is_group_call"`
	RoomID      string         `gorm:"size:100" json:"room_id,omitempty"`
	
	// WebRTC specific
	OfferSDP    string         `gorm:"type:text" json:"offer_sdp,omitempty"`
	AnswerSDP   string         `gorm:"type:text" json:"answer_sdp,omitempty"`
	
	// Relationships
	Caller       User             `gorm:"foreignKey:CallerID" json:"caller"`
	Callee       *User            `gorm:"foreignKey:CalleeID" json:"callee,omitempty"`
	Participants []CallParticipant `gorm:"foreignKey:CallID" json:"participants"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CallParticipant struct {
	ID       uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CallID   uint           `gorm:"not null;index" json:"call_id"`
	UserID   uint           `gorm:"not null;index" json:"user_id"`
	JoinedAt time.Time      `gorm:"not null" json:"joined_at"`
	LeftAt   *time.Time     `json:"left_at,omitempty"`
	IsActive bool           `gorm:"default:true" json:"is_active"`
	IsMuted  bool           `gorm:"default:false" json:"is_muted"`
	
	// WebRTC specific
	PeerID   string         `gorm:"size:100" json:"peer_id,omitempty"`
	
	// Relationships
	Call User `gorm:"foreignKey:CallID" json:"call"`
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CallType = constants.CallType
type CallStatus = constants.CallStatus

const (
	CallStatusPending  = constants.CallStatusPending
	CallStatusRinging  = constants.CallStatusRinging
	CallStatusOngoing  = constants.CallStatusOngoing
	CallStatusEnded    = constants.CallStatusEnded
	CallStatusDeclined = constants.CallStatusDeclined
	CallStatusMissed   = constants.CallStatusMissed
)

const (
	CallTypeVideo = constants.CallTypeVideo
	CallTypeAudio = constants.CallTypeAudio
)