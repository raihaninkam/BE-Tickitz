package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AllowedImageTypes defines allowed image MIME types
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// FileUploadConfig holds configuration for file uploads
type FileUploadConfig struct {
	MaxSize     int64    // Maximum file size in bytes
	UploadDir   string   // Directory to store uploaded files
	AllowedExts []string // Allowed file extensions
}

// UploadImageFile uploads an image file with validation
func UploadImageFile(ctx *gin.Context, fieldName, subDir string, config FileUploadConfig) (string, error) {
	file, header, err := ctx.Request.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %v", err)
	}
	defer file.Close()

	// Validate file size
	if header.Size > config.MaxSize {
		return "", fmt.Errorf("file size exceeds maximum limit of %d bytes", config.MaxSize)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := false
	for _, allowedExt := range config.AllowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}
	if !validExt {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	// Create directory if not exists
	uploadPath := filepath.Join(config.UploadDir, subDir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	filename := generateUniqueFilename(header.Filename)
	fullPath := filepath.Join(uploadPath, filename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return relative path for URL (always forward slash `/`)
	relativePath := fmt.Sprintf("/images/%s/%s", subDir, filename)
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	return relativePath, nil

}

// generateUniqueFilename generates a unique filename using UUID and timestamp
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("%d_%s%s", timestamp, uuid, ext)
}

// ValidateImageFile validates if the uploaded file is a valid image
func ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size (5MB limit)
	if header.Size > 5*1024*1024 {
		return fmt.Errorf("file size must be less than 5MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}
	if !validExt {
		return fmt.Errorf("only image files are allowed (jpg, jpeg, png, gif, webp)")
	}

	return nil
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	fullPath := filepath.Join("./", filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, no need to delete
	}

	return os.Remove(fullPath)
}
