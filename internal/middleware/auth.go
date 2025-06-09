package middleware

import (
	"context"
	"fmt"
	"net/http"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"
	"social_server/internal/services"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AuthMiddleware struct {
	authService        *services.AuthService
	sessionRepo        repositories.SessionRepository
	userRepo           repositories.UserRepository
	maxSessionsPerUser int
}

func NewAuthMiddleware(authService *services.AuthService, sessionRepo repositories.SessionRepository, userRepo repositories.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		authService:        authService,
		sessionRepo:        sessionRepo,
		userRepo:           userRepo,
		maxSessionsPerUser: 5, // Default max sessions per user
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := m.extractToken(c)
		if err != nil {
			m.logSecurityEvent(c, 0, postgres.EventInvalidToken, postgres.RiskMedium, "No valid token provided")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "No valid token provided",
			})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := m.sessionRepo.IsTokenBlacklisted(extractTokenID(token))
		if err != nil || isBlacklisted {
			m.logSecurityEvent(c, 0, postgres.EventInvalidToken, postgres.RiskHigh, "Blacklisted token used")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token is no longer valid",
			})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			println(fmt.Sprintf("Token: %s", token))
			m.logSecurityEvent(c, 0, postgres.EventInvalidToken, postgres.RiskMedium, "Invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": fmt.Sprintf("Invalid or expired token: %v", err),
			})
			c.Abort()
			return
		}

		// Validate session
		session, err := m.sessionRepo.GetSessionByTokenID(claims.TokenID)
		if err != nil || !session.IsValidActive() {
			m.logSecurityEvent(c, claims.UserID, postgres.EventInvalidToken, postgres.RiskHigh, "Invalid session")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Session is no longer valid",
			})
			c.Abort()
			return
		}

		// Get user info
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil || !user.IsAccountActive() {
			m.logSecurityEvent(c, claims.UserID, postgres.EventPermissionDenied, postgres.RiskHigh, "Account not active")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Account is not active",
			})
			c.Abort()
			return
		}

		// Check for suspicious activity
		if m.isSuspiciousActivity(c, session) {
			m.logSecurityEvent(c, claims.UserID, postgres.EventSuspiciousActivity, postgres.RiskHigh, "Suspicious activity detected")
			// Don't block but log the event
		}

		// Update session activity
		session.UpdateActivity()
		m.sessionRepo.UpdateSession(session)

		// Set user info in context
		c.Set("user_id", fmt.Sprintf("%d", claims.UserID))
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("session_id", claims.SessionID)
		c.Set("token_id", claims.TokenID)
		c.Set("user", user)

		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := m.extractToken(c)
		if err != nil {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", fmt.Sprintf("%d", claims.UserID))
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("session_id", claims.SessionID)
		c.Set("token_id", claims.TokenID)

		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", fmt.Errorf("authorization header must start with Bearer")
	}

	return strings.TrimPrefix(authHeader, bearerPrefix), nil
}

func GetUserID(c *gin.Context) (uint, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDString, ok := userIDStr.(string)
	if !ok {
		return 0, false
	}

	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(userID), true
}

func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}

	mail, ok := email.(string)
	return mail, ok
}

func GetSessionID(c *gin.Context) (string, bool) {
	sessionID, exists := c.Get("session_id")
	if !exists {
		return "", false
	}

	id, ok := sessionID.(string)
	return id, ok
}

func GetTokenID(c *gin.Context) (string, bool) {
	tokenID, exists := c.Get("token_id")
	if !exists {
		return "", false
	}

	id, ok := tokenID.(string)
	return id, ok
}

// Role-based access control middleware
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userModel, ok := user.(*postgres.User)
		if !ok || !userModel.CanAccess(requiredRole) {
			userID, _ := GetUserID(c)
			m.logSecurityEvent(c, userID, postgres.EventPermissionDenied, postgres.RiskMedium, fmt.Sprintf("Access denied for role: %s", requiredRole))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(postgres.RoleAdmin)
}

func (m *AuthMiddleware) RequireModerator() gin.HandlerFunc {
	return m.RequireRole(postgres.RoleModerator)
}

func (m *AuthMiddleware) RequireSuperUser() gin.HandlerFunc {
	return m.RequireRole(postgres.RoleSuperUser)
}

func RequireUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User authentication required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func SetUserContext(ctx context.Context, c *gin.Context) context.Context {
	if userID, exists := GetUserID(c); exists {
		ctx = context.WithValue(ctx, "user_id", userID)
	}
	if username, exists := GetUsername(c); exists {
		ctx = context.WithValue(ctx, "username", username)
	}
	if email, exists := GetEmail(c); exists {
		ctx = context.WithValue(ctx, "email", email)
	}
	if sessionID, exists := GetSessionID(c); exists {
		ctx = context.WithValue(ctx, "session_id", sessionID)
	}
	if tokenID, exists := GetTokenID(c); exists {
		ctx = context.WithValue(ctx, "token_id", tokenID)
	}
	return ctx
}

// Security helper methods
func (m *AuthMiddleware) isSuspiciousActivity(c *gin.Context, session *postgres.Session) bool {
	currentIP := c.ClientIP()
	currentUserAgent := c.GetHeader("User-Agent")

	return session.IsSuspicious(currentIP, currentUserAgent)
}

func (m *AuthMiddleware) logSecurityEvent(c *gin.Context, userID uint, eventType, risk, details string) {
	log := &postgres.SecurityLog{
		UserID:    userID,
		EventType: eventType,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Details:   details,
		Risk:      risk,
		CreatedAt: time.Now(),
	}

	if sessionID, exists := c.Get("session_id"); exists {
		if sid, ok := sessionID.(string); ok {
			log.SessionID = sid
		}
	}

	m.sessionRepo.LogSecurityEvent(log)
}

func extractTokenID(token string) string {
	// Parse JWT to get the token ID from claims
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		// Fallback to simplified extraction if parsing fails
		if len(token) > 10 {
			return token[:10]
		}
		return token
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if jti, exists := claims["jti"]; exists {
			if tokenID, ok := jti.(string); ok {
				return tokenID
			}
		}
	}

	// Fallback to simplified extraction
	if len(token) > 10 {
		return token[:10]
	}
	return token
}

// Device fingerprinting middleware
func (m *AuthMiddleware) DeviceFingerprint() gin.HandlerFunc {
	return func(c *gin.Context) {
		fingerprint := generateDeviceFingerprint(c)
		c.Set("device_fingerprint", fingerprint)
		c.Next()
	}
}

func generateDeviceFingerprint(c *gin.Context) string {
	userAgent := c.GetHeader("User-Agent")
	acceptLanguage := c.GetHeader("Accept-Language")
	acceptEncoding := c.GetHeader("Accept-Encoding")
	ip := c.ClientIP()

	return fmt.Sprintf("%s|%s|%s|%s", userAgent, acceptLanguage, acceptEncoding, ip)
}

// Rate limiting per user
func (m *AuthMiddleware) UserRateLimit(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.Next()
			return
		}

		// Check recent failed attempts (simple implementation)
		since := time.Now().Add(-time.Minute)
		failedCount, err := m.sessionRepo.CountFailedLogins(userID, since)
		if err == nil && failedCount > int64(requestsPerMinute) {
			m.logSecurityEvent(c, userID, postgres.EventSuspiciousActivity, postgres.RiskHigh, "Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Session limit middleware
func (m *AuthMiddleware) EnforceSessionLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.Next()
			return
		}

		activeCount, err := m.sessionRepo.GetActiveSessionsCount(userID)
		if err == nil && activeCount > int64(m.maxSessionsPerUser) {
			m.logSecurityEvent(c, userID, postgres.EventSuspiciousActivity, postgres.RiskMedium, "Too many active sessions")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "session_limit_exceeded",
				"message": "Too many active sessions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper method to get user model from context
func GetUser(c *gin.Context) (*postgres.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userModel, ok := user.(*postgres.User)
	return userModel, ok
}

// Check if user has specific role
func HasRole(c *gin.Context, role string) bool {
	user, exists := GetUser(c)
	if !exists {
		return false
	}
	return user.HasRole(role)
}

// Check if user is admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, postgres.RoleAdmin)
}

// Check if user is moderator
func IsModerator(c *gin.Context) bool {
	return HasRole(c, postgres.RoleModerator)
}
