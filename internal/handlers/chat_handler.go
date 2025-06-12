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

// // GetRoom gets a specific chat room
// // @Summary Get chat room
// // @Description Get details of a specific chat room
// // @Tags Chat
// // @Produce json
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} models.ChatRoom "Room details"
// // @Failure 400 {object} map[string]interface{} "Invalid room ID"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Failure 404 {object} map[string]interface{} "Room not found"
// // @Router /chat/rooms/{room_id} [get]
// func (h *ChatHandler) GetRoom(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	// Check access
// 	hasAccess, err := h.chatService.HasRoomAccess(c.Request.Context(), userID, roomID)
// 	if err != nil || !hasAccess {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
// 		return
// 	}

// 	room, err := h.chatService.GetRoom(c.Request.Context(), roomID)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, room)
// }

// // GetMessages gets messages from a chat room
// // @Summary Get room messages
// // @Description Get messages from a specific chat room
// // @Tags Chat
// // @Produce json
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Param limit query int false "Number of messages to return (1-100)" default(50)
// // @Param cursor query string false "Pagination cursor"
// // @Param created_at query string false "ISO timestamp for cursor"
// // @Success 200 {object} models.MessageResponse "Room messages"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Router /chat/rooms/{room_id}/messages [get]
// func (h *ChatHandler) GetMessages(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	limitStr := c.DefaultQuery("limit", "50")
// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 || limit > 100 {
// 		limit = 50
// 	}

// 	var cursor *models.MessageCursor
// 	if cursorParam := c.Query("cursor"); cursorParam != "" {
// 		if cursorID, err := primitive.ObjectIDFromHex(cursorParam); err == nil {
// 			if createdAtParam := c.Query("created_at"); createdAtParam != "" {
// 				if createdAt, err := time.Parse(time.RFC3339, createdAtParam); err == nil {
// 					cursor = &models.MessageCursor{
// 						ID:        cursorID,
// 						CreatedAt: createdAt,
// 					}
// 				}
// 			}
// 		}
// 	}

// 	response, err := h.chatService.GetRoomMessages(c.Request.Context(), roomID, userID, cursor, limit)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // SendMessage sends a message to a chat room
// // @Summary Send message
// // @Description Send a message to a chat room
// // @Tags Chat
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Param request body SendMessageRequest true "Message data"
// // @Success 201 {object} models.Message "Sent message"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Failure 500 {object} map[string]interface{} "Send failed"
// // @Router /chat/rooms/{room_id}/messages [post]
// func (h *ChatHandler) SendMessage(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	var req SendMessageRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Validate message
// 	if req.Content == "" && req.Media == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Message must have content or media"})
// 		return
// 	}

// 	message := &models.Message{
// 		ChatRoomID: roomID,
// 		SenderID:   userID,
// 		Type:       req.Type,
// 		Content:    req.Content,
// 		Media:      req.Media,
// 		ReplyToID:  req.ReplyToID,
// 		Mentions:   req.Mentions,
// 		Tags:       req.Tags,
// 		Location:   req.Location,
// 	}

// 	if message.Type == "" {
// 		message.Type = models.MessageTypeText
// 	}

// 	err = h.chatService.SendMessage(c.Request.Context(), message)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, message)
// }

// // DeleteMessage deletes a message
// // @Summary Delete message
// // @Description Delete a message from chat room
// // @Tags Chat
// // @Security BearerAuth
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} map[string]interface{} "Message deleted"
// // @Failure 400 {object} map[string]interface{} "Invalid message ID"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Failure 404 {object} map[string]interface{} "Message not found"
// // @Router /chat/messages/{message_id} [delete]
// func (h *ChatHandler) DeleteMessage(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	messageID, err := primitive.ObjectIDFromHex(c.Param("message_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
// 		return
// 	}

// 	err = h.chatService.DeleteMessage(c.Request.Context(), messageID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
// }

// // MarkMessageRead marks a message as read
// // @Summary Mark message as read
// // @Description Mark a message as read by the user
// // @Tags Chat
// // @Security BearerAuth
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} map[string]interface{} "Message marked as read"
// // @Failure 400 {object} map[string]interface{} "Invalid message ID"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 500 {object} map[string]interface{} "Mark failed"
// // @Router /chat/messages/{message_id}/read [post]
// func (h *ChatHandler) MarkMessageRead(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	messageID, err := primitive.ObjectIDFromHex(c.Param("message_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
// 		return
// 	}

// 	err = h.chatService.MarkMessageAsRead(c.Request.Context(), messageID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
// }

// // AddReaction adds a reaction to a message
// // @Summary Add reaction
// // @Description Add a reaction to a message
// // @Tags Chat
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param message_id path string true "Message ID"
// // @Param request body AddReactionRequest true "Reaction data"
// // @Success 200 {object} map[string]interface{} "Reaction added"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 500 {object} map[string]interface{} "Add failed"
// // @Router /chat/messages/{message_id}/reactions [post]
// func (h *ChatHandler) AddReaction(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	messageID, err := primitive.ObjectIDFromHex(c.Param("message_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
// 		return
// 	}

// 	var req AddReactionRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err = h.chatService.AddReaction(c.Request.Context(), messageID, userID, req.Emoji)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
// }

// // RemoveReaction removes a reaction from a message
// // @Summary Remove reaction
// // @Description Remove a reaction from a message
// // @Tags Chat
// // @Security BearerAuth
// // @Param message_id path string true "Message ID"
// // @Param emoji path string true "Emoji"
// // @Success 200 {object} map[string]interface{} "Reaction removed"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 500 {object} map[string]interface{} "Remove failed"
// // @Router /chat/messages/{message_id}/reactions/{emoji} [delete]
// func (h *ChatHandler) RemoveReaction(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	messageID, err := primitive.ObjectIDFromHex(c.Param("message_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
// 		return
// 	}

// 	emoji := c.Param("emoji")
// 	if emoji == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Emoji is required"})
// 		return
// 	}

// 	err = h.chatService.RemoveReaction(c.Request.Context(), messageID, userID, emoji)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
// }

// // GetParticipants gets room participants
// // @Summary Get room participants
// // @Description Get list of participants in a chat room
// // @Tags Chat
// // @Produce json
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} []models.ChatRoomMember "Room participants"
// // @Failure 400 {object} map[string]interface{} "Invalid room ID"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Router /chat/rooms/{room_id}/participants [get]
// func (h *ChatHandler) GetParticipants(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	participants, err := h.chatService.GetRoomParticipants(c.Request.Context(), roomID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, participants)
// }

// // AddParticipant adds a participant to a room
// // @Summary Add participant
// // @Description Add a user to a chat room
// // @Tags Chat
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Param request body AddParticipantRequest true "Participant data"
// // @Success 200 {object} map[string]interface{} "Participant added"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Failure 500 {object} map[string]interface{} "Add failed"
// // @Router /chat/rooms/{room_id}/participants [post]
// func (h *ChatHandler) AddParticipant(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	var req AddParticipantRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err = h.chatService.AddParticipant(c.Request.Context(), roomID, req.UserID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Participant added successfully"})
// }

// // RemoveParticipant removes a participant from a room
// // @Summary Remove participant
// // @Description Remove a user from a chat room
// // @Tags Chat
// // @Security BearerAuth
// // @Param room_id path string true "Room ID"
// // @Param user_id path string true "User ID"
// // @Success 200 {object} map[string]interface{} "Participant removed"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 403 {object} map[string]interface{} "Access denied"
// // @Failure 500 {object} map[string]interface{} "Remove failed"
// // @Router /chat/rooms/{room_id}/participants/{user_id} [delete]
// func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
// 	currentUserID := c.MustGet("user_id")

// 	roomID, err := primitive.ObjectIDFromHex(c.Param("room_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
// 		return
// 	}

// 	userID, err := primitive.ObjectIDFromHex(c.Param("user_id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
// 		return
// 	}

// 	err = h.chatService.RemoveParticipant(c.Request.Context(), roomID, userID, currentUserID)
// 	if err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Participant removed successfully"})
// }

// // SearchRooms searches chat rooms
// // @Summary Search rooms
// // @Description Search chat rooms by name or description
// // @Tags Chat
// // @Produce json
// // @Security BearerAuth
// // @Param q query string true "Search query"
// // @Param limit query int false "Number of rooms to return (1-100)" default(20)
// // @Param cursor query string false "Pagination cursor"
// // @Param last_activity query string false "ISO timestamp for cursor"
// // @Success 200 {object} models.ChatRoomResponse "Search results"
// // @Failure 400 {object} map[string]interface{} "Invalid request"
// // @Failure 401 {object} map[string]interface{} "Unauthorized"
// // @Failure 500 {object} map[string]interface{} "Search failed"
// // @Router /chat/rooms/search [get]
// func (h *ChatHandler) SearchRooms(c *gin.Context) {
// 	userID := c.MustGet("user_id")

// 	query := c.Query("q")
// 	if query == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
// 		return
// 	}

// 	limitStr := c.DefaultQuery("limit", "20")
// 	limit, err := strconv.Atoi(limitStr)
// 	if err != nil || limit < 1 || limit > 100 {
// 		limit = 20
// 	}

// 	var cursor *models.ChatRoomCursor
// 	if cursorParam := c.Query("cursor"); cursorParam != "" {
// 		if cursorID, err := primitive.ObjectIDFromHex(cursorParam); err == nil {
// 			if lastActivityParam := c.Query("last_activity"); lastActivityParam != "" {
// 				if lastActivity, err := time.Parse(time.RFC3339, lastActivityParam); err == nil {
// 					cursor = &models.ChatRoomCursor{
// 						ID:           cursorID,
// 						LastActivity: lastActivity,
// 					}
// 				}
// 			}
// 		}
// 	}

// 	response, err := h.chatService.SearchRooms(c.Request.Context(), query, userID, cursor, limit)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // Request/Response types
// // @Description Request body for creating a chat room
// type CreateRoomRequest struct {
// 	Name           string                  `json:"name" example:"Project Team"`
// 	Description    string                  `json:"description" example:"Discussion about the new project"`
// 	Type           models.ChatRoomType     `json:"type" binding:"required" example:"group"`
// 	Avatar         string                  `json:"avatar" example:"/uploads/avatars/room.png"`
// 	ParticipantIDs []uint                  `json:"participant_ids"`
// 	Settings       models.ChatRoomSettings `json:"settings"`
// }

// // @Description Request body for sending a message
// type SendMessageRequest struct {
// 	Content   string                  `json:"content" example:"Hello everyone!"`
// 	Type      models.MessageType      `json:"type" example:"text"`
// 	Media     *models.MessageMedia    `json:"media,omitempty"`
// 	ReplyToID *string                 `json:"reply_to_id,omitempty"`
// 	Mentions  []string                `json:"mentions,omitempty"`
// 	Tags      []string                `json:"tags,omitempty"`
// 	Location  *models.MessageLocation `json:"location,omitempty"`
// }

// // @Description Request body for adding a reaction to a message
// type AddReactionRequest struct {
// 	Emoji string `json:"emoji" binding:"required" example:"👍"`
// }

// // @Description Request body for adding a participant to a room
// type AddParticipantRequest struct {
// 	UserID string `json:"user_id" binding:"required"`
// }
