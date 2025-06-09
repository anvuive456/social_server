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

	if req.FirstName != nil {
		existingProfile.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		existingProfile.LastName = *req.LastName
	}
	if req.DisplayName != nil {
		existingProfile.DisplayName = *req.DisplayName
	}
	if req.Bio != nil {
		existingProfile.Bio = *req.Bio
	}
	if req.Phone != nil {
		existingProfile.Phone = *req.Phone
	}
	if avatarURL != "" {
		existingProfile.Avatar = avatarURL
		existingProfile.AvatarHash = avatarHash
	}
	if wallImageURL != "" {
		existingProfile.WallImage = wallImageURL
		existingProfile.WallImageHash = wallImageHash
	}
	if dateOfBirth != nil {
		existingProfile.DateOfBirth = dateOfBirth
	}

	updated, err := s.profileRepo.Update(existingProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	return updated, nil
}
