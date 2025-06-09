package handlers

import (
	"net/http"
	"strconv"

	"social_server/internal/middleware"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/services"

	"slices"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// SearchUsers searches for users
// @Summary Search users
// @Description Search for users by username, display name, or email
// @Tags Users
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Number of users to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	userId, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}
	query := requests.UserSearchRequest{}
	err := c.BindQuery(&query)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	response, err := h.userService.SearchUsers(userId, query.Search, query.Limit, query.Before, query.After)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// UploadAvatar uploads user avatar
// @Summary Upload avatar
// @Description Upload avatar image for current user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} map[string]interface{} "Upload success with avatar URL"
// @Failure 400 {object} map[string]interface{} "Invalid file"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file",
			"message": "Avatar file is required",
		})
		return
	}
	defer file.Close()

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "file_too_large",
			"message": "File size must be less than 5MB",
		})
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	isValidType := slices.Contains(allowedTypes, contentType)

	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file_type",
			"message": "Only JPEG, PNG, GIF, and WebP images are allowed",
		})
		return
	}

	avatarURL, err := h.userService.UploadAvatar(userID, file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "upload_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"avatar_url": avatarURL,
		},
		"message": "Avatar uploaded successfully",
	})
}

// GetUserStats gets user statistics
// @Summary Get user stats
// @Description Get user statistics like post count, friend count, etc.
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User statistics"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id}/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})
		return
	}

	stats, err := h.userService.GetUserStats(uint(userID))
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "get_stats_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// UpdateOnlineStatus updates user's online status
// @Summary Update online status
// @Description Update current user's online status
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status body map[string]bool true "Online status"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/online-status [put]
func (h *UserHandler) UpdateOnlineStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		IsOnline bool `json:"is_online" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	err := h.userService.UpdateOnlineStatus(userID, req.IsOnline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_status_failed",
			"message": err.Error(),
		})
		return
	}

	status := "offline"
	if req.IsOnline {
		status = "online"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Online status updated to " + status,
		"data": gin.H{
			"is_online": req.IsOnline,
		},
	})
}

// UpdateSettings updates user settings
// @Summary Update user settings
// @Description Update current user's privacy and notification settings
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body postgres.UserSettings true "User settings"
// @Success 200 {object} map[string]interface{} "Updated settings"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/settings [put]
func (h *UserHandler) UpdateSettings(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req postgres.UserSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	settings, err := h.userService.UpdateSettings(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_settings_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    settings,
		"message": "Settings updated successfully",
	})
}

// GetSettings gets user settings
// @Summary Get user settings
// @Description Get current user's privacy and notification settings
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} postgres.UserSettings "User settings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/settings [get]
func (h *UserHandler) GetSettings(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	settings, err := h.userService.GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_settings_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": settings,
	})
}
