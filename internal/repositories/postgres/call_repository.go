package postgres

import (
	"time"

	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
)

type callRepository struct {
	db *gorm.DB
}

func NewCallRepository(db *gorm.DB) repositories.CallRepository {
	return &callRepository{db: db}
}

func (r *callRepository) Create(call *postgres.Call) error {
	return r.db.Create(call).Error
}

func (r *callRepository) GetByID(id uint) (*postgres.Call, error) {
	var call postgres.Call
	err := r.db.
		Preload("Caller").
		Preload("Callee").
		Preload("Participants").
		Preload("Participants.User").
		First(&call, id).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *callRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.Call{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *callRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Call{}, id).Error
}

func (r *callRepository) GetUserCalls(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error) {
	var calls []postgres.Call
	query := r.db.
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Preload("Caller").
		Preload("Callee")

	order := paginator.DESC

	p := paginators.CreateCallPaginator(
		cursor,
		&order,
		&limit,
	)

	result, cursor, err := p.Paginate(query, &calls)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	// this is gorm error
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return calls, cursor, nil
}

func (r *callRepository) GetActiveCall(userID uint) (*postgres.Call, error) {
	var call postgres.Call
	err := r.db.
		Where("(caller_id = ? OR callee_id = ?) AND status IN (?)",
			userID, userID, []string{"ongoing", "ringing"}).
		Preload("Caller").
		Preload("Callee").
		Preload("Participants").
		Preload("Participants.User").
		First(&call).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *callRepository) EndCall(callID uint) error {
	now := time.Now()
	return r.db.
		Model(&postgres.Call{}).
		Where("id = ?", callID).
		Updates(map[string]interface{}{
			"status":     "ended",
			"ended_at":   &now,
			"updated_at": now,
		}).Error
}

func (r *callRepository) JoinCall(callID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if user is already a participant
		var existing postgres.CallParticipant
		err := tx.Where("call_id = ? AND user_id = ?", callID, userID).
			First(&existing).Error

		if err == nil {
			// Update existing participant
			return tx.Model(&existing).Updates(map[string]interface{}{
				"joined_at":  time.Now(),
				"is_active":  true,
				"updated_at": time.Now(),
			}).Error
		}

		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Create new participant
		participant := &postgres.CallParticipant{
			CallID:   callID,
			UserID:   userID,
			JoinedAt: time.Now(),
			IsActive: true,
		}
		return tx.Create(participant).Error
	})
}

func (r *callRepository) LeaveCall(callID, userID uint) error {
	now := time.Now()
	return r.db.
		Model(&postgres.CallParticipant{}).
		Where("call_id = ? AND user_id = ?", callID, userID).
		Updates(map[string]interface{}{
			"left_at":    &now,
			"is_active":  false,
			"updated_at": now,
		}).Error
}

func (r *callRepository) GetCallParticipants(callID uint) ([]postgres.CallParticipant, error) {
	var participants []postgres.CallParticipant
	err := r.db.
		Where("call_id = ?", callID).
		Preload("User").
		Find(&participants).Error
	return participants, err
}

func (r *callRepository) GetUserCallHistory(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error) {
	var calls []postgres.Call
	query := r.db.
		Where("(caller_id = ? OR callee_id = ?) AND status IN (?)",
			userID, userID, []string{"ended", "declined", "missed"}).
		Preload("Caller").
		Preload("Callee")

	order := paginator.DESC

	p := paginators.CreateCallPaginator(
		cursor,
		&order,
		&limit,
	)

	result, cursor, err := p.Paginate(query, &calls)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	// this is gorm error
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return calls, cursor, nil
}
