// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

/**
 * @title TransactionEventLogger
 * @dev Records and verifies transaction lifecycle events with cryptographic proofs
 */
contract TransactionEventLogger is AccessControl, Pausable {
    bytes32 public constant API_ROLE = keccak256("API_ROLE");
    
    enum EventType { CREATED, APPROVED, REJECTED, COMPLETED, FLAGGED }
    
    struct EventRecord {
        EventType eventType;
        uint256 timestamp;
        bytes32 actorHash;
        bytes32 detailsHash;
        string metadata;  // For storing short, non-sensitive data like AI flag reasons
    }
    
    // Primary storage: dbTransactionId -> array of events
    mapping(bytes32 => EventRecord[]) private transactionEvents;
    
    // Events
    event EventLogged(
        bytes32 indexed dbTransactionId,
        EventType indexed eventType,
        bytes32 actorHash,
        bytes32 detailsHash,
        uint256 timestamp
    );
    
    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(API_ROLE, msg.sender);
    }
    
    /**
     * @dev Records a new transaction creation event
     * @param _dbTransactionId The ID from the application database
     * @param _actorHash Hash of the actor's ID (creator)
     * @param _detailsHash Hash of critical transaction details
     */
    function recordTransactionCreation(
        bytes32 _dbTransactionId,
        bytes32 _actorHash,
        bytes32 _detailsHash
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(_dbTransactionId != bytes32(0), "Invalid transaction ID");
        
        _logEvent(
            _dbTransactionId,
            EventType.CREATED,
            _actorHash,
            _detailsHash,
            ""
        );
    }
    
    /**
     * @dev Records a transaction approval event
     */
    function recordApproval(
        bytes32 _dbTransactionId,
        bytes32 _approverHash
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(_dbTransactionId != bytes32(0), "Invalid transaction ID");
        require(_hasEvent(_dbTransactionId, EventType.CREATED), "Transaction not created");
        
        _logEvent(
            _dbTransactionId,
            EventType.APPROVED,
            _approverHash,
            bytes32(0),
            ""
        );
    }
    
    /**
     * @dev Records a transaction rejection event
     */
    function recordRejection(
        bytes32 _dbTransactionId,
        bytes32 _rejecterHash
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(_dbTransactionId != bytes32(0), "Invalid transaction ID");
        require(_hasEvent(_dbTransactionId, EventType.CREATED), "Transaction not created");
        
        _logEvent(
            _dbTransactionId,
            EventType.REJECTED,
            _rejecterHash,
            bytes32(0),
            ""
        );
    }
    
    /**
     * @dev Records a transaction completion event
     */
    function recordCompletion(
        bytes32 _dbTransactionId,
        bytes32 _completerHash
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(_dbTransactionId != bytes32(0), "Invalid transaction ID");
        require(_hasEvent(_dbTransactionId, EventType.APPROVED), "Transaction not approved");
        
        _logEvent(
            _dbTransactionId,
            EventType.COMPLETED,
            _completerHash,
            bytes32(0),
            ""
        );
    }
    
    /**
     * @dev Records an AI flagging event
     */
    function recordFlagging(
        bytes32 _dbTransactionId,
        string calldata _reason
    ) external whenNotPaused onlyRole(API_ROLE) {
        require(_dbTransactionId != bytes32(0), "Invalid transaction ID");
        require(bytes(_reason).length <= 256, "Reason too long");
        
        _logEvent(
            _dbTransactionId,
            EventType.FLAGGED,
            bytes32(0),
            bytes32(0),
            _reason
        );
    }
    
    /**
     * @dev Internal function to log an event
     */
    function _logEvent(
        bytes32 _dbTransactionId,
        EventType _eventType,
        bytes32 _actorHash,
        bytes32 _detailsHash,
        string memory _metadata
    ) private {
        EventRecord memory newEvent = EventRecord({
            eventType: _eventType,
            timestamp: block.timestamp,
            actorHash: _actorHash,
            detailsHash: _detailsHash,
            metadata: _metadata
        });
        
        transactionEvents[_dbTransactionId].push(newEvent);
        
        emit EventLogged(
            _dbTransactionId,
            _eventType,
            _actorHash,
            _detailsHash,
            block.timestamp
        );
    }
    
    /**
     * @dev Checks if a transaction has a specific event type
     */
    function _hasEvent(bytes32 _dbTransactionId, EventType _eventType) private view returns (bool) {
        EventRecord[] storage events = transactionEvents[_dbTransactionId];
        for (uint i = 0; i < events.length; i++) {
            if (events[i].eventType == _eventType) {
                return true;
            }
        }
        return false;
    }
    
    /**
     * @dev Gets all events for a transaction
     */
    function getTransactionEvents(bytes32 _dbTransactionId)
        external
        view
        returns (
            EventType[] memory eventTypes,
            uint256[] memory timestamps,
            bytes32[] memory actorHashes,
            bytes32[] memory detailsHashes,
            string[] memory metadatas
        )
    {
        EventRecord[] storage events = transactionEvents[_dbTransactionId];
        uint256 length = events.length;
        
        eventTypes = new EventType[](length);
        timestamps = new uint256[](length);
        actorHashes = new bytes32[](length);
        detailsHashes = new bytes32[](length);
        metadatas = new string[](length);
        
        for (uint i = 0; i < length; i++) {
            eventTypes[i] = events[i].eventType;
            timestamps[i] = events[i].timestamp;
            actorHashes[i] = events[i].actorHash;
            detailsHashes[i] = events[i].detailsHash;
            metadatas[i] = events[i].metadata;
        }
        
        return (eventTypes, timestamps, actorHashes, detailsHashes, metadatas);
    }
    
    /**
     * @dev Pauses the contract
     */
    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }
    
    /**
     * @dev Unpauses the contract
     */
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }
} 