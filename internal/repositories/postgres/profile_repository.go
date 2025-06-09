package postgres

import (
	"fmt"
	"social_server/internal/models/postgres"

	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{
		db: db,
	}
}

func (r *ProfileRepository) Create(profile *postgres.Profile) (*postgres.Profile, error) {
	if err := r.db.Create(profile).Error; err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) GetByID(id uint) (*postgres.Profile, error) {
	var profile postgres.Profile
	if err := r.db.First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return &profile, nil
}

func (r *ProfileRepository) GetByUserID(userID uint) (*postgres.Profile, error) {
	var profile postgres.Profile
	if err := r.db.First(&profile, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return &profile, nil
}

func (r *ProfileRepository) Update(profile *postgres.Profile) (*postgres.Profile, error) {
	if err := r.db.Save(profile).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) UpdatePartial(id uint, updates map[string]interface{}) (*postgres.Profile, error) {
	if err := r.db.Model(&postgres.Profile{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	var updated postgres.Profile
	if err := r.db.First(&updated, id).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve updated profile: %w", err)
	}

	return &updated, nil
}

func (r *ProfileRepository) CreateOrUpdate(profile *postgres.Profile) (*postgres.Profile, error) {
	// Try to get existing profile
	existingProfile, err := r.GetByUserID(profile.UserID)
	if err != nil {
		// Profile doesn't exist, create new one
		return r.Create(profile)
	}

	// Profile exists, update it
	profile.ID = existingProfile.ID
	return r.Update(profile)
}

func (r *ProfileRepository) Delete(id uint) error {
	if err := r.db.Delete(&postgres.Profile{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}
