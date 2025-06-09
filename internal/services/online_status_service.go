package services

import (
	"log"
	"social_server/internal/models/postgres"
	"sync"
	"time"

	"gorm.io/gorm"
)

type OnlineStatusService struct {
	db          *gorm.DB
	onlineUsers map[uint]*OnlineUserInfo // Cache in memory
	mutex       sync.RWMutex
	cleanupTicker *time.Ticker
	stopChan    chan bool
}

type OnlineUserInfo struct {
	UserID         uint      `json:"user_id"`
	LastSeen       time.Time `json:"last_seen"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	ConnectedAt    time.Time `json:"connected_at"`
	ConnectionID   string    `json:"connection_id"`
	IsActive       bool      `json:"is_active"`
}

type OnlineStatusUpdate struct {
	UserID    uint   `json:"user_id"`
	IsOnline  bool   `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
	Username  string `json:"username,omitempty"`
}

// OnlineStatusCallback is called when user online status changes
type OnlineStatusCallback func(update OnlineStatusUpdate)

var onlineStatusCallbacks []OnlineStatusCallback

const (
	HeartbeatTimeout = 90 * time.Second  // User considered offline after 90s without heartbeat
	CleanupInterval  = 30 * time.Second  // Cleanup check every 30s
)

func NewOnlineStatusService(db *gorm.DB) *OnlineStatusService {
	service := &OnlineStatusService{
		db:          db,
		onlineUsers: make(map[uint]*OnlineUserInfo),
		stopChan:    make(chan bool),
	}

	// Start background cleanup job
	service.startCleanupJob()

	return service
}

// SetUserOnline marks user as online
func (s *OnlineStatusService) SetUserOnline(userID uint, connectionID string) error {
	now := time.Now()
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update in-memory cache
	s.onlineUsers[userID] = &OnlineUserInfo{
		UserID:        userID,
		LastSeen:      now,
		LastHeartbeat: now,
		ConnectedAt:   now,
		ConnectionID:  connectionID,
		IsActive:      true,
	}

	// Update database
	err := s.db.Model(&postgres.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_online": true,
			"last_seen": now,
		}).Error

	if err != nil {
		delete(s.onlineUsers, userID)
		return err
	}

	// Broadcast status change
	s.broadcastStatusChange(userID, true, now)

	log.Printf("User %d set online with connection %s", userID, connectionID)
	return nil
}

// SetUserOffline marks user as offline
func (s *OnlineStatusService) SetUserOffline(userID uint) error {
	now := time.Now()
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove from in-memory cache
	delete(s.onlineUsers, userID)

	// Update database
	err := s.db.Model(&postgres.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_online": false,
			"last_seen": now,
		}).Error

	if err != nil {
		return err
	}

	// Broadcast status change
	s.broadcastStatusChange(userID, false, now)

	log.Printf("User %d set offline", userID)
	return nil
}

// UpdateHeartbeat updates user's last heartbeat
func (s *OnlineStatusService) UpdateHeartbeat(userID uint) error {
	now := time.Now()
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if userInfo, exists := s.onlineUsers[userID]; exists {
		userInfo.LastHeartbeat = now
		userInfo.LastSeen = now
		userInfo.IsActive = true
	} else {
		// User not in cache, add them
		s.onlineUsers[userID] = &OnlineUserInfo{
			UserID:        userID,
			LastSeen:      now,
			LastHeartbeat: now,
			ConnectedAt:   now,
			IsActive:      true,
		}
	}

	// Update database last_seen
	err := s.db.Model(&postgres.User{}).
		Where("id = ?", userID).
		Update("last_seen", now).Error

	return err
}

// IsUserOnline checks if user is online
func (s *OnlineStatusService) IsUserOnline(userID uint) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	userInfo, exists := s.onlineUsers[userID]
	if !exists {
		return false
	}

	// Check if heartbeat is recent
	return time.Since(userInfo.LastHeartbeat) < HeartbeatTimeout
}

// GetOnlineUsers returns list of online users
func (s *OnlineStatusService) GetOnlineUsers() []uint {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var onlineUserIDs []uint
	now := time.Now()

	for userID, userInfo := range s.onlineUsers {
		if now.Sub(userInfo.LastHeartbeat) < HeartbeatTimeout {
			onlineUserIDs = append(onlineUserIDs, userID)
		}
	}

	return onlineUserIDs
}

// GetOnlineUserInfo returns detailed info about online user
func (s *OnlineStatusService) GetOnlineUserInfo(userID uint) (*OnlineUserInfo, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	userInfo, exists := s.onlineUsers[userID]
	if !exists {
		return nil, false
	}

	// Check if still online
	if time.Since(userInfo.LastHeartbeat) >= HeartbeatTimeout {
		return nil, false
	}

	return userInfo, true
}

// GetOnlineUserCount returns count of online users
func (s *OnlineStatusService) GetOnlineUserCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	now := time.Now()

	for _, userInfo := range s.onlineUsers {
		if now.Sub(userInfo.LastHeartbeat) < HeartbeatTimeout {
			count++
		}
	}

	return count
}

// GetOnlineFriends returns online friends of a user
func (s *OnlineStatusService) GetOnlineFriends(userID uint) ([]uint, error) {
	// Get user's friends from database
	var friendIDs []uint
	err := s.db.Table("user_friends").
		Select("friend_id").
		Where("user_id = ? AND status = ?", userID, "accepted").
		Pluck("friend_id", &friendIDs).Error

	if err != nil {
		return nil, err
	}

	// Filter online friends
	var onlineFriends []uint
	for _, friendID := range friendIDs {
		if s.IsUserOnline(friendID) {
			onlineFriends = append(onlineFriends, friendID)
		}
	}

	return onlineFriends, nil
}

// CleanupOfflineUsers removes users who haven't sent heartbeat for a while
func (s *OnlineStatusService) CleanupOfflineUsers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	var usersToRemove []uint

	for userID, userInfo := range s.onlineUsers {
		if now.Sub(userInfo.LastHeartbeat) >= HeartbeatTimeout {
			usersToRemove = append(usersToRemove, userID)
		}
	}

	// Remove offline users
	for _, userID := range usersToRemove {
		delete(s.onlineUsers, userID)
		
		// Update database
		err := s.db.Model(&postgres.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"is_online": false,
				"last_seen": s.onlineUsers[userID].LastSeen,
			}).Error

		if err != nil {
			log.Printf("Error updating offline status for user %d: %v", userID, err)
		} else {
			// Broadcast status change
			s.broadcastStatusChange(userID, false, now)
		}
	}

	if len(usersToRemove) > 0 {
		log.Printf("Cleaned up %d offline users", len(usersToRemove))
	}
}

// startCleanupJob starts background job to cleanup offline users
func (s *OnlineStatusService) startCleanupJob() {
	s.cleanupTicker = time.NewTicker(CleanupInterval)
	
	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.CleanupOfflineUsers()
			case <-s.stopChan:
				s.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Stop stops the background cleanup job
func (s *OnlineStatusService) Stop() {
	close(s.stopChan)
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
}

// RegisterStatusCallback registers a callback for status changes
func (s *OnlineStatusService) RegisterStatusCallback(callback OnlineStatusCallback) {
	onlineStatusCallbacks = append(onlineStatusCallbacks, callback)
}

// broadcastStatusChange notifies all registered callbacks about status change
func (s *OnlineStatusService) broadcastStatusChange(userID uint, isOnline bool, lastSeen time.Time) {
	// Get username for the update
	var user postgres.User
	err := s.db.Select("email").First(&user, userID).Error
	if err != nil {
		log.Printf("Error getting user info for status broadcast: %v", err)
		return
	}

	update := OnlineStatusUpdate{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: lastSeen,
		Username: user.Email, // Using email as username for now
	}

	// Call all registered callbacks
	for _, callback := range onlineStatusCallbacks {
		go callback(update) // Run in goroutine to prevent blocking
	}
}

// GetUserLastSeen returns user's last seen time
func (s *OnlineStatusService) GetUserLastSeen(userID uint) (*time.Time, error) {
	s.mutex.RLock()
	userInfo, exists := s.onlineUsers[userID]
	s.mutex.RUnlock()

	if exists {
		return &userInfo.LastSeen, nil
	}

	// Get from database
	var user postgres.User
	err := s.db.Select("last_seen").First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	return user.LastSeen, nil
}

// GetConnectionStats returns statistics about online connections
func (s *OnlineStatusService) GetConnectionStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_cached_users": len(s.onlineUsers),
		"online_users_count": s.GetOnlineUserCount(),
		"cleanup_interval":   CleanupInterval.String(),
		"heartbeat_timeout":  HeartbeatTimeout.String(),
	}

	return stats
}