package handlers

import (
	"image/png"
	"net/http"
	"os"
	"social_server/internal/middleware"
	"social_server/internal/models/requests"
	"social_server/internal/services"
	"social_server/internal/utils"

	"github.com/buckket/go-blurhash"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// GetMyProfile gets current user's profile
// @Summary Get my profile
// @Description Get current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Profile not found"
// @Router /profile [get]
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	profile, err := h.profileService.GetMyProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "profile_not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

// CreateOrUpdateProfile creates or updates profile for current user
// @Summary Create or update profile
// @Description Create or update profile for authenticated user using multipart form data
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param first_name formData string true "First name"
// @Param last_name formData string true "Last name"
// @Param display_name formData string true "Display name"
// @Param bio formData string false "Bio"
// @Param phone formData string false "Phone number"
// @Param date_of_birth formData string false "Date of birth (YYYY-MM-DD)"
// @Param avatar formData file false "Avatar image"
// @Param wall_image formData file false "Wall image"
// @Success 200 {object} map[string]interface{} "Created or updated profile"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /profile [post]
func (h *ProfileHandler) CreateOrUpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.CreateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Handle avatar upload if provided
	var avatarURL string
	var avatarHash string
	if req.Avatar != nil {
		config := utils.DefaultAvatarConfig()
		uploadResult, err := utils.UploadImageAsPNG(req.Avatar, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_upload_failed",
				"message": err.Error(),
			})
			return
		}
		avatarURL = uploadResult.URL
		// Generate avatar hash
		imageFile, err := os.Open(uploadResult.FilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_open_failed",
				"message": err.Error(),
			})
			return
		}
		defer imageFile.Close()
		loadedImage, err := png.Decode(imageFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_decode_failed",
				"message": err.Error(),
			})
			return
		}
		str, err := blurhash.Encode(4, 3, loadedImage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_hash_failed",
				"message": err.Error(),
			})
			return
		}
		avatarHash = str
		// End avatar processing

	}

	var wallImageURL string
	var wallImageHash string

	if req.WallImage != nil {
		config := utils.DefaultAvatarConfig()
		uploadResult, err := utils.UploadImageAsPNG(req.WallImage, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_upload_failed",
				"message": err.Error(),
			})
			return
		}
		wallImageURL = uploadResult.URL
		// Generate avatar hash
		imageFile, err := os.Open(uploadResult.FilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_open_failed",
				"message": err.Error(),
			})
			return
		}
		defer imageFile.Close()
		loadedImage, err := png.Decode(imageFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_decode_failed",
				"message": err.Error(),
			})
			return
		}
		str, err := blurhash.Encode(4, 3, loadedImage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_hash_failed",
				"message": err.Error(),
			})
			return
		}
		wallImageHash = str
		// End avatar processing
	}

	profile, err := h.profileService.CreateOrUpdateProfile(userID, &req, avatarURL, avatarHash, wallImageURL, wallImageHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "create_or_update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    profile,
		"message": "Profile created or updated successfully",
	})
}



// UpdateProfile updates current user's profile
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param first_name formData string false "First name"
// @Param last_name formData string false "Last name"
// @Param display_name formData string false "Display name"
// @Param bio formData string false "Bio"
// @Param phone formData string false "Phone number"
// @Param date_of_birth formData string false "Date of birth (YYYY-MM-DD)"
// @Param avatar formData file false "Avatar image"
// @Param wall_image formData file false "Wall image"
// @Success 200 {object} map[string]interface{} "Updated profile"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /profile [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req requests.UpdateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Handle avatar upload if provided
	var avatarURL string
	var avatarHash string

	if req.Avatar != nil {
		config := utils.DefaultAvatarConfig()
		uploadResult, err := utils.UploadImageAsPNG(req.Avatar, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_upload_failed",
				"message": err.Error(),
			})
			return
		}
		avatarURL = uploadResult.URL
		// Generate avatar hash
		imageFile, err := os.Open(uploadResult.FilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_open_failed",
				"message": err.Error(),
			})
			return
		}
		defer imageFile.Close()
		loadedImage, err := png.Decode(imageFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_decode_failed",
				"message": err.Error(),
			})
			return
		}
		str, err := blurhash.Encode(4, 3, loadedImage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_hash_failed",
				"message": err.Error(),
			})
			return
		}
		avatarHash = str
		// End avatar processing
	}

	var wallImageURL string
	var wallImageHash string

	if req.WallImage != nil {
		config := utils.DefaultAvatarConfig()
		uploadResult, err := utils.UploadImageAsPNG(req.WallImage, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_upload_failed",
				"message": err.Error(),
			})
			return
		}
		wallImageURL = uploadResult.URL
		// Generate avatar hash
		imageFile, err := os.Open(uploadResult.FilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_open_failed",
				"message": err.Error(),
			})
			return
		}
		defer imageFile.Close()
		loadedImage, err := png.Decode(imageFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_decode_failed",
				"message": err.Error(),
			})
			return
		}
		str, err := blurhash.Encode(4, 3, loadedImage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "avatar_hash_failed",
				"message": err.Error(),
			})
			return
		}
		wallImageHash = str
		// End avatar processing
	}

	profile, err := h.profileService.UpdateProfile(userID, &req, avatarURL, avatarHash, wallImageURL, wallImageHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    profile,
		"message": "Profile updated successfully",
	})
}
