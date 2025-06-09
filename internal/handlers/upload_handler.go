package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"social_server/internal/middleware"
)

type UploadHandler struct {
	uploadDir    string
	maxFileSize  int64
	allowedTypes map[string][]string
}

type UploadResponse struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

type MultiUploadResponse struct {
	Files []UploadResponse `json:"files"`
	Count int              `json:"count"`
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	// Create upload directories if they don't exist
	dirs := []string{
		filepath.Join(uploadDir, "avatars"),
		filepath.Join(uploadDir, "images"),
		filepath.Join(uploadDir, "videos"),
		filepath.Join(uploadDir, "documents"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create upload directory %s: %v", dir, err))
		}
	}

	return &UploadHandler{
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024, // 50MB default
		allowedTypes: map[string][]string{
			"avatar": {"image/jpeg", "image/png", "image/gif", "image/webp"},
			"image":  {"image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp"},
			"video":  {"video/mp4", "video/avi", "video/mov", "video/wmv", "video/webm"},
			"document": {"application/pdf", "text/plain", "application/msword",
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		},
	}
}

func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file",
			"message": "No file provided or invalid file",
		})
		return
	}
	defer file.Close()

	// Validate file
	if err := h.validateFile(header, "avatar", 5*1024*1024); err != nil { // 5MB for avatars
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)

	// Save file
	uploadPath := filepath.Join(h.uploadDir, "avatars", filename)
	if err := h.saveFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "save_failed",
			"message": "Failed to save file",
		})
		return
	}

	// Return response
	response := UploadResponse{
		Filename: filename,
		URL:      fmt.Sprintf("/uploads/avatars/%s", filename),
		Size:     header.Size,
		Type:     header.Header.Get("Content-Type"),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UploadHandler) UploadImages(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_form",
			"message": "Invalid multipart form",
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no_files",
			"message": "No files provided",
		})
		return
	}

	if len(files) > 10 { // Limit to 10 images per upload
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "too_many_files",
			"message": "Maximum 10 files allowed per upload",
		})
		return
	}

	var uploadedFiles []UploadResponse
	var totalSize int64

	for _, header := range files {
		// Validate file
		if err := h.validateFile(header, "image", 10*1024*1024); err != nil { // 10MB for images
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"message": fmt.Sprintf("File %s: %s", header.Filename, err.Error()),
			})
			return
		}

		totalSize += header.Size
		if totalSize > 100*1024*1024 { // 100MB total limit
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "size_limit_exceeded",
				"message": "Total upload size exceeds 100MB limit",
			})
			return
		}

		// Open file
		file, err := header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "file_open_failed",
				"message": "Failed to open file",
			})
			return
		}
		defer file.Close()

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d_%s_%s%s", userID, time.Now().Format("20060102_150405"), uuid.New().String()[:8], ext)

		// Save file
		uploadPath := filepath.Join(h.uploadDir, "images", filename)
		if err := h.saveFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "save_failed",
				"message": "Failed to save file",
			})
			return
		}

		uploadedFiles = append(uploadedFiles, UploadResponse{
			Filename: filename,
			URL:      fmt.Sprintf("/uploads/images/%s", filename),
			Size:     header.Size,
			Type:     header.Header.Get("Content-Type"),
		})
	}

	response := MultiUploadResponse{
		Files: uploadedFiles,
		Count: len(uploadedFiles),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UploadHandler) UploadVideos(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_form",
			"message": "Invalid multipart form",
		})
		return
	}

	files := form.File["videos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no_files",
			"message": "No files provided",
		})
		return
	}

	if len(files) > 5 { // Limit to 5 videos per upload
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "too_many_files",
			"message": "Maximum 5 videos allowed per upload",
		})
		return
	}

	var uploadedFiles []UploadResponse

	for _, header := range files {
		// Validate file
		if err := h.validateFile(header, "video", 100*1024*1024); err != nil { // 100MB for videos
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"message": fmt.Sprintf("File %s: %s", header.Filename, err.Error()),
			})
			return
		}

		// Open file
		file, err := header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "file_open_failed",
				"message": "Failed to open file",
			})
			return
		}
		defer file.Close()

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d_%s_%s%s", userID, time.Now().Format("20060102_150405"), uuid.New().String()[:8], ext)

		// Save file
		uploadPath := filepath.Join(h.uploadDir, "videos", filename)
		if err := h.saveFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "save_failed",
				"message": "Failed to save file",
			})
			return
		}

		uploadedFiles = append(uploadedFiles, UploadResponse{
			Filename: filename,
			URL:      fmt.Sprintf("/uploads/videos/%s", filename),
			Size:     header.Size,
			Type:     header.Header.Get("Content-Type"),
		})
	}

	response := MultiUploadResponse{
		Files: uploadedFiles,
		Count: len(uploadedFiles),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UploadHandler) DeleteFile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	filename := c.Param("filename")
	fileType := c.Param("type") // avatar, image, video, document

	if filename == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Filename and type are required",
		})
		return
	}

	// Verify user owns the file (filename should start with user ID)
	if !strings.HasPrefix(filename, fmt.Sprintf("%d", userID)) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have permission to delete this file",
		})
		return
	}

	// Validate file type
	validTypes := map[string]bool{
		"avatars":   true,
		"images":    true,
		"videos":    true,
		"documents": true,
	}

	if !validTypes[fileType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_type",
			"message": "Invalid file type",
		})
		return
	}

	// Delete file
	filePath := filepath.Join(h.uploadDir, fileType, filename)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "file_not_found",
				"message": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete_failed",
			"message": "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

func (h *UploadHandler) GetFileInfo(c *gin.Context) {
	filename := c.Param("filename")
	fileType := c.Param("type")

	if filename == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Filename and type are required",
		})
		return
	}

	filePath := filepath.Join(h.uploadDir, fileType, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "file_not_found",
				"message": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "stat_failed",
			"message": "Failed to get file info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename":    filename,
		"size":        info.Size(),
		"modified_at": info.ModTime(),
		"url":         fmt.Sprintf("/uploads/%s/%s", fileType, filename),
	})
}

func (h *UploadHandler) validateFile(header *multipart.FileHeader, uploadType string, maxSize int64) error {
	// Check file size
	if header.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, maxSize)
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	allowedTypes, exists := h.allowedTypes[uploadType]
	if !exists {
		return fmt.Errorf("invalid upload type: %s", uploadType)
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("file type %s is not allowed for %s uploads", contentType, uploadType)
}

func (h *UploadHandler) saveFile(src multipart.File, dst string) error {
	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy file content
	_, err = io.Copy(out, src)
	return err
}

func (h *UploadHandler) ServeFile(c *gin.Context) {
	fileType := c.Param("type")
	filename := c.Param("filename")

	// Validate file type
	validTypes := map[string]bool{
		"avatars":   true,
		"images":    true,
		"videos":    true,
		"documents": true,
	}

	if !validTypes[fileType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_type",
			"message": "Invalid file type",
		})
		return
	}

	filePath := filepath.Join(h.uploadDir, fileType, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "file_not_found",
			"message": "File not found",
		})
		return
	}

	// Set proper headers for caching
	c.Header("Cache-Control", "public, max-age=31536000") // 1 year
	c.Header("ETag", fmt.Sprintf("\"%s\"", filename))

	// Serve file
	c.File(filePath)
}

func (h *UploadHandler) CleanupOldFiles(c *gin.Context) {
	// This endpoint should be protected and only accessible by admin
	// Implementation for cleaning up old unused files
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	var deletedCount int

	// Walk through upload directories
	for _, subdir := range []string{"avatars", "images", "videos", "documents"} {
		dirPath := filepath.Join(h.uploadDir, subdir)

		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking
			}

			if !info.IsDir() && info.ModTime().Before(cutoffTime) {
				if err := os.Remove(path); err == nil {
					deletedCount++
				}
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "cleanup_failed",
				"message": fmt.Sprintf("Failed to cleanup directory %s: %v", subdir, err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Cleanup completed",
		"deleted_count": deletedCount,
		"cutoff_date":   cutoffTime.Format("2006-01-02"),
	})
}
