package responses

import "time"

type AuthResponse struct {
	AccessToken      string   `json:"access_token"`
	RefreshToken     string   `json:"refresh_token"`
	ExpiresIn        int64    `json:"expires_in"`
	RefreshExpiresIn int64    `json:"refresh_expires_in"`
	TokenType        string   `json:"token_type"`
	User             UserAuth `json:"user"`
}

type UserAuth struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type TokenClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	SessionID string `json:"session_id"`
}

type LoginSessionResponse struct {
	ID           uint      `json:"id"`
	SessionID    string    `json:"session_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceInfo   string    `json:"device_info"`
	Location     string    `json:"location"`
	IsActive     bool      `json:"is_active"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type SessionListResponse struct {
	Sessions []LoginSessionResponse `json:"sessions"`
	Total    int                    `json:"total"`
}

type SecurityEventResponse struct {
	ID          uint      `json:"id"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
}

type SecurityEventListResponse struct {
	Events []SecurityEventResponse `json:"events"`
	Total  int                     `json:"total"`
}

type TwoFactorAuthResponse struct {
	Secret      string    `json:"secret"`
	QRCodeURL   string    `json:"qr_code_url"`
	BackupCodes []string  `json:"backup_codes"`
	IsEnabled   bool      `json:"is_enabled"`
	EnabledAt   time.Time `json:"enabled_at,omitempty"`
}

type VerificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}