package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/services"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketChatHandler struct {
	chatService  *services.ChatService
	userService  *services.UserService
	upgrader     websocket.Upgrader
	clients      map[string]*ChatClient
	rooms        map[string]map[string]*ChatClient
	clientsMutex sync.RWMutex
	roomsMutex   sync.RWMutex
	broadcast    chan *responses.WebSocketMessage
	register     chan *ChatClient
	unregister   chan *ChatClient
}

type ChatClient struct {
	ID       string
	UserID   string
	Username string
	Conn     *websocket.Conn
	Send     chan *responses.WebSocketMessage
	Rooms    map[string]bool
	IsOnline bool
	LastSeen time.Time
}

type WebSocketRequest struct {
	Type   string      `json:"type"`
	RoomID string      `json:"room_id,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type JoinRoomData struct {
	RoomID string `json:"room_id"`
}

type LeaveRoomData struct {
	RoomID string `json:"room_id"`
}

type SendMessageData struct {
	RoomID    string                            `json:"room_id"`
	Content   string                            `json:"content"`
	Type      requests.MessageType              `json:"type"`
	Media     *requests.SendMessageMediaRequest `json:"media,omitempty"`
	ReplyToID *string                           `json:"reply_to_id,omitempty"`
	Mentions  []string                          `json:"mentions,omitempty"`
	Tags      []string                          `json:"tags,omitempty"`
}

type TypingData struct {
	RoomID   string `json:"room_id"`
	IsTyping bool   `json:"is_typing"`
}

type MarkReadData struct {
	MessageID string `json:"message_id"`
}

func NewWebSocketChatHandler(chatService *services.ChatService, userService *services.UserService) *WebSocketChatHandler {
	handler := &WebSocketChatHandler{
		chatService: chatService,
		userService: userService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:    make(map[string]*ChatClient),
		rooms:      make(map[string]map[string]*ChatClient),
		broadcast:  make(chan *responses.WebSocketMessage, 256),
		register:   make(chan *ChatClient),
		unregister: make(chan *ChatClient),
	}

	go handler.run()
	return handler
}

func (h *WebSocketChatHandler) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// HandleWebSocket handles WebSocket connections for real-time chat
// @Summary Connect to chat WebSocket
// @Description Establish WebSocket connection for real-time chat messaging
// @Tags Chat
// @Security BearerAuth
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /ws/chat [get]
func (h *WebSocketChatHandler) HandleWebSocket(c *gin.Context) {
	// Get user from context (set by auth middleware)
	// userIDInterface, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	// userID, ok := userIDInterface
	// if !ok {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
	// 	return
	// }

	// // Get user info
	// user, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 	return
	// }

	// // Upgrade HTTP connection to WebSocket
	// conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	// if err != nil {
	// 	log.Printf("WebSocket upgrade failed: %v", err)
	// 	return
	// }

	// // Create client
	// client := &ChatClient{
	// 	ID:       primitive.NewObjectID().Hex(),
	// 	UserID:   userID,
	// 	Username: user.Username,
	// 	Conn:     conn,
	// 	Send:     make(chan *models.WebSocketMessage, 256),
	// 	Rooms:    make(map[string]bool),
	// 	IsOnline: true,
	// 	LastSeen: time.Now(),
	// }

	// // Register client
	// h.register <- client

	// // Set user online status
	// h.userService.UpdateOnlineStatus(c.Request.Context(), userID, true)

	// // Start goroutines
	// go h.writePump(client)
	// go h.readPump(client)
}

func (h *WebSocketChatHandler) registerClient(client *ChatClient) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	h.clients[client.ID] = client

	// Notify user is online
	h.broadcastUserStatus(client.UserID, true)

	log.Printf("Client %s connected (User: %s)", client.ID, client.Username)
}

func (h *WebSocketChatHandler) unregisterClient(client *ChatClient) {
	// h.clientsMutex.Lock()
	// defer h.clientsMutex.Unlock()

	// if _, ok := h.clients[client.ID]; ok {
	// 	// Remove from all rooms
	// 	for roomID := range client.Rooms {
	// 		h.leaveRoom(client, roomID)
	// 	}

	// 	// Close send channel and remove client
	// 	close(client.Send)
	// 	delete(h.clients, client.ID)

	// 	// Set user offline
	// 	h.userService.UpdateOnlineStatus(context.Background(), client.UserID, false)

	// 	// Notify user is offline
	// 	h.broadcastUserStatus(client.UserID, false)

	// 	log.Printf("Client %s disconnected (User: %s)", client.ID, client.Username)
	// }
}

func (h *WebSocketChatHandler) readPump(client *ChatClient) {
	// defer func() {
	// 	h.unregister <- client
	// 	client.Conn.Close()
	// }()

	// client.Conn.SetReadLimit(512)
	// client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// client.Conn.SetPongHandler(func(string) error {
	// 	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// 	return nil
	// })

	// for {
	// 	_, message, err := client.Conn.ReadMessage()
	// 	if err != nil {
	// 		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
	// 			log.Printf("WebSocket error: %v", err)
	// 		}
	// 		break
	// 	}

	// 	var req WebSocketRequest
	// 	if err := json.Unmarshal(message, &req); err != nil {
	// 		log.Printf("Invalid message format: %v", err)
	// 		continue
	// 	}

	// 	h.handleMessage(client, &req)
	// }
}

func (h *WebSocketChatHandler) writePump(client *ChatClient) {
	// ticker := time.NewTicker(54 * time.Second)
	// defer func() {
	// 	ticker.Stop()
	// 	client.Conn.Close()
	// }()

	// for {
	// 	select {
	// 	case message, ok := <-client.Send:
	// 		client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// 		if !ok {
	// 			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
	// 			return
	// 		}

	// 		if err := client.Conn.WriteJSON(message); err != nil {
	// 			log.Printf("WebSocket write error: %v", err)
	// 			return
	// 		}

	// 	case <-ticker.C:
	// 		client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// 		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
	// 			return
	// 		}
	// 	}
	// }
}

func (h *WebSocketChatHandler) handleMessage(client *ChatClient, req *WebSocketRequest) {
	// ctx := context.Background()

	// switch req.Type {
	// case "join_room":
	// 	var data JoinRoomData
	// 	if err := h.mapToStruct(req.Data, &data); err != nil {
	// 		log.Printf("Invalid join_room data: %v", err)
	// 		return
	// 	}
	// 	h.handleJoinRoom(ctx, client, data.RoomID)

	// case "leave_room":
	// 	var data LeaveRoomData
	// 	if err := h.mapToStruct(req.Data, &data); err != nil {
	// 		log.Printf("Invalid leave_room data: %v", err)
	// 		return
	// 	}
	// 	h.handleLeaveRoom(client, data.RoomID)

	// case "send_message":
	// 	var data SendMessageData
	// 	if err := h.mapToStruct(req.Data, &data); err != nil {
	// 		log.Printf("Invalid send_message data: %v", err)
	// 		return
	// 	}
	// 	h.handleSendMessage(ctx, client, &data)

	// case "typing":
	// 	var data TypingData
	// 	if err := h.mapToStruct(req.Data, &data); err != nil {
	// 		log.Printf("Invalid typing data: %v", err)
	// 		return
	// 	}
	// 	h.handleTyping(ctx, client, &data)

	// case "mark_read":
	// 	var data MarkReadData
	// 	if err := h.mapToStruct(req.Data, &data); err != nil {
	// 		log.Printf("Invalid mark_read data: %v", err)
	// 		return
	// 	}
	// 	h.handleMarkRead(ctx, client, &data)

	// default:
	// 	log.Printf("Unknown message type: %s", req.Type)
	// }
}

func (h *WebSocketChatHandler) handleJoinRoom(ctx context.Context, client *ChatClient, roomID string) {
	// roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	// if err != nil {
	// 	log.Printf("Invalid room ID: %s", roomID)
	// 	return
	// }

	// // Check if user has access to room
	// hasAccess, err := h.chatService.HasRoomAccess(ctx, client.UserID, roomObjectID)
	// if err != nil || !hasAccess {
	// 	log.Printf("User %s denied access to room %s", client.Username, roomID)
	// 	return
	// }

	// h.joinRoom(client, roomID)

	// // Send confirmation
	// client.Send <- &models.WebSocketMessage{
	// 	Type:      "room_joined",
	// 	RoomID:    roomID,
	// 	Data:      map[string]interface{}{"status": "success"},
	// 	Timestamp: time.Now(),
	// }
}

func (h *WebSocketChatHandler) handleLeaveRoom(client *ChatClient, roomID string) {
	// h.leaveRoom(client, roomID)

	// // Send confirmation
	// client.Send <- &models.WebSocketMessage{
	// 	Type:      "room_left",
	// 	RoomID:    roomID,
	// 	Data:      map[string]interface{}{"status": "success"},
	// 	Timestamp: time.Now(),
	// }
}

func (h *WebSocketChatHandler) handleSendMessage(ctx context.Context, client *ChatClient, data *SendMessageData) {
	// roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
	// if err != nil {
	// 	log.Printf("Invalid room ID: %s", data.RoomID)
	// 	return
	// }

	// // Create message
	// message := &models.Message{
	// 	ChatRoomID: roomObjectID,
	// 	SenderID:   client.UserID,
	// 	Type:       data.Type,
	// 	Content:    data.Content,
	// 	Media:      data.Media,
	// 	ReplyToID:  data.ReplyToID,
	// 	Mentions:   data.Mentions,
	// 	Tags:       data.Tags,
	// }

	// // Save message
	// err = h.chatService.SendMessage(ctx, message)
	// if err != nil {
	// 	log.Printf("Failed to send message: %v", err)
	// 	return
	// }

	// // Broadcast to room
	// h.broadcastToRoom(data.RoomID, &models.WebSocketMessage{
	// 	Type:      string(models.WSEventNewMessage),
	// 	RoomID:    data.RoomID,
	// 	Data:      message,
	// 	Timestamp: time.Now(),
	// })
}

func (h *WebSocketChatHandler) handleTyping(ctx context.Context, client *ChatClient, data *TypingData) {
	// roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
	// if err != nil {
	// 	log.Printf("Invalid room ID: %s", data.RoomID)
	// 	return
	// }

	// // Update typing status
	// if data.IsTyping {
	// 	h.chatService.SetTyping(ctx, roomObjectID, client.UserID, true)
	// } else {
	// 	h.chatService.SetTyping(ctx, roomObjectID, client.UserID, false)
	// }

	// // Broadcast typing status
	// eventType := models.WSEventTypingStart
	// if !data.IsTyping {
	// 	eventType = models.WSEventTypingStop
	// }

	// h.broadcastToRoomExcept(data.RoomID, client.ID, &models.WebSocketMessage{
	// 	Type:   string(eventType),
	// 	RoomID: data.RoomID,
	// 	Data: map[string]interface{}{
	// 		"user_id":   client.UserID,
	// 		"username":  client.Username,
	// 		"is_typing": data.IsTyping,
	// 	},
	// 	Timestamp: time.Now(),
	// })
}

func (h *WebSocketChatHandler) handleMarkRead(ctx context.Context, client *ChatClient, data *MarkReadData) {
	// messageObjectID, err := primitive.ObjectIDFromHex(data.MessageID)
	// if err != nil {
	// 	log.Printf("Invalid message ID: %s", data.MessageID)
	// 	return
	// }

	// // Mark message as read
	// err = h.chatService.MarkMessageAsRead(ctx, messageObjectID, client.UserID)
	// if err != nil {
	// 	log.Printf("Failed to mark message as read: %v", err)
	// 	return
	// }

	// // Get message to get room ID
	// message, err := h.chatService.GetMessage(ctx, messageObjectID)
	// if err != nil {
	// 	log.Printf("Failed to get message: %v", err)
	// 	return
	// }

	// // Broadcast read status
	// h.broadcastToRoom(message.ChatRoomID.Hex(), &models.WebSocketMessage{
	// 	Type: string(models.WSEventMessageRead),
	// 	Data: map[string]interface{}{
	// 		"message_id": data.MessageID,
	// 		"user_id":    client.UserID,
	// 		"read_at":    time.Now(),
	// 	},
	// 	Timestamp: time.Now(),
	// })
}

func (h *WebSocketChatHandler) joinRoom(client *ChatClient, roomID string) {
	// h.roomsMutex.Lock()
	// defer h.roomsMutex.Unlock()

	// if h.rooms[roomID] == nil {
	// 	h.rooms[roomID] = make(map[string]*ChatClient)
	// }

	// h.rooms[roomID][client.ID] = client
	// client.Rooms[roomID] = true

	// // Broadcast user joined
	// h.broadcastToRoomExcept(roomID, client.ID, &models.WebSocketMessage{
	// 	Type:   string(models.WSEventParticipantJoin),
	// 	RoomID: roomID,
	// 	Data: map[string]interface{}{
	// 		"user_id":  client.UserID,
	// 		"username": client.Username,
	// 	},
	// 	Timestamp: time.Now(),
	// })
}

func (h *WebSocketChatHandler) leaveRoom(client *ChatClient, roomID string) {
	// h.roomsMutex.Lock()
	// defer h.roomsMutex.Unlock()

	// if clients, ok := h.rooms[roomID]; ok {
	// 	delete(clients, client.ID)
	// 	if len(clients) == 0 {
	// 		delete(h.rooms, roomID)
	// 	}
	// }

	// delete(client.Rooms, roomID)

	// // Broadcast user left
	// h.broadcastToRoomExcept(roomID, client.ID, &models.WebSocketMessage{
	// 	Type:   string(models.WSEventParticipantLeave),
	// 	RoomID: roomID,
	// 	Data: map[string]interface{}{
	// 		"user_id":  client.UserID,
	// 		"username": client.Username,
	// 	},
	// 	Timestamp: time.Now(),
	// })
}

func (h *WebSocketChatHandler) broadcastMessage(message *responses.WebSocketMessage) {
	if message.RoomID != "" {
		h.broadcastToRoom(message.RoomID, message)
	} else {
		h.broadcastToAll(message)
	}
}

func (h *WebSocketChatHandler) broadcastToRoom(roomID string, message *responses.WebSocketMessage) {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(clients, client.ID)
			}
		}
	}
}

func (h *WebSocketChatHandler) broadcastToRoomExcept(roomID, exceptClientID string, message *responses.WebSocketMessage) {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for clientID, client := range clients {
			if clientID != exceptClientID {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(clients, clientID)
				}
			}
		}
	}
}

func (h *WebSocketChatHandler) broadcastToAll(message *responses.WebSocketMessage) {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client.ID)
		}
	}
}

func (h *WebSocketChatHandler) broadcastUserStatus(userID string, isOnline bool) {
	eventType := responses.WSEventUserOnline
	if !isOnline {
		eventType = responses.WSEventUserOffline
	}

	message := &responses.WebSocketMessage{
		Type: string(eventType),
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	h.broadcastToAll(message)
}

func (h *WebSocketChatHandler) mapToStruct(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// GetActiveUsers returns list of active users in a room
func (h *WebSocketChatHandler) GetActiveUsers(roomID string) []string {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()

	var users []string
	if clients, ok := h.rooms[roomID]; ok {
		for _, client := range clients {
			users = append(users, client.UserID)
		}
	}
	return users
}

// GetOnlineUsers returns list of all online users
func (h *WebSocketChatHandler) GetOnlineUsers() []string {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()

	var users []string
	seenUsers := make(map[string]bool)

	for _, client := range h.clients {
		if !seenUsers[client.UserID] {
			users = append(users, client.UserID)
			seenUsers[client.UserID] = true
		}
	}
	return users
}
