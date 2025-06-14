package handlers

import (
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/models/constants"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/services"

	"github.com/gin-gonic/gin"
)

var _ = postgres.ChatRoom{}
var _ = requests.CreateChatRoomRequest{}
var _ = responses.ChatRoomsResponse{}

type ChatHandler struct {
	chatService *services.ChatService
	userService *services.UserService
}

func NewChatHandler(chatService *services.ChatService, userService *services.UserService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		userService: userService,
	}
}

// CreateRoom creates a new chat room
// @Summary Create chat room
// @Description Create a new chat room (group or private)
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateChatRoomRequest true "Room creation request"
// @Success 201 {object} postgres.ChatRoom "Created room"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Creation failed"
// @Router /chat/rooms [post]
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req requests.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request",
			"message": err.Error()})
		return
	}

	if req.Type == constants.ChatRoomTypePrivate && len(req.Participants) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "participant_must_be_one_user",
			"message": "Private room must have exactly one participant",
		})
		return
	}

	room, err := h.chatService.CreateRoom(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_room_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

// GetRooms gets user's chat rooms
// @Summary Get user chat rooms
// @Description Get list of chat rooms for the authenticated user
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param query query requests.GetChatRoomsRequest true "Query parameters"
// @Success 200 {object} responses.ChatRoomsResponse "User rooms"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Fetch failed"
// @Router /chat/rooms [get]
func (h *ChatHandler) GetRooms(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req requests.GetChatRoomsRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	response, err := h.chatService.GetUserRooms(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteRoom deletes a chat room
// @Summary Delete chat room
// @Description Delete a specific chat room
// @Security BearerAuth
// @Tags Chat
// @Param id path string true "Room ID"
// @Success 200 {object} map[string]interface{} "Room deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Delete failed"
// @Router /chat/rooms/{id} [delete]
func (h *ChatHandler) DeleteRoom(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req requests.DeleteChatRoomRequest
	if err := c.BindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	err := h.chatService.DeleteRoom(userID, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_room_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}

// SyncRooms syncs the user's chat rooms
// @Summary Sync chat rooms
// @Description Sync the user's chat rooms
// @Security BearerAuth
// @Tags Chat
// @Param query query requests.SyncChatRoomsRequest true "Sync chat rooms"
// @Success 200 {object} map[string]interface{} "Rooms synced successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Sync failed"
// @Router /chat/rooms/sync [get]
func (h *ChatHandler) SyncRooms(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}
	var req requests.SyncChatRoomsRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "message": err.Error()})
		return
	}

	response, err := h.chatService.SyncRooms(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync_rooms_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}
