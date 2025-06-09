package requests

import (
	"mime/multipart"
	"time"
)

// CreateProfileRequest for creating user profile with multipart form
type CreateProfileRequest struct {
	FirstName   string `form:"first_name" binding:"required"`
	LastName    string `form:"last_name" binding:"required"`
	DisplayName string `form:"display_name" binding:"required"`
	Bio         string `form:"bio"`
	Phone       string `form:"phone"`
	DateOfBirth string `form:"date_of_birth"` // Format: 2006-01-02

	Avatar    *multipart.FileHeader `form:"avatar"`
	WallImage *multipart.FileHeader `form:"wall_image"`
}

// UpdateProfileRequest for updating user profile with multipart form
type UpdateProfileRequest struct {
	FirstName   *string `form:"first_name,omitempty"`
	LastName    *string `form:"last_name,omitempty"`
	DisplayName *string `form:"display_name,omitempty"`
	Bio         *string `form:"bio,omitempty"`
	Phone       *string `form:"phone,omitempty"`
	DateOfBirth *string `form:"date_of_birth,omitempty"` // Format: 2006-01-02

	Avatar    *multipart.FileHeader `form:"avatar,omitempty"`
	WallImage *multipart.FileHeader `form:"wall_image,omitempty"`
}

// ParseDateOfBirth parses date string to time.Time
func (r *CreateProfileRequest) ParseDateOfBirth() (*time.Time, error) {
	if r.DateOfBirth == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", r.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseDateOfBirth parses date string to time.Time
func (r *UpdateProfileRequest) ParseDateOfBirth() (*time.Time, error) {
	if r.DateOfBirth == nil {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", *r.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

type DeleteProfileRequest struct {
	Password string `json:"password" binding:"required"`
}
