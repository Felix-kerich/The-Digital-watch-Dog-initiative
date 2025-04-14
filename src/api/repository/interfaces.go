package repository

import (
	"time"

	"github.com/the-digital-watchdog-initiative/models"
)

// UserRepository handles data access operations for User entities
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(page, limit int, filter map[string]interface{}) ([]models.User, int64, error)
	UpdateLastLogin(id string, loginTime time.Time) error
	FindAll() ([]models.User, error)
}

// TransactionRepository handles data access operations for Transaction entities
type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	FindByID(id string) (*models.Transaction, error)
	Update(transaction *models.Transaction) error
	Delete(id string) error
	List(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error)
	FindByFundID(fundID string, page, limit int) ([]models.Transaction, int64, error)
	FindByEntityID(entityID string, page, limit int) ([]models.Transaction, int64, error)
	Approve(id string, approverID string) error
	Reject(id string, rejectedByID string, reason string) error
	Complete(id string) error
	Flag(id string, reason string) error
	GetByID(id string) (*models.Transaction, error)
	GetAll(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error)
	CompleteTransaction(transaction *models.Transaction, fund *models.Fund) error
}

// FundRepository handles data access operations for Fund entities
type FundRepository interface {
	Create(fund *models.Fund) error
	FindByID(id string) (*models.Fund, error)
	Update(fund *models.Fund) error
	Delete(id string) error
	List(page, limit int, filter map[string]interface{}) ([]models.Fund, int64, error)
	FindByEntityID(entityID string, page, limit int) ([]models.Fund, int64, error)
}

// EntityRepository handles data access operations for Entity entities
type EntityRepository interface {
	Create(entity *models.Entity) error
	FindByID(id string) (*models.Entity, error)
	Update(entity *models.Entity) error
	Delete(id string) error
	List(page, limit int, filter map[string]interface{}) ([]models.Entity, int64, error)
}

// FileRepository handles data access operations for File entities
type FileRepository interface {
	Create(file *models.File) error
	FindByID(id string) (*models.File, error)
	Update(file *models.File) error
	Delete(id string) error
	List(page, limit int, filter map[string]interface{}) ([]models.File, int64, error)
	FindByEntityID(entityID string, page, limit int) ([]models.File, int64, error)
	FindByTransactionID(transactionID string, page, limit int) ([]models.File, int64, error)
	FindByFundID(fundID string, page, limit int) ([]models.File, int64, error)
}

// AuditLogRepository handles data access operations for AuditLog entities
type AuditLogRepository interface {
	Create(auditLog *models.AuditLog) error
	FindByID(id string) (*models.AuditLog, error)
	List(page, limit int, filter map[string]interface{}) ([]models.AuditLog, int64, error)
	FindByUserID(userID string, page, limit int) ([]models.AuditLog, int64, error)
	FindByEntityID(entityID string, entityType string, page, limit int) ([]models.AuditLog, int64, error)
}

// SessionRepository handles data access operations for UserSession entities
type SessionRepository interface {
	Create(session *models.UserSession) error
	FindByID(id string) (*models.UserSession, error)
	FindByToken(token string) (*models.UserSession, error)
	FindByRefreshToken(refreshToken string) (*models.UserSession, error)
	RevokeByUserID(userID string) error
	RevokeByToken(token string) error
	RevokeExpiredSessions() error
}

// AdminRepository handles administrative data operations
type AdminRepository interface {
	GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error)
	GetUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	ResetUserPassword(id string, newPassword string) error
	GetSystemStats() (map[string]interface{}, error)
	GetRecentActivity(limit int) ([]models.AuditLog, error)
	GetUserRegistrationTrends(months int) ([]map[string]interface{}, error)
}

// AnalyticsRepository handles analytics data operations
type AnalyticsRepository interface {
	GetTransactionSummary(filter map[string]interface{}) (map[string]interface{}, error)
	GetUserActivitySummary(filter map[string]interface{}) (map[string]interface{}, error)
	GetFundUtilizationReport(fundID, entityID, fiscalYear string) (map[string]interface{}, error)
}
