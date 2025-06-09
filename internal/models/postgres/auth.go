package postgres

import (
	"time"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Token     string         `gorm:"uniqueIndex;size:2000;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	IsRevoked bool           `gorm:"default:false" json:"is_revoked"`
	DeviceInfo string        `gorm:"size:500" json:"device_info"`
	IPAddress  string        `gorm:"size:45" json:"ip_address"`
	UserAgent  string        `gorm:"size:1000" json:"user_agent"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type LoginSession struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	SessionID  string         `gorm:"uniqueIndex;size:255;not null" json:"session_id"`
	IPAddress  string         `gorm:"size:45;not null" json:"ip_address"`
	UserAgent  string         `gorm:"size:1000" json:"user_agent"`
	DeviceInfo string         `gorm:"size:500" json:"device_info"`
	Location   string         `gorm:"size:255" json:"location"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	LastActivity time.Time    `gorm:"index" json:"last_activity"`
	ExpiresAt  time.Time      `gorm:"not null;index" json:"expires_at"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PasswordReset struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Token     string         `gorm:"uniqueIndex;size:255;not null" json:"token"`
	Email     string         `gorm:"size:100;not null" json:"email"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	IsUsed    bool           `gorm:"default:false" json:"is_used"`
	UsedAt    *time.Time     `json:"used_at,omitempty"`
	IPAddress string         `gorm:"size:45" json:"ip_address"`
	UserAgent string         `gorm:"size:1000" json:"user_agent"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type EmailVerification struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	Token      string         `gorm:"uniqueIndex;size:255;not null" json:"token"`
	Email      string         `gorm:"size:100;not null" json:"email"`
	Type       string         `gorm:"size:50;not null" json:"type"` // registration, email_change
	ExpiresAt  time.Time      `gorm:"not null;index" json:"expires_at"`
	IsVerified bool           `gorm:"default:false" json:"is_verified"`
	VerifiedAt *time.Time     `json:"verified_at,omitempty"`
	IPAddress  string         `gorm:"size:45" json:"ip_address"`
	UserAgent  string         `gorm:"size:1000" json:"user_agent"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type LoginAttempt struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"size:100;not null;index" json:"email"`
	IPAddress    string    `gorm:"size:45;not null;index" json:"ip_address"`
	UserAgent    string    `gorm:"size:1000" json:"user_agent"`
	IsSuccessful bool      `gorm:"default:false" json:"is_successful"`
	FailureReason string   `gorm:"size:255" json:"failure_reason"`
	AttemptedAt  time.Time `gorm:"not null;index" json:"attempted_at"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

type RateLimitEntry struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Key        string    `gorm:"uniqueIndex;size:255;not null" json:"key"` // IP:endpoint or user_id:endpoint
	Count      int       `gorm:"not null;default:0" json:"count"`
	WindowStart time.Time `gorm:"not null;index" json:"window_start"`
	ExpiresAt  time.Time `gorm:"not null;index" json:"expires_at"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SecurityEvent struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      *uint     `gorm:"index" json:"user_id,omitempty"`
	EventType   string    `gorm:"size:100;not null;index" json:"event_type"`
	Description string    `gorm:"type:text" json:"description"`
	IPAddress   string    `gorm:"size:45;not null" json:"ip_address"`
	UserAgent   string    `gorm:"size:1000" json:"user_agent"`
	Severity    string    `gorm:"size:20;default:info" json:"severity"` // info, warning, critical
	Metadata    string    `gorm:"type:text" json:"metadata"` // JSON as string
	
	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

type TwoFactorAuth struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret       string         `gorm:"size:255;not null" json:"secret"`
	IsEnabled    bool           `gorm:"default:false" json:"is_enabled"`
	BackupCodes  string         `gorm:"type:text" json:"backup_codes"` // JSON array as string
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty"`
	EnabledAt    *time.Time     `json:"enabled_at,omitempty"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}



// Table names
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (LoginSession) TableName() string {
	return "login_sessions"
}

func (PasswordReset) TableName() string {
	return "password_resets"
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

func (RateLimitEntry) TableName() string {
	return "rate_limit_entries"
}

func (SecurityEvent) TableName() string {
	return "security_events"
}

func (TwoFactorAuth) TableName() string {
	return "two_factor_auths"
}