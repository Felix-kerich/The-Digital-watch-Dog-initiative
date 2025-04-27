package handlers

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
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FileController handles file-related operations
type FileController struct {
	fileService services.FileService
	auditService services.AuditService
	logger      *utils.NamedLogger
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
func NewFileController(fileService services.FileService, auditService services.AuditService) *FileController {
	return &FileController{
		fileService: fileService,
		auditService: auditService,
		logger:      utils.NewLogger("file-controller"),
	}
}

// Upload handles file uploads
func (fc *FileController) Upload(c *gin.Context) {
	fc.logger.Info("Processing file upload request", nil)

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		fc.logger.Error("No file provided", map[string]interface{}{"error": err.Error()})
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
		fc.logger.Warn("File too large", map[string]interface{}{
			"fileSize": file.Size,
			"maxSize": maxSize,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File too large (max %d bytes)", maxSize)})
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
	userID, exists := c.Get("userID")
	if !exists {
		fc.logger.Error("User ID not found in context", nil)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get entity ID from form data
	entityID := c.PostForm("entityId")
	transactionID := c.PostForm("transactionId")
	fundID := c.PostForm("fundId")
	description := c.PostForm("description")
	isPublic := c.PostForm("isPublic") == "true"

	// Create a temporary file to read the data
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		fc.logger.Error("Failed to create temp file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Save the uploaded file to the temp file
	if err := c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		fc.logger.Error("Failed to save uploaded file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Read the file data
	if _, err := tempFile.Seek(0, 0); err != nil {
		fc.logger.Error("Failed to seek temp file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	fileData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		fc.logger.Error("Failed to read temp file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	// Get file extension
	fileExt := filepath.Ext(file.Filename)

	// Create a database record for the file
	fileRecord := &models.File{
		ID:           uuid.New().String(),
		FileName:     file.Filename,
		FileSize:     file.Size,
		FileType:     fileExt[1:], // Remove the dot from extension
		ContentType:  contentType,
		EntityID:     entityID,
		TransactionID: transactionID,
		FundID:       fundID,
		Description:  description,
		IsPublic:     isPublic,
		UploadedByID: userID.(string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Use the file service to upload the file
	if err := fc.fileService.UploadFile(fileRecord, fileData); err != nil {
		fc.logger.Error("Failed to upload file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	// Log the activity
	metadata := map[string]interface{}{
		"fileName": fileRecord.FileName,
		"fileSize": fileRecord.FileSize,
		"fileType": fileRecord.FileType,
	}

	if err := fc.auditService.LogActivity(
		userID.(string),
		"FILE_UPLOADED",
		"File",
		fileRecord.ID,
		metadata,
	); err != nil {
		fc.logger.Warn("Failed to log file upload activity", map[string]interface{}{"error": err.Error()})
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
	fc.logger.Info("Getting file by ID", nil)

	// Get file ID from URL parameter
	fileID := c.Param("id")
	if fileID == "" {
		fc.logger.Warn("File ID is required", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Get file from service
	file, _, err := fc.fileService.GetFileByID(fileID)
	if err != nil {
		fc.logger.Error("File not found", map[string]interface{}{"fileID": fileID, "error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Return file information
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
	fc.logger.Info("Processing file download request", nil)

	// Get file ID from URL parameter
	fileID := c.Param("id")
	if fileID == "" {
		fc.logger.Warn("File ID is required", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Get file from service
	file, fileData, err := fc.fileService.GetFileByID(fileID)
	if err != nil {
		fc.logger.Error("File not found", map[string]interface{}{"fileID": fileID, "error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Create a temporary file to serve
	tempFile, err := os.CreateTemp("", "download-*")
	if err != nil {
		fc.logger.Error("Failed to create temp file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare file for download"})
		return
	}
	defer os.Remove(tempFile.Name())

	// Write the file data to the temp file
	if _, err := tempFile.Write(fileData); err != nil {
		tempFile.Close()
		fc.logger.Error("Failed to write to temp file", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare file for download"})
		return
	}
	tempFile.Close()

	// Log the download activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"fileName": file.FileName,
			"fileSize": file.FileSize,
		}

		if err := fc.auditService.LogActivity(
			userID.(string),
			"FILE_DOWNLOADED",
			"File",
			file.ID,
			metadata,
		); err != nil {
			fc.logger.Warn("Failed to log file download activity", map[string]interface{}{"error": err.Error()})
		}
	}

	// Serve the file
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.FileName))
	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.FileSize))
	c.File(tempFile.Name())
}

// Delete deletes a file
func (fc *FileController) Delete(c *gin.Context) {
	fc.logger.Info("Processing file deletion request", nil)

	// Get file ID from URL parameter
	fileID := c.Param("id")
	if fileID == "" {
		fc.logger.Warn("File ID is required", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Get current user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		fc.logger.Error("User not authenticated", nil)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get file from service
	file, _, err := fc.fileService.GetFileByID(fileID)
	if err != nil {
		fc.logger.Error("File not found", map[string]interface{}{"fileID": fileID, "error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if user has permission to delete the file
	userRole, _ := c.Get("userRole")
	if file.UploadedByID != userID.(string) && userRole != models.RoleAdmin {
		fc.logger.Warn("Permission denied for file deletion", map[string]interface{}{
			"fileID":        fileID,
			"requestUserID": userID.(string),
			"fileOwnerID":   file.UploadedByID,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this file"})
		return
	}

	// Delete the file using service
	if err := fc.fileService.DeleteFile(fileID); err != nil {
		fc.logger.Error("Failed to delete file", map[string]interface{}{"fileID": fileID, "error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Log the deletion activity
	metadata := map[string]interface{}{
		"fileName": file.FileName,
		"fileSize": file.FileSize,
	}

	if err := fc.auditService.LogActivity(
		userID.(string),
		"FILE_DELETED",
		"File",
		file.ID,
		metadata,
	); err != nil {
		fc.logger.Warn("Failed to log file deletion activity", map[string]interface{}{"error": err.Error()})
	}

	fc.logger.Info("File deleted successfully", map[string]interface{}{"fileID": fileID})
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
