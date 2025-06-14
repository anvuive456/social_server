package services

import (
	"fmt"
	"social_server/internal/models"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type ChatService struct {
	repos *repositories.Repositories
}

func NewChatService(repos *repositories.Repositories) *ChatService {
	return &ChatService{
		repos: repos,
	}
}

func (s *ChatService) SyncRooms(userID uint, req requests.SyncChatRoomsRequest) (*responses.SyncChatRoomsResponse, error) {
	rooms, count, err := s.repos.ChatRoom.GetUserChatRoomsByUserIDAndLastRoomID(userID, req.LastID)
	if err != nil {
		return nil, fmt.Errorf("failed to sync chat rooms: %w", err)
	}

	roomSummaries := make([]responses.ChatRoomSummary, len(rooms))
	for i, room := range rooms {
		name := room.Name
		if room.Type == postgres.ChatRoomTypeGroup {
			name = room.Name
		} else {
			for _, participant := range room.Participants {
				if participant.UserID != userID {
					name = participant.User.Profile.DisplayName
					break
				}
			}
		}

		avatar := room.Avatar
		if room.Type == postgres.ChatRoomTypePrivate {
			for _, participant := range room.Participants {
				if participant.UserID != userID {
					avatar = participant.User.Profile.Avatar
					break
				}
			}
		}

		unreadCount := 0
		for _, message := range room.Messages {
			read := false
			for _, readBy := range message.ReadBy {
				if readBy.UserID == userID {
					read = true
				}
			}
			if !read {
				unreadCount++
			}
		}
		var lastMessage *postgres.Message
		var lastActivity *time.Time
		if len(room.Messages) > 0 {
			lastMessage = &room.Messages[0]
			lastActivity = &lastMessage.UpdatedAt
		}

		roomSummaries[i] = responses.ChatRoomSummary{
			ID:               room.ID,
			LocalID:          room.LocalID,
			Avatar:           avatar,
			Name:             name,
			Type:             room.Type,
			LastMessage:      lastMessage,
			LastActivity:     lastActivity,
			CreatedAt:        room.CreatedAt,
			UpdatedAt:        room.UpdatedAt,
			ParticipantCount: len(room.Participants),
			UnreadCount:      unreadCount,
		}
	}

	return &responses.SyncChatRoomsResponse{
		Rooms: roomSummaries,
		Count: count,
	}, nil
}

func (s *ChatService) GetRoomByUserId(userID uint, roomID uint) (*responses.ChatRoomSummary, error) {
	room, err := s.repos.ChatRoom.GetByUserID(userID, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}
	var avatar string
	var name string
	var lastActivity time.Time
	var unreadCount int

	if room.Type == postgres.ChatRoomTypePrivate {
		for _, participant := range room.Participants {
			if participant.ID == userID {
				continue
			}
			avatar = participant.User.Profile.Avatar
			name = participant.User.Profile.DisplayName
			lastActivity = time.Now().UTC()
			unreadCount = 0
			break
		}
	} else if room.Type == postgres.ChatRoomTypeGroup {
		avatar = room.Avatar
		name = room.Name
		lastActivity = time.Now()
		unreadCount = 0
	}

	return &responses.ChatRoomSummary{
		ID:      room.ID,
		LocalID: room.LocalID,

		Avatar:           avatar,
		Name:             name,
		Type:             room.Type,
		LastMessage:      nil,
		LastActivity:     &lastActivity,
		CreatedAt:        room.CreatedAt,
		UpdatedAt:        room.UpdatedAt,
		ParticipantCount: len(room.Participants),
		UnreadCount:      unreadCount,
	}, nil
}

// Room operations
func (s *ChatService) CreateRoom(creatorID uint, req *requests.CreateChatRoomRequest) (*responses.ChatRoomSummary, error) {

	room, err := s.repos.ChatRoom.Create(req.Name, creatorID, req.Type, req.Participants, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}
	var avatar string
	var name string
	var lastActivity time.Time
	var unreadCount int

	if room.Type == postgres.ChatRoomTypePrivate {
		for _, participant := range room.Participants {
			if participant.ID == creatorID {
				continue
			}
			avatar = participant.User.Profile.Avatar
			name = participant.User.Profile.DisplayName
			lastActivity = time.Now().UTC()
			unreadCount = 0
			break
		}
	} else if room.Type == postgres.ChatRoomTypeGroup {
		avatar = room.Avatar
		name = room.Name
		lastActivity = time.Now()
		unreadCount = 0
	}

	return &responses.ChatRoomSummary{
		ID:      room.ID,
		LocalID: room.LocalID,

		Avatar:           avatar,
		Name:             name,
		Type:             room.Type,
		LastMessage:      nil,
		LastActivity:     &lastActivity,
		CreatedAt:        room.CreatedAt,
		UpdatedAt:        room.UpdatedAt,
		ParticipantCount: len(room.Participants),
		UnreadCount:      unreadCount,
	}, nil
}
func (s *ChatService) CreateRoomFromWs(creatorID uint, req *models.CreateChatRoomMessage) (*postgres.ChatRoom, error) {

	room, err := s.repos.ChatRoom.Create(req.Name, creatorID, req.Type, req.ParticipantIDs, &req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	return room, nil
}

func (s *ChatService) DeleteRoom(userID uint, roomID uint) error {
	err := s.repos.ChatRoom.Delete(userID, roomID)
	if err != nil {
		return fmt.Errorf("failed to delete chat room: %w", err)
	}

	return nil
}

func (s *ChatService) GetRoomMembers(roomID uint) ([]postgres.User, error) {
	members, err := s.repos.ChatRoom.GetParticipants(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat room members: %w", err)
	}

	users := make([]postgres.User, len(members))
	for i, member := range members {
		users[i] = member.User
	}

	return users, nil
}

func (s *ChatService) GetUserRooms(userID uint, req requests.GetChatRoomsRequest) (*responses.ChatRoomsResponse, error) {
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 5
	}
	cursor := paginator.Cursor{
		Before: &req.Before,
		After:  &req.After,
	}
	rooms, next, err := s.repos.ChatRoom.GetUserRooms(userID, req.Archive, cursor, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	return &responses.ChatRoomsResponse{
		Conversations: rooms,
		NextCursor:    &next,
	}, nil
}

func (s *ChatService) GetPrivateRoom(userID1, userID2 uint) (*postgres.ChatRoom, error) {
	room, err := s.repos.ChatRoom.GetPrivateRoom(userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get private room: %w", err)
	}
	return room, nil
}

// Message operations
// func (s *ChatService) SendMessage(message *postgres.Message) error {
// 	// Validate message
// 	if message.ChatRoomID == 0 || message.SenderID == 0 {
// 		return fmt.Errorf("invalid message: room ID and sender ID are required")
// 	}

// 	if message.Content == "" && message.Type == postgres.MessageTypeText {
// 		return fmt.Errorf("text message must have content")
// 	}

// 	// Set timestamps
// 	message.CreatedAt = time.Now()
// 	message.UpdatedAt = time.Now()

// 	// Create message
// 	created, err := s.repos.Message.Create(message)
// 	if err != nil {
// 		return fmt.Errorf("failed to send message: %w", err)
// 	}

// 	// Update room last activity
// 	err = s.repos.ChatRoom.UpdateLastActivity(message.ChatRoomID, time.Now())
// 	if err != nil {
// 		return fmt.Errorf("failed to update room activity: %w", err)
// 	}

// 	return nil
// }

func (s *ChatService) GetMessage(messageID uint) (*postgres.Message, error) {
	message, err := s.repos.Message.GetByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return message, nil
}

func (s *ChatService) GetRoomMessages(roomID, userID uint, limit int, before string, after string) (*responses.MessageResponse, error) {
	cursorObj := paginator.Cursor{
		Before: &before,
		After:  &after,
	}

	messages, next, err := s.repos.Message.GetRoomMessages(roomID, cursorObj, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get room messages: %w", err)
	}

	return &responses.MessageResponse{
		Messages:   messages,
		NextCursor: &next,
	}, nil
}

func (s *ChatService) MarkMessageAsRead(messageID, userID uint) error {
	err := s.repos.Message.MarkAsRead(messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}
	return nil
}

func (s *ChatService) MarkMessageAsDelivered(messageID, userID uint) error {
	err := s.repos.Message.MarkAsDelivered(messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark message as delivered: %w", err)
	}
	return nil
}

func (s *ChatService) DeleteMessage(messageID, userID uint) error {
	// Check if user owns the message
	message, err := s.repos.Message.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if message.SenderID != userID {
		// Check if user is admin
		participant, err := s.repos.Participant.GetByRoomAndUser(message.ChatRoomID, userID)
		if err != nil || participant.Role != postgres.ParticipantRoleAdmin {
			return fmt.Errorf("permission denied: only sender or admin can delete message")
		}
	}

	err = s.repos.Message.Delete(messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func (s *ChatService) AddReaction(messageID, userID uint, emoji string) error {
	err := s.repos.Message.AddReaction(messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

func (s *ChatService) RemoveReaction(messageID, userID uint, emoji string) error {
	err := s.repos.Message.RemoveReaction(messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	return nil
}

// Participant operations
func (s *ChatService) AddParticipant(roomID, userID, addedBy uint) error {
	// Check if user adding has permission
	participant, err := s.repos.Participant.GetByRoomAndUser(roomID, addedBy)
	if err != nil {
		return fmt.Errorf("permission denied: user not in room")
	}

	// Check room settings
	room, err := s.repos.ChatRoom.GetByID(roomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	if room.Settings.OnlyAdminsCanInvite && participant.Role != postgres.ParticipantRoleAdmin && participant.Role != postgres.ParticipantRoleOwner {
		return fmt.Errorf("permission denied: only admins can invite users")
	}

	// Add participant
	newParticipant := &postgres.Participant{
		ChatRoomID: roomID,
		UserID:     userID,
		Role:       postgres.ParticipantRoleMember,
		JoinedAt:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.repos.ChatRoom.AddParticipant(newParticipant)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

func (s *ChatService) RemoveParticipant(roomID, userID, removedBy uint) error {
	// Check if user removing has permission
	removerParticipant, err := s.repos.Participant.GetByRoomAndUser(roomID, removedBy)
	if err != nil {
		return fmt.Errorf("permission denied: user not in room")
	}

	targetParticipant, err := s.repos.Participant.GetByRoomAndUser(roomID, userID)
	if err != nil {
		return fmt.Errorf("user not found in room")
	}

	// Users can remove themselves
	if userID != removedBy {
		// Only admins and owners can remove others
		if removerParticipant.Role != postgres.ParticipantRoleAdmin && removerParticipant.Role != postgres.ParticipantRoleOwner {
			return fmt.Errorf("permission denied: only admins can remove users")
		}

		// Owners cannot be removed by admins
		if targetParticipant.Role == postgres.ParticipantRoleOwner && removerParticipant.Role != postgres.ParticipantRoleOwner {
			return fmt.Errorf("permission denied: cannot remove owner")
		}
	}

	err = s.repos.ChatRoom.RemoveParticipant(roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

func (s *ChatService) GetParticipants(roomID uint) ([]postgres.Participant, error) {
	participants, err := s.repos.ChatRoom.GetParticipants(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	return participants, nil
}

func (s *ChatService) UpdateParticipantRole(roomID, userID, updatedBy uint, role postgres.ParticipantRole) error {
	// Check if user updating has permission
	updaterParticipant, err := s.repos.Participant.GetByRoomAndUser(roomID, updatedBy)
	if err != nil {
		return fmt.Errorf("permission denied: user not in room")
	}

	// Only owners can change roles
	if updaterParticipant.Role != postgres.ParticipantRoleOwner {
		return fmt.Errorf("permission denied: only owners can change roles")
	}

	// Cannot change owner role
	if role == postgres.ParticipantRoleOwner {
		return fmt.Errorf("permission denied: cannot assign owner role")
	}

	err = s.repos.Participant.UpdateRole(roomID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update participant role: %w", err)
	}

	return nil
}

// Typing and presence
func (s *ChatService) SetTyping(roomID, userID uint, isTyping bool) error {
	err := s.repos.TypingIndicator.SetTyping(roomID, userID, isTyping)
	if err != nil {
		return fmt.Errorf("failed to set typing status: %w", err)
	}
	return nil
}

func (s *ChatService) GetTypingUsers(roomID uint) ([]uint, error) {
	userIDs, err := s.repos.TypingIndicator.GetTypingUsers(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get typing users: %w", err)
	}
	return userIDs, nil
}

func (s *ChatService) SetUserOnline(userID uint, isOnline bool) error {
	err := s.repos.OnlineStatus.SetOnline(userID, isOnline)
	if err != nil {
		return fmt.Errorf("failed to set user online status: %w", err)
	}
	return nil
}

func (s *ChatService) GetUserStatus(userID uint) (*postgres.OnlineStatus, error) {
	status, err := s.repos.OnlineStatus.GetStatus(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user status: %w", err)
	}
	return status, nil
}

func (s *ChatService) CheckUserPermission(roomID, userID uint, action string) error {
	participant, err := s.repos.Participant.GetByRoomAndUser(roomID, userID)
	if err != nil {
		return fmt.Errorf("user not in room")
	}

	switch action {
	case "send_message":
		// Check room settings
		// This would need room data to check settings
		return nil
	case "add_participant":
		if participant.Role != postgres.ParticipantRoleAdmin && participant.Role != postgres.ParticipantRoleOwner {
			return fmt.Errorf("permission denied")
		}
	case "remove_participant":
		if participant.Role != postgres.ParticipantRoleAdmin && participant.Role != postgres.ParticipantRoleOwner {
			return fmt.Errorf("permission denied")
		}
	}

	return nil
}

// Search operations
func (s *ChatService) SearchRooms(userID uint, query string, limit int) ([]responses.ChatRoomSummary, error) {
	rooms, err := s.repos.ChatRoom.SearchRooms(userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search rooms: %w", err)
	}
	return rooms, nil
}

func (s *ChatService) SearchMessages(roomID uint, query string, limit int) ([]responses.MessageSearchResult, error) {
	if roomID == 0 {
		return nil, fmt.Errorf("room ID is required")
	}

	results, err := s.repos.Message.SearchMessages(roomID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	return results, nil
}

func (s *ChatService) GetUnreadCount(roomID, userID uint) (int, error) {
	count, err := s.repos.Message.GetUnreadCount(roomID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// Room management
func (s *ChatService) ArchiveRoom(roomID, userID uint) error {
	// Check permission
	participant, err := s.repos.Participant.GetByRoomAndUser(roomID, userID)
	if err != nil {
		return fmt.Errorf("permission denied: user not in room")
	}

	if participant.Role != postgres.ParticipantRoleOwner {
		return fmt.Errorf("permission denied: only owner can archive room")
	}

	err = s.repos.ChatRoom.ArchiveRoom(roomID)
	if err != nil {
		return fmt.Errorf("failed to archive room: %w", err)
	}
	return nil
}

func (s *ChatService) UpdateRoom(roomID, userID uint, updates map[string]interface{}) error {
	// Check permission
	participant, err := s.repos.Participant.GetByRoomAndUser(roomID, userID)
	if err != nil {
		return fmt.Errorf("permission denied: user not in room")
	}

	if participant.Role != postgres.ParticipantRoleAdmin && participant.Role != postgres.ParticipantRoleOwner {
		return fmt.Errorf("permission denied: only admins can update room")
	}

	err = s.repos.ChatRoom.Update(roomID, updates)
	if err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}
	return nil
}

func (s *ChatService) CreateMessage(senderID uint, req models.SendChatMessageMessage) (*postgres.Message, error) {
	message, err := s.repos.Message.Create(req.Content, req.LocalID, senderID, req.RoomID, req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	return message, nil
}
