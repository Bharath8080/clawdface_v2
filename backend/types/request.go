package types

import (
	"encoding/json"
)

type StartConversationRequest struct {
	AgentId    string          `json:"agentId"`
	UserName   string          `json:"userName"`
	UserId     string          `json:"userId"`
	Context    json.RawMessage `json:"context"`
	MetaData   json.RawMessage `json:"metaData"`
	RoomId     string          `json:"roomId"`
	MeetingURL string          `json:"meetingURL"`
	Email      string          `json:"email"`
	AvatarId   string          `json:"avatarId"`
	Mode       string          `json:"mode"`
}

type StartUsageRequest struct {
	ConversationId string          `json:"conversation_Id"`
	Status         string          `json:"status"`
	JobId          string          `json:"job_Id"`
	Usage          json.RawMessage `json:"usage"`
	SessionId      string          `json:"session_Id"`
	ParticipantId  string          `json:"participant_Id"`
	Message        string          `json:"message"`
}

type RAGUpdateRequest struct {
	DocId   string `json:"doc_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Start Voice to Video Session Request
type SessionRequest struct {
	Input SessionInput `json:"input"`
}

// Start Voice to Video Session Request
type SessionRequest11Lab struct {
	UserId   string `json:"userId"`
	AvatarId string `json:"avatarId"`
}

type SessionInput struct {
	ConversationID string        `json:"conversation_id"`
	Avatar         VisualConfig  `json:"avatar"`
	Livekit        LivekitConfig `json:"lk"`
}

type VisualConfig struct {
	ID           string `json:"id"`
	Height       int    `json:"height"`
	Width        int    `json:"width"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	TargetWidth  int    `json:"target_width"`
	TargetHeight int    `json:"target_height"`
}

type LivekitConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
