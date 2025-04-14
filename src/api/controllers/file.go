package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FileController handles file-related operations
type FileController struct {
	fileRepo  repository.FileRepository
	auditRepo repository.AuditLogRepository
}

// FileResponse represents a file response
type FileResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"fileName"`
	Size        int64     `json:"fileSize"`
	ContentType string    `json:"contentType"`
	UploadedBy  string    `json:"uploadedById"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewFileController creates a new file controller
func NewFileController() *FileController {
	return &FileController{
		fileRepo:  repository.NewFileRepository(),
		auditRepo: repository.NewAuditLogRepository(),
	}
}

// Upload handles file uploads
func (fc *FileController) Upload(c *gin.Context) {
	// Get upload directory from environment or use default
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Check file size
	maxSizeStr := os.Getenv("MAX_UPLOAD_SIZE")
	maxSize := int64(10 * 1024 * 1024) // Default to 10MB
	if maxSizeStr != "" {
		if parsedSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = parsedSize
		}
	}

	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File too large (max %d bytes)", maxSize)})
		return
	}

	// Get file extension and generate a unique filename
	fileExt := filepath.Ext(file.Filename)
	fileID := uuid.New().String()
	fileName := fileID + fileExt
	filePath := filepath.Join(uploadDir, fileName)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get content type
	contentType := "application/octet-stream" // Default
	if file.Header != nil {
		if ct := file.Header.Get("Content-Type"); ct != "" {
			contentType = ct
		}
	}

	// Get user ID from context
	userID, _ := c.Get("userID")

	// Get entity ID from form data
	entityID := c.PostForm("entityId")
	transactionID := c.PostForm("transactionId")
	fundID := c.PostForm("fundId")
	description := c.PostForm("description")
	isPublic := c.PostForm("isPublic") == "true"

	// Create a database record for the file
	fileRecord := models.File{
		FileName:      file.Filename,
		FilePath:      filePath,
		FileSize:      file.Size,
		FileType:      fileExt[1:], // Remove the dot from extension
		ContentType:   contentType,
		EntityID:      entityID,
		TransactionID: transactionID,
		FundID:        fundID,
		Description:   description,
		IsPublic:      isPublic,
		UploadedByID:  userID.(string),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := fc.fileRepo.Create(&fileRecord); err != nil {
		// Try to remove the file if the record creation failed
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record file metadata"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "FILE_UPLOADED",
		UserID:     userID.(string),
		EntityID:   fileRecord.ID,
		EntityType: "File",
		Detail:     fmt.Sprintf("Uploaded file: %s (%d bytes)", file.Filename, file.Size),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for file upload")
	}

	// Return the file information
	c.JSON(http.StatusOK, gin.H{
		"id":            fileRecord.ID,
		"fileName":      fileRecord.FileName,
		"fileSize":      fileRecord.FileSize,
		"contentType":   fileRecord.ContentType,
		"entityId":      fileRecord.EntityID,
		"transactionId": fileRecord.TransactionID,
		"fundId":        fileRecord.FundID,
		"isPublic":      fileRecord.IsPublic,
		"createdAt":     fileRecord.CreatedAt,
	})
}

// GetByID retrieves a file by its ID
func (fc *FileController) GetByID(c *gin.Context) {
	id := c.Param("id")

	file, err := fc.fileRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, FileResponse{
		ID:          file.ID,
		Name:        file.FileName,
		Size:        file.FileSize,
		ContentType: file.ContentType,
		UploadedBy:  file.UploadedByID,
		CreatedAt:   file.CreatedAt,
	})
}

// Download downloads a file
func (fc *FileController) Download(c *gin.Context) {
	id := c.Param("id")

	file, err := fc.fileRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("userID")

	// Create audit log for download
	auditLog := models.AuditLog{
		Action:     "FILE_DOWNLOADED",
		UserID:     userID.(string),
		EntityID:   file.ID,
		EntityType: "File",
		Detail:     fmt.Sprintf("Downloaded file: %s", file.FileName),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for file download")
	}

	// Set appropriate headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.FileName))
	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Length", strconv.FormatInt(file.FileSize, 10))
	c.File(file.FilePath)
}

// Delete deletes a file
func (fc *FileController) Delete(c *gin.Context) {
	id := c.Param("id")

	file, err := fc.fileRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Check permissions
	if userRole != models.RoleAdmin && file.UploadedByID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this file"})
		return
	}

	// Delete the file
	if err := fc.fileRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Create audit log
	auditLog := models.AuditLog{
		Action:     "FILE_DELETED",
		UserID:     userID.(string),
		EntityID:   file.ID,
		EntityType: "File",
		Detail:     fmt.Sprintf("Deleted file: %s", file.FileName),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := fc.auditRepo.Create(&auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for file deletion")
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
