package postgres

import (
	"fmt"
	"strings"
	"time"

	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
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

func (r *chatRoomRepository) Create(name string, creatorID uint, chatRoomType postgres.ChatRoomType, participants []uint) (*postgres.ChatRoom, error) {
	// Check if private room already exists between two users
	if chatRoomType == postgres.ChatRoomTypePrivate && len(participants) == 1 {
		otherUserID := participants[0]
		existingRoom, err := r.GetPrivateRoom(creatorID, otherUserID)
		if err == nil {
			return existingRoom, nil
		}
	}

	room := &postgres.ChatRoom{
		Name:      name,
		CreatedBy: creatorID,
		Type:      chatRoomType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.Create(room).Error; err != nil {
		return nil, err
	}

	// Add Creator
	participant := &postgres.Participant{
		UserID:     creatorID,
		ChatRoomID: room.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := r.db.Create(participant).Error; err != nil {
		return nil, err
	}

	for _, participantID := range participants {
		participant := &postgres.Participant{
			UserID:     participantID,
			ChatRoomID: room.ID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := r.db.Create(participant).Error; err != nil {
			return nil, err
		}
	}

	return room, nil
}

func (r *chatRoomRepository) AddParticipant(participant *postgres.Participant) error {
	return r.db.Create(participant).Error
}

func (r *chatRoomRepository) GetByID(id uint) (*postgres.ChatRoom, error) {
	var room postgres.ChatRoom
	err := r.db.
		Preload("Creator").
		Preload("Participants").
		Preload("Participants.User").
		Preload("Participants.User.Profile").
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

func (r *chatRoomRepository) Delete(userID uint, roomID uint) error {
	if err := r.db.Delete(&postgres.ChatRoom{
		ID: roomID,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *chatRoomRepository) GetUserRooms(userID uint, archive bool, cursor paginator.Cursor, limit int) ([]responses.ChatRoomSummary, paginator.Cursor, error) {
	var rooms []postgres.ChatRoom
	db := r.db.Model(postgres.ChatRoom{}).Preload("Creator").Preload("Participants").Preload("Participants.User").Preload("Participants.User.Profile").
		Where(postgres.ChatRoom{
			Participants: []postgres.Participant{
				{
					ID: userID,
				},
			},
			DeletedAt:  nil,
			IsArchived: archive,
		})

	order := paginator.DESC
	p := paginators.CreateChatPaginator(cursor, &order, &limit)

	result, nextCursor, err := p.Paginate(db, &rooms)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	if len(rooms) == 0 {
		return []responses.ChatRoomSummary{}, nextCursor, nil
	}

	// Get room IDs for batch queries
	roomIDs := make([]uint, len(rooms))
	for i, room := range rooms {
		roomIDs[i] = room.ID
	}

	// Batch query for participant counts
	type ParticipantCount struct {
		ChatRoomID uint
		Count      int64
	}
	var participantCounts []ParticipantCount
	r.db.Model(&postgres.Participant{}).
		Select("chat_room_id, COUNT(*) as count").
		Where("chat_room_id IN ?", roomIDs).
		Group("chat_room_id").
		Scan(&participantCounts)

	// Create map for quick lookup
	participantCountMap := make(map[uint]int64)
	for _, pc := range participantCounts {
		participantCountMap[pc.ChatRoomID] = pc.Count
	}

	// Batch query for last messages
	type LastMessage struct {
		ChatRoomID uint
		Message    postgres.Message
	}
	var lastMessages []postgres.Message
	r.db.Raw(`
		SELECT DISTINCT ON (chat_room_id) *
		FROM messages
		WHERE chat_room_id IN ?
		ORDER BY chat_room_id, created_at DESC
	`, roomIDs).Scan(&lastMessages)

	// Create map for quick lookup
	lastMessageMap := make(map[uint]*postgres.Message)
	for _, msg := range lastMessages {
		lastMessageMap[msg.ChatRoomID] = &msg
	}

	// Batch query for unread counts
	type UnreadCount struct {
		ChatRoomID uint
		Count      int64
	}
	var unreadCounts []UnreadCount
	r.db.Raw(`
		SELECT
			m.chat_room_id,
			COUNT(*) as count
		FROM messages m
		LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = ?
		WHERE m.chat_room_id IN ? AND mr.id IS NULL
		GROUP BY m.chat_room_id
	`, userID, roomIDs).Scan(&unreadCounts)

	// Create map for quick lookup
	unreadCountMap := make(map[uint]int64)
	for _, uc := range unreadCounts {
		unreadCountMap[uc.ChatRoomID] = uc.Count
	}

	// Build summaries
	summaries := make([]responses.ChatRoomSummary, 0, len(rooms))
	for _, room := range rooms {
		var name string
		if room.Type == postgres.ChatRoomTypePrivate {
			for _, participant := range room.Participants {
				if participant.ID != userID {
					name = participant.User.Profile.DisplayName
					break
				}
			}
		} else {
			name = room.Name
		}

		summary := responses.ChatRoomSummary{
			ID:               room.ID,
			Name:             name,
			Type:             room.Type,
			Avatar:           room.Avatar,
			ParticipantCount: int(participantCountMap[room.ID]),
			LastMessage:      lastMessageMap[room.ID],
			LastActivity:     room.LastActivity,
			UnreadCount:      int(unreadCountMap[room.ID]),
			IsMuted:          false,
			CreatedAt:        room.CreatedAt,
		}

		summaries = append(summaries, summary)
	}

	return summaries, nextCursor, nil
}

func (r *chatRoomRepository) GetPrivateRoom(userID1, userID2 uint) (*postgres.ChatRoom, error) {
	var room postgres.ChatRoom
	err := r.db.
		Model(&postgres.ChatRoom{}).Preload("Participants").
		Where(&postgres.ChatRoom{
			DeletedAt: nil,
			Type:      postgres.ChatRoomTypePrivate,
			Participants: []postgres.Participant{
				{UserID: userID1},
				{UserID: userID2},
			},
		}).First(&room).Error
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
