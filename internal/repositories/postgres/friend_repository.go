package postgres

import (
	"fmt"
	"social_server/internal/models/paginators"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/repositories"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"
)

type friendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) repositories.FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) GetFriends(userID uint, cursor paginator.Cursor, limit int, search string) ([]postgres.UserFriend, paginator.Cursor, uint, error) {
	var userFriends []postgres.UserFriend
	var totalCount int64
	
	dbQuery := r.db.Model(&postgres.UserFriend{}).
		Preload("Friend").
		Preload("Friend.Profile").
		Where("user_id = ? AND status = ?", userID, postgres.FriendStatusActive)

	// Add search filter if provided
	if search != "" {
		dbQuery = dbQuery.Joins("JOIN users ON user_friends.friend_id = users.id").
			Joins("LEFT JOIN profiles ON users.id = profiles.user_id").
			Where("users.email ILIKE ? OR profiles.first_name ILIKE ? OR profiles.last_name ILIKE ? OR profiles.display_name ILIKE ?", 
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	order := paginator.DESC
	p := paginators.CreateUserFriendPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &userFriends)
	if err != nil {
		return nil, paginator.Cursor{}, 0, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, 0, result.Error
	}

	result.Count(&totalCount)

	return userFriends, nextCursor, uint(totalCount), nil
}

func (r *friendRepository) HasAlreadyFriendRequest(userID uint, targetID uint) (bool, error) {
	var count int64
	err := r.db.Table("friend_requests").Where("sender_id = ? AND receiver_id = ?", userID, targetID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *friendRepository) SendFriendRequest(fromID, toID uint, message string) error {
	// Check if already friends or request exists
	var count int64
	r.db.
		Model(&postgres.UserFriend{}).
		Where("((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = ?",
			fromID, toID, toID, fromID, postgres.FriendStatusActive).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("users are already friends")
	}

	r.db.
		Model(&postgres.FriendRequest{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			fromID, toID, toID, fromID).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("friend request already exists")
	}

	request := &postgres.FriendRequest{
		SenderID:   fromID,
		ReceiverID: toID,
		Status:     postgres.FriendRequestPending,
		Message:    message,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return r.db.Create(request).Error
}

func (r *friendRepository) GetFriendRequests(userID uint, requestType requests.GetFriendRequestType, cursor paginator.Cursor, limit int) ([]postgres.FriendRequest, paginator.Cursor, uint, error) {
	var data []postgres.FriendRequest
	var count int64
	var nextCursor paginator.Cursor

	order := paginator.DESC
	p := paginators.CreateFriendRequestPaginator(cursor, &order, &limit)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&postgres.FriendRequest{})
		if requestType == requests.GetFriendRequestTypeSent {
			query = query.Preload("Receiver").Preload("Receiver.Profile").Where("sender_id = ? AND status = ?", userID, postgres.FriendRequestPending)
		} else if requestType == requests.GetFriendRequestTypeReceived {
			query = query.Preload("Sender").Preload("Sender.Profile").Where("receiver_id = ? AND status = ?", userID, postgres.FriendRequestPending)
		} else {
			query = query.Preload("Sender").Preload("Sender.Profile").Preload("Receiver").Preload("Receiver.Profile").Where("(sender_id = ? AND status = ?) OR (receiver_id = ? AND status = ?)", userID, postgres.FriendRequestPending, userID, postgres.FriendRequestPending)
		}

		result, cursor, err := p.Paginate(query, &data)
		if err != nil {
			return err
		}
		if result.Error != nil {
			return result.Error
		}

		nextCursor = cursor
		result.Count(&count)

		return nil
	})

	if err != nil {
		return nil, paginator.Cursor{}, 0, err
	}

	return data, nextCursor, uint(count), nil
}

// GetFriendRequestStats implements repositories.FriendRepository.
func (r *friendRepository) GetFriendRequestStats(userID uint) (uint, uint, error) {
	var sentCount, receivedCount int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&postgres.FriendRequest{}).Where("sender_id = ? AND status = ?", userID, postgres.FriendRequestPending)
		result := query.Count(&sentCount)
		if result.Error != nil {
			return result.Error
		}

		query = tx.Model(&postgres.FriendRequest{}).Where("receiver_id = ? AND status = ?", userID, postgres.FriendRequestPending)
		result = query.Count(&receivedCount)
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return uint(sentCount), uint(receivedCount), nil
}

func (r *friendRepository) AcceptFriendRequest(userID, fromID uint) error {
	err := r.db.Model(&postgres.FriendRequest{}).
		Where("sender_id = ? AND receiver_id = ? AND status = ?", fromID, userID, "pending").
		Updates(map[string]interface{}{
			"status":     "accepted",
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return err
	}

	// Add friend relationship
	return r.AddFriend(userID, fromID)
}

func (r *friendRepository) DeclineFriendRequest(userID, fromID uint) error {
	return r.db.
		Model(&postgres.FriendRequest{}).
		Where("sender_id = ? AND receiver_id = ? AND status = ?", fromID, userID, "pending").
		Updates(map[string]interface{}{
			"status":     "rejected",
			"updated_at": time.Now(),
		}).Error
}

func (r *friendRepository) AddFriend(userID, friendID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Add friend relationship (both directions)
		userFriend1 := &postgres.UserFriend{
			UserID:   userID,
			FriendID: friendID,
			Status:   postgres.FriendStatusActive,
		}
		if err := tx.Create(userFriend1).Error; err != nil {
			return err
		}

		userFriend2 := &postgres.UserFriend{
			UserID:   friendID,
			FriendID: userID,
			Status:   postgres.FriendStatusActive,
		}
		if err := tx.Create(userFriend2).Error; err != nil {
			return err
		}

		// Remove friend request if exists
		return tx.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, friendID, friendID, userID).
			Delete(&postgres.FriendRequest{}).Error
	})
}
