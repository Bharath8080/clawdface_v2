package utils

import (
	"backend/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func StartVoiceSession(sessionReq types.SessionRequest) (error, types.StartConversationResponse) {
	client := &http.Client{}
	sessionReqBytes, err := json.Marshal(sessionReq)
	if err != nil {
		fmt.Println("Unable to parse json")
		return err, types.StartConversationResponse{}
	}
	//URL := "https://api.runpod.ai/v2/" + os.Getenv("RUNPOD_ENDPOINT_ID") + "/run"
	URL := os.Getenv("API_URL")
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(sessionReqBytes))
	if err != nil {
		fmt.Println("Error creating HTTP Request")
		return err, types.StartConversationResponse{}
	}
	//req.Header.Add("Authorization", "Bearer "+os.Getenv("RUNPOD_API_KEY"))
	req.Header.Add("Authorization", "Bearer "+os.Getenv("API_TOKEN"))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting response")
		return err, types.StartConversationResponse{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	response := types.StartConversationResponse{}
	json.Unmarshal(body, &response)
	return nil, response
}

func StartVoiceSession11Lab(sessionReq types.SessionRequest11Lab) (error, types.StartConversationResponse11Lab) {
	client := &http.Client{}
	sessionReqBytes, err := json.Marshal(sessionReq)
	if err != nil {
		fmt.Println("Unable to parse json")
		return err, types.StartConversationResponse11Lab{}
	}
	URL := os.Getenv("API_URL_11LAB")
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(sessionReqBytes))
	if err != nil {
		fmt.Println("Error creating HTTP Request")
		return err, types.StartConversationResponse11Lab{}
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting response")
		return err, types.StartConversationResponse11Lab{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	response := types.StartConversationResponse11Lab{}
	json.Unmarshal(body, &response)
	return nil, response
}

func CancelVoiceSession(runID string) error {
	client := &http.Client{}

	projectID := os.Getenv("API_PROJECT_ID")
	appID := os.Getenv("API_APP_ID")
	url := fmt.Sprintf(
		"https://rest.cerebrium.ai/v2/projects/%s/apps/%s/runs/%s",
		projectID,
		appID,
		runID,
	)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("API_TOKEN"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	// Optional: read response body for debugging/logging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Optional status check
	if resp.StatusCode >= 300 {
		return fmt.Errorf("cerebrium delete failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
