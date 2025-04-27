package services

import (
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/models"
)

// UserService handles business logic related to users
type UserService interface {
	// User profile operations
	GetUserProfile(userID string) (*models.User, error)
	UpdateUserProfile(userID string, userData map[string]interface{}) (*models.User, error)
	ChangeUserPassword(userID, currentPassword, newPassword string) error

	// User management operations
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	ResetUserPassword(id string, newPassword string) error
}

// AuthService handles authentication and authorization
type AuthService interface {
	Register(name, email, password, entityID string) (*models.User, error)
	Login(email, password string) (*models.UserSession, error)
	RefreshToken(refreshToken string) (*models.UserSession, error)
	Logout(token string) error
	ValidateToken(token string) (*models.User, error)
}

// TransactionService handles business logic related to transactions
type TransactionService interface {
	CreateTransaction(userID string, transactionType models.TransactionType, amount decimal.Decimal,
		currency, description, sourceID, destinationID, fundID, budgetLineItemID, documentRef string) (*models.Transaction, string, error)
	GetTransactionByID(id string) (*models.Transaction, error)
	GetAllTransactions(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error)
	GetTransactionsByFundID(fundID string, page, limit int) ([]models.Transaction, int64, error)
	GetTransactionsByEntityID(entityID string, page, limit int) ([]models.Transaction, int64, error)
	UpdateTransaction(transaction *models.Transaction) error
	ApproveTransaction(id string, approverID string) error
	RejectTransaction(id string, rejectedByID string, reason string) error
	CompleteTransaction(id string) error
	FlagTransaction(id string, reason string) error
}

// FundService handles business logic related to funds
type FundService interface {
	CreateFund(fund *models.Fund) error
	GetFundByID(id string) (*models.Fund, error)
	GetFunds(page, limit int, filter map[string]interface{}) ([]models.Fund, int64, error)
	GetFundsByEntityID(entityID string, page, limit int) ([]models.Fund, int64, error)
	UpdateFund(fund *models.Fund) error
	DeleteFund(id string) error
}

// EntityService handles business logic related to entities
type EntityService interface {
	CreateEntity(entity *models.Entity) error
	GetEntityByID(id string) (*models.Entity, error)
	GetEntities(page, limit int, filter map[string]interface{}) ([]models.Entity, int64, error)
	UpdateEntity(id string, updateData map[string]interface{}) (*models.Entity, error)
	DeleteEntity(id string) error
}

// FileService handles business logic related to files
type FileService interface {
	UploadFile(file *models.File, fileData []byte) error
	GetFileByID(id string) (*models.File, []byte, error)
	GetFiles(page, limit int, filter map[string]interface{}) ([]models.File, int64, error)
	GetFilesByEntityID(entityID string, page, limit int) ([]models.File, int64, error)
	GetFilesByTransactionID(transactionID string, page, limit int) ([]models.File, int64, error)
	GetFilesByFundID(fundID string, page, limit int) ([]models.File, int64, error)
	DeleteFile(id string) error
}

// AnalyticsService handles business logic related to analytics
type AnalyticsService interface {
	GetTransactionSummary(filter map[string]interface{}) (map[string]interface{}, error)
	GetUserActivitySummary(filter map[string]interface{}) (map[string]interface{}, error)
	GetFundUtilizationReport(fundID, entityID, fiscalYear string) (map[string]interface{}, error)
	GetSystemStats() (map[string]interface{}, error)
	GetRecentActivity(limit int) ([]models.AuditLog, error)
	GetUserRegistrationTrends(months int) ([]map[string]interface{}, error)
}

// AuditService handles business logic related to audit logging
type AuditService interface {
	LogActivity(userID, action, entityType, entityID string, metadata map[string]interface{}) error
	GetAuditLogs(page, limit int, filter map[string]interface{}) ([]models.AuditLog, int64, error)
	GetAuditLogsByUserID(userID string, page, limit int) ([]models.AuditLog, int64, error)
	GetAuditLogsByEntityID(entityID string, entityType string, page, limit int) ([]models.AuditLog, int64, error)
}

// BudgetLineItemService handles business logic related to budget line items
type BudgetLineItemService interface {
	CreateBudgetLineItem(item *models.BudgetLineItem) error
	GetBudgetLineItemByID(id string) (*models.BudgetLineItem, error)
	GetBudgetLineItems(page, limit int, filter map[string]interface{}) ([]models.BudgetLineItem, int64, error)
	GetBudgetLineItemsByFundID(fundID string, page, limit int) ([]models.BudgetLineItem, int64, error)
	UpdateBudgetLineItem(item *models.BudgetLineItem) error
	DeleteBudgetLineItem(id string) error
}

// AdminService handles business logic related to administrative operations
type AdminService interface {
	// User management operations
	GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error)
	GetUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	ResetUserPassword(id string, newPassword string) error

	// System operations
	GetSystemInfo() (map[string]interface{}, error)
}
