package utils

import (
	"os"
)

// Config represents application configuration
type Config struct {
	AppEnv               string
	LogLevel             string
	BlockchainServiceURL string
	BlockchainRPCURL     string
	ChainID              string
	NetworkName          string
	Port                 string
	TransactionLogger    string
	FundManager          string
	BlockchainPrivateKey string
	GasLimit             string
	GasPrice             string
}

var config *Config

// GetConfig returns the application configuration
func GetConfig() *Config {
	if config == nil {
		config = &Config{
			AppEnv:               getEnvWithDefault("APP_ENV", "development"),
			LogLevel:             getEnvWithDefault("LOG_LEVEL", "info"),
			BlockchainServiceURL: getEnvWithDefault("BLOCKCHAIN_SERVICE_URL", "https://sepolia.infura.io/v3/"),
			BlockchainRPCURL:     getEnvWithDefault("BLOCKCHAIN_RPC_URL", "https://sepolia.infura.io/v3/"),
			ChainID:              getEnvWithDefault("CHAIN_ID", "11155111"),
			NetworkName:          getEnvWithDefault("NETWORK_NAME", "Sepolia"),
			Port:                 getEnvWithDefault("PORT", "8080"),
			TransactionLogger:    getEnvWithDefault("TRANSACTION_LOGGER_ADDRESS", "0x87009637ff84b57dED05ceEABac3ccc44c4E1E7c"),
			FundManager:          getEnvWithDefault("FUND_MANAGER_ADDRESS", "0xC0b6BfC3F32564D92d6c14D3B489d6Bc9f16b256"),
			BlockchainPrivateKey: getEnvWithDefault("BLOCKCHAIN_PRIVATE_KEY", ""),
			GasLimit:             getEnvWithDefault("GAS_LIMIT", "3000000"),
			GasPrice:             getEnvWithDefault("GAS_PRICE", "3000000000"),
		}
	}
	return config
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsSepoliaNetwork returns true if configured for Sepolia testnet
func IsSepoliaNetwork() bool {
	return GetConfig().NetworkName == "Sepolia" || GetConfig().ChainID == "11155111"
}

// GetNetworkConfig returns a map of network configuration details
func GetNetworkConfig() map[string]string {
	config := GetConfig()
	return map[string]string{
		"network":     config.NetworkName,
		"chainId":     config.ChainID,
		"rpcUrl":      config.BlockchainRPCURL,
		"serviceUrl":  config.BlockchainServiceURL,
		"logger":      config.TransactionLogger,
		"fundManager": config.FundManager,
	}
}
