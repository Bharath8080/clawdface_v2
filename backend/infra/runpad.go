package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response structure matching Runpod's health API
type HealthResponse struct {
	Jobs struct {
		Completed  int `json:"completed"`
		Failed     int `json:"failed"`
		InProgress int `json:"inProgress"`
		InQueue    int `json:"inQueue"`
		Retried    int `json:"retried"`
	} `json:"jobs"`
	Workers struct {
		Idle         int `json:"idle"`
		Initializing int `json:"initializing"`
		Ready        int `json:"ready"`
		Running      int `json:"running"`
		Throttled    int `json:"throttled"`
		Unhealthy    int `json:"unhealthy"`
	} `json:"workers"`
}

type JobStatus struct {
	DelayTime     int    `json:"delayTime"`
	ExecutionTime int    `json:"executionTime"`
	Id            string `json:"id"`
	Status        string `json:"status"`
	WorkerId      string `json:"workerId"`
	Output        struct {
		Status string `json:"status"`
	} `json:"output"`
}

func GetHealthResponse(urls, token string) (*HealthResponse, error) {
	url := urls + "health"

	// Create new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Return the full response object
	return &health, nil
}

func GetHealth(url string, token string) (*HealthResponse, error) {
	health, err := GetHealthResponse(url, token)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return health, nil
}

func IsReady(url string, token string) (bool, error) {
	health, err := GetHealthResponse(url, token)
	if err != nil {
		fmt.Println("Error:", err)
		return false, err
	}
	return health.Workers.Ready > 0, nil
}

func PostJob(url string, configJSON map[string]interface{}, token string) (string, error) {

	payloadBytes, err := json.Marshal(configJSON)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // Pass Bearer token here

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", string(body))
	}

	return string(body), nil
}

func DeleteJob(url string, token string) (string, error) {
	// Create DELETE request
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating delete request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Check for success (202)
	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func GetJobStatus(urls string, token string, jobId string) (*JobStatus, error) {
	url := urls + "status/" + jobId
	// Create new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var status JobStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return &status, nil
}
