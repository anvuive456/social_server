package services

import (
	"crypto/md5"
	"fmt"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SessionSecurityService struct {
	sessionRepo repositories.SessionRepository
	userRepo    repositories.UserRepository

	// Security settings
	maxFailedAttempts   int
	lockoutDuration     time.Duration
	sessionTimeout      time.Duration
	maxSessionsPerUser  int
	suspiciousThreshold int
}

type SecurityConfig struct {
	MaxFailedAttempts   int
	LockoutDuration     time.Duration
	SessionTimeout      time.Duration
	MaxSessionsPerUser  int
	SuspiciousThreshold int
}

func NewSessionSecurityService(sessionRepo repositories.SessionRepository, userRepo repositories.UserRepository, config *SecurityConfig) *SessionSecurityService {
	if config == nil {
		config = &SecurityConfig{
			MaxFailedAttempts:   5,
			LockoutDuration:     15 * time.Minute,
			SessionTimeout:      24 * time.Hour,
			MaxSessionsPerUser:  5,
			SuspiciousThreshold: 3,
		}
	}

	return &SessionSecurityService{
		sessionRepo:         sessionRepo,
		userRepo:            userRepo,
		maxFailedAttempts:   config.MaxFailedAttempts,
		lockoutDuration:     config.LockoutDuration,
		sessionTimeout:      config.SessionTimeout,
		maxSessionsPerUser:  config.MaxSessionsPerUser,
		suspiciousThreshold: config.SuspiciousThreshold,
	}
}

// Session Management
func (s *SessionSecurityService) CreateSecureSession(userID uint, tokenID, sessionID string, c *gin.Context) (*postgres.Session, error) {
	// Check session limits
	activeCount, err := s.sessionRepo.GetActiveSessionsCount(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active sessions: %w", err)
	}

	if activeCount >= int64(s.maxSessionsPerUser) {
		// Deactivate oldest session
		sessions, err := s.sessionRepo.GetUserSessions(userID)
		if err == nil && len(sessions) > 0 {
			oldestSession := sessions[len(sessions)-1]
			s.sessionRepo.DeactivateSession(oldestSession.TokenID)
			s.LogSecurityEvent(userID, oldestSession.SessionID, postgres.EventLogout, postgres.RiskLow, "Session terminated due to limit")
		}
	}

	session := &postgres.Session{
		UserID:       userID,
		TokenID:      tokenID,
		SessionID:    sessionID,
		IsActive:     true,
		DeviceInfo:   s.extractDeviceInfo(c),
		UserAgent:    c.GetHeader("User-Agent"),
		IPAddress:    c.ClientIP(),
		DeviceID:     s.generateDeviceID(c),
		Platform:     s.extractPlatform(c.GetHeader("User-Agent")),
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(s.sessionTimeout),
		Country:      s.extractCountry(c.ClientIP()),
		City:         s.extractCity(c.ClientIP()),
	}

	err = s.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.LogSecurityEvent(userID, sessionID, postgres.EventLogin, postgres.RiskLow, "Session created successfully")

	return session, nil
}

func (s *SessionSecurityService) ValidateSession(tokenID string, c *gin.Context) (*postgres.Session, error) {
	// Check if token is blacklisted
	isBlacklisted, err := s.sessionRepo.IsTokenBlacklisted(tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}

	if isBlacklisted {
		s.LogSecurityEvent(0, "", postgres.EventInvalidToken, postgres.RiskHigh, "Blacklisted token used")
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Get session
	session, err := s.sessionRepo.GetSessionByTokenID(tokenID)
	if err != nil {
		s.LogSecurityEvent(0, "", postgres.EventInvalidToken, postgres.RiskMedium, "Invalid session token")
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	if !session.IsValidActive() {
		s.LogSecurityEvent(session.UserID, session.SessionID, postgres.EventInvalidToken, postgres.RiskMedium, "Expired session used")
		return nil, fmt.Errorf("session expired or inactive")
	}

	// Check for suspicious activity
	if s.isSuspiciousActivity(session, c) {
		s.LogSecurityEvent(session.UserID, session.SessionID, postgres.EventSuspiciousActivity, postgres.RiskHigh, "Suspicious activity detected")

		// Count recent suspicious activities
		since := time.Now().Add(-1 * time.Hour)
		suspiciousLogs, err := s.sessionRepo.GetSuspiciousActivities(session.UserID, since)
		if err == nil && len(suspiciousLogs) >= s.suspiciousThreshold {
			// Auto-logout user due to suspicious activity
			s.TerminateSession(tokenID, "suspicious_activity")
			return nil, fmt.Errorf("session terminated due to suspicious activity")
		}
	}

	// Update session activity
	session.LastActivity = time.Now()
	if c != nil && session.IPAddress != c.ClientIP() {
		session.IPAddress = c.ClientIP()
	}

	err = s.sessionRepo.UpdateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

func (s *SessionSecurityService) TerminateSession(tokenID, reason string) error {
	session, err := s.sessionRepo.GetSessionByTokenID(tokenID)
	if err != nil {
		return err
	}

	// Deactivate session
	err = s.sessionRepo.DeactivateSession(tokenID)
	if err != nil {
		return err
	}

	// Blacklist token
	expiresAt := time.Now().Add(24 * time.Hour) // Keep in blacklist for 24 hours
	err = s.sessionRepo.BlacklistToken(tokenID, session.UserID, reason, expiresAt)
	if err != nil {
		return err
	}

	s.LogSecurityEvent(session.UserID, session.SessionID, postgres.EventLogout, postgres.RiskLow, fmt.Sprintf("Session terminated: %s", reason))

	return nil
}

func (s *SessionSecurityService) TerminateAllUserSessions(userID uint, reason string) error {
	sessions, err := s.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return err
	}

	// Deactivate all sessions
	err = s.sessionRepo.DeactivateUserSessions(userID)
	if err != nil {
		return err
	}

	// Blacklist all tokens
	expiresAt := time.Now().Add(24 * time.Hour)
	for _, session := range sessions {
		s.sessionRepo.BlacklistToken(session.TokenID, userID, reason, expiresAt)
	}

	s.LogSecurityEvent(userID, "", postgres.EventLogout, postgres.RiskMedium, fmt.Sprintf("All sessions terminated: %s", reason))

	return nil
}

// Security Analysis
func (s *SessionSecurityService) AnalyzeLoginAttempt(userID uint, success bool, c *gin.Context) error {
	risk := postgres.RiskLow
	details := "Login successful"

	if !success {
		risk = postgres.RiskMedium
		details = "Login failed"

		// Check failed login count
		since := time.Now().Add(-15 * time.Minute)
		failedCount, err := s.sessionRepo.CountFailedLogins(userID, since)
		if err == nil && failedCount >= int64(s.maxFailedAttempts-1) {
			risk = postgres.RiskHigh
			details = "Multiple failed login attempts detected"

			// Temporarily lock account or take other action
			if failedCount >= int64(s.maxFailedAttempts) {
				s.LogSecurityEvent(userID, "", postgres.EventAccountLocked, postgres.RiskCritical, "Account locked due to multiple failed attempts")
			}
		}
	}

	return s.LogSecurityEvent(userID, "", postgres.EventLogin, risk, details)
}

func (s *SessionSecurityService) CheckAccountSecurity(userID uint) (*SecurityReport, error) {
	// Get recent security logs
	logs, err := s.sessionRepo.GetUserSecurityLogs(userID, 50)
	if err != nil {
		return nil, err
	}

	// Get active sessions
	sessions, err := s.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return nil, err
	}

	// Analyze security
	report := &SecurityReport{
		UserID:         userID,
		ActiveSessions: len(sessions),
		RecentLogs:     len(logs),
		RiskLevel:      postgres.RiskLow,
		GeneratedAt:    time.Now(),
	}

	// Count high-risk events in last 24 hours
	highRiskCount := 0
	since := time.Now().Add(-24 * time.Hour)

	for _, log := range logs {
		if log.CreatedAt.After(since) && log.IsHighRisk() {
			highRiskCount++
		}
	}

	if highRiskCount > 3 {
		report.RiskLevel = postgres.RiskHigh
		report.Recommendations = append(report.Recommendations, "Consider changing password due to suspicious activity")
	} else if highRiskCount > 1 {
		report.RiskLevel = postgres.RiskMedium
		report.Recommendations = append(report.Recommendations, "Monitor account activity closely")
	}

	if len(sessions) > 3 {
		report.Recommendations = append(report.Recommendations, "Consider terminating unused sessions")
	}

	return report, nil
}

// Security Logging
func (s *SessionSecurityService) LogSecurityEvent(userID uint, sessionID, eventType, risk, details string) error {
	log := &postgres.SecurityLog{
		UserID:    userID,
		SessionID: sessionID,
		EventType: eventType,
		Risk:      risk,
		Details:   details,
		CreatedAt: time.Now(),
	}

	return s.sessionRepo.LogSecurityEvent(log)
}

// Cleanup Operations
func (s *SessionSecurityService) CleanupExpiredSessions() error {
	err := s.sessionRepo.DeleteExpiredSessions()
	if err != nil {
		return err
	}

	return s.sessionRepo.CleanupExpiredBlacklist()
}

// Helper methods
func (s *SessionSecurityService) isSuspiciousActivity(session *postgres.Session, c *gin.Context) bool {
	if c == nil {
		return false
	}

	currentIP := c.ClientIP()
	currentUserAgent := c.GetHeader("User-Agent")

	return session.IsSuspicious(currentIP, currentUserAgent)
}

func (s *SessionSecurityService) generateDeviceID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	userAgent := c.GetHeader("User-Agent")
	acceptLanguage := c.GetHeader("Accept-Language")
	acceptEncoding := c.GetHeader("Accept-Encoding")

	data := fmt.Sprintf("%s|%s|%s", userAgent, acceptLanguage, acceptEncoding)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func (s *SessionSecurityService) extractDeviceInfo(c *gin.Context) string {
	if c == nil {
		return "Unknown Device"
	}

	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return "Unknown Device"
	}

	// Simple device detection
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "mobile"):
		return "Mobile Device"
	case strings.Contains(ua, "tablet"):
		return "Tablet"
	case strings.Contains(ua, "windows"):
		return "Windows Computer"
	case strings.Contains(ua, "mac"):
		return "Mac Computer"
	case strings.Contains(ua, "linux"):
		return "Linux Computer"
	default:
		return "Unknown Device"
	}
}

func (s *SessionSecurityService) extractPlatform(userAgent string) string {
	if userAgent == "" {
		return "Unknown"
	}

	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac"):
		return "MacOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "ios"):
		return "iOS"
	default:
		return "Unknown"
	}
}

func (s *SessionSecurityService) extractCountry(ip string) string {
	// This is a placeholder. In production, you'd use a GeoIP service
	return "Unknown"
}

func (s *SessionSecurityService) extractCity(ip string) string {
	// This is a placeholder. In production, you'd use a GeoIP service
	return "Unknown"
}

// Data structures
type SecurityReport struct {
	UserID          uint      `json:"user_id"`
	ActiveSessions  int       `json:"active_sessions"`
	RecentLogs      int       `json:"recent_logs"`
	RiskLevel       string    `json:"risk_level"`
	Recommendations []string  `json:"recommendations"`
	GeneratedAt     time.Time `json:"generated_at"`
}
