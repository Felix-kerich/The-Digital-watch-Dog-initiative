package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FileServiceImpl implements FileService interface
type FileServiceImpl struct {
	fileRepo  repository.FileRepository
	auditRepo repository.AuditLogRepository
	logger    *utils.NamedLogger
}

// NewFileService creates a new file service
func NewFileService(fileRepo repository.FileRepository, auditRepo repository.AuditLogRepository) FileService {
	return &FileServiceImpl{
		fileRepo:  fileRepo,
		auditRepo: auditRepo,
		logger:    utils.NewLogger("file-service"),
	}
}

// UploadFile uploads a new file
func (s *FileServiceImpl) UploadFile(file *models.File, fileData []byte) error {
	s.logger.Info("Uploading file", map[string]interface{}{
		"fileName": file.FileName,
		"fileSize": len(fileData),
	})

	// Generate ID if not provided
	if file.ID == "" {
		file.ID = uuid.New().String()
	}

	// Set creation timestamp if not set
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}

	// Create upload directory if it doesn't exist
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		s.logger.Error("Failed to create upload directory", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Generate file path
	filePath := filepath.Join(uploadDir, file.ID)
	
	// Write file to disk
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		s.logger.Error("Failed to write file to disk", map[string]interface{}{
			"fileName": file.FileName,
			"error":    err.Error(),
		})
		return err
	}

	// Set file path
	file.FilePath = filePath

	// Create the file record in database
	if err := s.fileRepo.Create(file); err != nil {
		s.logger.Error("Failed to create file record", map[string]interface{}{
			"fileName": file.FileName,
			"error":    err.Error(),
		})
		
		// Clean up file if database insertion fails
		os.Remove(filePath)
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     file.UploadedByID,
		Action:     "UPLOAD_FILE",
		EntityType: "FILE",
		EntityID:   file.ID,
		Timestamp:  time.Now(),
		Detail:     fmt.Sprintf("Uploaded file: %s", file.FileName),
	})

	return nil
}

// GetFileByID retrieves a file by ID
func (s *FileServiceImpl) GetFileByID(id string) (*models.File, []byte, error) {
	s.logger.Info("Getting file by ID", map[string]interface{}{"fileID": id})

	// Get file metadata
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get file metadata", map[string]interface{}{
			"fileID": id,
			"error":  err.Error(),
		})
		return nil, nil, err
	}

	// Read file from disk
	fileData, err := os.ReadFile(file.FilePath)
	if err != nil {
		s.logger.Error("Failed to read file from disk", map[string]interface{}{
			"fileID": id,
			"path":   file.FilePath,
			"error":  err.Error(),
		})
		return nil, nil, err
	}

	return file, fileData, nil
}

// GetFiles retrieves files with pagination and filtering
func (s *FileServiceImpl) GetFiles(page, limit int, filter map[string]interface{}) ([]models.File, int64, error) {
	s.logger.Info("Getting files list", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	files, total, err := s.fileRepo.List(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get files list", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return files, total, nil
}

// GetFilesByEntityID retrieves files for a specific entity
func (s *FileServiceImpl) GetFilesByEntityID(entityID string, page, limit int) ([]models.File, int64, error) {
	s.logger.Info("Getting files by entity ID", map[string]interface{}{
		"entityID": entityID,
		"page":     page,
		"limit":    limit,
	})

	files, total, err := s.fileRepo.FindByEntityID(entityID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get files by entity ID", map[string]interface{}{
			"entityID": entityID,
			"page":     page,
			"limit":    limit,
			"error":    err.Error(),
		})
		return nil, 0, err
	}

	return files, total, nil
}

// GetFilesByTransactionID retrieves files for a specific transaction
func (s *FileServiceImpl) GetFilesByTransactionID(transactionID string, page, limit int) ([]models.File, int64, error) {
	s.logger.Info("Getting files by transaction ID", map[string]interface{}{
		"transactionID": transactionID,
		"page":          page,
		"limit":         limit,
	})

	files, total, err := s.fileRepo.FindByTransactionID(transactionID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get files by transaction ID", map[string]interface{}{
			"transactionID": transactionID,
			"page":          page,
			"limit":         limit,
			"error":         err.Error(),
		})
		return nil, 0, err
	}

	return files, total, nil
}

// GetFilesByFundID retrieves files for a specific fund
func (s *FileServiceImpl) GetFilesByFundID(fundID string, page, limit int) ([]models.File, int64, error) {
	s.logger.Info("Getting files by fund ID", map[string]interface{}{
		"fundID": fundID,
		"page":   page,
		"limit":  limit,
	})

	files, total, err := s.fileRepo.FindByFundID(fundID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get files by fund ID", map[string]interface{}{
			"fundID": fundID,
			"page":   page,
			"limit":  limit,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return files, total, nil
}

// DeleteFile deletes a file
func (s *FileServiceImpl) DeleteFile(id string) error {
	s.logger.Info("Deleting file", map[string]interface{}{"fileID": id})

	// Get file metadata
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get file metadata for deletion", map[string]interface{}{
			"fileID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Delete file from disk
	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("Failed to delete file from disk", map[string]interface{}{
			"fileID": id,
			"path":   file.FilePath,
			"error":  err.Error(),
		})
		// Continue with database deletion even if file removal fails
	}

	// Delete file record from database
	if err := s.fileRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete file record", map[string]interface{}{
			"fileID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		Action:     "DELETE_FILE",
		EntityType: "FILE",
		EntityID:   id,
		Timestamp:  time.Now(),
	})

	return nil
}
