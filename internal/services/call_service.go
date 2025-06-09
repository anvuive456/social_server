package services

import (
	"fmt"
	"social_server/internal/models/constants"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/repositories"
	"time"

	"github.com/google/uuid"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type CallService struct {
	callRepo repositories.CallRepository
	userRepo repositories.UserRepository
}

func NewCallService(callRepo repositories.CallRepository, userRepo repositories.UserRepository) *CallService {
	return &CallService{
		callRepo: callRepo,
		userRepo: userRepo,
	}
}

func (s *CallService) CreateCall(callerID, calleeID uint, callType, roomID string) (*postgres.Call, error) {
	// Validate users exist
	_, err := s.userRepo.GetByID(callerID)
	if err != nil {
		return nil, fmt.Errorf("caller not found: %w", err)
	}

	_, err = s.userRepo.GetByID(calleeID)
	if err != nil {
		return nil, fmt.Errorf("callee not found: %w", err)
	}

	// Check if users are friends
	isFriend, err := s.userRepo.IsFriend(callerID, calleeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check friendship: %w", err)
	}
	if !isFriend {
		return nil, fmt.Errorf("users are not friends")
	}

	// Check if caller has active call
	activeCall, err := s.callRepo.GetActiveCall(callerID)
	if err == nil && activeCall != nil {
		return nil, fmt.Errorf("caller already in active call")
	}

	// Check if callee has active call
	activeCall, err = s.callRepo.GetActiveCall(calleeID)
	if err == nil && activeCall != nil {
		return nil, fmt.Errorf("callee already in active call")
	}

	// Generate room ID if not provided
	if roomID == "" {
		roomID = s.generateRoomID()
	}

	// Create call record
	call := &postgres.Call{
		CallerID:    callerID,
		CalleeID:    &calleeID,
		Type:        callType,
		Status:      string(constants.CallStatusRinging),
		IsGroupCall: false,
		RoomID:      roomID,
	}

	err = s.callRepo.Create(call)
	if err != nil {
		return nil, fmt.Errorf("failed to create call: %w", err)
	}

	// Add caller as participant
	err = s.callRepo.JoinCall(call.ID, callerID)
	if err != nil {
		return nil, fmt.Errorf("failed to add caller as participant: %w", err)
	}

	return call, nil
}

func (s *CallService) AcceptCall(callID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.Status != string(constants.CallStatusRinging) {
		return fmt.Errorf("call is not in ringing state")
	}

	// Update call status
	updates := map[string]interface{}{
		"status":     string(constants.CallStatusOngoing),
		"started_at": time.Now(),
	}

	err = s.callRepo.Update(callID, updates)
	if err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	// Add callee as participant if not already added
	if call.CalleeID != nil {
		err = s.callRepo.JoinCall(callID, *call.CalleeID)
		if err != nil {
			// Ignore error if already joined
		}
	}

	return nil
}

func (s *CallService) DeclineCall(callID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.Status != string(constants.CallStatusRinging) {
		return fmt.Errorf("call is not in ringing state")
	}

	// Update call status
	updates := map[string]interface{}{
		"status":   string(constants.CallStatusDeclined),
		"ended_at": time.Now(),
	}

	err = s.callRepo.Update(callID, updates)
	if err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	return nil
}

func (s *CallService) EndCall(callID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.Status == string(constants.CallStatusEnded) {
		return nil // Already ended
	}

	// Calculate duration if call was ongoing
	var duration int
	if call.StartedAt != nil {
		duration = int(time.Since(*call.StartedAt).Seconds())
	}

	// Update call status
	updates := map[string]interface{}{
		"status":   string(constants.CallStatusEnded),
		"ended_at": time.Now(),
		"duration": duration,
	}

	err = s.callRepo.Update(callID, updates)
	if err != nil {
		return fmt.Errorf("failed to update call status: %w", err)
	}

	return nil
}

func (s *CallService) GetCallByID(callID uint) (*postgres.Call, error) {
	return s.callRepo.GetByID(callID)
}

func (s *CallService) GetUserCalls(userID uint, cursor paginator.Cursor, limit int) ([]postgres.Call, paginator.Cursor, error) {
	return s.callRepo.GetUserCalls(userID, cursor, limit)
}

func (s *CallService) GetActiveCall(userID uint) (*postgres.Call, error) {
	return s.callRepo.GetActiveCall(userID)
}

func (s *CallService) GetCallParticipants(callID uint) ([]postgres.CallParticipant, error) {
	return s.callRepo.GetCallParticipants(callID)
}

func (s *CallService) JoinCall(callID, userID uint) error {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	if call.Status != string(constants.CallStatusOngoing) && call.Status != string(constants.CallStatusRinging) {
		return fmt.Errorf("call is not active")
	}

	return s.callRepo.JoinCall(callID, userID)
}

func (s *CallService) LeaveCall(callID, userID uint) error {
	_, err := s.callRepo.GetByID(callID)
	if err != nil {
		return fmt.Errorf("call not found: %w", err)
	}

	err = s.callRepo.LeaveCall(callID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave call: %w", err)
	}

	// Check if all participants have left
	participants, err := s.callRepo.GetCallParticipants(callID)
	if err != nil {
		return err
	}

	activeParticipants := 0
	for _, participant := range participants {
		if participant.IsActive {
			activeParticipants++
		}
	}

	// End call if no active participants
	if activeParticipants == 0 {
		return s.EndCall(callID)
	}

	return nil
}

func (s *CallService) UpdateCallSDPs(callID uint, offerSDP, answerSDP string) error {
	updates := map[string]interface{}{}

	if offerSDP != "" {
		updates["offer_sdp"] = offerSDP
	}

	if answerSDP != "" {
		updates["answer_sdp"] = answerSDP
	}

	if len(updates) == 0 {
		return nil
	}

	return s.callRepo.Update(callID, updates)
}

func (s *CallService) CanUserAccessCall(callID, userID uint) (bool, error) {
	call, err := s.callRepo.GetByID(callID)
	if err != nil {
		return false, err
	}

	if call.CallerID == userID {
		return true, nil
	}

	if call.CalleeID != nil && *call.CalleeID == userID {
		return true, nil
	}

	// Check if user is a participant (for group calls)
	participants, err := s.callRepo.GetCallParticipants(callID)
	if err != nil {
		return false, err
	}

	for _, participant := range participants {
		if participant.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (s *CallService) GetCallHistory(userID uint, req *requests.CallHistoryQueryRequest) ([]postgres.Call, paginator.Cursor, error) {
	cursor := paginator.Cursor{
		Before: &req.Before,
		After:  &req.After,
	}

	return s.callRepo.GetUserCallHistory(userID, cursor, req.Limit)
}

func (s *CallService) generateRoomID() string {
	return fmt.Sprintf("room_%s", uuid.New().String())
}

func (s *CallService) GetCallStats(userID uint) (map[string]interface{}, error) {
	calls, _, err := s.callRepo.GetUserCalls(userID, paginator.Cursor{}, 1000)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_calls":     len(calls),
		"completed_calls": 0,
		"missed_calls":    0,
		"declined_calls":  0,
		"total_duration":  0,
	}

	for _, call := range calls {
		switch call.Status {
		case string(constants.CallStatusEnded):
			stats["completed_calls"] = stats["completed_calls"].(int) + 1
			stats["total_duration"] = stats["total_duration"].(int) + call.Duration
		case string(constants.CallStatusMissed):
			stats["missed_calls"] = stats["missed_calls"].(int) + 1
		case string(constants.CallStatusDeclined):
			stats["declined_calls"] = stats["declined_calls"].(int) + 1
		}
	}

	return stats, nil
}
