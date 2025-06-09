package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social_server/internal/middleware"
	"social_server/internal/models"
	"social_server/internal/services"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	authService *services.AuthService
	callService *services.CallService
	// onlineStatusService *services.OnlineStatusService
	connections map[string]*WebSocketConnection
	rooms       map[string]*models.Room
	userConns   map[uint]string // userID -> connectionID
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
}

type WebSocketConnection struct {
	Conn      *websocket.Conn
	UserID    uint
	ConnID    string
	RoomID    string
	IsActive  bool
	LastPing  time.Time
	SendChan  chan []byte
	CloseChan chan bool
}

func NewWebSocketHandler(authService *services.AuthService, callService *services.CallService,

// onlineStatusService *services.OnlineStatusService
) *WebSocketHandler {
	handler := &WebSocketHandler{
		authService: authService,
		callService: callService,
		// onlineStatusService: onlineStatusService,
		connections: make(map[string]*WebSocketConnection),
		rooms:       make(map[string]*models.Room),
		userConns:   make(map[uint]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	// Register callback for online status changes
	// onlineStatusService.RegisterStatusCallback(handler.handleOnlineStatusChange)

	return handler
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"error":   "cannot_upgrade_websocket",
			"message": fmt.Sprintf("failed to upgrade to WebSocket: %v", err),
		})
		return
	}

	// Create WebSocket connection
	connID := h.generateConnectionID()
	wsConn := &WebSocketConnection{
		Conn:      conn,
		UserID:    userID,
		ConnID:    connID,
		IsActive:  true,
		LastPing:  time.Now(),
		SendChan:  make(chan []byte, 256),
		CloseChan: make(chan bool),
	}

	// Register connection
	h.registerConnection(wsConn)

	// Set user online in status service
	// err = h.onlineStatusService.SetUserOnline(userID, connID)
	// if err != nil {
	// 	log.Printf("Failed to set user %d online: %v", userID, err)
	// }

	// Start connection handlers
	go h.handleConnection(wsConn)
	go h.handleSender(wsConn)

	log.Printf("WebSocket connection established for user %d with conn ID %s", userID, connID)
	h.sendToConnection(wsConn, models.WSMessage{
		Type: models.MessageTypeConnected,
	})
}

func (h *WebSocketHandler) registerConnection(conn *WebSocketConnection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Close existing connection if exists
	if existingConnID, exists := h.userConns[conn.UserID]; exists {
		if existingConn, ok := h.connections[existingConnID]; ok {
			existingConn.CloseChan <- true
		}
	}

	h.connections[conn.ConnID] = conn
	h.userConns[conn.UserID] = conn.ConnID
}

func (h *WebSocketHandler) unregisterConnection(connID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if conn, exists := h.connections[connID]; exists {
		// Remove from room if in one
		if conn.RoomID != "" {
			h.removeFromRoom(conn.RoomID, conn.UserID)
		}

		// Set user offline in status service
		// err := h.onlineStatusService.SetUserOffline(conn.UserID)
		// if err != nil {
		// 	log.Printf("Failed to set user %d offline: %v", conn.UserID, err)
		// }

		delete(h.connections, connID)
		delete(h.userConns, conn.UserID)
		conn.Conn.Close()
	}
}

func (h *WebSocketHandler) handleConnection(conn *WebSocketConnection) {
	defer h.unregisterConnection(conn.ConnID)

	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.LastPing = time.Now()
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-conn.CloseChan:
			println("Connection closed: CloseChan")
			return
		default:
			_, messageBytes, err := conn.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error on handle connection: %v", err)
				}
				return
			}

			var message models.WSMessage
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				println(fmt.Sprintf("Invalid message format: %v", messageBytes))
				h.sendError(conn, "invalid_message", "Invalid message format")
				continue
			}

			message.From = conn.UserID
			message.Timestamp = time.Now().Format(time.RFC3339)

			h.handleMessage(conn, &message)
		}
	}
}

func (h *WebSocketHandler) handleSender(conn *WebSocketConnection) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message := <-conn.SendChan:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-conn.CloseChan:
			return
		}
	}
}

func (h *WebSocketHandler) handleMessage(conn *WebSocketConnection, message *models.WSMessage) {
	switch message.Type {
	case models.MessageTypeJoinRoom:
		h.handleJoinRoom(conn, message)
	case models.MessageTypeLeaveRoom:
		h.handleLeaveRoom(conn, message)
	case models.MessageTypeCallRequest:
		h.handleCallRequest(conn, message)
	case models.MessageTypeCallResponse:
		h.handleCallResponse(conn, message)
	case models.MessageTypeOffer:
		h.handleOffer(conn, message)
	case models.MessageTypeAnswer:
		h.handleAnswer(conn, message)
	case models.MessageTypeICECandidate:
		h.handleICECandidate(conn, message)
	case models.MessageTypeCallEnd:
		h.handleCallEnd(conn, message)
	case models.MessageTypeHeartbeat:
		h.handleHeartbeat(conn, message)
	default:
		h.sendError(conn, "unknown_message_type", "Unknown message type")
	}
}

func (h *WebSocketHandler) handleJoinRoom(conn *WebSocketConnection, message *models.WSMessage) {
	var joinMsg models.JoinRoomMessage
	if err := json.Unmarshal(message.Data, &joinMsg); err != nil {
		h.sendError(conn, "invalid_join_room_data", "Invalid join room data")
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Leave current room if in one
	if conn.RoomID != "" {
		h.removeFromRoom(conn.RoomID, conn.UserID)
	}

	// Join new room
	room, exists := h.rooms[joinMsg.RoomID]
	if !exists {
		room = &models.Room{
			ID:           joinMsg.RoomID,
			Type:         joinMsg.CallType,
			Status:       "waiting",
			Participants: []uint{},
			CreatedAt:    time.Now(),
			MaxUsers:     10, // Default max users
		}
		h.rooms[joinMsg.RoomID] = room
	}

	// Add user to room
	room.Participants = append(room.Participants, conn.UserID)
	conn.RoomID = joinMsg.RoomID

	// Notify other participants
	h.broadcastToRoom(joinMsg.RoomID, models.WSMessage{
		Type:      models.MessageTypeUserJoined,
		From:      conn.UserID,
		RoomID:    joinMsg.RoomID,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: h.marshalData(models.UserJoinedMessage{
			UserID: conn.UserID,
			RoomID: joinMsg.RoomID,
		}),
	}, conn.UserID)

	log.Printf("User %d joined room %s", conn.UserID, joinMsg.RoomID)
}

func (h *WebSocketHandler) handleLeaveRoom(conn *WebSocketConnection, message *models.WSMessage) {
	if conn.RoomID == "" {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.removeFromRoom(conn.RoomID, conn.UserID)
	conn.RoomID = ""
}

func (h *WebSocketHandler) handleCallRequest(conn *WebSocketConnection, message *models.WSMessage) {

	var callReq models.CallRequestMessage
	if err := json.Unmarshal(message.Data, &callReq); err != nil {
		h.sendError(conn, "invalid_call_request", "Invalid call request data")
		return
	}

	// Create call in database
	caller, err := h.callService.CreateCall(conn.UserID, callReq.CalleeID, callReq.CallType, callReq.RoomID)
	if err != nil {
		h.sendError(conn, "call_creation_failed", err.Error())
		return
	}

	// Send call request to callee
	h.sendToUser(callReq.CalleeID, models.WSMessage{
		Type:      models.MessageTypeCallRequest,
		From:      conn.UserID,
		To:        callReq.CalleeID,
		RoomID:    callReq.RoomID,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: h.marshalData(models.CallRequestMessage{
			CallID:   caller.ID,
			CallerID: caller.CallerID,
			CalleeID: *caller.CalleeID,
			CallType: caller.Type,
			RoomID:   caller.RoomID,
		}),
	})

	log.Printf("Call request sent from user %d to user %d", conn.UserID, callReq.CalleeID)
}

func (h *WebSocketHandler) handleCallResponse(conn *WebSocketConnection, message *models.WSMessage) {
	var callResp models.CallResponseMessage
	if err := json.Unmarshal(message.Data, &callResp); err != nil {
		h.sendError(conn, "invalid_call_response", "Invalid call response data")
		return
	}

	// Update call status
	if callResp.Response == "accept" {
		err := h.callService.AcceptCall(callResp.CallID)
		if err != nil {
			h.sendError(conn, "call_accept_failed", err.Error())
			return
		}
	} else {
		err := h.callService.DeclineCall(callResp.CallID)
		if err != nil {
			h.sendError(conn, "call_decline_failed", err.Error())
			return
		}
	}

	// Broadcast response to room
	h.broadcastToRoom(callResp.RoomID, models.WSMessage{
		Type:      models.MessageTypeCallResponse,
		From:      conn.UserID,
		RoomID:    callResp.RoomID,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      message.Data,
	}, 0)
}

func (h *WebSocketHandler) handleOffer(conn *WebSocketConnection, message *models.WSMessage) {
	if message.To == 0 {
		h.sendError(conn, "missing_recipient", "Recipient user ID is required")
		return
	}

	// Forward offer to recipient
	h.sendToUser(message.To, *message)
}

func (h *WebSocketHandler) handleAnswer(conn *WebSocketConnection, message *models.WSMessage) {
	if message.To == 0 {
		h.sendError(conn, "missing_recipient", "Recipient user ID is required")
		return
	}

	// Forward answer to recipient
	h.sendToUser(message.To, *message)
}

func (h *WebSocketHandler) handleICECandidate(conn *WebSocketConnection, message *models.WSMessage) {
	if message.To == 0 {
		h.sendError(conn, "missing_recipient", "Recipient user ID is required")
		return
	}

	// Forward ICE candidate to recipient
	h.sendToUser(message.To, *message)
}

func (h *WebSocketHandler) handleCallEnd(conn *WebSocketConnection, message *models.WSMessage) {
	var callEnd models.CallEndMessage
	if err := json.Unmarshal(message.Data, &callEnd); err != nil {
		h.sendError(conn, "invalid_call_end", "Invalid call end data")
		return
	}

	// End call in database
	err := h.callService.EndCall(callEnd.CallID)
	if err != nil {
		h.sendError(conn, "call_end_failed", err.Error())
		return
	}

	// Broadcast call end to room
	h.broadcastToRoom(callEnd.RoomID, models.WSMessage{
		Type:      models.MessageTypeCallEnd,
		From:      conn.UserID,
		RoomID:    callEnd.RoomID,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      message.Data,
	}, 0)
}

func (h *WebSocketHandler) handleHeartbeat(conn *WebSocketConnection, message *models.WSMessage) {
	conn.LastPing = time.Now()

	// Update heartbeat in online status service
	// err := h.onlineStatusService.UpdateHeartbeat(conn.UserID)
	// if err != nil {
	// 	log.Printf("Failed to update heartbeat for user %d: %v", conn.UserID, err)
	// }

	// Send heartbeat response
	response := models.WSMessage{
		Type:      models.MessageTypeHeartbeat,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: h.marshalData(models.HeartbeatMessage{
			Status:    "ok",
			Message:   "Tình tịch",
			Timestamp: time.Now(),
		}),
	}

	h.sendToConnection(conn, response)
}

func (h *WebSocketHandler) sendToUser(userID uint, message models.WSMessage) {
	h.mutex.RLock()
	connID, exists := h.userConns[userID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	h.mutex.RLock()
	conn, exists := h.connections[connID]
	h.mutex.RUnlock()

	if !exists || !conn.IsActive {
		return
	}

	h.sendToConnection(conn, message)
}

func (h *WebSocketHandler) sendToConnection(conn *WebSocketConnection, message models.WSMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case conn.SendChan <- messageBytes:
	default:
		close(conn.SendChan)
		h.unregisterConnection(conn.ConnID)
	}
}

func (h *WebSocketHandler) broadcastToRoom(roomID string, message models.WSMessage, exclude uint) {
	h.mutex.RLock()
	room, exists := h.rooms[roomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	for _, userID := range room.Participants {
		if exclude != 0 && userID == exclude {
			continue
		}
		h.sendToUser(userID, message)
	}
}

func (h *WebSocketHandler) removeFromRoom(roomID string, userID uint) {
	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	// Remove user from participants
	for i, participantID := range room.Participants {
		if participantID == userID {
			room.Participants = append(room.Participants[:i], room.Participants[i+1:]...)
			break
		}
	}

	// Notify other participants
	h.broadcastToRoom(roomID, models.WSMessage{
		Type:      models.MessageTypeUserLeft,
		From:      userID,
		RoomID:    roomID,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: h.marshalData(models.UserLeftMessage{
			UserID: userID,
			RoomID: roomID,
		}),
	}, userID)

	// Remove room if empty
	if len(room.Participants) == 0 {
		delete(h.rooms, roomID)
	}

	log.Printf("User %d left room %s", userID, roomID)
}

func (h *WebSocketHandler) sendError(conn *WebSocketConnection, code, message string) {
	errorMsg := models.WSMessage{
		Type:      models.MessageTypeError,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: h.marshalData(models.ErrorMessage{
			Code:    code,
			Message: message,
		}),
	}
	h.sendToConnection(conn, errorMsg)
}

func (h *WebSocketHandler) marshalData(data interface{}) json.RawMessage {
	bytes, _ := json.Marshal(data)
	return json.RawMessage(bytes)
}

func (h *WebSocketHandler) generateConnectionID() string {
	return fmt.Sprintf("conn_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// GetWebRTCConfig returns STUN/TURN server configuration
func (h *WebSocketHandler) GetWebRTCConfig() models.WebRTCConfig {
	return models.WebRTCConfig{
		ICEServers: []models.ICEServerConfig{
			{
				URLs: []string{
					"stun:stun.l.google.com:19302",
					"stun:stun1.l.google.com:19302",
					"stun:stun2.l.google.com:19302",
				},
			},
			// Add TURN servers here if needed
			{
				URLs: []string{
					"turn:14.224.203.223:3478?transport=udp",
					"turn:14.224.203.223:3478?transport=tcp",
				},
				Username:   "webrtcuser",
				Credential: "webrtccredential",
			},
		},
	}
}

// Health check for WebSocket connections
func (h *WebSocketHandler) GetConnectionStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(h.connections),
		"active_rooms":      len(h.rooms),
		"connected_users":   len(h.userConns),
	}

	// Add online status service stats
	// if h.onlineStatusService != nil {
	// 	onlineStats := h.onlineStatusService.GetConnectionStats()
	// 	for key, value := range onlineStats {
	// 		stats[key] = value
	// 	}
	// }

	return stats
}

// handleOnlineStatusChange broadcasts online status changes to relevant users
func (h *WebSocketHandler) handleOnlineStatusChange(update services.OnlineStatusUpdate) {
	// Create websocket message for online status change
	// message := models.WSMessage{
	// 	Type:      models.MessageTypeUserOnlineStatus,
	// 	From:      update.UserID,
	// 	Timestamp: time.Now().Format(time.RFC3339),
	// 	Data: h.marshalData(models.UserOnlineStatusMessage{
	// 		UserID:   update.UserID,
	// 		IsOnline: update.IsOnline,
	// 		LastSeen: update.LastSeen,
	// 		Username: update.Username,
	// 	}),
	// }

	// Get online friends of this user to notify them
	// friendIDs, err := h.onlineStatusService.GetOnlineFriends(update.UserID)
	// if err != nil {
	// 	log.Printf("Failed to get online friends for user %d: %v", update.UserID, err)
	// 	return
	// }

	// Send status update to online friends
	// for _, friendID := range friendIDs {
	// 	h.sendToUser(friendID, message)
	// }

	// log.Printf("Broadcasted online status change for user %d (online: %v) to %d friends",
	// update.UserID, update.IsOnline, len(friendIDs))
}
