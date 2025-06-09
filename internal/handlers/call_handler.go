package handlers

import (
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CallHandler struct {
	callService *services.CallService
	wsHandler   *WebSocketHandler
}

func NewCallHandler(callService *services.CallService, wsHandler *WebSocketHandler) *CallHandler {
	return &CallHandler{
		callService: callService,
		wsHandler:   wsHandler,
	}
}

// InitiateCall godoc
// @Summary Initiate a new call
// @Description Create a new call and send call request to the callee
// @Tags calls
// @Accept json
// @Produce json
// @Param request body requests.InitiateCallRequest true "Call initiation request"
// @Success 201 {object} responses.CallResponse
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 500 {object} interface{}
// @Security BearerAuth
// @Router /calls [post]
func (h *CallHandler) InitiateCall(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	callerID := userIDInterface.(uint)

	var req requests.InitiateCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate call type
	if req.Type != "video" && req.Type != "audio" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_call_type",
			"message": "Call type must be 'video' or 'audio'",
		})
		return
	}

	call, err := h.callService.CreateCall(callerID, req.CalleeID, req.Type, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "call_creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, responses.CallResponse{
		ID:       call.ID,
		CallerID: call.CallerID,
		CalleeID: call.CalleeID,
		Type:     call.Type,
		Status:   call.Status,
		RoomID:   call.RoomID,
	})
}

// GetCall godoc
// @Summary Get call details
// @Description Get details of a specific call
// @Tags calls
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} responses.CallDetailResponse
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 404 {object} interface{}
// @Security BearerAuth
// @Router /calls/{id} [get]
func (h *CallHandler) GetCall(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	callIDStr := c.Param("id")
	callID, err := strconv.ParseUint(callIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_call_id",
			"message": "Invalid call ID format",
		})
		return
	}

	// Check if user can access this call
	canAccess, err := h.callService.CanUserAccessCall(uint(callID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to check call access",
		})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have access to this call",
		})
		return
	}

	call, err := h.callService.GetCallByID(uint(callID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "call_not_found",
			"message": "Call not found",
		})
		return
	}

	participants, err := h.callService.GetCallParticipants(uint(callID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to get call participants",
		})
		return
	}

	c.JSON(http.StatusOK, responses.CallDetailResponse{
		ID:           call.ID,
		CallerID:     call.CallerID,
		CalleeID:     call.CalleeID,
		Type:         call.Type,
		Status:       call.Status,
		Duration:     call.Duration,
		StartedAt:    call.StartedAt,
		EndedAt:      call.EndedAt,
		IsGroupCall:  call.IsGroupCall,
		RoomID:       call.RoomID,
		Participants: participants,
		CreatedAt:    call.CreatedAt,
		UpdatedAt:    call.UpdatedAt,
	})
}

// AcceptCall godoc
// @Summary Accept a call
// @Description Accept an incoming call
// @Tags calls
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 404 {object} interface{}
// @Security BearerAuth
// @Router /calls/{id}/accept [post]
func (h *CallHandler) AcceptCall(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	callIDStr := c.Param("id")
	callID, err := strconv.ParseUint(callIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_call_id",
			"message": "Invalid call ID format",
		})
		return
	}

	// Check if user can access this call
	canAccess, err := h.callService.CanUserAccessCall(uint(callID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to check call access",
		})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have access to this call",
		})
		return
	}

	err = h.callService.AcceptCall(uint(callID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "call_accept_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Call accepted successfully",
	})
}

// DeclineCall godoc
// @Summary Decline a call
// @Description Decline an incoming call
// @Tags calls
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 404 {object} interface{}
// @Security BearerAuth
// @Router /calls/{id}/decline [post]
func (h *CallHandler) DeclineCall(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	callIDStr := c.Param("id")
	callID, err := strconv.ParseUint(callIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_call_id",
			"message": "Invalid call ID format",
		})
		return
	}

	// Check if user can access this call
	canAccess, err := h.callService.CanUserAccessCall(uint(callID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to check call access",
		})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have access to this call",
		})
		return
	}

	err = h.callService.DeclineCall(uint(callID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "call_decline_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Call declined successfully",
	})
}

// EndCall godoc
// @Summary End a call
// @Description End an ongoing call
// @Tags calls
// @Produce json
// @Param id path int true "Call ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Failure 404 {object} interface{}
// @Security BearerAuth
// @Router /calls/{id}/end [post]
func (h *CallHandler) EndCall(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	callIDStr := c.Param("id")
	callID, err := strconv.ParseUint(callIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_call_id",
			"message": "Invalid call ID format",
		})
		return
	}

	// Check if user can access this call
	canAccess, err := h.callService.CanUserAccessCall(uint(callID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to check call access",
		})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have access to this call",
		})
		return
	}

	err = h.callService.EndCall(uint(callID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "call_end_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Call ended successfully",
	})
}

// GetCallHistory godoc
// @Summary Get call history
// @Description Get user's call history with pagination
// @Tags calls
// @Produce json
// @Param before query string false "Pagination cursor"
// @Param after query string false "Pagination cursor"
// @Param limit query int false "Number of calls to return" default(20)
// @Success 200 {object} responses.CallHistoriesResponse
// @Failure 400 {object} interface{}
// @Failure 401 {object} interface{}
// @Security BearerAuth
// @Router /calls/history [get]
func (h *CallHandler) GetCallHistory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}
	var req requests.CallHistoryQueryRequest
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_query",
			"message": "Invalid query parameters",
		})
		return
	}

	calls, nextCursor, err := h.callService.GetCallHistory(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to get call history",
		})
		return
	}

	c.JSON(http.StatusOK, responses.CallHistoriesResponse{
		Calls:      calls,
		NextCursor: nextCursor,
	})
}

// GetActiveCalls godoc
// @Summary Get active calls
// @Description Get user's currently active calls
// @Tags calls
// @Produce json
// @Success 200 {object} responses.ActiveCallsResponse
// @Failure 401 {object} interface{}
// @Failure 500 {object} interface{}
// @Security BearerAuth
// @Router /calls/active [get]
func (h *CallHandler) GetActiveCalls(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	activeCall, err := h.callService.GetActiveCall(userID)
	if err != nil {
		c.JSON(http.StatusOK, responses.ActiveCallsResponse{
			Calls: []responses.CallResponse{},
		})
		return
	}

	calls := []responses.CallResponse{}
	if activeCall != nil {
		calls = append(calls, responses.CallResponse{
			ID:       activeCall.ID,
			CallerID: activeCall.CallerID,
			CalleeID: activeCall.CalleeID,
			Type:     activeCall.Type,
			Status:   activeCall.Status,
			RoomID:   activeCall.RoomID,
		})
	}

	c.JSON(http.StatusOK, responses.ActiveCallsResponse{
		Calls: calls,
	})
}

// GetCallStats godoc
// @Summary Get call statistics
// @Description Get user's call statistics
// @Tags calls
// @Produce json
// @Success 200 {object} responses.CallStatsResponse
// @Failure 401 {object} interface{}
// @Failure 500 {object} interface{}
// @Security BearerAuth
// @Router /calls/stats [get]
func (h *CallHandler) GetCallStats(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)

	stats, err := h.callService.GetCallStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "database_error",
			"message": "Failed to get call statistics",
		})
		return
	}

	c.JSON(http.StatusOK, responses.CallStatsResponse{
		Stats: stats,
	})
}

// GetWebRTCConfig godoc
// @Summary Get WebRTC configuration
// @Description Get STUN/TURN server configuration for WebRTC
// @Tags calls
// @Produce json
// @Success 200 {object} models.WebRTCConfig
// @Failure 401 {object} interface{}
// @Security BearerAuth
// @Router /calls/webrtc-config [get]
func (h *CallHandler) GetWebRTCConfig(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	config := h.wsHandler.GetWebRTCConfig()
	c.JSON(http.StatusOK, config)
}

// GetConnectionStats godoc
// @Summary Get WebSocket connection statistics
// @Description Get statistics about WebSocket connections (admin only)
// @Tags calls
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} interface{}
// @Security BearerAuth
// @Router /calls/connection-stats [get]
func (h *CallHandler) GetConnectionStats(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	stats := h.wsHandler.GetConnectionStats()
	c.JSON(http.StatusOK, stats)
}
