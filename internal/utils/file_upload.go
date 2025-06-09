package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileUploadResult contains information about uploaded file
type FileUploadResult struct {
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	URL          string `json:"url"`
	// Image processing info (optional)
	ImageInfo    *ImageProcessInfo `json:"image_info,omitempty"`
}

// ImageProcessInfo contains information about processed image
type ImageProcessInfo struct {
	Width         int  `json:"width"`
	Height        int  `json:"height"`
	HasRotation   bool `json:"has_rotation"`
	Orientation   int  `json:"orientation"`
	ConvertedToPNG bool `json:"converted_to_png"`
}

// FileUploadConfig contains configuration for file upload
type FileUploadConfig struct {
	UploadDir    string
	MaxFileSize  int64
	AllowedTypes []string
	BaseURL      string
}

// DefaultAvatarConfig returns default configuration for avatar uploads
func DefaultAvatarConfig() *FileUploadConfig {
	return &FileUploadConfig{
		UploadDir:    "./uploads/avatars",
		MaxFileSize:  5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/jpg", "image/gif", "image/webp"},
		BaseURL:      "/uploads/avatars",
	}
}

// UploadFile uploads a file with given configuration
func UploadFile(fileHeader *multipart.FileHeader, config *FileUploadConfig) (*FileUploadResult, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("no file provided")
	}

	// Check file size
	if fileHeader.Size > config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of %d bytes", config.MaxFileSize)
	}

	// Check file type
	if len(config.AllowedTypes) > 0 {
		contentType := fileHeader.Header.Get("Content-Type")
		if !isAllowedType(contentType, config.AllowedTypes) {
			return nil, fmt.Errorf("file type %s is not allowed", contentType)
		}
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(config.UploadDir, fileName)

	// Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	// Return upload result
	return &FileUploadResult{
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		FilePath:     filePath,
		FileSize:     fileHeader.Size,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		URL:          filepath.Join(config.BaseURL, fileName),
	}, nil
}

// UploadImageAsPNG uploads an image file, processes EXIF rotation and converts to PNG
func UploadImageAsPNG(fileHeader *multipart.FileHeader, config *FileUploadConfig) (*FileUploadResult, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("no file provided")
	}

	// Check file size
	if fileHeader.Size > config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of %d bytes", config.MaxFileSize)
	}

	// Check if it's an image file
	contentType := fileHeader.Header.Get("Content-Type")
	if !IsImageFile(contentType) {
		return nil, fmt.Errorf("file type %s is not a supported image format", contentType)
	}

	// Check file type against allowed types
	if len(config.AllowedTypes) > 0 {
		if !isAllowedType(contentType, config.AllowedTypes) {
			return nil, fmt.Errorf("file type %s is not allowed", contentType)
		}
	}

	// Process image and save as PNG
	uploadResult, processResult, err := ProcessAndSaveImageAsPNG(fileHeader, config)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %v", err)
	}

	// Add image processing info to result
	uploadResult.ImageInfo = &ImageProcessInfo{
		Width:         processResult.Width,
		Height:        processResult.Height,
		HasRotation:   processResult.HasRotation,
		Orientation:   processResult.Orientation,
		ConvertedToPNG: true,
	}

	return uploadResult, nil
}

// DeleteFile deletes a file from the upload directory
func DeleteFile(fileName string, uploadDir string) error {
	if fileName == "" {
		return nil
	}

	// Clean the filename to prevent path traversal
	fileName = filepath.Base(fileName)
	filePath := filepath.Join(uploadDir, fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	// Delete the file
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// isAllowedType checks if the content type is in the allowed types list
func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
	}
	return false
}

// GetFileURL constructs the full URL for a file
func GetFileURL(fileName string, baseURL string) string {
	if fileName == "" {
		return ""
	}
	return filepath.Join(baseURL, fileName)
}

// ExtractFileNameFromURL extracts filename from URL
func ExtractFileNameFromURL(url string) string {
	if url == "" {
		return ""
	}
	return filepath.Base(url)
}
