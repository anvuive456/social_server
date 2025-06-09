package postgres

import (
	"context"
	"time"

	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/repositories"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
)

// TypingIndicator Repository Implementation
type typingIndicatorRepository struct {
	db *gorm.DB
}

func NewTypingIndicatorRepository(db *gorm.DB) repositories.TypingIndicatorRepository {
	return &typingIndicatorRepository{db: db}
}

func (r *typingIndicatorRepository) SetTyping(roomID, userID uint, isTyping bool) error {
	indicator := &postgres.TypingIndicator{
		ChatRoomID: roomID,
		UserID:     userID,
		IsTyping:   isTyping,
		LastTyped:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return r.db.
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Assign(map[string]interface{}{
			"is_typing":  isTyping,
			"last_typed": time.Now(),
			"updated_at": time.Now(),
		}).
		FirstOrCreate(indicator).Error
}

func (r *typingIndicatorRepository) GetTypingUsers(roomID uint) ([]uint, error) {
	var indicators []postgres.TypingIndicator
	err := r.db.
		Where("chat_room_id = ? AND is_typing = ? AND last_typed > ?",
			roomID, true, time.Now().Add(-30*time.Second)).
		Find(&indicators).Error

	if err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(indicators))
	for _, indicator := range indicators {
		userIDs = append(userIDs, indicator.UserID)
	}

	return userIDs, nil
}

func (r *typingIndicatorRepository) ClearTyping(roomID, userID uint) error {
	return r.db.
		Model(&postgres.TypingIndicator{}).
		Where("chat_room_id = ? AND user_id = ?", roomID, userID).
		Updates(map[string]interface{}{
			"is_typing":  false,
			"updated_at": time.Now(),
		}).Error
}

func (r *typingIndicatorRepository) ClearExpiredTyping(ctx context.Context) error {
	return r.db.
		Model(&postgres.TypingIndicator{}).
		Where("is_typing = ? AND last_typed < ?", true, time.Now().Add(-30*time.Second)).
		Updates(map[string]interface{}{
			"is_typing":  false,
			"updated_at": time.Now(),
		}).Error
}

// OnlineStatus Repository Implementation
type onlineStatusRepository struct {
	db *gorm.DB
}

func NewOnlineStatusRepository(db *gorm.DB) repositories.OnlineStatusRepository {
	return &onlineStatusRepository{db: db}
}

func (r *onlineStatusRepository) SetOnline(userID uint, isOnline bool) error {
	now := time.Now()
	status := &postgres.OnlineStatus{
		UserID:    userID,
		IsOnline:  isOnline,
		LastSeen:  now,
		Status:    "online",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if !isOnline {
		status.Status = "offline"
	}

	return r.db.
		Where("user_id = ?", userID).
		Assign(map[string]interface{}{
			"is_online":  isOnline,
			"last_seen":  now,
			"status":     status.Status,
			"updated_at": now,
		}).
		FirstOrCreate(status).Error
}

func (r *onlineStatusRepository) GetStatus(userID uint) (*postgres.OnlineStatus, error) {
	var status postgres.OnlineStatus
	err := r.db.
		Where("user_id = ?", userID).
		Preload("User").
		First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *onlineStatusRepository) GetMultipleStatus(userIDs []uint) ([]postgres.OnlineStatus, error) {
	var statuses []postgres.OnlineStatus
	err := r.db.
		Where("user_id IN ?", userIDs).
		Preload("User").
		Find(&statuses).Error
	return statuses, err
}

func (r *onlineStatusRepository) UpdateLastSeen(userID uint) error {
	return r.db.
		Model(&postgres.OnlineStatus{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"last_seen":  time.Now(),
			"updated_at": time.Now(),
		}).Error
}

func (r *onlineStatusRepository) SetCustomStatus(userID uint, status string) error {
	return r.db.
		Model(&postgres.OnlineStatus{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// ChatInvite Repository Implementation
type chatInviteRepository struct {
	db *gorm.DB
}

func NewChatInviteRepository(db *gorm.DB) repositories.ChatInviteRepository {
	return &chatInviteRepository{db: db}
}

func (r *chatInviteRepository) Create(invite *postgres.ChatInvite) error {
	return r.db.Create(invite).Error
}

func (r *chatInviteRepository) GetByID(id uint) (*postgres.ChatInvite, error) {
	var invite postgres.ChatInvite
	err := r.db.
		Preload("ChatRoom").
		Preload("Inviter").
		Preload("Invitee").
		First(&invite, id).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *chatInviteRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.ChatInvite{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *chatInviteRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.ChatInvite{}, id).Error
}

func (r *chatInviteRepository) GetUserInvites(userID uint, status string, cursor paginator.Cursor, limit int) ([]postgres.ChatInvite, paginator.Cursor, error) {
	var invites []postgres.ChatInvite
	dbQuery := r.db.
		Where("invitee_id = ?", userID).
		Preload("ChatRoom").
		Preload("Inviter")

	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	order := paginator.DESC
	p := paginators.CreateNotificationPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &invites)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return invites, nextCursor, nil
}

func (r *chatInviteRepository) GetRoomInvites(roomID uint, cursor paginator.Cursor, limit int) ([]postgres.ChatInvite, paginator.Cursor, error) {
	var invites []postgres.ChatInvite
	dbQuery := r.db.
		Where("chat_room_id = ?", roomID).
		Preload("Inviter").
		Preload("Invitee")

	order := paginator.DESC
	p := paginators.CreateNotificationPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &invites)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return invites, nextCursor, nil
}

func (r *chatInviteRepository) AcceptInvite(inviteID uint) error {
	return r.db.
		Model(&postgres.ChatInvite{}).
		Where("id = ?", inviteID).
		Updates(map[string]interface{}{
			"status":     "accepted",
			"updated_at": time.Now(),
		}).Error
}

func (r *chatInviteRepository) DeclineInvite(inviteID uint) error {
	return r.db.
		Model(&postgres.ChatInvite{}).
		Where("id = ?", inviteID).
		Updates(map[string]interface{}{
			"status":     "declined",
			"updated_at": time.Now(),
		}).Error
}

func (r *chatInviteRepository) ExpireOldInvites(ctx context.Context) error {
	return r.db.
		Model(&postgres.ChatInvite{}).
		Where("status = ? AND expires_at < ?", "pending", time.Now()).
		Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": time.Now(),
		}).Error
}

// ChatNotification Repository Implementation
type chatNotificationRepository struct {
	db *gorm.DB
}

func NewChatNotificationRepository(db *gorm.DB) repositories.ChatNotificationRepository {
	return &chatNotificationRepository{db: db}
}

func (r *chatNotificationRepository) Create(notification *postgres.ChatNotification) error {
	return r.db.Create(notification).Error
}

func (r *chatNotificationRepository) GetByID(id uint) (*postgres.ChatNotification, error) {
	var notification postgres.ChatNotification
	err := r.db.
		Preload("User").
		Preload("ChatRoom").
		Preload("Message").
		First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *chatNotificationRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.ChatNotification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *chatNotificationRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.ChatNotification{}, id).Error
}

func (r *chatNotificationRepository) GetUserNotifications(userID uint, cursor paginator.Cursor, limit int) ([]postgres.ChatNotification, paginator.Cursor, error) {
	var notifications []postgres.ChatNotification
	dbQuery := r.db.
		Where("user_id = ?", userID).
		Preload("ChatRoom").
		Preload("Message")

	order := paginator.DESC
	p := paginators.CreateNotificationPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &notifications)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return notifications, nextCursor, nil
}

func (r *chatNotificationRepository) MarkAsRead(notificationID uint) error {
	now := time.Now()
	return r.db.
		Model(&postgres.ChatNotification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *chatNotificationRepository) MarkAllAsRead(userID uint) error {
	now := time.Now()
	return r.db.
		Model(&postgres.ChatNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

func (r *chatNotificationRepository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.
		Model(&postgres.ChatNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *chatNotificationRepository) DeleteOldNotifications(olderThan time.Time) error {
	return r.db.
		Where("created_at < ?", olderThan).
		Delete(&postgres.ChatNotification{}).Error
}
