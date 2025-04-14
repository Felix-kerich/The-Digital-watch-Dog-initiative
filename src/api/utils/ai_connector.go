package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// AnomalyDetectionRequest represents a request to the AI service for anomaly detection
type AnomalyDetectionRequest struct {
	Transaction    map[string]interface{}   `json:"transaction"`
	HistoricalData []map[string]interface{} `json:"historical_data,omitempty"`
}

// AnomalyDetectionResponse represents the response from the AI service
type AnomalyDetectionResponse struct {
	TransactionID string  `json:"transaction_id"`
	IsAnomaly     bool    `json:"is_anomaly"`
	AnomalyScore  float64 `json:"anomaly_score"`
	Reason        string  `json:"reason,omitempty"`
}

// DetectAnomaly sends a transaction to the AI service for anomaly detection
func DetectAnomaly(transaction map[string]interface{}, historicalData []map[string]interface{}) (*AnomalyDetectionResponse, error) {
	// Get the AI service URL from environment variables
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		return nil, errors.New("AI_SERVICE_URL environment variable not set")
	}

	// Create the request
	request := AnomalyDetectionRequest{
		Transaction:    transaction,
		HistoricalData: historicalData,
	}

	// Convert request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send the request to the AI service
	resp, err := client.Post(
		aiServiceURL+"/detect",
		"application/json",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to AI service: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned non-OK status: %d", resp.StatusCode)
	}

	// Decode the response
	var response AnomalyDetectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

// RetrainModel sends a request to the AI service to retrain its model with new data
func RetrainModel(transactions []map[string]interface{}) error {
	// Get the AI service URL from environment variables
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		return errors.New("AI_SERVICE_URL environment variable not set")
	}

	// Create the request
	request := map[string]interface{}{
		"transactions": transactions,
	}

	// Convert request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with longer timeout for retraining
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Send the request to the AI service
	resp, err := client.Post(
		aiServiceURL+"/retrain",
		"application/json",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to send retrain request to AI service: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service returned non-OK status for retraining: %d", resp.StatusCode)
	}

	return nil
}

// CheckAIServiceHealth checks if the AI service is up and running
func CheckAIServiceHealth() (bool, error) {
	// Get the AI service URL from environment variables
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		return false, errors.New("AI_SERVICE_URL environment variable not set")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Send a GET request to the health endpoint
	resp, err := client.Get(aiServiceURL + "/health")
	if err != nil {
		return false, fmt.Errorf("failed to connect to AI service: %v", err)
	}
	defer resp.Body.Close()

	// Check if status is OK
	return resp.StatusCode == http.StatusOK, nil
}
