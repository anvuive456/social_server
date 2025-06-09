package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// ImageProcessResult contains information about processed image
type ImageProcessResult struct {
	ProcessedPath string `json:"processed_path"`
	OriginalPath  string `json:"original_path"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	HasRotation   bool   `json:"has_rotation"`
	Orientation   int    `json:"orientation"`
}

// ProcessImageToPNG processes an uploaded image file, handles EXIF rotation and converts to PNG
func ProcessImageToPNG(fileHeader *multipart.FileHeader, outputDir string) (*ImageProcessResult, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("no file provided")
	}

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Read file content into buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Get EXIF orientation
	orientation := 1
	hasRotation := false
	
	// Only try to read EXIF for JPEG images
	if format == "jpeg" {
		if exifOrientation, err := getEXIFOrientation(bytes.NewReader(buf.Bytes())); err == nil {
			orientation = exifOrientation
			hasRotation = orientation != 1
		}
	}

	// Apply rotation based on EXIF orientation
	if hasRotation {
		img = applyOrientation(img, orientation)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate output filename (PNG)
	originalName := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	outputFilename := fmt.Sprintf("%s_processed.png", originalName)
	outputPath := filepath.Join(outputDir, outputFilename)

	// Save as PNG
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}

	bounds := img.Bounds()
	
	return &ImageProcessResult{
		ProcessedPath: outputPath,
		OriginalPath:  fileHeader.Filename,
		Width:         bounds.Dx(),
		Height:        bounds.Dy(),
		HasRotation:   hasRotation,
		Orientation:   orientation,
	}, nil
}

// getEXIFOrientation reads EXIF orientation from image data
func getEXIFOrientation(r io.Reader) (int, error) {
	exifData, err := exif.Decode(r)
	if err != nil {
		return 1, err
	}

	orientationTag, err := exifData.Get(exif.Orientation)
	if err != nil {
		return 1, err
	}

	orientation, err := orientationTag.Int(0)
	if err != nil {
		return 1, err
	}

	return orientation, nil
}

// applyOrientation applies the correct rotation/flip based on EXIF orientation
func applyOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		// Flip horizontal
		return imaging.FlipH(img)
	case 3:
		// Rotate 180°
		return imaging.Rotate180(img)
	case 4:
		// Flip vertical
		return imaging.FlipV(img)
	case 5:
		// Rotate 90° clockwise and flip horizontal
		return imaging.FlipH(imaging.Rotate90(img))
	case 6:
		// Rotate 90° clockwise
		return imaging.Rotate90(img)
	case 7:
		// Rotate 90° counter-clockwise and flip horizontal
		return imaging.FlipH(imaging.Rotate270(img))
	case 8:
		// Rotate 90° counter-clockwise
		return imaging.Rotate270(img)
	default:
		// Orientation 1 or unknown - no transformation needed
		return img
	}
}

// ProcessAndSaveImageAsPNG processes uploaded image and saves as PNG with EXIF rotation correction
func ProcessAndSaveImageAsPNG(fileHeader *multipart.FileHeader, config *FileUploadConfig) (*FileUploadResult, *ImageProcessResult, error) {
	if fileHeader == nil {
		return nil, nil, fmt.Errorf("no file provided")
	}

	// Check if it's an image file
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, nil, fmt.Errorf("file is not an image")
	}

	// Process image to PNG
	processResult, err := ProcessImageToPNG(fileHeader, config.UploadDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process image: %v", err)
	}

	// Get file info for FileUploadResult
	fileInfo, err := os.Stat(processResult.ProcessedPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get processed file info: %v", err)
	}

	// Create FileUploadResult
	fileName := filepath.Base(processResult.ProcessedPath)
	uploadResult := &FileUploadResult{
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		FilePath:     processResult.ProcessedPath,
		FileSize:     fileInfo.Size(),
		ContentType:  "image/png",
		URL:          filepath.Join(config.BaseURL, fileName),
	}

	return uploadResult, processResult, nil
}

// IsImageFile checks if the file is an image based on content type
func IsImageFile(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}
	
	for _, imgType := range imageTypes {
		if strings.EqualFold(contentType, imgType) {
			return true
		}
	}
	return false
}