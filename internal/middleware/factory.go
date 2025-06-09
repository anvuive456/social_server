package middleware

import (
	"fmt"
	"social_server/internal/repositories"
	"social_server/internal/services"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
)

type MiddlewareFactory struct {
	authService       *services.AuthService
	sessionSecService *services.SessionSecurityService
	sessionRepo       repositories.SessionRepository
	userRepo          repositories.UserRepository
	authMiddleware    *AuthMiddleware
}

type MiddlewareConfig struct {
	MaxSessionsPerUser  int
	RateLimitPerMinute  int
	SessionTimeout      time.Duration
	MaxFailedAttempts   int
	LockoutDuration     time.Duration
	SuspiciousThreshold int
}

func NewMiddlewareFactory(
	authService *services.AuthService,
	sessionSecService *services.SessionSecurityService,
	sessionRepo repositories.SessionRepository,
	userRepo repositories.UserRepository,
	config *MiddlewareConfig,
) *MiddlewareFactory {
	if config == nil {
		config = &MiddlewareConfig{
			MaxSessionsPerUser:  5,
			RateLimitPerMinute:  60,
			SessionTimeout:      24 * time.Hour,
			MaxFailedAttempts:   5,
			LockoutDuration:     15 * time.Minute,
			SuspiciousThreshold: 3,
		}
	}

	authMiddleware := NewAuthMiddleware(authService, sessionRepo, userRepo)
	authMiddleware.maxSessionsPerUser = config.MaxSessionsPerUser

	return &MiddlewareFactory{
		authService:       authService,
		sessionSecService: sessionSecService,
		sessionRepo:       sessionRepo,
		userRepo:          userRepo,
		authMiddleware:    authMiddleware,
	}
}

// Auth middleware methods
func (f *MiddlewareFactory) RequireAuth() gin.HandlerFunc {
	return f.authMiddleware.RequireAuth()
}

func (f *MiddlewareFactory) OptionalAuth() gin.HandlerFunc {
	return f.authMiddleware.OptionalAuth()
}

func (f *MiddlewareFactory) RequireRole(role string) gin.HandlerFunc {
	return f.authMiddleware.RequireRole(role)
}

func (f *MiddlewareFactory) RequireAdmin() gin.HandlerFunc {
	return f.authMiddleware.RequireAdmin()
}

func (f *MiddlewareFactory) RequireModerator() gin.HandlerFunc {
	return f.authMiddleware.RequireModerator()
}

func (f *MiddlewareFactory) RequireSuperUser() gin.HandlerFunc {
	return f.authMiddleware.RequireSuperUser()
}

// Security middleware methods
func (f *MiddlewareFactory) DeviceFingerprint() gin.HandlerFunc {
	return f.authMiddleware.DeviceFingerprint()
}

func (f *MiddlewareFactory) UserRateLimit(requestsPerMinute int) gin.HandlerFunc {
	return f.authMiddleware.UserRateLimit(requestsPerMinute)
}

func (f *MiddlewareFactory) EnforceSessionLimit() gin.HandlerFunc {
	return f.authMiddleware.EnforceSessionLimit()
}

// Middleware chains
func (f *MiddlewareFactory) AuthChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.DeviceFingerprint(),
		f.RequireAuth(),
		f.EnforceSessionLimit(),
	}
}

func (f *MiddlewareFactory) AdminChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.DeviceFingerprint(),
		f.RequireAuth(),
		f.RequireAdmin(),
		f.UserRateLimit(100), // Higher rate limit for admins
	}
}

func (f *MiddlewareFactory) ModeratorChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.DeviceFingerprint(),
		f.RequireAuth(),
		f.RequireModerator(),
		f.UserRateLimit(80), // Higher rate limit for moderators
	}
}

func (f *MiddlewareFactory) PublicChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.DeviceFingerprint(),
		f.OptionalAuth(),
		f.UserRateLimit(80), // Public rate limiting
	}
}

func (f *MiddlewareFactory) SecureChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.DeviceFingerprint(),
		f.RequireAuth(),
		f.EnforceSessionLimit(),
		f.UserRateLimit(60),
	}
}

// Role-based middleware helpers
func (f *MiddlewareFactory) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetUser(c)
		if !exists {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		hasRole := slices.ContainsFunc(roles, user.HasRole)

		if !hasRole {
			userID, _ := GetUserID(c)
			f.authMiddleware.logSecurityEvent(c, userID, "permission_denied", "medium", "Access denied for any of required roles")
			c.JSON(403, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Security monitoring middleware
func (f *MiddlewareFactory) SecurityMonitoring() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Log security events based on response
		userID, _ := GetUserID(c)
		status := c.Writer.Status()
		duration := time.Since(start)

		if status >= 400 {
			risk := "low"
			if status >= 500 {
				risk = "medium"
			}
			if status == 401 || status == 403 {
				risk = "high"
			}

			f.authMiddleware.logSecurityEvent(c, userID, "api_error", risk,
				fmt.Sprintf("HTTP %d - %s - Duration: %v", status, c.Request.URL.Path, duration))
		}
	}
}

// Session cleanup middleware (for background jobs)
func (f *MiddlewareFactory) SessionCleanupJob() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				f.sessionSecService.CleanupExpiredSessions()
			}
		}
	}()
}

// Maintenance mode middleware
func (f *MiddlewareFactory) MaintenanceMode(enabled bool, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled {
			// Allow admin access during maintenance
			if user, exists := GetUser(c); exists && user.IsAdmin() {
				c.Next()
				return
			}

			c.JSON(503, gin.H{
				"error":   "maintenance",
				"message": message,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Account status middleware
func (f *MiddlewareFactory) CheckAccountStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetUser(c)
		if !exists {
			c.Next()
			return
		}

		if !user.IsAccountActive() {
			userID, _ := GetUserID(c)
			f.authMiddleware.logSecurityEvent(c, userID, "account_disabled", "high", "Access attempt with disabled account")

			message := "Account is not active"
			if user.IsBanned {
				if user.BannedUntil != nil && user.BannedUntil.After(time.Now()) {
					message = fmt.Sprintf("Account is temporarily banned until %v", user.BannedUntil.Format("2006-01-02 15:04:05"))
				} else {
					message = "Account is permanently banned"
				}
				if user.BanReason != "" {
					message += fmt.Sprintf(". Reason: %s", user.BanReason)
				}
			}

			c.JSON(403, gin.H{
				"error":   "account_disabled",
				"message": message,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Global instance for convenience
var globalFactory *MiddlewareFactory

func InitializeMiddlewareFactory(
	authService *services.AuthService,
	sessionSecService *services.SessionSecurityService,
	sessionRepo repositories.SessionRepository,
	userRepo repositories.UserRepository,
	config *MiddlewareConfig,
) {
	globalFactory = NewMiddlewareFactory(authService, sessionSecService, sessionRepo, userRepo, config)
}

// Global convenience functions
func Auth() gin.HandlerFunc {
	if globalFactory == nil {
		return func(c *gin.Context) {
			c.JSON(500, gin.H{
				"error":   "middleware_not_initialized",
				"message": "Auth middleware not properly initialized",
			})
			c.Abort()
		}
	}
	return globalFactory.RequireAuth()
}

func OptionalAuth() gin.HandlerFunc {
	if globalFactory == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return globalFactory.OptionalAuth()
}
