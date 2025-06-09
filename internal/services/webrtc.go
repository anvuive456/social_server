package services

import (
	"fmt"
	"log"
	"net/http"
	"social_server/internal/models/constants"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type WebRTCService struct {
	callRepo    repositories.CallRepository
	userRepo    repositories.UserRepository
	upgrader    websocket.Upgrader
	connections map[string]*WebRTCConnection
	rooms       map[string]*CallRoom
	mutex       sync.RWMutex
}

type WebRTCConnection struct {
	ID       string
	UserID   uint
	Conn     *websocket.Conn
	CallID   *uint
	RoomID   string
	IsActive bool
}

type CallRoom struct {
	ID           string
	CallID       uint
	Participants map[uint]*WebRTCConnection
	CreatedAt    time.Time
	IsActive     bool
	mutex        sync.RWMutex
}

type WebRTCMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	CallID    *uint       `json:"call_id,omitempty"`
	UserID    *uint       `json:"user_id,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type CallInitRequest struct {
	CalleeID uint   `json:"callee_id"`
	Type     string `json:"type"` // "video" or "audio"
}

type CallResponse struct {
	CallID uint   `json:"call_id"`
	RoomID string `json:"room_id"`
	Status string `json:"status"`
}

func NewWebRTCService(callRepo repositories.CallRepository, userRepo repositories.UserRepository) *WebRTCService {
	return &WebRTCService{
		callRepo: callRepo,
		userRepo: userRepo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		connections: make(map[string]*WebRTCConnection),
		rooms:       make(map[string]*CallRoom),
	}
}

func (s *WebRTCService) InitiateCall(callerID, calleeID uint, callType string) (*CallResponse, error) {
	// Validate users exist
	caller, err := s.userRepo.GetByID(callerID)
	if err != nil {
		return nil, fmt.Errorf("caller not found: %w", err)
	}

	_, err = s.userRepo.GetByID(calleeID)
	if err != nil {
		return nil, fmt.Errorf("callee not found: %w", err)
	}

	// Check if users are friends
	isFriend, err := s.userRepo.IsFriend(callerID, calleeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check friendship: %w", err)
	}
	if !isFriend {
		return nil, fmt.Errorf("users are not friends")
	}

	// Check if caller has active call
	activeCall, err := s.callRepo.GetActiveCall(callerID)
	if err == nil && activeCall != nil {
		return nil, fmt.Errorf("caller already in active call")
	}

	// Check if callee has active call
	activeCall, err = s.callRepo.GetActiveCall(calleeID)
	if err == nil && activeCall != nil {
		return nil, fmt.Errorf("callee already in active call")
	}

	// Create call record
	roomID := uuid.New().String()
	call := &postgres.Call{
		CallerID:    callerID,
		CalleeID:    &calleeID,
		Type:        callType,
		Status:      string(constants.CallStatusRinging),
		IsGroupCall: false,
		RoomID:      roomID,
	}

	err = s.callRepo.Create(call)
	if err != nil {
		return nil, fmt.Errorf("failed to create call: %w", err)
	}

	// Create room
	room := &CallRoom{
		ID:           roomID,
		CallID:       call.ID,
		Participants: make(map[uint]*WebRTCConnection),
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	s.mutex.Lock()
	s.rooms[roomID] = room
	s.mutex.Unlock()

	// Add caller as participant
	err = s.callRepo.JoinCall(call.ID, callerID)
	if err != nil {
		return nil, fmt.Errorf("failed to add caller as participant: %w", err)
	}

	// Notify callee
	s.notifyUser(callerID, "call_incoming", map[string]interface{}{
		"call_id": call.ID,
		"room_id": roomID,
		"caller":  caller,
		"type":    callType,
		"status":  "ringing",
	})

	return &CallResponse{
		CallID: call.ID,
		RoomID: roomID,
		Status: "ringing",
	}, nil
}

func (s *WebRTCService) AcceptCall(callID, userID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.CalleeID == nil || *call.CalleeID != userID {
		return fmt.Errorf("unauthorized to accept this call")
	}

	// Update call status
	err = s.callRepo.Update(callID, map[string]interface{}{
		"status":     postgres.CallStatusOngoing,
		"started_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	// Add callee as participant
	err = s.callRepo.JoinCall(callID, userID)
	if err != nil {
		return fmt.Errorf("failed to add callee as participant: %w", err)
	}

	// Notify caller
	s.notifyUser(call.CallerID, "call_accepted", map[string]interface{}{
		"call_id": callID,
		"room_id": call.RoomID,
		"status":  "ongoing",
	})

	return nil
}

func (s *WebRTCService) DeclineCall(callID, userID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.CalleeID == nil || *call.CalleeID != userID {
		return fmt.Errorf("unauthorized to decline this call")
	}

	// Update call status
	err = s.callRepo.Update(callID, map[string]interface{}{
		"status":   postgres.CallStatusDeclined,
		"ended_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	// Notify caller
	s.notifyUser(call.CallerID, "call_declined", map[string]interface{}{
		"call_id": callID,
		"status":  "declined",
	})

	// Clean up room
	s.mutex.Lock()
	delete(s.rooms, call.RoomID)
	s.mutex.Unlock()

	return nil
}

func (s *WebRTCService) EndCall(callID, userID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.CallerID != userID && (call.CalleeID == nil || *call.CalleeID != userID) {
		return fmt.Errorf("unauthorized to end this call")
	}

	// End call
	err = s.callRepo.EndCall(callID)
	if err != nil {
		return fmt.Errorf("failed to end call: %w", err)
	}

	// Determine other user
	var otherUserID uint
	if userID == call.CallerID {
		if call.CalleeID != nil {
			otherUserID = *call.CalleeID
		}
	} else {
		otherUserID = call.CallerID
	}

	// Notify other user
	if otherUserID != 0 {
		s.notifyUser(otherUserID, "call_ended", map[string]interface{}{
			"call_id": callID,
			"status":  "ended",
		})
	}

	// Clean up room
	s.mutex.Lock()
	delete(s.rooms, call.RoomID)
	s.mutex.Unlock()

	return nil
}

func (s *WebRTCService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get user ID from query parameters or authentication
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		log.Printf("User ID not provided")
		return
	}

	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		log.Printf("Invalid user ID: %v", err)
		return
	}

	// Create connection
	connectionID := uuid.New().String()
	webrtcConn := &WebRTCConnection{
		ID:       connectionID,
		UserID:   userID,
		Conn:     conn,
		IsActive: true,
	}

	s.mutex.Lock()
	s.connections[connectionID] = webrtcConn
	s.mutex.Unlock()

	// Handle messages
	s.handleConnection(webrtcConn)

	// Clean up
	s.mutex.Lock()
	delete(s.connections, connectionID)
	s.mutex.Unlock()
}

func (s *WebRTCService) handleConnection(conn *WebRTCConnection) {
	for {
		var message WebRTCMessage
		err := conn.Conn.ReadJSON(&message)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch message.Type {
		case "offer":
			s.handleOffer(conn, message)
		case "answer":
			s.handleAnswer(conn, message)
		case "ice-candidate":
			s.handleICECandidate(conn, message)
		case "join-room":
			s.handleJoinRoom(conn, message)
		case "leave-room":
			s.handleLeaveRoom(conn, message)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

func (s *WebRTCService) handleOffer(conn *WebRTCConnection, message WebRTCMessage) {
	// Handle WebRTC offer
	s.broadcastToRoom(message.RoomID, message, conn.UserID)
}

func (s *WebRTCService) handleAnswer(conn *WebRTCConnection, message WebRTCMessage) {
	// Handle WebRTC answer
	s.broadcastToRoom(message.RoomID, message, conn.UserID)
}

func (s *WebRTCService) handleICECandidate(conn *WebRTCConnection, message WebRTCMessage) {
	// Handle ICE candidate
	s.broadcastToRoom(message.RoomID, message, conn.UserID)
}

func (s *WebRTCService) handleJoinRoom(conn *WebRTCConnection, message WebRTCMessage) {

	if message.CallID == nil {
		log.Printf("Call ID not provided for join room")
		return
	}

	call, err := s.callRepo.GetByID(*message.CallID)
	if err != nil {
		log.Printf("Failed to get call: %v", err)
		return
	}

	if call.CallerID != conn.UserID && (call.CalleeID == nil || *call.CalleeID != conn.UserID) {
		log.Printf("Unauthorized to join call")
		return
	}

	// Join call as participant
	err = s.callRepo.JoinCall(*message.CallID, conn.UserID)
	if err != nil {
		log.Printf("Failed to join call: %v", err)
		return
	}

	// Add to room
	s.mutex.Lock()
	if room, exists := s.rooms[call.RoomID]; exists {
		room.mutex.Lock()
		room.Participants[conn.UserID] = conn
		room.mutex.Unlock()
		conn.CallID = message.CallID
		conn.RoomID = call.RoomID
	}
	s.mutex.Unlock()
}

func (s *WebRTCService) handleLeaveRoom(conn *WebRTCConnection, message WebRTCMessage) {

	if message.CallID == nil {
		return
	}

	call, err := s.callRepo.GetByID(*message.CallID)
	if err != nil {
		log.Printf("Failed to get call: %v", err)
		return
	}

	if call.CallerID != conn.UserID && (call.CalleeID == nil || *call.CalleeID != conn.UserID) {
		return
	}

	// Leave call
	err = s.callRepo.LeaveCall(*message.CallID, conn.UserID)
	if err != nil {
		log.Printf("Failed to leave call: %v", err)
		return
	}

	// Remove from room
	s.mutex.Lock()
	if room, exists := s.rooms[call.RoomID]; exists {
		room.mutex.Lock()
		delete(room.Participants, conn.UserID)
		room.mutex.Unlock()
	}
	s.mutex.Unlock()

	conn.CallID = nil
	conn.RoomID = ""
}

func (s *WebRTCService) GetUserCalls(userID uint, cursor *paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error) {
	return s.callRepo.GetUserCalls(userID, *cursor, limit)
}

func (s *WebRTCService) GetActiveCalls(userID uint) ([]postgres.Call, error) {
	// Get active call for user
	activeCall, err := s.callRepo.GetActiveCall(userID)
	if err != nil {
		return []postgres.Call{}, nil // No active calls
	}

	return []postgres.Call{*activeCall}, nil
}

func (s *WebRTCService) broadcastToRoom(roomID string, message WebRTCMessage, excludeUserID uint) {
	s.mutex.RLock()
	room, exists := s.rooms[roomID]
	s.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	message.Timestamp = time.Now()

	for userID, conn := range room.Participants {
		if userID != excludeUserID && conn.IsActive {
			err := conn.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("Failed to send message to user %d: %v", userID, err)
				conn.IsActive = false
			}
		}
	}
}

func (s *WebRTCService) notifyUser(userID uint, messageType string, data interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	message := WebRTCMessage{
		Type:      messageType,
		Data:      data,
		UserID:    &userID,
		Timestamp: time.Now(),
	}

	for _, conn := range s.connections {
		if conn.UserID == userID && conn.IsActive {
			err := conn.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("Failed to notify user %d: %v", userID, err)
				conn.IsActive = false
			}
		}
	}
}

func (s *WebRTCService) GetCallByID(callID, userID uint) (*postgres.Call, error) {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return nil, fmt.Errorf("call not found: %w", err)
	}

	if call.CallerID != userID && (call.CalleeID == nil || *call.CalleeID != userID) {
		return nil, fmt.Errorf("unauthorized to access this call")
	}

	return call, nil
}

func (s *WebRTCService) UpdateCallStatus(callID, userID uint, status string) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.CallerID != userID && (call.CalleeID == nil || *call.CalleeID != userID) {
		return fmt.Errorf("unauthorized to update this call")
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if status == string(constants.CallStatusEnded) {
		updates["ended_at"] = time.Now()
	} else if status == string(constants.CallStatusOngoing) {
		updates["started_at"] = time.Now()
	}

	return s.callRepo.Update(callID, updates)
}

func (s *WebRTCService) GetCallParticipants(callID, userID uint) ([]postgres.CallParticipant, error) {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return nil, fmt.Errorf("call not found: %w", err)
	}

	if call.CallerID != userID && (call.CalleeID == nil || *call.CalleeID != userID) {
		return nil, fmt.Errorf("unauthorized to access call participants")
	}

	return s.callRepo.GetCallParticipants(callID)
}
