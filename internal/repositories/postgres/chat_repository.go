package postgres

import (
	"fmt"
	"strings"
	"time"

	"social_server/internal/models/postgres"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"

	"gorm.io/gorm"
)

type chatRoomRepository struct {
	db *gorm.DB
}

func NewChatRoomRepository(db *gorm.DB) repositories.ChatRoomRepository {
	return &chatRoomRepository{
		db: db,
	}
}

func (r *chatRoomRepository) AddParticipant(participant *postgres.Participant) error {
	return r.db.Create(participant).Error
}

func (r *chatRoomRepository) Create(room *postgres.ChatRoom) error {
	return r.db.Create(room).Error
}

func (r *chatRoomRepository) GetByID(id uint) (*postgres.ChatRoom, error) {
	var room postgres.ChatRoom
	err := r.db.
		Preload("Creator").
		Preload("Participants").
		Preload("Participants.User").
		First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *chatRoomRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.ChatRoom{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *chatRoomRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.ChatRoom{}, id).Error
}

func (r *chatRoomRepository) GetUserRooms(userID uint) ([]responses.ChatRoomSummary, error) {
	var rooms []postgres.ChatRoom
	db := r.db.
		Joins("JOIN participants ON chat_rooms.id = participants.chat_room_id").
		Where("participants.user_id = ? AND chat_rooms.is_archived = ?", userID, false).
		Preload("Participants").
		Order("chat_rooms.last_activity DESC")

	err := db.Find(&rooms).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]responses.ChatRoomSummary, 0, len(rooms))

	for _, room := range rooms {
		var participantCount int64
		r.db.
			Model(&postgres.Participant{}).
			Where("chat_room_id = ?", room.ID).
			Count(&participantCount)

		var lastMessage *postgres.Message
		r.db.
			Where("chat_room_id = ?", room.ID).
			Order("created_at DESC").
			First(&lastMessage)

		var unreadCount int64
		r.db.
			Model(&postgres.Message{}).
			Joins("LEFT JOIN message_reads ON messages.id = message_reads.message_id AND message_reads.user_id = ?", userID).
			Where("messages.chat_room_id = ? AND message_reads.id IS NULL", room.ID).
			Count(&unreadCount)

		summary := responses.ChatRoomSummary{
			ID:               room.ID,
			Name:             room.Name,
			Type:             room.Type,
			Avatar:           room.Avatar,
			ParticipantCount: int(participantCount),
			LastMessage:      lastMessage,
			LastActivity:     room.LastActivity,
			UnreadCount:      int(unreadCount),
			IsMuted:          false,
			CreatedAt:        room.CreatedAt,
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r *chatRoomRepository) GetPrivateRoom(userID1, userID2 uint) (*postgres.ChatRoom, error) {
	var room postgres.ChatRoom
	err := r.db.
		Where("type = ?", postgres.ChatRoomTypePrivate).
		Joins("JOIN participants p1 ON chat_rooms.id = p1.chat_room_id AND p1.user_id = ?", userID1).
		Joins("JOIN participants p2 ON chat_rooms.id = p2.chat_room_id AND p2.user_id = ?", userID2).
		First(&room).Error

	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *chatRoomRepository) RemoveParticipant(roomID, userID uint) error {
	return r.db.
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Delete(&postgres.Participant{}).Error
}

func (r *chatRoomRepository) GetParticipants(roomID uint) ([]postgres.User, error) {
	var participants []postgres.User
	err := r.db.Model(&postgres.User{}).
		Where("chat_room_id = ?", roomID).
		Preload("User").
		Preload("User.Profiles").
		Find(&participants).Error

	if err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *chatRoomRepository) UpdateLastActivity(roomID uint, lastActivity time.Time) error {
	return r.db.
		Model(&postgres.ChatRoom{}).
		Where("id = ?", roomID).
		Update("last_activity", lastActivity).Error
}

func (r *chatRoomRepository) ArchiveRoom(roomID uint) error {
	return r.db.
		Model(&postgres.ChatRoom{}).
		Where("id = ?", roomID).
		Update("is_archived", true).Error
}

func (r *chatRoomRepository) SearchRooms(userID uint, query string, limit int) ([]responses.ChatRoomSummary, error) {
	var rooms []postgres.ChatRoom
	db := r.db.
		Joins("JOIN participants ON chat_rooms.id = participants.chat_room_id").
		Where("participants.user_id = ? AND chat_rooms.is_archived = ?", userID, false)

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		db = db.Where("LOWER(chat_rooms.name) LIKE ? OR LOWER(chat_rooms.description) LIKE ?", searchQuery, searchQuery)
	}

	err := db.Order("chat_rooms.last_activity DESC, chat_rooms.id DESC").
		Limit(limit + 1).
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	summaries := make([]responses.ChatRoomSummary, 0, len(rooms))

	for _, room := range rooms {
		summary := responses.ChatRoomSummary{
			ID:           room.ID,
			Name:         room.Name,
			Type:         room.Type,
			Avatar:       room.Avatar,
			LastActivity: room.LastActivity,
			CreatedAt:    room.CreatedAt,
		}

		// Add participant count
		summary.ParticipantCount = len(room.Participants)

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Participant Repository Implementation
type participantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) repositories.ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) Create(participant *postgres.Participant) error {
	return r.db.Create(participant).Error
}

func (r *participantRepository) GetByID(id uint) (*postgres.Participant, error) {
	var participant postgres.Participant
	err := r.db.
		Preload("User").
		Preload("ChatRoom").
		First(&participant, id).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

func (r *participantRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.Participant{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *participantRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Participant{}, id).Error
}

func (r *participantRepository) GetByRoomAndUser(roomID, userID uint) (*postgres.Participant, error) {
	var participant postgres.Participant
	err := r.db.
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Preload("User").
		Preload("ChatRoom").
		First(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

func (r *participantRepository) GetRoomParticipants(roomID uint) ([]postgres.Participant, error) {
	var participants []postgres.Participant
	err := r.db.
		Where("chat_room_id = ?", roomID).
		Preload("User").
		Find(&participants).Error
	return participants, err
}

func (r *participantRepository) GetUserParticipations(userID uint) ([]postgres.Participant, error) {
	var participants []postgres.Participant
	err := r.db.
		Where("user_id = ?", userID).
		Preload("ChatRoom").
		Find(&participants).Error
	return participants, err
}

func (r *participantRepository) UpdateRole(roomID, userID uint, role postgres.ParticipantRole) error {
	return r.db.
		Model(&postgres.Participant{}).
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Updates(map[string]interface{}{
			"role":       role,
			"updated_at": time.Now(),
		}).Error
}

func (r *participantRepository) UpdateLastRead(roomID, userID uint) error {
	return r.db.
		Model(&postgres.Participant{}).
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Updates(map[string]interface{}{
			"last_read_at": time.Now(),
			"updated_at":   time.Now(),
		}).Error
}

func (r *participantRepository) MuteParticipant(roomID, userID uint, isMuted bool) error {
	return r.db.
		Model(&postgres.Participant{}).
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Updates(map[string]interface{}{
			"is_muted":   isMuted,
			"updated_at": time.Now(),
		}).Error
}

func (r *participantRepository) BlockParticipant(roomID, userID uint, isBlocked bool) error {
	return r.db.
		Model(&postgres.Participant{}).
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Updates(map[string]interface{}{
			"is_blocked": isBlocked,
			"updated_at": time.Now(),
		}).Error
}
