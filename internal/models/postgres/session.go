package postgres

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	TokenID     string     `gorm:"uniqueIndex;size:255;not null" json:"token_id"`
	SessionID   string     `gorm:"uniqueIndex;size:255;not null" json:"session_id"`
	RefreshToken string    `gorm:"uniqueIndex;size:2000" json:"refresh_token"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	
	// Device information
	DeviceInfo   string `gorm:"type:text" json:"device_info,omitempty"`
	UserAgent    string `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress    string `gorm:"size:45" json:"ip_address,omitempty"`
	DeviceID     string `gorm:"size:255;index" json:"device_id,omitempty"`
	Platform     string `gorm:"size:50" json:"platform,omitempty"`
	
	// Security tracking
	LoginTime    time.Time  `json:"login_time"`
	LastActivity time.Time  `json:"last_activity"`
	ExpiresAt    time.Time  `json:"expires_at"`
	
	// Location tracking (optional)
	Country      string `gorm:"size:50" json:"country,omitempty"`
	City         string `gorm:"size:100" json:"city,omitempty"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
}

type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TokenID   string    `gorm:"uniqueIndex;size:255;not null" json:"token_id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Reason    string    `gorm:"size:100" json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SecurityLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	SessionID   string    `gorm:"size:255;index" json:"session_id"`
	EventType   string    `gorm:"size:50;not null" json:"event_type"` // login, logout, token_refresh, suspicious_activity
	IPAddress   string    `gorm:"size:45" json:"ip_address"`
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	Details     string    `gorm:"type:text" json:"details"`
	Risk        string    `gorm:"size:20;default:low" json:"risk"` // low, medium, high, critical
	CreatedAt   time.Time `json:"created_at"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Table names
func (Session) TableName() string {
	return "sessions"
}

func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

func (SecurityLog) TableName() string {
	return "security_logs"
}

// Session methods
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValidActive() bool {
	return s.IsActive && !s.IsExpired()
}

func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

func (s *Session) Deactivate() {
	s.IsActive = false
}

// Security methods
func (s *Session) IsSuspicious(currentIP, currentUserAgent string) bool {
	// Basic suspicious activity detection
	if s.IPAddress != "" && s.IPAddress != currentIP {
		return true
	}
	
	if s.UserAgent != "" && s.UserAgent != currentUserAgent {
		return true
	}
	
	// Check for unusual activity patterns
	timeSinceLastActivity := time.Since(s.LastActivity)
	if timeSinceLastActivity > 24*time.Hour {
		return true
	}
	
	return false
}

func (s *Session) GetDeviceFingerprint() string {
	return s.UserAgent + "|" + s.Platform + "|" + s.DeviceID
}

// Security log helper methods
func (sl *SecurityLog) IsHighRisk() bool {
	return sl.Risk == "high" || sl.Risk == "critical"
}

func (sl *SecurityLog) IsCritical() bool {
	return sl.Risk == "critical"
}

// Constants for event types
const (
	EventLogin              = "login"
	EventLogout             = "logout"
	EventTokenRefresh       = "token_refresh"
	EventPasswordChange     = "password_change"
	EventSuspiciousActivity = "suspicious_activity"
	EventAccountLocked      = "account_locked"
	EventPermissionDenied   = "permission_denied"
	EventInvalidToken       = "invalid_token"
)

// Constants for risk levels
const (
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// Constants for blacklist reasons
const (
	BlacklistReasonLogout           = "logout"
	BlacklistReasonPasswordChange   = "password_change"
	BlacklistReasonSuspicious       = "suspicious_activity"
	BlacklistReasonExpired          = "expired"
	BlacklistReasonRevoked          = "revoked"
	BlacklistReasonAccountDisabled  = "account_disabled"
)