package services

import (
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// ServiceProvider manages all service instances
type ServiceProvider struct {
	UserService           UserService
	AuthService           AuthService
	TransactionService    TransactionService
	FundService           FundService
	EntityService         EntityService
	FileService           FileService
	AnalyticsService      AnalyticsService
	AuditService          AuditService
	BudgetLineItemService BudgetLineItemService
	AdminService          AdminService
	logger                *utils.NamedLogger
}

// NewServiceProvider creates a new service provider with all services initialized
func NewServiceProvider() *ServiceProvider {
	logger := utils.NewLogger("service-provider")
	logger.Info("Initializing service provider", nil)

	// Create repositories
	userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewSessionRepository()
	auditRepo := repository.NewAuditLogRepository()
	transactionRepo := repository.NewTransactionRepository()
	fundRepo := repository.NewFundRepository()
	entityRepo := repository.NewEntityRepository()
	fileRepo := repository.NewFileRepository()
	analyticsRepo := repository.NewAnalyticsRepository()
	budgetLineItemRepo := repository.NewBudgetLineItemRepository()
	adminRepo := repository.NewAdminRepository()

	// Create services
	userService := NewUserService(userRepo, auditRepo)
	authService := NewAuthService(userRepo, sessionRepo, auditRepo)
	transactionService := NewTransactionService(transactionRepo, fundRepo, entityRepo, auditRepo)
	fundService := NewFundService(fundRepo, auditRepo)
	entityService := NewEntityService(entityRepo, auditRepo)
	fileService := NewFileService(fileRepo, auditRepo)
	analyticsService := NewAnalyticsService(analyticsRepo, auditRepo)
	auditService := NewAuditService(auditRepo)
	budgetLineItemService := NewBudgetLineItemService(budgetLineItemRepo, fundRepo)
	adminService := NewAdminService(adminRepo, userRepo)

	return &ServiceProvider{
		UserService:           userService,
		AuthService:           authService,
		TransactionService:    transactionService,
		FundService:           fundService,
		EntityService:         entityService,
		FileService:           fileService,
		AnalyticsService:      analyticsService,
		AuditService:          auditService,
		BudgetLineItemService: budgetLineItemService,
		AdminService:          adminService,
		logger:                logger,
	}
}

// Initialize initializes all services and performs any necessary setup
func (sp *ServiceProvider) Initialize() error {
	sp.logger.Info("Initializing services", nil)
	// Perform any additional initialization if needed
	return nil
}

// Shutdown performs cleanup for all services
func (sp *ServiceProvider) Shutdown() error {
	sp.logger.Info("Shutting down services", nil)
	// Perform any cleanup if needed
	return nil
}
