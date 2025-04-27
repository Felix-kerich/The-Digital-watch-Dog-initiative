package services

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// TransactionServiceImpl implements TransactionService interface
type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
	fundRepo        repository.FundRepository
	entityRepo      repository.EntityRepository
	auditRepo       repository.AuditLogRepository
	blockchain      *utils.BlockchainService
	logger          *utils.NamedLogger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	fundRepo repository.FundRepository,
	entityRepo repository.EntityRepository,
	auditRepo repository.AuditLogRepository,
) TransactionService {
	// Initialize blockchain service
	blockchain, err := utils.NewBlockchainService()
	if err != nil {
		log.Printf("Warning: Failed to initialize blockchain service: %v", err)
	}

	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		fundRepo:        fundRepo,
		entityRepo:      entityRepo,
		auditRepo:       auditRepo,
		blockchain:      blockchain,
		logger:          utils.NewLogger("transaction-service"),
	}
}

// CreateTransaction creates a new transaction
func (s *TransactionServiceImpl) CreateTransaction(userID string, transactionType models.TransactionType, amount decimal.Decimal,
	currency, description, sourceID, destinationID, fundID, budgetLineItemID, documentRef string) (*models.Transaction, string, error) {
	s.logger.Info("Creating new transaction", map[string]interface{}{
		"type":          transactionType,
		"amount":        amount,
		"fundID":        fundID,
		"sourceID":      sourceID,
		"destinationID": destinationID,
		"userID":        userID,
	})

	// Validate fund exists
	fund, err := s.fundRepo.FindByID(fundID)
	if err != nil {
		s.logger.Error("Failed to find fund for transaction", map[string]interface{}{
			"fundID": fundID,
			"error":  err.Error(),
		})
		return nil, "", &utils.ValidationError{Message: "Invalid fund ID"}
	}

	// Validate source entity exists
	_, err = s.entityRepo.FindByID(sourceID)
	if err != nil {
		s.logger.Error("Failed to find source entity for transaction", map[string]interface{}{
			"sourceID": sourceID,
			"error":    err.Error(),
		})
		return nil, "", &utils.ValidationError{Message: "Invalid source entity ID"}
	}

	// Validate destination entity exists
	_, err = s.entityRepo.FindByID(destinationID)
	if err != nil {
		s.logger.Error("Failed to find destination entity for transaction", map[string]interface{}{
			"destinationID": destinationID,
			"error":         err.Error(),
		})
		return nil, "", &utils.ValidationError{Message: "Invalid destination entity ID"}
	}

	// Validate transaction type
	if !isValidTransactionType(transactionType) {
		s.logger.Error("Invalid transaction type", map[string]interface{}{
			"type": transactionType,
		})
		return nil, "", &utils.ValidationError{Message: "Invalid transaction type"}
	}

	// Check fund balance for withdrawals
	if transactionType == models.TransactionDisbursement || transactionType == models.TransactionExpenditure {
		// Calculate available balance: Amount - (Allocated + Disbursed + Utilized)
		availableBalance := fund.Amount.Sub(fund.Allocated.Add(fund.Disbursed).Add(fund.Utilized))
		if availableBalance.LessThan(amount) {
			s.logger.Warn("Insufficient fund balance for transaction", map[string]interface{}{
				"fundID":            fund.ID,
				"availableBalance":  availableBalance,
				"transactionAmount": amount,
			})
			return nil, "", &utils.ValidationError{Message: "Insufficient fund balance"}
		}
	}

	// Create the transaction object
	transactionID := uuid.New().String()
	createdByIDPtr := &userID

	transaction := &models.Transaction{
		ID:               transactionID,
		TransactionType:  transactionType,
		Amount:           amount,
		Currency:         currency,
		Status:           models.TransactionPending,
		Description:      description,
		SourceID:         sourceID,
		DestinationID:    destinationID,
		FundID:           fundID,
		BudgetLineItemID: budgetLineItemID,
		DocumentRef:      documentRef,
		CreatedByID:      createdByIDPtr,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Record transaction in blockchain
	if s.blockchain != nil {
		blockchainData := &utils.TransactionData{
			ID:               transaction.ID,
			Amount:           amount.InexactFloat64(),
			Currency:         currency,
			Type:             string(transactionType),
			Description:      description,
			SourceID:         sourceID,
			DestinationID:    destinationID,
			FundID:           fundID,
			BudgetLineItemID: budgetLineItemID,
			DocumentRef:      documentRef,
			CreatedByID:      userID,
		}

		txHash, err := s.blockchain.RecordTransaction(blockchainData)
		if err != nil {
			s.logger.Error("Failed to record transaction on blockchain", map[string]interface{}{
				"error": err.Error(),
				"txID":  transaction.ID,
			})
			// Continue with database transaction even if blockchain fails
		} else {
			transaction.BlockchainTxHash = txHash
			transaction.BlockchainStatus = "RECORDED"
		}
	}

	// Create the transaction in database
	if err := s.transactionRepo.Create(transaction); err != nil {
		s.logger.Error("Failed to create transaction", map[string]interface{}{
			"transactionID": transactionID,
			"error":         err.Error(),
		})
		return nil, "", err
	}

	// Log the activity
	err = s.auditRepo.Create(&models.AuditLog{
		UserID:     *transaction.CreatedByID,
		Action:     "CREATE_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   transaction.ID,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"type":     transaction.TransactionType,
			"amount":   transaction.Amount,
			"fundID":   transaction.FundID,
			"sourceID": transaction.SourceID,
		},
	})
	if err != nil {
		s.logger.Error("Failed to log transaction creation", map[string]interface{}{
			"transactionID": transaction.ID,
			"error":         err.Error(),
		})
		// Continue even if logging fails
	}

	return transaction, transaction.BlockchainTxHash, nil
}

// GetTransactionByID retrieves a transaction by ID
func (s *TransactionServiceImpl) GetTransactionByID(id string) (*models.Transaction, error) {
	s.logger.Info("Getting transaction by ID", map[string]interface{}{"transactionID": id})

	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get transaction", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return nil, err
	}

	return transaction, nil
}

// GetAllTransactions retrieves transactions with pagination and filtering
func (s *TransactionServiceImpl) GetAllTransactions(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error) {
	s.logger.Info("Getting transactions", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	transactions, total, err := s.transactionRepo.GetAll(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get transactions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetTransactionsByFundID retrieves transactions for a specific fund
func (s *TransactionServiceImpl) GetTransactionsByFundID(fundID string, page, limit int) ([]models.Transaction, int64, error) {
	s.logger.Info("Getting transactions by fund ID", map[string]interface{}{
		"fundID": fundID,
		"page":   page,
		"limit":  limit,
	})

	// Validate fund exists
	_, err := s.fundRepo.FindByID(fundID)
	if err != nil {
		s.logger.Error("Failed to find fund", map[string]interface{}{
			"fundID": fundID,
			"error":  err.Error(),
		})
		return nil, 0, errors.New("invalid fund ID")
	}

	transactions, total, err := s.transactionRepo.FindByFundID(fundID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get transactions by fund ID", map[string]interface{}{
			"fundID": fundID,
			"page":   page,
			"limit":  limit,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetTransactionsByEntityID retrieves transactions for a specific entity
func (s *TransactionServiceImpl) GetTransactionsByEntityID(entityID string, page, limit int) ([]models.Transaction, int64, error) {
	s.logger.Info("Getting transactions by entity ID", map[string]interface{}{
		"entityID": entityID,
		"page":     page,
		"limit":    limit,
	})

	// Validate entity exists
	_, err := s.entityRepo.FindByID(entityID)
	if err != nil {
		s.logger.Error("Failed to find entity", map[string]interface{}{
			"entityID": entityID,
			"error":    err.Error(),
		})
		return nil, 0, errors.New("invalid entity ID")
	}

	transactions, total, err := s.transactionRepo.FindByEntityID(entityID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get transactions by entity ID", map[string]interface{}{
			"entityID": entityID,
			"page":     page,
			"limit":    limit,
			"error":    err.Error(),
		})
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdateTransaction updates a transaction
func (s *TransactionServiceImpl) UpdateTransaction(transaction *models.Transaction) error {
	s.logger.Info("Updating transaction", map[string]interface{}{"transactionID": transaction.ID})

	// Get the current transaction to verify it exists
	existingTransaction, err := s.transactionRepo.FindByID(transaction.ID)
	if err != nil {
		s.logger.Error("Failed to find transaction for update", map[string]interface{}{
			"transactionID": transaction.ID,
			"error":         err.Error(),
		})
		return err
	}

	// Only allow updates to pending transactions
	if existingTransaction.Status != models.TransactionPending {
		s.logger.Warn("Cannot update non-pending transaction", map[string]interface{}{
			"transactionID": transaction.ID,
			"status":        existingTransaction.Status,
		})
		return errors.New("cannot update transaction that is not in pending status")
	}

	// Update the transaction
	transaction.UpdatedAt = time.Now()
	if err := s.transactionRepo.Update(transaction); err != nil {
		s.logger.Error("Failed to update transaction", map[string]interface{}{
			"transactionID": transaction.ID,
			"error":         err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     *transaction.CreatedByID,
		Action:     "UPDATE_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   transaction.ID,
		Timestamp:  time.Now(),
	})

	return nil
}

// ApproveTransaction approves a transaction
func (s *TransactionServiceImpl) ApproveTransaction(id string, approverID string) error {
	s.logger.Info("Approving transaction", map[string]interface{}{
		"transactionID": id,
		"approverID":    approverID,
	})

	// Get the transaction
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find transaction for approval", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Only allow approval of pending transactions
	if transaction.Status != models.TransactionPending {
		s.logger.Warn("Cannot approve non-pending transaction", map[string]interface{}{
			"transactionID": id,
			"status":        transaction.Status,
		})
		return errors.New("cannot approve transaction that is not in pending status")
	}

	// Record approval in blockchain
	if s.blockchain != nil {
		txHash, err := s.blockchain.ApproveTransaction(id, approverID)
		if err != nil {
			s.logger.Error("Failed to record approval on blockchain", map[string]interface{}{
				"error": err.Error(),
				"txID":  id,
			})
			// Continue with database update even if blockchain fails
		} else {
			transaction.BlockchainTxHash = txHash
			transaction.BlockchainStatus = "APPROVED"
		}
	}

	// Approve the transaction in database
	if err := s.transactionRepo.Approve(id, approverID); err != nil {
		s.logger.Error("Failed to approve transaction", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     approverID,
		Action:     "APPROVE_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   id,
		Timestamp:  time.Now(),
	})

	return nil
}

// RejectTransaction rejects a transaction
func (s *TransactionServiceImpl) RejectTransaction(id string, rejectedByID string, reason string) error {
	s.logger.Info("Rejecting transaction", map[string]interface{}{
		"transactionID": id,
		"rejectedByID":  rejectedByID,
		"reason":        reason,
	})

	// Get the transaction
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find transaction for rejection", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Only allow rejection of pending transactions
	if transaction.Status != models.TransactionPending {
		s.logger.Warn("Cannot reject non-pending transaction", map[string]interface{}{
			"transactionID": id,
			"status":        transaction.Status,
		})
		return errors.New("cannot reject transaction that is not in pending status")
	}

	// Record rejection in blockchain
	if s.blockchain != nil {
		txHash, err := s.blockchain.RejectTransaction(id, rejectedByID)
		if err != nil {
			s.logger.Error("Failed to record rejection on blockchain", map[string]interface{}{
				"error": err.Error(),
				"txID":  id,
			})
			// Continue with database update even if blockchain fails
		} else {
			transaction.BlockchainTxHash = txHash
			transaction.BlockchainStatus = "REJECTED"
		}
	}

	// Reject the transaction in database
	if err := s.transactionRepo.Reject(id, rejectedByID, reason); err != nil {
		s.logger.Error("Failed to reject transaction", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     rejectedByID,
		Action:     "REJECT_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   id,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	})

	return nil
}

// CompleteTransaction marks a transaction as completed
func (s *TransactionServiceImpl) CompleteTransaction(id string) error {
	s.logger.Info("Completing transaction", map[string]interface{}{"transactionID": id})

	// Get the transaction
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find transaction for completion", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Only allow completion of approved transactions
	if transaction.Status != models.TransactionApproved {
		s.logger.Warn("Cannot complete non-approved transaction", map[string]interface{}{
			"transactionID": id,
			"status":        transaction.Status,
		})
		return errors.New("cannot complete transaction that is not in approved status")
	}

	// Get the fund
	fund, err := s.fundRepo.FindByID(transaction.FundID)
	if err != nil {
		s.logger.Error("Failed to find fund for transaction completion", map[string]interface{}{
			"fundID": transaction.FundID,
			"error":  err.Error(),
		})
		return err
	}

	// Record completion in blockchain
	if s.blockchain != nil {
		txHash, err := s.blockchain.CompleteTransaction(id, *transaction.CreatedByID)
		if err != nil {
			s.logger.Error("Failed to record completion on blockchain", map[string]interface{}{
				"error": err.Error(),
				"txID":  id,
			})
			// Continue with database update even if blockchain fails
		} else {
			transaction.BlockchainTxHash = txHash
			transaction.BlockchainStatus = "COMPLETED"
		}
	}

	// Complete the transaction (this will update fund balances)
	if err := s.transactionRepo.CompleteTransaction(transaction, fund); err != nil {
		s.logger.Error("Failed to complete transaction", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		Action:     "COMPLETE_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   id,
		Timestamp:  time.Now(),
	})

	return nil
}

// FlagTransaction flags a transaction for review
func (s *TransactionServiceImpl) FlagTransaction(id string, reason string) error {
	s.logger.Info("Flagging transaction", map[string]interface{}{
		"transactionID": id,
		"reason":        reason,
	})

	// Get the transaction
	_, err := s.transactionRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find transaction for flagging", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Flag the transaction
	if err := s.transactionRepo.Flag(id, reason); err != nil {
		s.logger.Error("Failed to flag transaction", map[string]interface{}{
			"transactionID": id,
			"error":         err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		Action:     "FLAG_TRANSACTION",
		EntityType: "TRANSACTION",
		EntityID:   id,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	})

	return nil
}

// Helper function to validate transaction type
func isValidTransactionType(transactionType models.TransactionType) bool {
	validTypes := []models.TransactionType{
		models.TransactionAllocation,
		models.TransactionDisbursement,
		models.TransactionExpenditure,
		models.TransactionReturns,
	}

	for _, t := range validTypes {
		if t == transactionType {
			return true
		}
	}

	return false
}
