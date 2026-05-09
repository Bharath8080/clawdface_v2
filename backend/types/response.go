package types

import "time"

type StartConversationResponse struct {
	ConversationId string `json:"conversationId"`
	URL            string `json:"url"`
	LivekitURL     string `json:"livekit_url"`
	Token          string `json:"token"`
	ImageURL       string `json:"image_url"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	RunID          string `json:"run_id"`
	OpenClawURL    string `json:"openclaw_url"`
	GatewayToken   string `json:"gateway_token"`
}

type StartConversationResponse11Lab struct {
	Success                bool      `json:"success"`
	SignedUrl              string    `json:"signedUrl"`
	SessionId              string    `json:"sessionId"`
	ExpiresAt              time.Time `json:"expiresAt"`
	SessionDurationMinutes int       `json:"sessionDurationMinutes"`
}

type StartUsageResponse struct {
	Status string `json:"status"`
}
