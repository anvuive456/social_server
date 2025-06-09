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

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *postgres.User) (*postgres.User, error) {
	err := r.db.Create(user).Error
	return user, err
}

func (r *userRepository) GetByID(id uint) (*postgres.User, error) {
	var user postgres.User
	err := r.db.
		Preload("UserFriends").
		Preload("SentFriendRequests").
		Preload("ReceivedFriendRequests").
		Preload("BlockedUsers").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*postgres.User, error) {
	var user postgres.User
	err := r.db.
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*postgres.User, error) {
	var user postgres.User
	err := r.db.
		Where("username = ? AND is_active = ?", username, true).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(id uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.
		Model(&postgres.User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&postgres.User{}, id).Error
}

func (r *userRepository) UpdateOnlineStatus(id uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online":  isOnline,
		"updated_at": time.Now(),
	}
	if !isOnline {
		updates["last_seen"] = time.Now()
	}
	return r.db.
		Model(&postgres.User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *userRepository) SearchUsers(currentUserID uint, query string, cursor paginator.Cursor, limit int) ([]responses.UserSearchSimpleResponse, paginator.Cursor, error) {
	var results []struct {
		UserID           uint       `gorm:"column:user_id"`
		Email            string     `gorm:"column:email"`
		ProfileID        uint       `gorm:"column:profile_id"`
		FirstName        string     `gorm:"column:first_name"`
		LastName         string     `gorm:"column:last_name"`
		DisplayName      string     `gorm:"column:display_name"`
		Avatar           string     `gorm:"column:avatar"`
		AvatarHash       string     `gorm:"column:avatar_hash"`
		Bio              string     `gorm:"column:bio"`
		DateOfBirth      *time.Time `gorm:"column:date_of_birth"`
		Phone            string     `gorm:"column:phone"`
		ProfileCreatedAt time.Time  `gorm:"column:profile_created_at"`
		ProfileUpdatedAt time.Time  `gorm:"column:profile_updated_at"`
	}

	dbQuery := r.db.
		Table("users").
		Select("users.id as user_id, users.email, profiles.id as profile_id, profiles.first_name, profiles.last_name, profiles.display_name, profiles.avatar, profiles.avatar_hash, profiles.bio, profiles.date_of_birth, profiles.phone, profiles.created_at as profile_created_at, profiles.updated_at as profile_updated_at").
		Joins("INNER JOIN profiles ON users.id = profiles.user_id").
		Where("users.is_active = ?", true).
		Where("users.id != ?", currentUserID)

	// Add search conditions
	if query != "" {
		searchQuery := fmt.Sprintf("%%%s%%", strings.ToLower(query))
		dbQuery = dbQuery.Where(
			"LOWER(users.email) LIKE ? OR LOWER(profiles.first_name) LIKE ? OR LOWER(profiles.last_name) LIKE ? OR LOWER(profiles.phone) LIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery,
		)
	}

	order := paginator.ASC
	p := paginators.CreateSearchUsersPaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &results)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	// Convert to UserSearchSimpleResponse
	users := make([]responses.UserSearchSimpleResponse, 0, len(results))
	for _, result := range results {
		userSearchResponse := responses.UserSearchSimpleResponse{
			ID:    result.UserID,
			Email: result.Email,
			Profile: responses.ProfileResponse{
				ID:          result.ProfileID,
				UserID:      result.UserID,
				FirstName:   result.FirstName,
				LastName:    result.LastName,
				DisplayName: result.DisplayName,
				Avatar:      result.Avatar,
				AvatarHash:  result.AvatarHash,
				Bio:         result.Bio,
				DateOfBirth: result.DateOfBirth,
				Phone:       result.Phone,
				CreatedAt:   result.ProfileCreatedAt,
				UpdatedAt:   result.ProfileUpdatedAt,
			},
		}

		users = append(users, userSearchResponse)
	}

	return users, nextCursor, nil
}

func (r *userRepository) GetUserProfile(id uint) (*responses.UserProfile, error) {
	var user postgres.User
	err := r.db.
		Select("id, email, is_online, is_verified").
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	// Get friend count
	var friendCount int64
	r.db.
		Model(&postgres.UserFriend{}).
		Where("user_id = ? AND status = ?", id, postgres.FriendStatusActive).
		Count(&friendCount)

	// Get post count
	var postCount int64
	r.db.
		Model(&postgres.Post{}).
		Where("author_id = ? AND deleted_at IS NULL", id).
		Count(&postCount)

	// Get default profile (oldest profile)
	var defaultProfile postgres.Profile
	profileErr := r.db.
		Where("user_id = ?", id).
		Order("created_at ASC").
		First(&defaultProfile).Error

	profile := &responses.UserProfile{
		ID:          user.ID,
		IsOnline:    user.IsOnline,
		IsVerified:  user.IsVerified,
		FriendCount: int(friendCount),
		PostCount:   int(postCount),
	}

	// Include default profile if found
	if profileErr == nil {
		profile.Profile = responses.ProfileResponse{
			ID:          defaultProfile.ID,
			UserID:      defaultProfile.UserID,
			FirstName:   defaultProfile.FirstName,
			LastName:    defaultProfile.LastName,
			DisplayName: defaultProfile.DisplayName,
			Avatar:      defaultProfile.Avatar,
			Bio:         defaultProfile.Bio,
			DateOfBirth: defaultProfile.DateOfBirth,
			Phone:       defaultProfile.Phone,
			CreatedAt:   defaultProfile.CreatedAt,
			UpdatedAt:   defaultProfile.UpdatedAt,
		}
	}

	return profile, nil
}

func (r *userRepository) RemoveFriend(userID, friendID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update status to inactive instead of deleting
		if err := tx.Model(&postgres.UserFriend{}).
			Where("user_id = ? AND friend_id = ?", userID, friendID).
			Update("status", postgres.FriendStatusInactive).Error; err != nil {
			return err
		}

		return tx.Model(&postgres.UserFriend{}).
			Where("user_id = ? AND friend_id = ?", friendID, userID).
			Update("status", postgres.FriendStatusInactive).Error
	})
}

func (r *userRepository) GetFriendRequests(userID uint, requestType string, cursor paginator.Cursor, limit int) ([]postgres.FriendRequest, paginator.Cursor, error) {
	var requests []postgres.FriendRequest
	dbQuery := r.db.
		Where("status = ?", "pending").
		Preload("Sender").
		Preload("Receiver")

	// Filter by request type
	if requestType == "sent" {
		dbQuery = dbQuery.Where("sender_id = ?", userID)
	} else {
		dbQuery = dbQuery.Where("receiver_id = ?", userID)
	}

	order := paginator.DESC
	p := paginators.CreateProfilePaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &requests)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	return requests, nextCursor, nil
}

func (r *userRepository) BlockUser(userID, targetID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Add to blocked users
		if err := tx.Exec("INSERT INTO user_blocks (user_id, blocked_user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			userID, targetID).Error; err != nil {
			return err
		}

		// Remove friend relationship if exists
		if err := r.RemoveFriend(userID, targetID); err != nil {
			// Ignore error if they weren't friends
		}

		// Remove/decline any pending friend requests
		return tx.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, targetID, targetID, userID).
			Delete(&postgres.FriendRequest{}).Error
	})
}

func (r *userRepository) UnblockUser(userID, targetID uint) error {
	return r.db.
		Exec("DELETE FROM user_blocks WHERE user_id = ? AND blocked_user_id = ?", userID, targetID).Error
}

func (r *userRepository) IsFriend(userID, targetID uint) (bool, error) {
	var count int64
	err := r.db.
		Model(&postgres.UserFriend{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", userID, targetID, postgres.FriendStatusActive).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) AreFriends(userID, targetID uint) (bool, error) {
	var count int64
	err := r.db.
		Model(&postgres.UserFriend{}).
		Where("((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = ?",
			userID, targetID, targetID, userID, postgres.FriendStatusActive).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) GetBlockedUsers(userID uint, cursor paginator.Cursor, limit int) ([]responses.UserProfile, paginator.Cursor, error) {
	var users []postgres.User
	dbQuery := r.db.
		Table("users").
		Joins("JOIN user_blocks ON users.id = user_blocks.blocked_user_id").
		Where("user_blocks.user_id = ? AND users.is_active = ?", userID, true).
		Select("users.id, users.email, users.is_online, users.is_verified, users.created_at")

	order := paginator.DESC
	p := paginators.CreateProfilePaginator(cursor, &order, &limit)
	result, nextCursor, err := p.Paginate(dbQuery, &users)
	if err != nil {
		return nil, paginator.Cursor{}, err
	}

	if result.Error != nil {
		return nil, paginator.Cursor{}, result.Error
	}

	profiles := make([]responses.UserProfile, 0, len(users))
	for _, user := range users {
		// Get friend count
		var friendCount int64
		r.db.
			Model(&postgres.UserFriend{}).
			Where("user_id = ? AND status = ?", user.ID, postgres.FriendStatusActive).
			Count(&friendCount)

		// Get post count
		var postCount int64
		r.db.
			Model(&postgres.Post{}).
			Where("author_id = ? AND deleted_at IS NULL", user.ID).
			Count(&postCount)

		profiles = append(profiles, responses.UserProfile{
			ID:          user.ID,
			IsOnline:    user.IsOnline,
			IsVerified:  user.IsVerified,
			FriendCount: int(friendCount),
			PostCount:   int(postCount),
		})
	}

	return profiles, nextCursor, nil
}

func (r *userRepository) IsBlocked(userID, targetID uint) (bool, error) {
	var count int64
	err := r.db.
		Table("user_blocks").
		Where("user_id = ? AND blocked_user_id = ?", userID, targetID).
		Count(&count).Error

	return count > 0, err
}
