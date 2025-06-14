package services

import (
	"fmt"
	"social_server/internal/config"
	"social_server/internal/models/postgres"
	"social_server/internal/models/requests"
	"social_server/internal/repositories"
)

type ProfileService struct {
	profileRepo repositories.ProfileRepository
	userRepo    repositories.UserRepository
	config      *config.AuthConfig
}

func NewProfileService(profileRepo repositories.ProfileRepository, userRepo repositories.UserRepository, config *config.AuthConfig) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		userRepo:    userRepo,
		config:      config,
	}
}

func (s *ProfileService) GetMyProfile(userID uint) (*postgres.Profile, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Get user's profile
	profile, err := s.profileRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %v", err)
	}

	return profile, nil
}

func (s *ProfileService) CreateOrUpdateProfile(
	userID uint,
	req *requests.CreateProfileRequest,
	avatarURL string,
	avatarHash string,
	wallImageURL string,
	wallImageHash string,
) (*postgres.Profile, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Parse date of birth
	dateOfBirth, err := req.ParseDateOfBirth()
	if err != nil {
		return nil, fmt.Errorf("invalid date of birth format: %v", err)
	}

	// Create or update profile
	profile := &postgres.Profile{
		UserID:        userID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DisplayName:   req.DisplayName,
		Bio:           req.Bio,
		Avatar:        avatarURL,
		AvatarHash:    avatarHash,
		Phone:         req.Phone,
		DateOfBirth:   dateOfBirth,
		WallImage:     wallImageURL,
		WallImageHash: wallImageHash,
	}

	created, err := s.profileRepo.CreateOrUpdate(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update profile: %v", err)
	}

	return created, nil
}

func (s *ProfileService) UpdateProfile(
	userID uint,
	req *requests.UpdateProfileRequest,
	avatarURL string,
	avatarHash string,
	wallImageURL string,
	wallImageHash string,
) (*postgres.Profile, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Get existing profile
	existingProfile, err := s.profileRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %v", err)
	}

	// Parse date of birth
	dateOfBirth, err := req.ParseDateOfBirth()
	if err != nil {
		return nil, fmt.Errorf("invalid date of birth format: %v", err)
	}
	data := map[string]any{}
	if req.FirstName != nil {
		data["first_name"] = *req.FirstName
	}

	if req.LastName != nil {
		data["last_name"] = *req.LastName
	}
	if req.DisplayName != nil {
		data["display_name"] = *req.DisplayName
	}
	if req.Bio != nil {
		data["bio"] = *req.Bio
	}
	if req.Phone != nil {
		data["phone"] = *req.Phone
	}
	if avatarURL != "" {
		data["avatar"] = avatarURL
	}
	if avatarHash != "" {
		data["avatar_hash"] = avatarHash
	}
	if wallImageURL != "" {
		data["wall_image"] = wallImageURL
	}
	if wallImageHash != "" {
		data["wall_image_hash"] = wallImageHash
	}
	if dateOfBirth != nil {
		data["date_of_birth"] = dateOfBirth
	}

	updated, err := s.profileRepo.UpdatePartial(existingProfile.ID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	return updated, nil
}
