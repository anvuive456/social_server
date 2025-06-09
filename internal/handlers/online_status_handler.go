package handlers

import (
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OnlineStatusHandler struct {
	onlineStatusService *services.OnlineStatusService
	authService         *services.AuthService
}

func NewOnlineStatusHandler(onlineStatusService *services.OnlineStatusService, authService *services.AuthService) *OnlineStatusHandler {
	return &OnlineStatusHandler{
		onlineStatusService: onlineStatusService,
		authService:         authService,
	}
}

// GetUserOnlineStatus returns online status of a specific user
// GET /api/online-status/user/:user_id
func (h *OnlineStatusHandler) GetUserOnlineStatus(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "User ID must be a valid number",
		})
		return
	}

	targetUserID := uint(userID)

	// Check if current user can view this user's online status
	// (For now, any authenticated user can check others' status)
	// TODO: Add privacy settings check here

	isOnline := h.onlineStatusService.IsUserOnline(targetUserID)
	userInfo, _ := h.onlineStatusService.GetOnlineUserInfo(targetUserID)
	lastSeen, err := h.onlineStatusService.GetUserLastSeen(targetUserID)

	response := gin.H{
		"user_id":   targetUserID,
		"is_online": isOnline,
	}

	if userInfo != nil {
		response["connected_at"] = userInfo.ConnectedAt
		response["last_heartbeat"] = userInfo.LastHeartbeat
		response["connection_id"] = userInfo.ConnectionID
	}

	if lastSeen != nil {
		response["last_seen"] = lastSeen
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetOnlineFriends returns list of online friends for current user
// GET /api/online-status/friends
func (h *OnlineStatusHandler) GetOnlineFriends(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	onlineFriends, err := h.onlineStatusService.GetOnlineFriends(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_get_online_friends",
			"message": err.Error(),
		})
		return
	}

	// Get detailed info for each online friend
	var friendsInfo []gin.H
	for _, friendID := range onlineFriends {
		userInfo, exists := h.onlineStatusService.GetOnlineUserInfo(friendID)
		if exists {
			friendInfo := gin.H{
				"user_id":        friendID,
				"is_online":      true,
				"last_seen":      userInfo.LastSeen,
				"connected_at":   userInfo.ConnectedAt,
				"last_heartbeat": userInfo.LastHeartbeat,
			}
			friendsInfo = append(friendsInfo, friendInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"online_friends_count": len(onlineFriends),
			"online_friends":       friendsInfo,
		},
	})
}

// GetOnlineUsers returns list of all online users (admin only)
// GET /api/online-status/users
func (h *OnlineStatusHandler) GetOnlineUsers(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if user is admin (you might want to implement this check)
	// For now, allowing all authenticated users
	// TODO: Add admin role check

	onlineUserIDs := h.onlineStatusService.GetOnlineUsers()

	var usersInfo []gin.H
	for _, onlineUserID := range onlineUserIDs {
		userInfo, exists := h.onlineStatusService.GetOnlineUserInfo(onlineUserID)
		if exists {
			info := gin.H{
				"user_id":        onlineUserID,
				"is_online":      true,
				"last_seen":      userInfo.LastSeen,
				"connected_at":   userInfo.ConnectedAt,
				"last_heartbeat": userInfo.LastHeartbeat,
				"connection_id":  userInfo.ConnectionID,
			}
			usersInfo = append(usersInfo, info)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"online_users_count": len(onlineUserIDs),
			"online_users":       usersInfo,
		},
	})
}

// GetOnlineUsersCount returns count of online users
// GET /api/online-status/count
func (h *OnlineStatusHandler) GetOnlineUsersCount(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count := h.onlineStatusService.GetOnlineUserCount()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"online_users_count": count,
		},
	})
}

// GetMyOnlineStatus returns current user's online status
// GET /api/online-status/me
func (h *OnlineStatusHandler) GetMyOnlineStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isOnline := h.onlineStatusService.IsUserOnline(userID)
	userInfo, _ := h.onlineStatusService.GetOnlineUserInfo(userID)
	lastSeen, err := h.onlineStatusService.GetUserLastSeen(userID)

	response := gin.H{
		"user_id":   userID,
		"is_online": isOnline,
	}

	if userInfo != nil {
		response["connected_at"] = userInfo.ConnectedAt
		response["last_heartbeat"] = userInfo.LastHeartbeat
		response["connection_id"] = userInfo.ConnectionID
		response["is_active"] = userInfo.IsActive
	}

	if lastSeen != nil {
		response["last_seen"] = lastSeen
	}

	if err != nil {
		response["last_seen_error"] = err.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetConnectionStats returns statistics about online connections (admin only)
// GET /api/online-status/stats
func (h *OnlineStatusHandler) GetConnectionStats(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// TODO: Add admin role check here

	stats := h.onlineStatusService.GetConnectionStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserLastSeen returns user's last seen time
// GET /api/online-status/user/:user_id/last-seen
func (h *OnlineStatusHandler) GetUserLastSeen(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "User ID must be a valid number",
		})
		return
	}

	targetUserID := uint(userID)
	lastSeen, err := h.onlineStatusService.GetUserLastSeen(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_get_last_seen",
			"message": err.Error(),
		})
		return
	}

	response := gin.H{
		"user_id": targetUserID,
	}

	if lastSeen != nil {
		response["last_seen"] = lastSeen
		response["has_last_seen"] = true
	} else {
		response["has_last_seen"] = false
		response["message"] = "User has never been online"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// SetUserOffline manually sets user offline (admin only)
// POST /api/online-status/user/:user_id/offline
func (h *OnlineStatusHandler) SetUserOffline(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// TODO: Add admin role check here

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "User ID must be a valid number",
		})
		return
	}

	targetUserID := uint(userID)
	err = h.onlineStatusService.SetUserOffline(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_set_offline",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User set offline successfully",
		"data": gin.H{
			"user_id":   targetUserID,
			"is_online": false,
		},
	})
}
