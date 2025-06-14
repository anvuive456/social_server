package postgres

import (
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"
	"time"

	"gorm.io/gorm"
)

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
