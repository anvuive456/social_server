package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"social_server/internal/middleware"
	"social_server/internal/models/requests"
	"social_server/internal/services"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendService *services.FriendService
}

func NewFriendHandler(friendService *services.FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

// GetFriends gets user's friends list
// @Summary Get friends list
// @Description Get list of user's friends with pagination
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of friends to return" default(20)
// @Param before query string false "Cursor for pagination"
// @Param after query string false "Cursor for pagination"
// @Success 200 {object} map[string]interface{} "Friends list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends [get]
func (h *FriendHandler) GetFriends(c *gin.Context) {
	currentID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse req
	var req requests.GetFriendsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	response, err := h.friendService.GetFriends(currentID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_friends_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// SendFriendRequest sends a friend request
// @Summary Send friend request
// @Description Send a friend request to another user
// @Tags Friends
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Target user ID"
// @Param request body map[string]string false "Optional message"
// @Success 200 {object} map[string]interface{} "Friend request sent"
// @Failure 400 {object} map[string]interface{} "Invalid request or user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /friends/send-request [post]
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.SendFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request format",
		})
		return
	}

	if err := h.friendService.SendFriendRequest(userID, uint(req.TargetID), req.Message); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "send_request_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend request sent successfully",
	})
}

// AcceptFriendRequest accepts a friend request
// @Summary Accept friend request
// @Description Accept a friend request from another user
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param request body requests.AcceptFriendRequest true "Sender user ID"
// @Success 200 {object} map[string]interface{} "Friend request accepted"
// @Failure 400 {object} map[string]interface{} "Invalid user ID or request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Request not found"
// @Router /friends/accept-request [post]
func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.AcceptFriendRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.friendService.AcceptFriendRequest(userID, req.UserID); err != nil {
		if err.Error() == "friend request not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "request_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "accept_request_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend request accepted successfully",
	})
}

// DeclineFriendRequest declines a friend request
// @Summary Decline friend request
// @Description Decline a friend request from another user
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param request body requests.DeclineFriendRequest true "Sender user ID"
// @Success 200 {object} map[string]interface{} "Friend request declined"
// @Failure 400 {object} map[string]interface{} "Invalid user ID or request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Request not found"
// @Router /friends/decline-request [post]
func (h *FriendHandler) DeclineFriendRequest(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}
	var req requests.DeclineFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request format",
		})
		return
	}

	if err := h.friendService.DeclineFriendRequest(userID, uint(req.UserID)); err != nil {
		if err.Error() == "friend request not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "request_not_found",
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "decline_request_failed",
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend request declined successfully",
	})
}

// GetFriendRequests gets user's friend requests
// @Summary Get friend requests
// @Description Get list of received or sent friend requests
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param type query requests.GetFriendRequestType true "Request type: received, sent, or all" default("all")
// @Param limit query int true "Number of requests to return" default(5)
// @Param before query string false "Cursor for pagination"
// @Param after query string false "Cursor for pagination"
// @Success 200 {object} map[string]interface{} "Friend requests list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/requests [get]
func (h *FriendHandler) GetFriendRequests(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}
	var req requests.GetFriendRequestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	println(fmt.Sprintf("Request: %v", req))

	response, err := h.friendService.GetFriendRequests(userID, req.Type, req.Limit, req.Before, req.After)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_requests_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetFriendRequestStats gets user's friend requests stats
// @Summary Get friend requests stats
// @Description Get stats of received or sent friend requests
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Friend requests stats"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/requests-stats [get]
func (h *FriendHandler) GetFriendRequestStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	stats, err := h.friendService.GetFriendRequestStats(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_stats_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// RemoveFriend removes a friend
// @Summary Remove friend
// @Description Remove a user from friends list
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param id path int true "Friend user ID"
// @Success 200 {object} map[string]interface{} "Friend removed"
// @Failure 400 {object} map[string]interface{} "Invalid friend ID or not friends"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/{id} [delete]
func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	friendIDParam := c.Param("id")
	friendID, err := strconv.ParseUint(friendIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_friend_id",
			"message": "Invalid friend ID format",
		})
		return
	}

	err = h.friendService.RemoveFriend(userID, uint(friendID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "remove_friend_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend removed successfully",
	})
}

// BlockUser blocks a user
// @Summary Block user
// @Description Block a user and remove from friends if applicable
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID to block"
// @Success 200 {object} map[string]interface{} "User blocked"
// @Failure 400 {object} map[string]interface{} "Invalid user ID or already blocked"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/{id}/block [post]
func (h *FriendHandler) BlockUser(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_target_id",
			"message": "Invalid target user ID format",
		})
		return
	}

	err = h.friendService.BlockUser(userID, uint(targetID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "block_user_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User blocked successfully",
	})
}

// UnblockUser unblocks a user
// @Summary Unblock user
// @Description Unblock a previously blocked user
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID to unblock"
// @Success 200 {object} map[string]interface{} "User unblocked"
// @Failure 400 {object} map[string]interface{} "Invalid user ID or not blocked"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/{id}/unblock [post]
func (h *FriendHandler) UnblockUser(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_target_id",
			"message": "Invalid target user ID format",
		})
		return
	}

	err = h.friendService.UnblockUser(userID, uint(targetID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unblock_user_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User unblocked successfully",
	})
}

// GetBlockedUsers gets list of blocked users
// @Summary Get blocked users
// @Description Get list of users blocked by current user
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of users to return" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} map[string]interface{} "Blocked users list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/blocked [get]
func (h *FriendHandler) GetBlockedUsers(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Parse limit
	limitParam := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	cursor := c.Query("cursor")

	response, err := h.friendService.GetBlockedUsers(userID, limit, cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "get_blocked_users_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// CheckFriendship checks friendship status between users
// @Summary Check friendship status
// @Description Check friendship status between current user and another user
// @Tags Friends
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID to check"
// @Success 200 {object} map[string]interface{} "Friendship status"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /friends/{id}/status [get]
func (h *FriendHandler) CheckFriendship(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_target_id",
			"message": "Invalid target user ID format",
		})
		return
	}

	status, err := h.friendService.CheckFriendshipStatus(userID, uint(targetID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "check_status_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}
