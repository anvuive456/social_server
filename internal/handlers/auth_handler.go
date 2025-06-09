package handlers

import (
	"net/http"

	"social_server/internal/middleware"
	"social_server/internal/models/requests"
	"social_server/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	mailService *services.MailService
}

func NewAuthHandler(authService *services.AuthService, mailService *services.MailService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		mailService: mailService,
	}
}

// Register creates a new user account
// @Summary Register new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body requests.RegisterRequest true "Registration data"
// @Success 201 {object} map[string]interface{} "Registration successful with tokens"
// @Failure 400 {object} map[string]interface{} "Invalid request or registration failed"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.Register(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "registration_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "Registration successful",
	})
}

// Login authenticates a user
// @Summary User login
// @Description Authenticate user with email/username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body requests.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with tokens"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "login_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Login successful",
	})
}

// Logout invalidates user session
// @Summary User logout
// @Description Invalidate current user session and tokens
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Logout failed"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	sessionID := c.GetString("session_id")

	err := h.authService.Logout(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "logout_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// RefreshToken generates new access token
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh body requests.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]interface{} "New tokens generated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid or expired refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.RefreshToken(req.RefreshToken, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "refresh_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "Token refreshed successfully",
	})
}

// ChangePassword changes user password
// @Summary Change password
// @Description Change current user's password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body requests.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]interface{} "Password changed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or current password incorrect"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	err := h.authService.ChangePassword(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "change_password_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// ForgotPassword initiates password reset process
// @Summary Forgot password
// @Description Send password reset email to user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email body map[string]string true "Email address"
// @Success 200 {object} map[string]interface{} "Reset email sent"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req requests.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	userAgent := c.GetHeader("User-Agent")

	err := h.authService.ForgotPassword(&req, c.ClientIP(), userAgent)
	if err != nil {
		// Always return success to prevent email enumeration
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, a password reset link has been sent",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, a password reset link has been sent",
	})
}

// ResetPassword resets password using reset token
// @Summary Reset password
// @Description Reset password using token from email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset body requests.ResetPasswordRequest true "Reset password data"
// @Success 200 {object} map[string]interface{} "Password reset successful"
// @Failure 400 {object} map[string]interface{} "Invalid request or token"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req requests.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	err := h.authService.ResetPassword(&req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "reset_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}

// VerifyEmail verifies user email
// @Summary Verify email
// @Description Verify user email address using token
// @Tags Authentication
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]interface{} "Email verified successfully"
// @Failure 400 {object} map[string]interface{} "Invalid or expired token"
// @Router /auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_token",
			"message": "Verification token is required",
		})
		return
	}
	req := requests.VerifyEmailRequest{
		Token: token,
	}

	err := h.authService.VerifyEmail(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "verification_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ResendVerification resends verification email
// @Summary Resend verification email
// @Description Resend email verification link
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email body map[string]string true "Email address"
// @Success 200 {object} map[string]interface{} "Verification email sent"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	err := h.mailService.SendEmailVerification(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "resend_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent",
	})
}

// GetSessions gets user's active sessions
// @Summary Get active sessions
// @Description Get list of user's active login sessions
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Active sessions"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/sessions [get]
//TODO: Implement GetSessions method
// func (h *AuthHandler) GetSessions(c *gin.Context) {
// 	userID, exists := middleware.GetUserID(c)
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error":   "unauthorized",
// 			"message": "User not authenticated",
// 		})
// 		return
// 	}

// 	sessions, err := h.authService.GetUserSessions( userID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "get_sessions_failed",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"data": sessions,
// 	})
// }

// RevokeSession revokes a specific session
// @Summary Revoke session
// @Description Revoke a specific login session
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} map[string]interface{} "Session revoked"
// @Failure 400 {object} map[string]interface{} "Invalid session ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/sessions/{session_id} [delete]
// TODO: Implement session revocation logic
// func (h *AuthHandler) RevokeSession(c *gin.Context) {
// 	userID, exists := middleware.GetUserID(c)
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error":   "unauthorized",
// 			"message": "User not authenticated",
// 		})
// 		return
// 	}

// 	sessionID := c.Param("session_id")
// 	if sessionID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "missing_session_id",
// 			"message": "Session ID is required",
// 		})
// 		return
// 	}

// 	err := h.authService.RevokeSession( userID, sessionID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "revoke_failed",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Session revoked successfully",
// 	})
// }

// RevokeAllSessions revokes all user sessions except current
// @Summary Revoke all sessions
// @Description Revoke all login sessions except current one
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "All sessions revoked"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/sessions/revoke-all [post]
// TODO: Implement RevokeAllSessions
// func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
// 	userID, exists := middleware.GetUserID(c)
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error":   "unauthorized",
// 			"message": "User not authenticated",
// 		})
// 		return
// 	}

// 	currentSessionID := c.GetString("session_id")

// 	err := h.authService.RevokeAllSessions( userID, currentSessionID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "revoke_all_failed",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "All other sessions revoked successfully",
// 	})
// }
