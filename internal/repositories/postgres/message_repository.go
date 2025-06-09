package postgres

import (
	"encoding/json"
	"fmt"
	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"
	"strings"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
)

// Message Repository Implementation
type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repositories.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *postgres.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) GetByID(id uint) (*postgres.Message, error) {
	var message postgres.Message
	err := r.db.
		Preload("Sender").
		Preload("ReadBy").
		Preload("Reactions").
		First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.Message{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.Message{}, id).Error
}

func (r *messageRepository) GetRoomMessages(roomID uint, cursor paginator.Cursor, limit int) ([]postgres.Message, paginator.Cursor, error) {
	var messages []postgres.Message
	db := r.db.
		Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Preload("Reactions").
		Preload("Reactions.User")

	order := paginator.DESC
	p := paginators.CreateMessagesPaginator(
		cursor,
		&order,
		&limit,
	)

	result, nextCursor, err := p.Paginate(db, &messages)

	if err != nil {
		return nil, paginator.Cursor{}, err
	}
	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return messages, nextCursor, nil
}

func (r *messageRepository) MarkAsRead(messageID, userID uint) error {
	read := &postgres.MessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
		CreatedAt: time.Now(),
	}
	return r.db.
		Where("message_id = ? AND user_id = ?", messageID, userID).
		FirstOrCreate(read).Error
}

func (r *messageRepository) MarkAsDelivered(messageID, userID uint) error {
	// Update delivery status in the message
	var message postgres.Message
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err
	}

	var deliveryStatus map[string]postgres.MessageStatus
	if message.DeliveryStatus != "" {
		json.Unmarshal([]byte(message.DeliveryStatus), &deliveryStatus)
	} else {
		deliveryStatus = make(map[string]postgres.MessageStatus)
	}

	deliveryStatus[fmt.Sprintf("%d", userID)] = postgres.MessageStatusDelivered

	statusBytes, _ := json.Marshal(deliveryStatus)
	return r.db.
		Model(&message).
		Update("delivery_status", string(statusBytes)).Error
}

func (r *messageRepository) GetUnreadCount(roomID, userID uint) (int, error) {
	var count int64
	err := r.db.
		Table("messages").
		Joins("LEFT JOIN message_reads ON messages.id = message_reads.message_id AND message_reads.user_id = ?", userID).
		Where("messages.chat_room_id = ? AND message_reads.id IS NULL", roomID).
		Count(&count).Error
	return int(count), err
}

func (r *messageRepository) SearchMessages(roomID uint, query string, limit int) ([]responses.MessageSearchResult, error) {
	var messages []postgres.Message
	db := r.db.
		Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Preload("ChatRoom")

	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		db = db.Where("LOWER(content) LIKE ?", searchQuery)
	}

	err := db.Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	results := make([]responses.MessageSearchResult, 0, len(messages))

	for _, message := range messages {
		result := responses.MessageSearchResult{
			Message:  &message,
			ChatRoom: &message.ChatRoom,
			Score:    1.0,
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *messageRepository) GetMessagesByType(roomID uint, messageType postgres.MessageType, cursor paginator.Cursor, limit int) ([]postgres.Message, paginator.Cursor, error) {
	var messages []postgres.Message
	db := r.db.
		Where("chat_room_id = ? AND type = ?", roomID, messageType).
		Preload("Sender")

	order := paginator.DESC
	p := paginators.CreateMessagesPaginator(cursor, &order, &limit)

	result, nextCursor, err := p.Paginate(db, &messages)

	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return messages, nextCursor, nil
}

func (r *messageRepository) GetRecentMedia(roomID uint, mediaType postgres.MessageType, limit int) ([]postgres.Message, error) {
	var messages []postgres.Message
	err := r.db.
		Where("chat_room_id = ? AND type = ?", roomID, mediaType).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) AddReaction(messageID, userID uint, emoji string) error {
	reaction := &postgres.MessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
		ReactedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	return r.db.
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		FirstOrCreate(reaction).Error
}

func (r *messageRepository) RemoveReaction(messageID, userID uint, emoji string) error {
	return r.db.
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&postgres.MessageReaction{}).Error
}
