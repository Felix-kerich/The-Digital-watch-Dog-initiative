// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title FundTransactionLedger
 * @dev Smart contract for recording and managing public fund transactions
 */
contract FundTransactionLedger is AccessControl, Pausable {
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant API_ROLE = keccak256("API_ROLE");

    // Transaction statuses
    enum TransactionStatus { PENDING, APPROVED, REJECTED, COMPLETED, FLAGGED }

    // Transaction types
    enum TransactionType { ALLOCATION, DISBURSEMENT, EXPENDITURE, RETURNS }

    // Structure to store transaction details
    struct TransactionRecord {
        bytes32 id;                  // UUID from the application database
        uint256 amount;             // Transaction amount
        string currency;            // Currency code (e.g., "KES")
        TransactionType txType;     // Type of transaction
        string description;         // Transaction description
        bytes32 sourceId;           // Source entity ID
        bytes32 destinationId;      // Destination entity ID
        bytes32 fundId;            // Associated fund ID
        bytes32 budgetLineItemId;  // Associated budget line item ID
        string documentRef;        // Reference to supporting documents
        bytes32 createdById;       // ID of user who created the transaction
        TransactionStatus status;   // Current status
        uint256 timestamp;         // Timestamp of creation
        bytes32 approvedById;      // ID of user who approved (if applicable)
        bytes32 rejectedById;      // ID of user who rejected (if applicable)
        string rejectionReason;    // Reason for rejection (if applicable)
        bool aiFlagged;           // Whether transaction was flagged by AI
        string aiReasonDetails;    // AI flagging reason details
    }

    // Mapping to store transactions
    mapping(bytes32 => TransactionRecord) public transactions;
    
    // Events
    event TransactionCreated(bytes32 indexed id, TransactionType txType, uint256 amount);
    event TransactionApproved(bytes32 indexed id, bytes32 approvedById);
    event TransactionRejected(bytes32 indexed id, bytes32 rejectedById, string reason);
    event TransactionCompleted(bytes32 indexed id);
    event TransactionFlagged(bytes32 indexed id, string reason);

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(ADMIN_ROLE, msg.sender);
    }

    /**
     * @dev Creates a new transaction record
     */
    function createTransaction(
        bytes32 id,
        uint256 amount,
        string memory currency,
        uint8 txType,
        string memory description,
        bytes32 sourceId,
        bytes32 destinationId,
        bytes32 fundId,
        bytes32 budgetLineItemId,
        string memory documentRef,
        bytes32 createdById
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(transactions[id].id == 0, "Transaction already exists");
        require(amount > 0, "Amount must be greater than 0");

        transactions[id] = TransactionRecord({
            id: id,
            amount: amount,
            currency: currency,
            txType: TransactionType(txType),
            description: description,
            sourceId: sourceId,
            destinationId: destinationId,
            fundId: fundId,
            budgetLineItemId: budgetLineItemId,
            documentRef: documentRef,
            createdById: createdById,
            status: TransactionStatus.PENDING,
            timestamp: block.timestamp,
            approvedById: 0,
            rejectedById: 0,
            rejectionReason: "",
            aiFlagged: false,
            aiReasonDetails: ""
        });

        emit TransactionCreated(id, TransactionType(txType), amount);
    }

    /**
     * @dev Approves a transaction
     */
    function approveTransaction(bytes32 id, bytes32 approvedById) 
        external 
        whenNotPaused 
        onlyRole(API_ROLE) 
    {
        require(transactions[id].id != 0, "Transaction does not exist");
        require(transactions[id].status == TransactionStatus.PENDING, "Transaction not in pending state");

        transactions[id].status = TransactionStatus.APPROVED;
        transactions[id].approvedById = approvedById;

        emit TransactionApproved(id, approvedById);
    }

    /**
     * @dev Rejects a transaction
     */
    function rejectTransaction(bytes32 id, bytes32 rejectedById, string memory reason) 
        external 
        whenNotPaused 
        onlyRole(API_ROLE) 
    {
        require(transactions[id].id != 0, "Transaction does not exist");
        require(transactions[id].status == TransactionStatus.PENDING, "Transaction not in pending state");

        transactions[id].status = TransactionStatus.REJECTED;
        transactions[id].rejectedById = rejectedById;
        transactions[id].rejectionReason = reason;

        emit TransactionRejected(id, rejectedById, reason);
    }

    /**
     * @dev Marks a transaction as completed
     */
    function completeTransaction(bytes32 id) 
        external 
        whenNotPaused 
        onlyRole(API_ROLE) 
    {
        require(transactions[id].id != 0, "Transaction does not exist");
        require(transactions[id].status == TransactionStatus.APPROVED, "Transaction not approved");

        transactions[id].status = TransactionStatus.COMPLETED;

        emit TransactionCompleted(id);
    }

    /**
     * @dev Flags a transaction based on AI detection
     */
    function flagTransaction(bytes32 id, string memory reason) 
        external 
        whenNotPaused 
        onlyRole(API_ROLE) 
    {
        require(transactions[id].id != 0, "Transaction does not exist");
        
        transactions[id].status = TransactionStatus.FLAGGED;
        transactions[id].aiFlagged = true;
        transactions[id].aiReasonDetails = reason;

        emit TransactionFlagged(id, reason);
    }

    /**
     * @dev Gets a transaction by ID
     */
    function getTransaction(bytes32 id) 
        external 
        view 
        returns (
            uint256 amount,
            string memory currency,
            TransactionType txType,
            TransactionStatus status,
            uint256 timestamp,
            bool aiFlagged
        ) 
    {
        require(transactions[id].id != 0, "Transaction does not exist");
        
        TransactionRecord storage txn = transactions[id];
        return (
            txn.amount,
            txn.currency,
            txn.txType,
            txn.status,
            txn.timestamp,
            txn.aiFlagged
        );
    }

    /**
     * @dev Pauses the contract
     */
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpauses the contract
     */
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
} 