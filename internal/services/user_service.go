package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"social_server/internal/config"
	"social_server/internal/models/postgres"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"
	"strings"
	"time"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type UserService struct {
	userRepo repositories.UserRepository
	config   *config.AuthConfig
}

func NewUserService(userRepo repositories.UserRepository, config *config.AuthConfig) *UserService {
	return &UserService{
		userRepo: userRepo,
		config:   config,
	}
}

type UserSearchResponse struct {
	Users      []responses.UserSearchSimpleResponse `json:"users"`
	NextCursor *paginator.Cursor                    `json:"next_cursor,omitempty"`
}

func (s *UserService) SearchUsers(currentUserID uint, query string, limit int, before, after string) (*UserSearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if len(query) < 2 {
		return nil, fmt.Errorf("search query must be at least 2 characters")
	}

	// For now, implement basic search without cursor (can be enhanced later)
	cursorObj := paginator.Cursor{
		Before: &before,
		After:  &after,
	}
	users, next, err := s.userRepo.SearchUsers(currentUserID, query, cursorObj, limit)
	if err != nil {
		return nil, err
	}

	return &UserSearchResponse{
		Users:      users,
		NextCursor: &next,
	}, nil
}

func (s *UserService) UploadAvatar(userID uint, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return "", fmt.Errorf("user account is inactive")
	}

	// Generate filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)

	// In a real implementation, you would upload to cloud storage (S3, etc.)
	// For now, we'll simulate by saving to uploads directory
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	// Update user avatar in database
	updates := map[string]interface{}{
		"avatar":     avatarURL,
		"updated_at": time.Now(),
	}

	err = s.userRepo.Update(userID, updates)
	if err != nil {
		return "", fmt.Errorf("failed to update avatar")
	}

	return avatarURL, nil
}

type UserStats struct {
	UserID       uint       `json:"user_id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"display_name"`
	FriendsCount int        `json:"friends_count"`
	PostsCount   int        `json:"posts_count"`
	IsVerified   bool       `json:"is_verified"`
	IsOnline     bool       `json:"is_online"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	JoinedAt     time.Time  `json:"joined_at"`
}

func (s *UserService) GetUserStats(userID uint) (*UserStats, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get friends count
	friendsCount := 0

	// TODO: Get posts count from post repository
	postsCount := 0

	stats := &UserStats{
		UserID:       user.ID,
		Username:     user.Email, // Use email since username field removed
		DisplayName:  user.Email, // Use email as fallback
		FriendsCount: friendsCount,
		PostsCount:   postsCount,
		IsVerified:   user.IsVerified,
		IsOnline:     user.IsOnline,
		LastSeen:     user.LastSeen,
		JoinedAt:     user.CreatedAt,
	}

	return stats, nil
}

func (s *UserService) UpdateOnlineStatus(userID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online":  isOnline,
		"updated_at": time.Now(),
	}

	if !isOnline {
		updates["last_seen"] = time.Now()
	}

	err := s.userRepo.Update(userID, updates)
	if err != nil {
		return fmt.Errorf("failed to update online status")
	}

	return nil
}

func (s *UserService) UpdateSettings(userID uint, settings *postgres.UserSettings) (*postgres.UserSettings, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Validate settings values
	if settings.PrivacyProfileVisibility != "" {
		validVisibility := []string{"public", "friends", "private"}
		isValid := false
		for _, v := range validVisibility {
			if settings.PrivacyProfileVisibility == v {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, fmt.Errorf("invalid profile visibility setting")
		}
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// Update privacy settings
	if settings.PrivacyProfileVisibility != "" {
		updates["settings_privacy_profile_visibility"] = settings.PrivacyProfileVisibility
	}
	updates["settings_privacy_show_online_status"] = settings.PrivacyShowOnlineStatus
	updates["settings_privacy_allow_friend_requests"] = settings.PrivacyAllowFriendRequests

	// Update notification settings
	updates["settings_notifications_email"] = settings.NotificationsEmail
	updates["settings_notifications_push"] = settings.NotificationsPush
	updates["settings_notifications_friend_requests"] = settings.NotificationsFriendRequests
	updates["settings_notifications_messages"] = settings.NotificationsMessages
	updates["settings_notifications_posts"] = settings.NotificationsPosts

	err = s.userRepo.Update(userID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update settings")
	}

	// Get updated user to return current settings
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated settings")
	}

	return &user.Settings, nil
}

func (s *UserService) GetSettings(userID uint) (*postgres.UserSettings, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &user.Settings, nil
}

func (s *UserService) GetUserByID(userID uint) (*postgres.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(email string) (*postgres.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (*postgres.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) DeactivateAccount(userID uint) error {
	updates := map[string]interface{}{
		"is_active":  false,
		"is_online":  false,
		"updated_at": time.Now(),
	}

	err := s.userRepo.Update(userID, updates)
	if err != nil {
		return fmt.Errorf("failed to deactivate account")
	}

	return nil
}

func (s *UserService) ReactivateAccount(userID uint) error {
	updates := map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}

	err := s.userRepo.Update(userID, updates)
	if err != nil {
		return fmt.Errorf("failed to reactivate account")
	}

	return nil
}

func (s *UserService) VerifyUser(userID uint) error {
	updates := map[string]interface{}{
		"is_verified": true,
		"updated_at":  time.Now(),
	}

	err := s.userRepo.Update(userID, updates)
	if err != nil {
		return fmt.Errorf("failed to verify user")
	}

	return nil
}

func (s *UserService) CheckUserPermissions(currentUserID, targetUserID uint, action string) error {
	if currentUserID == targetUserID {
		return nil // Users can always access their own content
	}

	// Check if current user is blocked by target user
	isBlocked, err := s.userRepo.IsBlocked(targetUserID, currentUserID)
	if err != nil {
		return fmt.Errorf("failed to check blocked status")
	}
	if isBlocked {
		return fmt.Errorf("access denied")
	}

	switch action {
	case "view_profile":
		return s.checkProfileViewPermission(currentUserID, targetUserID)
	case "send_message":
		return s.checkMessagePermission(currentUserID, targetUserID)
	default:
		return fmt.Errorf("unknown action")
	}
}

func (s *UserService) checkProfileViewPermission(currentUserID, targetUserID uint) error {
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	switch targetUser.Settings.PrivacyProfileVisibility {
	case "private":
		return fmt.Errorf("profile is private")
	case "friends":
		areFriends, err := s.userRepo.IsFriend(currentUserID, targetUserID)
		if err != nil {
			return fmt.Errorf("failed to check friendship")
		}
		if !areFriends {
			return fmt.Errorf("profile is friends only")
		}
	case "public":
		// Public profile, anyone can view
	}

	return nil
}

func (s *UserService) checkMessagePermission(currentUserID, targetUserID uint) error {
	// Check if they are friends (for now, only friends can message each other)
	areFriends, err := s.userRepo.IsFriend(currentUserID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to check friendship")
	}
	if !areFriends {
		return fmt.Errorf("can only message friends")
	}

	return nil
}

func (s *UserService) ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must be less than 50 characters")
	}

	// Check if username contains only allowed characters
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '.') {
			return fmt.Errorf("username can only contain letters, numbers, underscore and dot")
		}
	}

	// Username cannot start or end with special characters
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, ".") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, ".") {
		return fmt.Errorf("username cannot start or end with underscore or dot")
	}

	return nil
}

func (s *UserService) ValidateEmail(email string) error {
	if len(email) < 5 {
		return fmt.Errorf("email must be at least 5 characters")
	}
	if len(email) > 100 {
		return fmt.Errorf("email must be less than 100 characters")
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
