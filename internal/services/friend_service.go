package services

import (
	"fmt"
	"social_server/internal/models/requests"
	"social_server/internal/models/responses"
	"social_server/internal/repositories"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type FriendService struct {
	userRepo   repositories.UserRepository
	friendRepo repositories.FriendRepository
}

func NewFriendService(userRepo repositories.UserRepository, friendRepo repositories.FriendRepository) *FriendService {
	return &FriendService{
		userRepo:   userRepo,
		friendRepo: friendRepo,
	}
}

func (s *FriendService) GetFriendRequestStats(userID uint) (*responses.FriendStatsResponse, error) {
	sentCount, receivedCount, err := s.friendRepo.GetFriendRequestStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count friends")
	}

	return &responses.FriendStatsResponse{
		TotalSent:     sentCount,
		TotalReceived: receivedCount,
	}, nil
}

func (s *FriendService) GetFriends(userID uint, req *requests.GetFriendsRequest) (*responses.FriendsResponse, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// For now, implement basic friends list without cursor (can be enhanced later)
	cursorObj := paginator.Cursor{
		Before: &req.Before,
		After:  &req.After,
	}
	userFriends, next, totalCount, err := s.friendRepo.GetFriends(userID, cursorObj, req.Limit, req.Search)
	if err != nil {
		return nil, err
	}

	return &responses.FriendsResponse{
		Friends:    userFriends,
		NextCursor: &next,
		TotalCount: totalCount,
	}, nil
}

func (s *FriendService) GetUserFriends(targetUserID uint, currentUserID *uint, limit int, before string, after string) (*responses.FriendsResponse, error) {
	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if current user can view target user's friends
	if currentUserID != nil && *currentUserID != targetUserID {
		// Check privacy settings
		if targetUser.Settings.PrivacyProfileVisibility == "private" {
			return nil, fmt.Errorf("access denied")
		}

		if targetUser.Settings.PrivacyProfileVisibility == "friends" {
			// Check if they are friends
			areFriends, err := s.userRepo.IsFriend(*currentUserID, targetUserID)
			if err != nil || !areFriends {
				return nil, fmt.Errorf("access denied")
			}
		}

		// Check if current user is blocked
		isBlocked, err := s.userRepo.IsBlocked(targetUserID, *currentUserID)
		if err == nil && isBlocked {
			return nil, fmt.Errorf("access denied")
		}
	}

	return nil, err

	// res, err := s.GetFriends(targetUserID, limit, before, after)
	// if err != nil {
	// 	return nil, err
	// }
	// return res, nil
}

func (s *FriendService) SendFriendRequest(fromID, toID uint, message string) error {
	// Validate sender exists
	_, err := s.userRepo.GetByID(fromID)
	if err != nil {
		return fmt.Errorf("sender not found")
	}

	// Validate receiver exists
	receiver, err := s.userRepo.GetByID(toID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !receiver.IsActive {
		return fmt.Errorf("user account is inactive")
	}

	// Check if users are the same
	if fromID == toID {
		return fmt.Errorf("cannot send friend request to yourself")
	}

	// Check if already friends
	areFriends, err := s.userRepo.IsFriend(fromID, toID)
	if err == nil && areFriends {
		return fmt.Errorf("users are already friends")
	}

	// Check if sender is blocked by receiver
	isBlocked, err := s.userRepo.IsBlocked(toID, fromID)
	if err == nil && isBlocked {
		return fmt.Errorf("cannot send friend request")
	}

	// Check if receiver blocked sender
	isBlockedBySender, err := s.userRepo.IsBlocked(fromID, toID)
	if err == nil && isBlockedBySender {
		return fmt.Errorf("cannot send friend request")
	}

	// Check receiver's privacy settings
	if !receiver.Settings.PrivacyAllowFriendRequests {
		return fmt.Errorf("user does not accept friend requests")
	}

	// Check if there's already a pending request
	hasRequest, err := s.friendRepo.HasAlreadyFriendRequest(fromID, toID)
	if err == nil && hasRequest {
		return fmt.Errorf("friend request already sent")
	}

	err = s.friendRepo.SendFriendRequest(fromID, toID, message)
	if err != nil {
		return fmt.Errorf("failed to send friend request")
	}

	return nil
}

func (s *FriendService) AcceptFriendRequest(userID, fromID uint) error {
	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_, err = s.userRepo.GetByID(fromID)
	if err != nil {
		return fmt.Errorf("sender not found")
	}

	hasRequest, err := s.friendRepo.HasAlreadyFriendRequest(fromID, userID)
	if err != nil {
		return fmt.Errorf("failed to check friend requests")
	}

	if !hasRequest {
		return fmt.Errorf("friend request not found")
	}

	err = s.friendRepo.AcceptFriendRequest(userID, fromID)
	if err != nil {
		return fmt.Errorf("failed to accept friend request")
	}

	return nil
}

func (s *FriendService) DeclineFriendRequest(userID, fromID uint) error {
	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_, err = s.userRepo.GetByID(fromID)
	if err != nil {
		return fmt.Errorf("sender not found")
	}

	//  Check if request exists
	hasRequest, err := s.friendRepo.HasAlreadyFriendRequest(fromID, userID)
	if err != nil {
		return fmt.Errorf("failed to check friend requests")
	}

	if !hasRequest {
		return fmt.Errorf("friend request not found")
	}

	err = s.friendRepo.DeclineFriendRequest(userID, fromID)
	if err != nil {
		return fmt.Errorf("failed to decline friend request")
	}

	return nil
}

func (s *FriendService) GetFriendRequests(userID uint, requestType requests.GetFriendRequestType, limit int, before string, after string) (*responses.FriendRequestsResponse, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	cursor := paginator.Cursor{
		Before: &before,
		After:  &after,
	}
	requests, nextCursor, totalCount, err := s.friendRepo.GetFriendRequests(userID, requestType, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend requests")
	}

	return &responses.FriendRequestsResponse{
		Requests:   requests,
		NextCursor: &nextCursor,
		TotalCount: totalCount,
	}, nil
}

func (s *FriendService) RemoveFriend(userID, friendID uint) error {
	if userID == friendID {
		return fmt.Errorf("cannot remove yourself as friend")
	}

	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_, err = s.userRepo.GetByID(friendID)
	if err != nil {
		return fmt.Errorf("friend not found")
	}

	// Check if they are actually friends
	areFriends, err := s.userRepo.IsFriend(userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship")
	}
	if !areFriends {
		return fmt.Errorf("users are not friends")
	}

	err = s.userRepo.RemoveFriend(userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friend")
	}

	return nil
}

func (s *FriendService) BlockUser(userID, targetID uint) error {
	if userID == targetID {
		return fmt.Errorf("cannot block yourself")
	}

	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_, err = s.userRepo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("target user not found")
	}

	// Check if already blocked
	isBlocked, err := s.userRepo.IsBlocked(userID, targetID)
	if err == nil && isBlocked {
		return fmt.Errorf("user is already blocked")
	}

	// Remove friendship if exists
	areFriends, err := s.userRepo.IsFriend(userID, targetID)
	if err == nil && areFriends {
		err = s.userRepo.RemoveFriend(userID, targetID)
		if err != nil {
			return fmt.Errorf("failed to remove friendship before blocking")
		}
	}

	err = s.userRepo.BlockUser(userID, targetID)
	if err != nil {
		return fmt.Errorf("failed to block user")
	}

	return nil
}

func (s *FriendService) UnblockUser(userID, targetID uint) error {
	if userID == targetID {
		return fmt.Errorf("cannot unblock yourself")
	}

	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	_, err = s.userRepo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("target user not found")
	}

	// Check if user is actually blocked
	isBlocked, err := s.userRepo.IsBlocked(userID, targetID)
	if err != nil {
		return fmt.Errorf("failed to check blocked status")
	}
	if !isBlocked {
		return fmt.Errorf("user is not blocked")
	}

	err = s.userRepo.UnblockUser(userID, targetID)
	if err != nil {
		return fmt.Errorf("failed to unblock user")
	}

	return nil
}

type BlockedUsersResponse struct {
	Users      []responses.UserProfile `json:"users"`
	NextCursor *string                 `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
}

func (s *FriendService) GetBlockedUsers(userID uint, limit int, cursor string) (*BlockedUsersResponse, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	cursorObj := paginator.Cursor{}
	blockedUsers, _, err := s.userRepo.GetBlockedUsers(userID, cursorObj, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked users")
	}

	// Simple pagination implementation
	start := 0
	if cursor != "" {
		for i, user := range blockedUsers {
			if fmt.Sprintf("%d", user.ID) == cursor {
				start = i + 1
				break
			}
		}
	}

	end := start + limit
	if end > len(blockedUsers) {
		end = len(blockedUsers)
	}

	var paginatedUsers []responses.UserProfile
	if start < len(blockedUsers) {
		paginatedUsers = blockedUsers[start:end]
	}

	var nextCursor *string
	hasMore := end < len(blockedUsers)
	if hasMore && len(paginatedUsers) > 0 {
		lastUserID := fmt.Sprintf("%d", paginatedUsers[len(paginatedUsers)-1].ID)
		nextCursor = &lastUserID
	}

	return &BlockedUsersResponse{
		Users:      paginatedUsers,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

type FriendshipStatus struct {
	Status         string `json:"status"` // none, friends, pending_sent, pending_received, blocked, blocked_by
	IsFriend       bool   `json:"is_friend"`
	IsBlocked      bool   `json:"is_blocked"`
	IsBlockedBy    bool   `json:"is_blocked_by"`
	PendingRequest bool   `json:"pending_request"`
	CanSendRequest bool   `json:"can_send_request"`
	CanMessage     bool   `json:"can_message"`
}

func (s *FriendService) CheckFriendshipStatus(userID, targetID uint) (*FriendshipStatus, error) {
	if userID == targetID {
		return &FriendshipStatus{
			Status:         "self",
			IsFriend:       false,
			IsBlocked:      false,
			IsBlockedBy:    false,
			PendingRequest: false,
			CanSendRequest: false,
			CanMessage:     true,
		}, nil
	}

	// Validate users exist
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	targetUser, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return nil, fmt.Errorf("target user not found")
	}

	status := &FriendshipStatus{}

	// Check if friends
	areFriends, err := s.userRepo.IsFriend(userID, targetID)
	if err == nil {
		status.IsFriend = areFriends
	}

	// Check if blocked
	isBlocked, err := s.userRepo.IsBlocked(userID, targetID)
	if err == nil {
		status.IsBlocked = isBlocked
	}

	// Check if blocked by target
	isBlockedBy, err := s.userRepo.IsBlocked(targetID, userID)
	if err == nil {
		status.IsBlockedBy = isBlockedBy
	}

	// Check pending requests
	if !status.IsFriend && !status.IsBlocked && !status.IsBlockedBy {
		has, err := s.friendRepo.HasAlreadyFriendRequest(userID, targetID)
		if err == nil && has {
			status.PendingRequest = true
		}
	}

	// Determine overall status
	if status.IsFriend {
		status.Status = "friends"
		status.CanMessage = true
	} else if status.IsBlocked {
		status.Status = "blocked"
	} else if status.IsBlockedBy {
		status.Status = "blocked_by"
	} else if status.PendingRequest {
		status.Status = "pending"
	} else {
		status.Status = "none"
		// Can send request if target allows friend requests
		if targetUser.Settings.PrivacyAllowFriendRequests {
			status.CanSendRequest = true
		}
	}

	return status, nil
}
