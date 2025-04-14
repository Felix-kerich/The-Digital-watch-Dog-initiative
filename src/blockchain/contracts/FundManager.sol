// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title FundManager
 * @dev Smart contract for transparent public fund management
 */
contract FundManager {
    /// Transaction type enum
    enum TransactionType { ALLOCATION, DISBURSEMENT, EXPENDITURE, RETURNS }
    
    /// Transaction status enum
    enum TransactionStatus { PENDING, APPROVED, REJECTED, COMPLETED, FLAGGED }

    /// Represents a fund transaction on the blockchain
    struct Transaction {
        bytes32 id;                   // Unique transaction ID (UUID from backend)
        uint256 amount;               // Amount in smallest unit (e.g., cents)
        string currency;              // Currency code (e.g., "KES")
        TransactionType txType;       // Type of transaction
        TransactionStatus status;     // Current status
        string description;           // Brief description
        bytes32 sourceId;             // Source entity ID
        bytes32 destinationId;        // Destination entity ID
        bytes32 fundId;               // Fund ID
        bytes32 budgetLineItemId;     // Budget line item ID (optional)
        string documentRef;           // Reference to supporting document hash/IPFS
        bool aiFlagged;               // Whether the AI flagged this transaction
        string aiFlagReason;          // Reason for AI flag
        bytes32 approvedById;         // ID of user who approved
        bytes32 createdById;          // ID of user who created
        uint256 createdAt;            // Creation timestamp
        uint256 updatedAt;            // Last updated timestamp
        uint256 completedAt;          // Completion timestamp (0 if not completed)
    }

    /// Represents a fund on the blockchain
    struct Fund {
        bytes32 id;                   // Unique fund ID (UUID from backend)
        string name;                  // Fund name
        string code;                  // Fund code
        string category;              // Fund category
        string fiscalYear;            // Fiscal year
        uint256 totalAmount;          // Total amount allocated to fund
        uint256 allocated;            // Amount already allocated
        uint256 disbursed;            // Amount already disbursed
        uint256 utilized;             // Amount already utilized/spent
        uint256 createdAt;            // Creation timestamp
    }

    /// Fund creation event
    event FundCreated(
        bytes32 indexed id,
        string name,
        string code,
        uint256 totalAmount,
        uint256 timestamp
    );

    /// Transaction created event
    event TransactionCreated(
        bytes32 indexed id,
        TransactionType txType,
        uint256 amount,
        TransactionStatus status,
        bytes32 indexed fundId,
        uint256 timestamp
    );

    /// Transaction status change event
    event TransactionStatusChanged(
        bytes32 indexed id,
        TransactionStatus oldStatus,
        TransactionStatus newStatus,
        bytes32 indexed changedById,
        uint256 timestamp
    );

    // State variables
    address public owner;
    address[] public authorizedUpdaters;
    mapping(bytes32 => Transaction) public transactions;
    mapping(bytes32 => Fund) public funds;
    bytes32[] public transactionIds;
    bytes32[] public fundIds;
    
    /// Restricts function to contract owner
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can perform this action");
        _;
    }
    
    /// Restricts function to authorized updaters
    modifier onlyAuthorized() {
        bool isAuthorized = false;
        if (msg.sender == owner) {
            isAuthorized = true;
        } else {
            for (uint i = 0; i < authorizedUpdaters.length; i++) {
                if (msg.sender == authorizedUpdaters[i]) {
                    isAuthorized = true;
                    break;
                }
            }
        }
        require(isAuthorized, "Not authorized to perform this action");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev Adds an authorized updater address
     * @param updater Address of the updater to add
     */
    function addAuthorizedUpdater(address updater) external onlyOwner {
        // Check if already authorized
        for (uint i = 0; i < authorizedUpdaters.length; i++) {
            if (authorizedUpdaters[i] == updater) {
                return; // Already authorized
            }
        }
        authorizedUpdaters.push(updater);
    }

    /**
     * @dev Removes an authorized updater address
     * @param updater Address of the updater to remove
     */
    function removeAuthorizedUpdater(address updater) external onlyOwner {
        for (uint i = 0; i < authorizedUpdaters.length; i++) {
            if (authorizedUpdaters[i] == updater) {
                // Move last element to the place of the removed element
                authorizedUpdaters[i] = authorizedUpdaters[authorizedUpdaters.length - 1];
                authorizedUpdaters.pop();
                return;
            }
        }
    }

    /**
     * @dev Creates a new fund
     * @param id Unique ID for the fund
     * @param name Name of the fund
     * @param code Fund code
     * @param category Fund category
     * @param fiscalYear Fiscal year of the fund
     * @param totalAmount Total amount allocated to the fund
     */
    function createFund(
        bytes32 id,
        string calldata name,
        string calldata code,
        string calldata category,
        string calldata fiscalYear,
        uint256 totalAmount
    ) external onlyAuthorized {
        // Ensure fund doesn't already exist
        require(funds[id].id == bytes32(0), "Fund already exists");
        
        // Create the fund
        funds[id] = Fund({
            id: id,
            name: name,
            code: code,
            category: category,
            fiscalYear: fiscalYear,
            totalAmount: totalAmount,
            allocated: 0,
            disbursed: 0,
            utilized: 0,
            createdAt: block.timestamp
        });
        
        // Add to fund IDs array
        fundIds.push(id);
        
        // Emit event
        emit FundCreated(id, name, code, totalAmount, block.timestamp);
    }

    /**
     * @dev Creates a new transaction
     * @param id Unique ID for the transaction
     * @param amount Transaction amount
     * @param currency Currency code
     * @param txType Transaction type
     * @param description Description of the transaction
     * @param sourceId Source entity ID
     * @param destinationId Destination entity ID
     * @param fundId Fund ID
     * @param budgetLineItemId Budget line item ID (optional)
     * @param documentRef Reference to supporting document
     * @param createdById ID of the user who created the transaction
     */
    function createTransaction(
        bytes32 id,
        uint256 amount,
        string calldata currency,
        uint8 txType,
        string calldata description,
        bytes32 sourceId,
        bytes32 destinationId,
        bytes32 fundId,
        bytes32 budgetLineItemId,
        string calldata documentRef,
        bytes32 createdById
    ) external onlyAuthorized {
        // Ensure transaction doesn't already exist
        require(transactions[id].id == bytes32(0), "Transaction already exists");
        
        // Ensure fund exists
        require(funds[fundId].id != bytes32(0), "Fund does not exist");
        
        // Create the transaction
        transactions[id] = Transaction({
            id: id,
            amount: amount,
            currency: currency,
            txType: TransactionType(txType),
            status: TransactionStatus.PENDING,
            description: description,
            sourceId: sourceId,
            destinationId: destinationId,
            fundId: fundId,
            budgetLineItemId: budgetLineItemId,
            documentRef: documentRef,
            aiFlagged: false,
            aiFlagReason: "",
            approvedById: bytes32(0),
            createdById: createdById,
            createdAt: block.timestamp,
            updatedAt: block.timestamp,
            completedAt: 0
        });
        
        // Add to transaction IDs array
        transactionIds.push(id);
        
        // Emit event
        emit TransactionCreated(
            id,
            TransactionType(txType),
            amount,
            TransactionStatus.PENDING,
            fundId,
            block.timestamp
        );
    }

    /**
     * @dev Approves a transaction
     * @param id ID of the transaction to approve
     * @param approvedById ID of the user approving the transaction
     */
    function approveTransaction(bytes32 id, bytes32 approvedById) external onlyAuthorized {
        // Ensure transaction exists
        require(transactions[id].id != bytes32(0), "Transaction does not exist");
        
        // Ensure transaction is in a state that can be approved
        require(
            transactions[id].status == TransactionStatus.PENDING ||
            transactions[id].status == TransactionStatus.FLAGGED,
            "Transaction cannot be approved in its current state"
        );
        
        // Update status
        TransactionStatus oldStatus = transactions[id].status;
        transactions[id].status = TransactionStatus.APPROVED;
        transactions[id].approvedById = approvedById;
        transactions[id].updatedAt = block.timestamp;
        
        // Emit event
        emit TransactionStatusChanged(
            id,
            oldStatus,
            TransactionStatus.APPROVED,
            approvedById,
            block.timestamp
        );
    }

    /**
     * @dev Rejects a transaction
     * @param id ID of the transaction to reject
     * @param rejectedById ID of the user rejecting the transaction
     */
    function rejectTransaction(bytes32 id, bytes32 rejectedById) external onlyAuthorized {
        // Ensure transaction exists
        require(transactions[id].id != bytes32(0), "Transaction does not exist");
        
        // Ensure transaction is in a state that can be rejected
        require(
            transactions[id].status == TransactionStatus.PENDING ||
            transactions[id].status == TransactionStatus.FLAGGED,
            "Transaction cannot be rejected in its current state"
        );
        
        // Update status
        TransactionStatus oldStatus = transactions[id].status;
        transactions[id].status = TransactionStatus.REJECTED;
        transactions[id].updatedAt = block.timestamp;
        
        // Emit event
        emit TransactionStatusChanged(
            id,
            oldStatus,
            TransactionStatus.REJECTED,
            rejectedById,
            block.timestamp
        );
    }

    /**
     * @dev Completes a transaction and updates fund balances
     * @param id ID of the transaction to complete
     * @param completedById ID of the user completing the transaction
     */
    function completeTransaction(bytes32 id, bytes32 completedById) external onlyAuthorized {
        // Ensure transaction exists
        require(transactions[id].id != bytes32(0), "Transaction does not exist");
        
        // Ensure transaction is approved
        require(
            transactions[id].status == TransactionStatus.APPROVED,
            "Transaction must be approved before it can be completed"
        );
        
        // Get the fund
        bytes32 fundId = transactions[id].fundId;
        require(funds[fundId].id != bytes32(0), "Fund does not exist");
        
        // Update fund amounts based on transaction type
        if (transactions[id].txType == TransactionType.ALLOCATION) {
            funds[fundId].allocated += transactions[id].amount;
        } else if (transactions[id].txType == TransactionType.DISBURSEMENT) {
            funds[fundId].disbursed += transactions[id].amount;
        } else if (transactions[id].txType == TransactionType.EXPENDITURE) {
            funds[fundId].utilized += transactions[id].amount;
        } else if (transactions[id].txType == TransactionType.RETURNS) {
            // Return funds to the fund
            if (transactions[id].amount <= funds[fundId].allocated) {
                funds[fundId].allocated -= transactions[id].amount;
            }
        }
        
        // Update transaction status
        TransactionStatus oldStatus = transactions[id].status;
        transactions[id].status = TransactionStatus.COMPLETED;
        transactions[id].updatedAt = block.timestamp;
        transactions[id].completedAt = block.timestamp;
        
        // Emit event
        emit TransactionStatusChanged(
            id,
            oldStatus,
            TransactionStatus.COMPLETED,
            completedById,
            block.timestamp
        );
    }

    /**
     * @dev Flags a transaction based on AI anomaly detection
     * @param id ID of the transaction to flag
     * @param reason Reason for flagging
     */
    function flagTransaction(bytes32 id, string calldata reason) external onlyAuthorized {
        // Ensure transaction exists
        require(transactions[id].id != bytes32(0), "Transaction does not exist");
        
        // Ensure transaction is in a state that can be flagged
        require(
            transactions[id].status == TransactionStatus.PENDING,
            "Only pending transactions can be flagged"
        );
        
        // Update status
        TransactionStatus oldStatus = transactions[id].status;
        transactions[id].status = TransactionStatus.FLAGGED;
        transactions[id].aiFlagged = true;
        transactions[id].aiFlagReason = reason;
        transactions[id].updatedAt = block.timestamp;
        
        // Emit event
        emit TransactionStatusChanged(
            id,
            oldStatus,
            TransactionStatus.FLAGGED,
            bytes32(0), // System action
            block.timestamp
        );
    }

    /**
     * @dev Returns the total number of transactions
     * @return Number of transactions
     */
    function getTransactionCount() external view returns (uint256) {
        return transactionIds.length;
    }

    /**
     * @dev Returns the total number of funds
     * @return Number of funds
     */
    function getFundCount() external view returns (uint256) {
        return fundIds.length;
    }

    /**
     * @dev Returns a batch of transaction IDs
     * @param start Starting index
     * @param limit Maximum number of IDs to return
     * @return Array of transaction IDs
     */
    function getTransactionIds(uint256 start, uint256 limit) external view returns (bytes32[] memory) {
        // Calculate how many IDs to return
        uint256 end = start + limit;
        if (end > transactionIds.length) {
            end = transactionIds.length;
        }
        
        uint256 count = end - start;
        bytes32[] memory result = new bytes32[](count);
        
        for (uint256 i = 0; i < count; i++) {
            result[i] = transactionIds[start + i];
        }
        
        return result;
    }

    /**
     * @dev Returns a batch of fund IDs
     * @param start Starting index
     * @param limit Maximum number of IDs to return
     * @return Array of fund IDs
     */
    function getFundIds(uint256 start, uint256 limit) external view returns (bytes32[] memory) {
        // Calculate how many IDs to return
        uint256 end = start + limit;
        if (end > fundIds.length) {
            end = fundIds.length;
        }
        
        uint256 count = end - start;
        bytes32[] memory result = new bytes32[](count);
        
        for (uint256 i = 0; i < count; i++) {
            result[i] = fundIds[start + i];
        }
        
        return result;
    }
} 