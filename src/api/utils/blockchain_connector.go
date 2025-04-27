package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

// EventType represents the type of transaction event
type EventType uint8

const (
	EventTypeCreated EventType = iota
	EventTypeApproved
	EventTypeRejected
	EventTypeCompleted
	EventTypeFlagged
)

// BlockchainService represents a service to interact with the blockchain
type BlockchainService struct {
	client             *ethclient.Client
	privateKey         *ecdsa.PrivateKey
	loggerAddress      common.Address
	fundManagerAddress common.Address
	loggerABI          abi.ABI
	fundManagerABI     abi.ABI
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService() (*BlockchainService, error) {
	config := GetConfig()

	// Get blockchain service URL from environment
	blockchainURL := config.BlockchainServiceURL
	if blockchainURL == "" {
		return nil, fmt.Errorf("BLOCKCHAIN_SERVICE_URL environment variable not set")
	}

	// Get contract addresses from config
	loggerAddressStr := config.TransactionLogger
	fundManagerAddressStr := config.FundManager
	if loggerAddressStr == "" || fundManagerAddressStr == "" {
		return nil, fmt.Errorf("contract addresses not set in environment variables")
	}

	loggerAddress := common.HexToAddress(loggerAddressStr)
	fundManagerAddress := common.HexToAddress(fundManagerAddressStr)

	// Get private key from environment
	privateKeyStr := config.BlockchainPrivateKey
	if privateKeyStr == "" {
		return nil, fmt.Errorf("BLOCKCHAIN_PRIVATE_KEY environment variable not set")
	}

	// Remove '0x' prefix if present
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = privateKeyStr[2:]
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Connect to blockchain
	client, err := ethclient.Dial(blockchainURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %v", err)
	}

	// Verify we're connected to the right network
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	expectedChainID := new(big.Int)
	expectedChainID.SetString(config.ChainID, 10)
	if chainID.Cmp(expectedChainID) != 0 {
		Logger.Warnf("Connected to chain ID %v, expected %v", chainID.String(), expectedChainID.String())
	}

	// Parse Logger ABI
	loggerABI, err := abi.JSON(strings.NewReader(TransactionLoggerABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse logger ABI: %v", err)
	}

	// Parse FundManager ABI
	fundManagerABI, err := abi.JSON(strings.NewReader(FundManagerABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse fund manager ABI: %v", err)
	}

	Logger.Infof("Connected to %s blockchain network (Chain ID: %s)", config.NetworkName, chainID.String())

	return &BlockchainService{
		client:             client,
		privateKey:         privateKey,
		loggerAddress:      loggerAddress,
		fundManagerAddress: fundManagerAddress,
		loggerABI:          loggerABI,
		fundManagerABI:     fundManagerABI,
	}, nil
}

// Constants for contract ABIs
const TransactionLoggerABI = `[
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "creatorHash", "type": "bytes32"},
			{"internalType": "bytes32", "name": "detailsHash", "type": "bytes32"}
		],
		"name": "recordTransactionCreation",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "approverHash", "type": "bytes32"}
		],
		"name": "recordApproval",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "rejectorHash", "type": "bytes32"}
		],
		"name": "recordRejection",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "completerHash", "type": "bytes32"}
		],
		"name": "recordCompletion",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "string", "name": "reason", "type": "string"}
		],
		"name": "recordFlagging",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const FundManagerABI = `[
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "uint256", "name": "amount", "type": "uint256"},
			{"internalType": "string", "name": "currency", "type": "string"},
			{"internalType": "uint8", "name": "txType", "type": "uint8"},
			{"internalType": "string", "name": "description", "type": "string"},
			{"internalType": "bytes32", "name": "sourceId", "type": "bytes32"},
			{"internalType": "bytes32", "name": "destinationId", "type": "bytes32"},
			{"internalType": "bytes32", "name": "fundId", "type": "bytes32"},
			{"internalType": "bytes32", "name": "budgetLineItemId", "type": "bytes32"},
			{"internalType": "string", "name": "documentRef", "type": "string"},
			{"internalType": "bytes32", "name": "createdById", "type": "bytes32"}
		],
		"name": "createTransaction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "approvedById", "type": "bytes32"}
		],
		"name": "approveTransaction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "rejectedById", "type": "bytes32"}
		],
		"name": "rejectTransaction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "bytes32", "name": "completedById", "type": "bytes32"}
		],
		"name": "completeTransaction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "bytes32", "name": "id", "type": "bytes32"},
			{"internalType": "string", "name": "reason", "type": "string"}
		],
		"name": "flagTransaction",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// Close closes the blockchain connection
func (bs *BlockchainService) Close() {
	bs.client.Close()
}

// TransactionData represents transaction data to be recorded on the blockchain
type TransactionData struct {
	ID               string
	Amount           float64
	Currency         string
	Type             string
	Description      string
	SourceID         string
	DestinationID    string
	FundID           string
	BudgetLineItemID string
	DocumentRef      string
	CreatedByID      string
}

// hashUserID creates a bytes32 hash of a user ID for blockchain storage
func hashUserID(userID string) ([32]byte, error) {
	if userID == "" {
		return [32]byte{}, nil
	}

	// Remove '0x' prefix if present
	if strings.HasPrefix(userID, "0x") {
		userID = userID[2:]
	}

	// Hash the user ID using Keccak256
	hash := crypto.Keccak256([]byte(userID))
	var bytes32 [32]byte
	copy(bytes32[:], hash)
	return bytes32, nil
}

// hashTransactionDetails creates a bytes32 hash of critical transaction details
func hashTransactionDetails(tx *TransactionData) ([32]byte, error) {
	// Create a string combining critical fields
	details := fmt.Sprintf("%f:%s:%s:%s:%s:%s:%s",
		tx.Amount,
		tx.Currency,
		tx.Type,
		tx.SourceID,
		tx.DestinationID,
		tx.FundID,
		tx.BudgetLineItemID,
	)

	// Hash the details
	hash := crypto.Keccak256([]byte(details))
	var bytes32 [32]byte
	copy(bytes32[:], hash)
	return bytes32, nil
}

// RecordTransaction records a transaction creation event on the blockchain
func (bs *BlockchainService) RecordTransaction(transaction *TransactionData) (string, error) {
	// Get the auth for transaction
	auth, err := bs.getTransactionAuth()
	if err != nil {
		return "", err
	}

	// Convert transaction ID to bytes32
	id := stringToBytes32(transaction.ID)

	// Hash the creator ID
	creatorHash, err := hashUserID(transaction.CreatedByID)
	if err != nil {
		return "", fmt.Errorf("failed to hash creator ID: %v", err)
	}

	// Hash transaction details
	detailsHash, err := hashTransactionDetails(transaction)
	if err != nil {
		return "", fmt.Errorf("failed to hash transaction details: %v", err)
	}

	// Pack the method call for recordTransactionCreation
	input, err := bs.loggerABI.Pack("recordTransactionCreation",
		id, creatorHash, detailsHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create and send the transaction
	tx, err := bs.sendTransaction(auth, input, bs.loggerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to record transaction on blockchain: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// ApproveTransaction records a transaction approval event on the blockchain
func (bs *BlockchainService) ApproveTransaction(transactionID, approverID string) (string, error) {
	auth, err := bs.getTransactionAuth()
	if err != nil {
		return "", err
	}

	id := stringToBytes32(transactionID)
	approverHash, err := hashUserID(approverID)
	if err != nil {
		return "", fmt.Errorf("failed to hash approver ID: %v", err)
	}

	input, err := bs.loggerABI.Pack("recordApproval", id, approverHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack approval data: %v", err)
	}

	tx, err := bs.sendTransaction(auth, input, bs.loggerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to record approval on blockchain: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// RejectTransaction records a transaction rejection event on the blockchain
func (bs *BlockchainService) RejectTransaction(transactionID, rejectorID string) (string, error) {
	auth, err := bs.getTransactionAuth()
	if err != nil {
		return "", err
	}

	id := stringToBytes32(transactionID)
	rejectorHash, err := hashUserID(rejectorID)
	if err != nil {
		return "", fmt.Errorf("failed to hash rejector ID: %v", err)
	}

	input, err := bs.loggerABI.Pack("recordRejection", id, rejectorHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack rejection data: %v", err)
	}

	tx, err := bs.sendTransaction(auth, input, bs.loggerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to record rejection on blockchain: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// CompleteTransaction records a transaction completion event on the blockchain
func (bs *BlockchainService) CompleteTransaction(transactionID, completerID string) (string, error) {
	auth, err := bs.getTransactionAuth()
	if err != nil {
		return "", err
	}

	id := stringToBytes32(transactionID)
	completerHash, err := hashUserID(completerID)
	if err != nil {
		return "", fmt.Errorf("failed to hash completer ID: %v", err)
	}

	input, err := bs.loggerABI.Pack("recordCompletion", id, completerHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack completion data: %v", err)
	}

	tx, err := bs.sendTransaction(auth, input, bs.loggerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to record completion on blockchain: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// FlagTransaction records an AI flagging event on the blockchain
func (bs *BlockchainService) FlagTransaction(transactionID, reason string) (string, error) {
	auth, err := bs.getTransactionAuth()
	if err != nil {
		return "", err
	}

	id := stringToBytes32(transactionID)

	input, err := bs.loggerABI.Pack("recordFlagging", id, reason)
	if err != nil {
		return "", fmt.Errorf("failed to pack flagging data: %v", err)
	}

	tx, err := bs.sendTransaction(auth, input, bs.loggerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to record flagging on blockchain: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// GetTransactionEvents retrieves all events for a transaction from the blockchain
func (bs *BlockchainService) GetTransactionEvents(transactionID string) ([]EventRecord, error) {
	id := stringToBytes32(transactionID)

	result, err := bs.loggerABI.Pack("getTransactionEvents", id)
	if err != nil {
		return nil, fmt.Errorf("failed to pack get events data: %v", err)
	}

	msg := ethereum.CallMsg{
		To:   &bs.loggerAddress,
		Data: result,
	}

	output, err := bs.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	// Unpack the results
	results, err := bs.loggerABI.Unpack("getTransactionEvents", output)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack events: %v", err)
	}

	// Convert the results to EventRecord structs
	events := make([]EventRecord, 0)
	if len(results) >= 5 {
		eventTypes, _ := results[0].([]uint8)
		timestamps, _ := results[1].([]*big.Int)
		actorHashes, _ := results[2].([][32]byte)
		detailsHashes, _ := results[3].([][32]byte)
		metadatas, _ := results[4].([]string)

		for i := 0; i < len(eventTypes); i++ {
			events = append(events, EventRecord{
				EventType:   EventType(eventTypes[i]),
				Timestamp:   timestamps[i].Uint64(),
				ActorHash:   actorHashes[i],
				DetailsHash: detailsHashes[i],
				Metadata:    metadatas[i],
			})
		}
	}

	return events, nil
}

// EventRecord represents a blockchain event record
type EventRecord struct {
	EventType   EventType
	Timestamp   uint64
	ActorHash   [32]byte
	DetailsHash [32]byte
	Metadata    string
}

// getTransactionAuth creates and returns a transaction auth object
func (bs *BlockchainService) getTransactionAuth() (*bind.TransactOpts, error) {
	// Get chain ID
	chainID, err := bs.client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Get nonce for address
	publicKey := bs.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := bs.client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Get gas price
	gasPrice, err := bs.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	// Create auth
	auth, err := bind.NewKeyedTransactorWithChainID(bs.privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // No ether transfer
	auth.GasLimit = uint64(3000000) // Gas limit
	auth.GasPrice = gasPrice

	return auth, nil
}

// createTransaction calls the createTransaction method on the smart contract
func (bs *BlockchainService) createTransaction(
	auth *bind.TransactOpts,
	id [32]byte,
	amount *big.Int,
	currency string,
	txType uint8,
	description string,
	sourceId [32]byte,
	destinationId [32]byte,
	fundId [32]byte,
	budgetLineItemId [32]byte,
	documentRef string,
	createdById [32]byte,
) (*types.Transaction, error) {
	// Pack the method call
	input, err := bs.fundManagerABI.Pack("createTransaction",
		id, amount, currency, txType, description, sourceId,
		destinationId, fundId, budgetLineItemId, documentRef, createdById)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create transaction data
	txData := &types.LegacyTx{
		To:       &bs.fundManagerAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// approveTransactionOnChain calls the approveTransaction method on the smart contract
func (bs *BlockchainService) approveTransactionOnChain(
	auth *bind.TransactOpts,
	id [32]byte,
	approvedById [32]byte,
) (*types.Transaction, error) {
	// Pack the method call
	input, err := bs.fundManagerABI.Pack("approveTransaction", id, approvedById)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create transaction data
	txData := &types.LegacyTx{
		To:       &bs.fundManagerAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// rejectTransactionOnChain calls the rejectTransaction method on the smart contract
func (bs *BlockchainService) rejectTransactionOnChain(
	auth *bind.TransactOpts,
	id [32]byte,
	rejectedById [32]byte,
) (*types.Transaction, error) {
	// Pack the method call
	input, err := bs.fundManagerABI.Pack("rejectTransaction", id, rejectedById)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create transaction data
	txData := &types.LegacyTx{
		To:       &bs.fundManagerAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// completeTransactionOnChain calls the completeTransaction method on the smart contract
func (bs *BlockchainService) completeTransactionOnChain(
	auth *bind.TransactOpts,
	id [32]byte,
	completedById [32]byte,
) (*types.Transaction, error) {
	// Pack the method call
	input, err := bs.fundManagerABI.Pack("completeTransaction", id, completedById)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create transaction data
	txData := &types.LegacyTx{
		To:       &bs.fundManagerAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// flagTransactionOnChain calls the flagTransaction method on the smart contract
func (bs *BlockchainService) flagTransactionOnChain(
	auth *bind.TransactOpts,
	id [32]byte,
	reason string,
) (*types.Transaction, error) {
	// Pack the method call
	input, err := bs.fundManagerABI.Pack("flagTransaction", id, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %v", err)
	}

	// Create transaction data
	txData := &types.LegacyTx{
		To:       &bs.fundManagerAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// stringToBytes32 converts a string to a [32]byte array
// If the string is a UUID, it converts it to a bytes32 directly
// Otherwise, it pads the string with zeros
func stringToBytes32(s string) [32]byte {
	var bytes32 [32]byte

	// Check if the string is a UUID
	if u, err := uuid.Parse(s); err == nil {
		// Convert UUID to bytes and copy to bytes32
		b, _ := u.MarshalBinary()
		copy(bytes32[:], b)
	} else {
		// Pad the string with zeros
		copy(bytes32[:], s)
	}

	return bytes32
}

// sendTransaction sends a transaction to the blockchain
func (bs *BlockchainService) sendTransaction(auth *bind.TransactOpts, input []byte, contractAddress common.Address) (*types.Transaction, error) {
	// Create transaction data
	txData := &types.LegacyTx{
		To:       &contractAddress,
		Gas:      auth.GasLimit,
		GasPrice: auth.GasPrice,
		Value:    auth.Value,
		Nonce:    uint64(auth.Nonce.Int64()),
		Data:     input,
	}

	// Send the transaction
	tx := types.NewTx(txData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	err = bs.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
