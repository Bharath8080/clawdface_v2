package utils

import (
	"backend/configs"
	"backend/infra"
	"backend/langchain"
	vectors "backend/pinecone"
	"backend/types"
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"math/rand"
	"sort"
	"time"

	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"baliance.com/gooxml/document"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"google.golang.org/api/option"
)

var DB *sql.DB

const dateLayoutYYYYMMDD = "2006-01-02"

type ScheduledJob struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Cron         string `json:"cron"`
	AgentEmailID string `json:"agentEmailID"`
	MeetingURL   string `json:"meetingUrl"`
	Expiry       string `json:"expiry"`
	Status       string `json:"status"`
	StartTime    string `json:"start_time,omitempty"`
	UID          string `json:"uid,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	VideoURL     string `json:"video_url,omitempty"`
}

type JoinExternalMeetingsRequest struct {
	AgentID               string `json:"agent_id"`
	MeetingURL            string `json:"meeting_url"`
	ConversationalContext string `json:"context"`
	UserName              string `json:"user_name"`
	UserID                string `json:"user_id"`
	WakePhrase            string `json:"wake_phrase,omitempty"`
}

type TokenSourceRequest struct {
	RoomName              string                     `json:"room_name"`
	ParticipantName       string                     `json:"participant_name"`
	ParticipantIdentity   string                     `json:"participant_identity"`
	ParticipantMetadata   string                     `json:"participant_metadata"`
	ParticipantAttributes map[string]string          `json:"participant_attributes"`
	RoomConfig            *livekit.RoomConfiguration `json:"room_config"`
}

func getRequestClientIP(r *http.Request) string {
	if r == nil {
		log.Printf("getRequestClientIP: nil request")
		return ""
	}

	if cf := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cf != "" {
		log.Printf("getRequestClientIP: using CF-Connecting-IP")
		return cf
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		log.Printf("getRequestClientIP: using X-Real-IP")
		return xr
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		log.Printf("getRequestClientIP: using X-Forwarded-For")
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		log.Printf("getRequestClientIP: using RemoteAddr host")
		return host
	}

	log.Printf("getRequestClientIP: using RemoteAddr raw")
	return r.RemoteAddr
}

func getUserCreateClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := strings.TrimSpace(r.Header.Get("IP")); ip != "" {
		return ip
	}

	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

func HandleGetClientIP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("HandleGetClientIP: remote_addr=%q cf=%q x_real=%q xff=%q", r.RemoteAddr, r.Header.Get("CF-Connecting-IP"), r.Header.Get("X-Real-IP"), r.Header.Get("X-Forwarded-For"))

	ip := getRequestClientIP(r)
	log.Printf("HandleGetClientIP: resolved_ip=%q", ip)

	resp := map[string]string{
		"ip":               ip,
		"remote_addr":      r.RemoteAddr,
		"cf_connecting_ip": r.Header.Get("CF-Connecting-IP"),
		"x_real_ip":        r.Header.Get("X-Real-IP"),
		"x_forwarded_for":  r.Header.Get("X-Forwarded-For"),
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Handle request payload validation
func (r *JoinExternalMeetingsRequest) Validate() error {
	if r.AgentID == "" {
		return errors.New("agent_id is required")
	}
	if r.MeetingURL == "" {
		return errors.New("meeting_url is required")
	}
	if r.UserName == "" {
		return errors.New("user_name is required")
	}
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

type Transcript struct {
	Timestamp         string          `json:"timestamp"`
	Role              string          `json:"role"`
	Content           string          `json:"content"`
	Conversation_id   string          `json:"conversation_id"`
	Message_timestamp int32           `json:"message_timestamp"`
	Name              string          `json:"name"`
	Arguments         string          `json:"arguments"`
	Output            string          `json:"output"`
	IsError           bool            `json:"is_error"`
	Extra             json.RawMessage `json:"extra"`
	CallID            string          `json:"call_id"`
	Type              string          `json:"type"`
}

type InputFile struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Role      string                 `json:"role"`
	Content   []string               `json:"content"`
	Metrics   map[string]interface{} `json:"metrics"`
	Name      string                 `json:"name"`
	Arguments string                 `json:"arguments"`
	Output    string                 `json:"output"`
	IsError   bool                   `json:"is_error"`
	Extra     json.RawMessage        `json:"extra"`
	CallID    string                 `json:"call_id"`
}

type IdentifyRequest struct {
	UserID string `json:"userId"`
}

type LLMConfig struct {
	Model         string `json:"model"`
	Provider      string `json:"provider"`
	UseNLTK       bool   `json:"use_nltk"`
	FallbackModel string `json:"fallback_model"`
	URL           string `json:"url"`
	Token         string `json:"token"`
}

type STTConfig struct {
	Model                          string  `json:"model"`
	Language                       string  `json:"language"`
	Provider                       string  `json:"provider"`
	FallbackModel                  string  `json:"fallback_model"`
	MaxEndpointingDelay            float64 `json:"max_endpointing_delay"`
	MinEndpointingDelay            float64 `json:"min_endpointing_delay"`
	AllowIntermResultsInterruption bool    `json:"allow_interm_results_interruption"`
}

type LLMAgent struct {
	Model           string `json:"model"`
	Provider        string `json:"provider"`
	UseNLTK         bool   `json:"use_nltk"`
	FallbackModel   string `json:"fallback_model"`
	ReasoningEffort string `json:"reasoning_effort"`
	URL             string `json:"url"`
	Token           string `json:"token"`
}

type STTAgent struct {
	Model         string `json:"model"`
	Language      string `json:"language"`
	Provider      string `json:"provider"`
	FallbackModel string `json:"fallback_model"`
	WakePhrase    string `json:"wake_phrase,omitempty"`
}

type TTSAgent struct {
	Pitch               int                   `json:"pitch"`
	Gender              string                `json:"gender"`
	Encoding            string                `json:"encoding"`
	Language            string                `json:"language"`
	ModelID             string                `json:"model_id"`
	Provider            string                `json:"provider"`
	VoiceID             string                `json:"voice_id"`
	Stability           float64               `json:"stability"`
	SampleRate          int                   `json:"sample_rate"`
	SpeakingRate        float64               `json:"speaking_rate"`
	SimilarityBoost     float64               `json:"similarity_boost"`
	FallbackVoiceID     string                `json:"fallback_voice_id"`
	EffectsProfileID    string                `json:"effects_profile_id"`
	CustomPronunciation []CustomPronunciation `json:"customPronounciation"`
}

type VAD struct {
	Enabled             bool    `json:"enabled"`
	Provider            string  `json:"provider"`
	MinSilenceDuration  float64 `json:"min_silence_duration"`
	ActivationThreshold float64 `json:"activation_threshold"`
}

type TTSConfig struct {
	Pitch               int                   `json:"pitch"`
	Gender              string                `json:"gender"`
	Encoding            string                `json:"encoding"`
	Language            string                `json:"language"`
	ModelID             string                `json:"model_id"`
	Provider            string                `json:"provider"`
	VoiceID             string                `json:"voice_id"`
	Stability           float64               `json:"stability"`
	SampleRate          int                   `json:"sample_rate"`
	SpeakingRate        float64               `json:"speaking_rate"`
	SimilarityBoost     float64               `json:"similarity_boost"`
	FallbackVoiceID     string                `json:"fallback_voice_id"`
	EffectsProfileID    string                `json:"effects_profile_id"`
	CustomPronunciation []CustomPronunciation `json:"customPronounciation"`
}

type CustomPronunciation struct {
	Word          string `json:"word"`
	Pronunciation string `json:"pronounciation"`
}

type ProtocolConfig struct {
	Simulcast    bool   `json:"simulcast"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate int    `json:"video_bitrate"`
}

type SuperResolutionConfig struct {
	Scale   float64 `json:"scale"`
	Enabled bool    `json:"enabled"`
}

type NoiseCancellationConfig struct {
	Provider string `json:"provider"`
}

type WelcomeMessage struct {
	Messages []string `json:"messages"`
	WaitTime int      `json:"wait_time"`
}

type ExitMessage struct {
	Messages        []string `json:"messages"`
	CalloutBefore   int      `json:"callout_before"`
	MaxCallDuration int      `json:"max_call_duration"`
}

type IdleTimeout struct {
	Timeout       int      `json:"timeout"`
	FillerPhrases []string `json:"filler_phrases"`
	Messages      []string `json:"messages"`
}

type Config struct {
	LLM               LLMConfig               `json:"llm"`
	STT               STTConfig               `json:"stt"`
	TTS               TTSConfig               `json:"tts"`
	Protocol          ProtocolConfig          `json:"protocol"`
	TurnDetector      bool                    `json:"turn_detector"`
	SuperResolution   SuperResolutionConfig   `json:"super_resolution"`
	NoiseCancellation NoiseCancellationConfig `json:"noise_cancellation"`
}

type KnowledgeBase struct {
	Enabled       bool   `json:"enabled"`
	RetrievalTopK int    `json:"retrieval_top_k"`
	IndexHost     string `json:"index_host"`
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Mode          string `json:"mode"`
}

type Memory struct {
	Enabled      bool   `json:"isEnabled"`
	Instructions string `json:"instruction"`
}
type EventMessageDelay struct {
	Delay   int    `json:"delay"`
	Message string `json:"message"`
}

type EventMessages struct {
	OnDelay EventMessageDelay `json:"on_delay"`
	OnError struct {
		Message string `json:"message"`
	} `json:"on_error"`
}

type RequestConfigMCP struct {
	URL     string          `json:"url"`
	Headers json.RawMessage `json:"headers"`
}

type RequestConfigTool struct {
	Name    string          `json:"name"`
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	Headers json.RawMessage `json:"headers"`
}

// ---------------- Scene Context Engine ----------------
type SceneAction struct {
	Type                string `json:"Type"`
	ActionName          string `json:"Action_Name"`
	AnalysisInstruction string `json:"Analysis_Instruction"`
	ActionNeedsObserved string `json:"Action_Needs_To_Be_Observed"`
}

type LLMPrompts struct {
	FirstQuery                          string        `json:"first_query"`
	ActionsList                         []SceneAction `json:"actions_list"`
	AnalyzeAction                       string        `json:"analyze_action"`
	GetUserAppearance                   string        `json:"get_user_appearance"`
	SyntheticUserQuery                  string        `json:"synthetic_user_query"`
	AnalyzeSceneCtxResponse             string        `json:"analyze_scene_ctx_response"`
	AnalyzeActionsSystemPrompt          string        `json:"analyze_actions_system_prompt"`
	UserQueryAnalysisSystemPrompt       string        `json:"user_query_analysis_system_prompt"`
	AddActionRecognitionSyntheticUserQy bool          `json:"add_action_recognition_synthetic_user_query"`
}

type SceneContextEngine struct {
	VisionLLM         string     `json:"vision_llm"`
	LLMPrompts        LLMPrompts `json:"llm_prompts"`
	SnapshotScale     float64    `json:"snapshot_scale"`
	OnSnapshotTimeout int        `json:"on_snapshot_timeout"`
	Interpolation     struct {
		Exp     int  `json:"exp"`
		Enabled bool `json:"enabled"`
	} `json:"interpolation_config"`
}

// ---------------- Scene Analyzer Prompt ----------------
type SceneAnalyzerPrompt struct {
	TaskPrompt   string `json:"task_prompt"`
	SystemPrompt string `json:"system_prompt"`
}

type Callback struct {
	URL            string   `json:"url"`
	EventsToListen []string `json:"events_to_listen"`
}

type Communication struct {
	Provider   string `json:"provider"`
	MeetingURL string `json:"meeting_url"`
}

type ConversationalContext struct {
	Text       string `json:"text"`
	WakePhrase string `json:"wake_phrase,omitempty"`
}

type CreateCollectionResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	TotalChunks int    `json:"total_chunks"`
	TotalFiles  int    `json:"total_files"`
	Files       []struct {
		Filename   string `json:"filename"`
		DocumentID string `json:"document_id"`
		ChunkCount int    `json:"chunk_count"`
	} `json:"files"`
}

type TextDoc struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}

type URLDoc struct {
	Text string `json:"text"`
	ID   string `json:"id"`
	URL  string `json:"url,omitempty"`
}

type Avatar struct {
	AvatarID               string              `json:"avatar_key_id"`
	AvatarName             string              `json:"avatar_name"`
	AvatarProfilePic       string              `json:"profile_pic"`
	Gender                 string              `json:"gender"`
	PersonaName            string              `json:"persona_name"`
	PersonaPrompt          string              `json:"persona_prompt"`
	Config                 Config              `json:"config"`
	Timeout                int                 `json:"timeout"`
	FrameRate              int                 `json:"frame_rate"`
	SilencePadding         float64             `json:"silence_padding"`
	AvatarDataSource       string              `json:"avatar_data_source"`
	AudioFeaturesType      string              `json:"audio_features_type"`
	EyeMaskReplacement     bool                `json:"eye_mask_replacement"`
	IsFaceEnhancerEnabled  bool                `json:"is_face_enhancer_enabled"`
	AudioFeaturesWindowLen int                 `json:"audio_features_window_length"`
	ConversationalContext  string              `json:"conversational_context"`
	WelcomeMessage         WelcomeMessage      `json:"welcome_message"`
	IdleTimeout            IdleTimeout         `json:"idle_timeout"`
	ExitMessage            ExitMessage         `json:"exit_message"`
	ExitHeadsUpMessage     ExitMessage         `json:"exit_heads_up_message"`
	WarningExitMessage     ExitMessage         `json:"warning_exit_message"`
	KnowledgeBase          []KnowledgeBase     `json:"knowledge_base"`
	MCP                    []MCP               `json:"mcp_servers"`
	SceneContextEngine     SceneContextEngine  `json:"scene_context_engine"`
	SceneAnalyzerPrompt    SceneAnalyzerPrompt `json:"scene_analyzer_prompt"`
	RecordRoom             bool                `json:"record_room"`
	Callback               *Callback           `json:"callback,omitempty"`
	Communication          *Communication      `json:"communication,omitempty"`
}

type Agent struct {
	AvatarID       string `json:"avatar_key_id"`
	AvatarName     string `json:"avatar_name"`
	AvatarImageURL string `json:"image_url"`
	Mode           string `json:"mode"`
	Avatar         struct {
		ID           string `json:"id"`
		Height       int    `json:"height"`
		Width        int    `json:"width"`
		X            int    `json:"x"`
		Y            int    `json:"y"`
		TargetWidth  int    `json:"target_width"`
		TargetHeight int    `json:"target_height"`
	} `json:"avatar"`
	AgentId            string          `json:"agent_id"`
	PersonaName        string          `json:"persona_name"`
	PersonaPrompt      string          `json:"persona_prompt"`
	IdleTimeout        IdleTimeout     `json:"idle_timeout"`
	Timeout            int             `json:"timeout"`
	WarningExitMessage ExitMessage     `json:"warning_exit_message"`
	ExitMessage        ExitMessage     `json:"exit_message"`
	WelcomeMessage     WelcomeMessage  `json:"welcome_message"`
	Config             AgentsConfig    `json:"config"`
	KnowledgeBase      []KnowledgeBase `json:"knowledge_base"`
	Memory             Memory          `json:"memory"`
	Actions            Actions         `json:"actions"`
	RecordRoom         bool            `json:"record_room"`
	Callback           *Callback       `json:"callback,omitempty"`
	Communication      *Communication  `json:"communication,omitempty"`
	OpenClawURL        string          `json:"openclaw_url"`
	GatewayToken       string          `json:"gateway_token"`
	SessionKey         string          `json:"sessionKey"`
	ConnectionType     string          `json:"connection_type"`
	AvatarId           string          `json:"avatarId"`
	ConversationId     string          `json:"conversation_id"`
}

// Actions
type Actions struct {
	MCPServers  []MCP  `json:"mcp_servers"`
	HTTPTools   []Tool `json:"http_tools"`
	ClientTools []Tool `json:"client_tools"`
}

type MCP struct {
	Type          string           `json:"type"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	RequestConfig RequestConfigMCP `json:"request_config"`
	EventMessages json.RawMessage  `json:"event_messages"`
}

type Tool struct {
	Type          string            `json:"type"`
	Schema        json.RawMessage   `json:"schema"`
	RequestConfig RequestConfigTool `json:"request_config"`
	EventMessages json.RawMessage   `json:"event_messages"`
}

type Schema struct {
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string                       `json:"type"`
	Properties map[string]ParameterProperty `json:"properties"`
	Required   []string                     `json:"required"`
}

type ParameterProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type RequestConfig struct {
	Name string `json:"name"`
}

type SessionResponse struct {
	Participants []Participant `json:"participants"`
}

type Participant struct {
	ParticipantIdentity string    `json:"participantIdentity"`
	Location            string    `json:"location"`
	Region              string    `json:"region"`
	Sessions            []Session `json:"sessions"`
}

type Session struct {
	ParticipantID string `json:"participantId"`
}

type AgentsConfig struct {
	LLM LLMAgent `json:"llm"`
	STT STTAgent `json:"stt"`
	TTS TTSAgent `json:"tts"`
	//Protocol            ProtocolConfig          `json:"protocol"`
	PreemptiveSynthesis bool                    `json:"preemptive_synthesis"`
	TurnDetector        bool                    `json:"turn_detector"`
	VAD                 VAD                     `json:"vad"`
	NoiseCancellation   NoiseCancellationConfig `json:"noise_cancellation"`
}

type AgentConfig struct {
	AgentName           string          `json:"agent_name"`
	AgentSystemPrompt   string          `json:"agent_system_prompt"`
	DefaultSystemPrompt bool            `json:"default_system_prompt"`
	Config              json.RawMessage `json:"config"`
	IsActive            bool            `json:"is_active"`
	Avatars             json.RawMessage `json:"avatars"`
	IsPublic            bool            `json:"is_public"`
	KnowledgeBase       []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Mode string `json:"mode"`
	} `json:"knowledge_base"`
	MCP []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"mcp"`
	Tool []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tool"`
	Record          bool     `json:"record"`
	Callback_url    string   `json:"callback_url"`
	Callback_events []string `json:"callback_events" db:"callback_events"`
}

type AvatarCon struct {
	ID             string `json:"id"`
	AvatarKeyID    string `json:"avatar_key_id"`
	AvatarName     string `json:"avatar_name"`
	DisplayPicture string `json:"display_picture"`
	ImageURL       string `json:"image_url"`
	Gender         string `json:"gender"`
	DefaultPrompt  string `json:"default_prompt"`
}

type Avatars struct {
	ID             string          `json:"id"`
	AvatarKeyID    string          `json:"avatar_key_id"`
	AvatarName     string          `json:"avatar_name"`
	Status         string          `json:"status"`
	DisplayPicture string          `json:"display_picture"`
	ImageURL       string          `json:"image_url"`
	Gender         string          `json:"gender"`
	MetaData       json.RawMessage `json:"meta_data"`
	DefaultPrompt  string          `json:"default_prompt"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type AvatarsList struct {
	ID             string    `json:"id"`
	AvatarKeyID    string    `json:"avatar_key_id"`
	AvatarName     string    `json:"avatar_name"`
	Status         string    `json:"status"`
	DisplayPicture string    `json:"display_picture"`
	ImageURL       string    `json:"image_url"`
	Gender         string    `json:"gender"`
	MetaData       string    `json:"meta_data"`
	DefaultPrompt  string    `json:"default_prompt"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type KBInputDoc struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	URL      string `json:"url,omitempty"`
	WebURL   string `json:"web_url,omitempty"`
}

type EmailTemplate struct {
	ID           string
	TemplateName string
	Description  string
	FromEmail    string
	CCEmail      sql.NullString // ✅ safe for NULL
	EmailContent string
	Status       bool
}

type KBInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsActive    bool         `json:"is_active"`
	Input       []KBInputDoc `json:"input"`
}

type DocumentRecord struct {
	ID       string
	Name     string
	FileType string
	URL      string
	WebURL   string
	Content  string
}

type URLPayload struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

type ScriptToVideoRequest struct {
	AvatarID     string `json:"avatar_id"`
	VoiceID      string `json:"voice_id"`
	ProviderName string `json:"provider_name"`
	ModelName    string `json:"model_name"`
	Script       string `json:"script"`
	CallbackURL  string `json:"callback_url"`
}

// ---------------------------
// Shared Struct
// ---------------------------
type AgentConfigInput struct {
	AgentName           string          `json:"agent_name"`
	AgentSystemPrompt   string          `json:"agent_system_prompt"`
	DefaultSystemPrompt bool            `json:"default_system_prompt"`
	Config              json.RawMessage `json:"config"`
	IsActive            bool            `json:"is_active"`
	Avatars             json.RawMessage `json:"avatars"`
	IsPublic            bool            `json:"is_public"`
	KnowledgeBase       []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Mode string `json:"mode"`
	} `json:"knowledge_base"`
	MCP []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"mcp"`
	Tool []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tool"`
	Integration []struct {
		ID string `json:"id"`
	} `json:"integration"`
	Record         bool            `json:"record"`
	CallbackURL    string          `json:"callback_url"`
	CallbackEvents []string        `json:"callback_events"`
	Email          string          `json:"email"`
	Type           string          `json:"type"`
	AddOns         json.RawMessage `json:"add_on"`
}

type PaymentStatusResponse struct {
	HasPaymentFailed bool   `json:"hasPaymentFailed"`
	Error            string `json:"error,omitempty"`
}

type GraphData struct {
	Label      string   `json:"label"`
	Sessions   int      `json:"sessions"`
	Seconds    float64  `json:"seconds"`
	Concurrent int      `json:"concurrent"`
	Countries  []string `json:"countries"`
}
type StatsResponse struct {
	TotalSessions   int         `json:"total_sessions"`
	TotalSeconds    float64     `json:"total_seconds"`
	TotalCountries  int         `json:"total_countries"`
	PeakConcurrency int         `json:"peak_concurrency"`
	Concurrent      int         `json:"concurrent"`
	GraphView       []GraphData `json:"graphView"`
}

type GlobalConversationDetail struct {
	ConversationID   string    `json:"conversation_id"`
	OrganizationID   string    `json:"org_id,omitempty"`
	OrganizationName string    `json:"organization_name,omitempty"`
	UserID           string    `json:"user_id,omitempty"`
	UserName         string    `json:"user_name,omitempty"`
	AgentID          string    `json:"agent_id,omitempty"`
	AgentName        string    `json:"agent_name,omitempty"`
	Status           string    `json:"status,omitempty"`
	Type             string    `json:"type,omitempty"`
	AvatarID         string    `json:"avatar_id,omitempty"`
	Region           string    `json:"region,omitempty"`
	Location         string    `json:"location,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	Seconds          float64   `json:"seconds"`
}

type GlobalGraphData struct {
	Label             string                     `json:"label"`
	Sessions          int                        `json:"sessions"`
	Seconds           float64                    `json:"seconds"`
	Concurrent        int                        `json:"concurrent"`
	Countries         []string                   `json:"countries"`
	Organizations     []string                   `json:"org_ids"`
	OrganizationNames []string                   `json:"org_names"`
	Conversations     []GlobalConversationDetail `json:"conversations"`
}

type GlobalMetricsResponse struct {
	StartDate          string            `json:"start_date"`
	EndDate            string            `json:"end_date"`
	GroupBy            string            `json:"group_by"`
	AppliedFilters     map[string]string `json:"applied_filters"`
	Countries          []string          `json:"countries"`
	OrgIDs             []string          `json:"org_ids"`
	OrgNames           []string          `json:"org_names"`
	TotalSessions      int               `json:"total_sessions"`
	TotalSeconds       float64           `json:"total_seconds"`
	TotalCountries     int               `json:"total_countries"`
	TotalOrganizations int               `json:"total_organizations"`
	TotalUsers         int               `json:"total_users"`
	PeakConcurrency    int               `json:"peak_concurrency"`
	Concurrent         int               `json:"concurrent"`
	GraphView          []GlobalGraphData `json:"graphView"`
}

type Organization struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	PreSignLogo string `json:"pre_sign_logo"`
	Owner       string `json:"owner"`
}

type AgentOrganization struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	PreSignLogo string `json:"pre_sign_logo"`
	Owner       string `json:"owner"`
	AvatarLogo  string `json:"avatar_logo"`
	AvatarType  string `json:"avatar_type"`
}

type AgentAvatars struct {
	Name     string `json:"name"`
	ID       string `json:"key_id"`
	ImageURL string `json:"image_url"`
	Timeout  int    `json:"timeout"`
}

// InitDB initializes the database connection
func InitDB(connectionString string) error {
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return err
	}

	log.Println("Attempting to ping database...")
	err = DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	log.Println("Database connection established successfully.")
	return nil
}

func GetEmailTemplateByName(templateName string) (*EmailTemplate, error) {

	query := `
		SELECT 
			id,
			template_name,
			description,
			from_email,
			cc_email,
			email_content,
			status
		FROM public.email_templates
		WHERE LOWER(template_name) = LOWER($1)
		AND status = true
		LIMIT 1
	`

	var tmpl EmailTemplate

	err := DB.QueryRow(query, templateName).Scan(
		&tmpl.ID,
		&tmpl.TemplateName,
		&tmpl.Description,
		&tmpl.FromEmail,
		&tmpl.CCEmail,
		&tmpl.EmailContent,
		&tmpl.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tmpl, nil
}

func HandleGetOrganizationByAgentID(w http.ResponseWriter, r *http.Request, agentID string) error {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching Organization for Agent ID: %s", agentID)

	query := `
		SELECT
			COALESCE(o.name, ''),
			COALESCE(o.description, ''),
			COALESCE(o.logo, ''),
			COALESCE(o.owner, ''),
			a.avatars::text
		FROM agents a
		JOIN api_keys ak2
			ON a.created_by = ak2.id
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id
		JOIN organizations o
			ON w2.organization_id = o.id
		JOIN api_keys ak1
			ON ak1.workspace_id = ak2.workspace_id
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id
		   AND w1.organization_id = w2.organization_id
		WHERE a.id = $1
		ORDER BY a.created_at DESC
		LIMIT 1;
	`

	var org AgentOrganization
	var avatars string
	err := DB.QueryRow(query, agentID).Scan(
		&org.Name,
		&org.Description,
		&org.Logo,
		&org.Owner,
		&avatars,
	)

	if err == sql.ErrNoRows {
		log.Printf("No organization found for agent ID: %s", agentID)
		return fmt.Errorf("organization not found for agent ID: %s", agentID)
	} else if err != nil {
		log.Printf("Error retrieving Organization: %v", err)
		return fmt.Errorf("failed to retrieve organization: %v", err)
	}

	if org.Logo != "" {
		bucket := configs.GetEnv("AWS_BUCKET")
		region := configs.GetEnv("AWS_REGION")
		presign, presignDownload, err := PreSignURL(bucket, org.Logo, region)
		if err != nil {
			log.Printf("Error presigning logo URL: %v", err)
			log.Printf("Error presigning logo URL: %v", presignDownload)
		}
		org.PreSignLogo = presign
	}

	selectedAgent, selectedCount, agentErr := getRandomAgent(avatars, "")
	if agentErr != nil {
		return fmt.Errorf("getRandomAgent failed: %v, %v", agentErr, selectedCount)
	}

	org.AvatarType = "random"
	if selectedCount == 1 {
		var imageURL string
		err = DB.QueryRow(`
        SELECT image_url
        FROM avatars
        WHERE avatar_key_id = $1
    `, selectedAgent.AvatarID).Scan(&imageURL)
		if err != nil {
			return fmt.Errorf("avatar config load failed: %v", err)
		}
		org.AvatarLogo = imageURL
		org.AvatarType = "single"
	}

	return json.NewEncoder(w).Encode(org)
}

func HandleGetOrganizationByAgentID1(w http.ResponseWriter, r *http.Request, agentID string) error {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching Organization for Agent ID: %s", agentID)

	query := `
		SELECT
			COALESCE(o.name, ''),
			COALESCE(o.description, ''),
			COALESCE(o.logo, ''),
			COALESCE(o.owner, ''),
			a.avatars::text,
			a.agent_name,
			COALESCE(NULLIF(a.config->>'timeout', ''), '0')::INTEGER AS timeout
		FROM agents a
		JOIN api_keys ak2
			ON a.created_by = ak2.id
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id
		JOIN organizations o
			ON w2.organization_id = o.id
		JOIN api_keys ak1
			ON ak1.workspace_id = ak2.workspace_id
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id
		   AND w1.organization_id = w2.organization_id
		WHERE a.id = $1
		ORDER BY a.created_at DESC
		LIMIT 1;
	`

	var org Organization
	var avatarsList []AgentAvatars
	var avatarjson, agentName string
	var timeout int
	err := DB.QueryRow(query, agentID).Scan(
		&org.Name,
		&org.Description,
		&org.Logo,
		&org.Owner,
		&avatarjson,
		&agentName,
		&timeout,
	)

	if err == sql.ErrNoRows {
		log.Printf("No organization found for agent ID: %s", agentID)
		return fmt.Errorf("organization not found for agent ID: %s", agentID)
	} else if err != nil {
		log.Printf("Error retrieving Organization: %v", err)
		return fmt.Errorf("failed to retrieve organization: %v", err)
	}

	if org.Logo != "" {
		bucket := configs.GetEnv("AWS_BUCKET")
		region := configs.GetEnv("AWS_REGION")
		presign, presignDownload, err := PreSignURL(bucket, org.Logo, region)
		if err != nil {
			log.Printf("Error presigning logo URL: %v", err)
			log.Printf("Error presigning logo URL: %v", presignDownload)
		}
		org.PreSignLogo = presign
	}

	var avatars []Agent

	// Unmarshal JSON string into slice of Avatar
	err = json.Unmarshal([]byte(avatarjson), &avatars)
	if err != nil {
		return fmt.Errorf("failed to unmarshal avatars array: %v", err)
	}
	for _, avatar := range avatars {
		var imageURL sql.NullString
		var avatarName sql.NullString

		err = DB.QueryRow(`
		SELECT image_url, avatar_name
		FROM avatars
		WHERE avatar_key_id = $1
	`, avatar.AvatarID).Scan(&imageURL, &avatarName)

		if err != nil {
			if err == sql.ErrNoRows {
				// optional: skip instead of failing
				continue
			}
			return fmt.Errorf("avatar config load failed: %w", err)
		}

		if timeout == 0 {
			timeout = avatar.Timeout
		}

		avatarsList = append(avatarsList, AgentAvatars{
			ID:       avatar.AvatarID,
			ImageURL: imageURL.String,
			Name:     avatarName.String,
			Timeout:  timeout,
		})
	}

	avatarType := "random"
	if len(avatars) == 1 {
		avatarType = "single"
	}

	response := map[string]interface{}{
		"org":         org,
		"avatars":     avatarsList,
		"agent_name":  agentName,
		"avatar_type": avatarType,
	}

	return json.NewEncoder(w).Encode(response)
}

// HandleGetAPIKeys retrieves all API keys from the database
func HandleGetAPIKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		WriteBadRequestError(w, "User ID is required")
		return
	}

	query := `
        SELECT id, key_hash, name, description, expire_at, created_at, is_default, workspace_id
        FROM api_keys
        WHERE user_id = $1
        AND is_active = true
		AND workspace_id IS NOT NULL
        ORDER BY created_at DESC;
    `
	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error querying API keys: %v", err)
		WriteInternalServerError(w, "Failed to retrieve API keys")
		return
	}
	defer rows.Close()

	var apiKeys []map[string]interface{}
	for rows.Next() {
		var id, keyHash, name, description, workspace_id string
		var createdAt, expireAt time.Time
		var isDefault bool

		if err := rows.Scan(&id, &keyHash, &name, &description, &expireAt, &createdAt, &isDefault, &workspace_id); err != nil {
			log.Printf("Error scanning API key row: %v", err)
			WriteInternalServerError(w, "Error scanning API keys")
			return
		}

		apiKeys = append(apiKeys, map[string]interface{}{
			"id":           id,
			"key_hash":     keyHash,
			"name":         name,
			"description":  description,
			"is_default":   isDefault,
			"workspace_id": workspace_id,
			"expire_at":    expireAt.Format("2006-01-02 15:04:05"),
			"created":      createdAt.Format("2006-01-02 15:04:05"),
		})
	}

	json.NewEncoder(w).Encode(apiKeys)
}

func GetPaymentFailureStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-API-Key", http.StatusUnauthorized)
		return
	}

	const query = `
		SELECT
			cl.payment_failed,
			COALESCE(cl.error_message, '')
		FROM api_keys ak
		JOIN workspaces w ON w.id = ak.workspace_id
		JOIN credit_limits cl ON cl.organization_id = w.organization_id
		WHERE ak.key_hash = $1
		LIMIT 1;
	`

	var resp PaymentStatusResponse
	var errorMsg string

	err := DB.QueryRow(query, apiKey).Scan(
		&resp.HasPaymentFailed,
		&errorMsg,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println("DB error:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if errorMsg != "" {
		resp.Error = errorMsg
	} else {
		resp.Error = ""
	}

	json.NewEncoder(w).Encode(resp)
}

// HandleCreateAPIKey inserts a new API key into the database
func HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var requestBody struct {
		Name        string `json:"name"`
		UserID      string `json:"user_id"`
		WorkspaceID string `json:"workspace_id"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %v", err)

		WriteBadRequestError(w, "Invalid request body")
		return
	}

	if requestBody.IsDefault {

		query_update := `
        UPDATE api_keys
        SET	is_default = false
        WHERE workspace_id = $2 and user_id = $1`

		DB.QueryRow(
			query_update,
			requestBody.UserID,
			requestBody.WorkspaceID,
		)
	}

	apiKeyId := strings.ReplaceAll(uuid.New().String(), "-", "")

	id := uuid.New().String()

	query := `
        INSERT INTO api_keys (
            id, user_id, key_hash, name, is_active, description, expire_at,
            created_at, updated_at, is_default, workspace_id
        ) VALUES (
            $1, $2, $3, $4, true, $5, CURRENT_TIMESTAMP + INTERVAL '30 years',
            CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6, $7
        ) RETURNING created_at`

	var createdAt time.Time
	err := DB.QueryRow(
		query,
		id,
		requestBody.UserID,
		apiKeyId,
		requestBody.Name,
		requestBody.Description,
		requestBody.IsDefault,
		requestBody.WorkspaceID,
	).Scan(&createdAt)

	if err != nil {
		log.Printf("Error creating API key: %v", err)
		WriteInternalServerError(w, "Failed to create API key")
		return
	}

	response := map[string]interface{}{
		"id":           id,
		"name":         requestBody.Name,
		"key_hash":     apiKeyId,
		"created":      createdAt,
		"is_active":    true,
		"is_default":   requestBody.IsDefault,
		"description":  requestBody.Description,
		"workspace_id": requestBody.WorkspaceID,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleUpdateAPIKey updates an existing API Key
func HandleUpdateAPIKey(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating API Key with ID: %s", apiKeyId)

	var requestBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
		UserID      string `json:"user_id"`
		WorkspaceID string `json:"workspace_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	if requestBody.IsDefault {

		query_update := `
        UPDATE api_keys
        SET	is_default = false
        WHERE workspace_id = $2 and user_id = $1`

		DB.QueryRow(
			query_update,
			requestBody.UserID,
			requestBody.WorkspaceID,
		)
	}

	query := `
        UPDATE api_keys
        SET name = $1,
            description = $2,
            updated_at = CURRENT_TIMESTAMP,
			is_default = $3,
			workspace_id = $5
        WHERE id = $4
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		requestBody.Name,
		requestBody.Description,
		requestBody.IsDefault,
		apiKeyId,
		requestBody.WorkspaceID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No API key found with ID: %s", apiKeyId)
		WriteNotFoundError(w, "API Key not found")
		return
	} else if err != nil {
		log.Printf("Error updating API Key config: %v", err)
		WriteInternalServerError(w, "Failed to update API Key")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "API Key updated successfully",
	})
}

// HandleDeleteAPIKey deletes an API key from the database
func HandleDeleteAPIKey(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting API Key with ID: %s", apiKeyId)

	query := `
        UPDATE api_keys
        SET is_active = false,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		apiKeyId,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No API key found with ID: %s", apiKeyId)
		WriteNotFoundError(w, "API Key not found")
		return
	} else if err != nil {
		log.Printf("Error deleting API Key config: %v", err)
		WriteInternalServerError(w, "Failed to delete API Key")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "API Key deleted successfully",
	})
}

// HandleCreateInbox create a new Inbox
func HandleCreateInbox(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var requestBody struct {
		UserName    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %v", err)

		WriteBadRequestError(w, "Invalid request body")
		return
	}

	url := configs.GetEnv("INBOX_URL")
	domain := configs.GetEnv("INBOX_DOMAIN")
	token := configs.GetEnv("INBOX_TOKEN")

	result := map[string]interface{}{
		"username":     requestBody.UserName,
		"display_name": requestBody.DisplayName,
		"domain":       domain,
	}

	responseBody, err := infra.PostJob(url, result, token)
	if err != nil {
		fmt.Printf("Inbox Create API Error: %v\n", err)
		WriteInternalServerError(w, err.Error())
		return
	}
	fmt.Println("Inbox Create API Success. Response:", responseBody)

	json.NewEncoder(w).Encode(responseBody)
}

// HandleDeleteInbox deletes an inbox
func HandleDeleteInbox(w http.ResponseWriter, inboxId string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	url := configs.GetEnv("INBOX_URL")
	token := configs.GetEnv("INBOX_TOKEN")

	responseBody, err := infra.DeleteJob(url+"/"+inboxId, token)
	if err != nil {
		fmt.Printf("Inbox Delete API Error: %v\n", err)
		WriteInternalServerError(w, err.Error())
		return
	}
	fmt.Println("Inbox Delete API Success. Response:", responseBody)

	json.NewEncoder(w).Encode(responseBody)
}

// HandleGetAvatarConfig retrieves a specific avatar by ID
func HandleGetAvatarConfig(w http.ResponseWriter, avatarConfigID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching avatar config with ID: %s", avatarConfigID)

	query := `
        SELECT id, avatar_key_id, avatar_name, status, display_picture, image_url, gender, created_at, updated_at, meta_data
        FROM avatars
        WHERE id = $1`

	var (
		id, avatarKey, name, status, displayPicture, ImageURL, gender, metaData string
		createdAt, updatedAt                                                    time.Time
	)

	err := DB.QueryRow(query, avatarConfigID).Scan(
		&id, &avatarKey, &name, &status, &displayPicture, &ImageURL, &gender, &createdAt, &updatedAt, &metaData,
	)

	if err == sql.ErrNoRows {
		log.Printf("No avatar found with ID: %s", avatarConfigID)
		WriteNotFoundError(w, "Avatar not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving avatar config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve avatar")
		return
	}

	avatarConfig := map[string]interface{}{
		"id":              id,
		"avatar_key_id":   avatarKey,
		"avatar_name":     name,
		"status":          status,
		"display_picture": displayPicture,
		"image_url":       ImageURL,
		"gender":          gender,
		"meta_data":       json.RawMessage(metaData),
		"created_at":      createdAt,
		"updated_at":      updatedAt,
	}

	json.NewEncoder(w).Encode(avatarConfig)
}

// HandleGetAllAvatarConfigs retrieves all avatars
func HandleGetAllAvatarConfigs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	avatarConfig, err := GetAllAvatarConfigs()
	if err != nil {
		if err.Error() == "not found" {
			WriteNotFoundError(w, "Avatar not found")
			return
		}
		log.Printf("Error retrieving avatar config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve avatar")
		return
	}

	var avatarList []AvatarsList
	for _, a := range avatarConfig {
		avatarList = append(avatarList, AvatarsList{
			ID:             a.ID,
			AvatarKeyID:    a.AvatarKeyID,
			AvatarName:     a.AvatarName,
			DisplayPicture: a.ImageURL,
			ImageURL:       a.ImageURL,
			Gender:         a.Gender,
			DefaultPrompt:  a.DefaultPrompt,
			Status:         a.Status,
			MetaData:       string(a.MetaData),
			CreatedAt:      a.CreatedAt,
			UpdatedAt:      a.UpdatedAt,
		})
	}

	json.NewEncoder(w).Encode(avatarList)
}

// GetAllAvatarConfigs retrieves all avatars
func GetAllAvatarConfigs() ([]Avatars, error) {
	query := `
        SELECT id, avatar_key_id, avatar_name, status, display_picture, image_url, gender, created_at, updated_at, meta_data
        FROM avatars
        ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()

	var avatarConfigs []Avatars
	for rows.Next() {
		var (
			id, avatarKey, name, status, displayPicture, ImageURL, gender string
			createdAt, updatedAt                                          time.Time
			metaData                                                      json.RawMessage
		)

		if err := rows.Scan(
			&id, &avatarKey, &name, &status, &displayPicture, &ImageURL, &gender, &createdAt, &updatedAt, &metaData,
		); err != nil {
			return nil, fmt.Errorf("error scan avatar data: %w", err)
		}

		avatarConfigs = append(avatarConfigs, Avatars{
			ID:             id,
			AvatarKeyID:    avatarKey,
			AvatarName:     name,
			Status:         status,
			DisplayPicture: displayPicture,
			ImageURL:       ImageURL,
			Gender:         gender,
			MetaData:       metaData,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}

	return avatarConfigs, nil
}

// HandleGetAllAvatarsConfigs retrieves all avatars without metadata
func HandleGetAllAvatarsConfigs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	avatarCons, shouldReturn := GetAllAvatarsFunction(w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(avatarCons)
}

// HandleGetAllAvatarsConfigs retrieves all avatars without metadata
func HandleGetAllAvatarsConfigsAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	avatarCons, shouldReturn := GetAllAvatarsFunction(w)
	if shouldReturn {
		return
	}

	var avatarConfigsAPI []map[string]interface{}

	for _, avatar := range avatarCons {

		AgentConfigAPI := map[string]interface{}{
			"avatar_key_id":   avatar.AvatarKeyID,
			"avatar_name":     avatar.AvatarName,
			"gender":          avatar.Gender,
			"display_picture": avatar.DisplayPicture,
			"image_url":       avatar.ImageURL,
		}

		avatarConfigsAPI = append(avatarConfigsAPI, AgentConfigAPI)
	}

	json.NewEncoder(w).Encode(avatarConfigsAPI)
}

func GetAllAvatarsFunction(w http.ResponseWriter) ([]AvatarCon, bool) {
	avatarConfig, err := GetAllAvatarConfigs()
	if err != nil {
		if err.Error() == "not found" {
			WriteNotFoundError(w, "Avatar not found")
			return nil, true
		}
		log.Printf("Error retrieving avatar config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve avatar")
		return nil, true
	}

	var avatarCons []AvatarCon
	for _, a := range avatarConfig {
		avatarCons = append(avatarCons, AvatarCon{
			ID:             a.ID,
			AvatarKeyID:    a.AvatarKeyID,
			AvatarName:     a.AvatarName,
			DisplayPicture: a.ImageURL,
			ImageURL:       a.ImageURL,
			Gender:         a.Gender,
			DefaultPrompt:  a.DefaultPrompt,
		})
	}
	return avatarCons, false
}

// HandleCreateUser creates a new user and assigns workspace membership
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user struct {
		ID                string `json:"id"`
		FirstName         string `json:"first_name"`
		LastName          string `json:"last_name"`
		Email             string `json:"email"`
		Company           string `json:"company"`
		Password          string `json:"password_hash"`
		SubscribeToMailer bool   `json:"subscribe_to_mailer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	clientIP := getUserCreateClientIP(r)

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Step 1: Create user
	var userID string
	createUserQuery := `
		INSERT INTO users (
			id, email, first_name, last_name, status, password_hash,
			created_at, updated_at, subscribe_to_mailer, client_ip
		) VALUES (
			$1, $2, $3, $4, 'Active', $5,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6, $7
		) RETURNING id`

	err = tx.QueryRow(
		createUserQuery,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Password,
		user.SubscribeToMailer,
		clientIP,
	).Scan(&userID)

	if err != nil {
		// Handle unique constraint violations (user already exists) gracefully
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("User already exists (constraint: %s), returning success", pqErr.Constraint)
			tx.Rollback()
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "user already exists", "userId": user.ID})
			return
		}
		log.Printf("Error creating user: %v", err)
		WriteInternalServerError(w, "Failed to create user")
		return
	}

	// Step 2: Check if user has an invitation
	var invitation struct {
		ID          string
		WorkspaceID string
		RoleID      string
		Status      string
		CreatedBy   uuid.UUID
	}
	checkInviteQuery := `
		SELECT id, workspace_id, role_id, status, created_by
		FROM invitations
		WHERE user_email = $1
		LIMIT 1`
	err = tx.QueryRow(checkInviteQuery, user.Email).Scan(&invitation.ID, &invitation.WorkspaceID, &invitation.RoleID, &invitation.Status, &invitation.CreatedBy)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking invitations: %v", err)
		WriteInternalServerError(w, "Failed to check invitations")
		return
	}

	if err == nil {
		// Step 3a: Add user to workspace_members based on invitation
		insertMemberQuery := `
			INSERT INTO workspace_members (
				id, user_id, role_id, workspace_id, status,
				created_at, created_by
			) VALUES (
				gen_random_uuid(), $1, $2, $3, 'Active',
				CURRENT_TIMESTAMP, $4
			)`
		_, err = tx.Exec(insertMemberQuery, userID, invitation.RoleID, invitation.WorkspaceID, invitation.CreatedBy)
		if err != nil {
			log.Printf("Error adding user to workspace_members: %v", err)
			WriteInternalServerError(w, "Failed to add user to workspace")
			return
		}

		// Step 3b: Update invitation status to Accepted
		_, err = tx.Exec(`UPDATE invitations SET status = 'Accepted' WHERE id = $1`, invitation.ID)
		if err != nil {
			log.Printf("Error updating invitation status: %v", err)
			WriteInternalServerError(w, "Failed to update invitation")
			return
		}
	} else {
		// Step 4a: No invitation → Create new Organization
		var organizationID string
		createOrgQuery := `
			INSERT INTO organizations (
				id, name, description, status, owner,
				created_at, updated_at
			) VALUES (
				gen_random_uuid(), $1, $2, 'Active', $3,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			) RETURNING id`
		err = tx.QueryRow(createOrgQuery,
			user.Company,
			"Default Organization for new user",
			user.Email,
		).Scan(&organizationID)
		if err != nil {
			log.Printf("Error creating Organization: %v", err)
			WriteInternalServerError(w, "Failed to create Organization")
			return
		}

		// Step 4b: No invitation → Create credit limit
		insertCreditLimitQuery := `
			INSERT INTO credit_limits (
				id, organization_id
			) VALUES (
				gen_random_uuid(), $1
			)`
		_, err = tx.Exec(insertCreditLimitQuery, organizationID)
		if err != nil {
			log.Printf("Error Credit Limit Add: %v", err)
			WriteInternalServerError(w, "Failed to add Credit Limit")
			return
		}

		company := user.Company
		if company == "" {
			company = user.FirstName
		}

		// Step 4c: No invitation → Create new workspace and add as Admin
		var workspaceID string
		createWorkspaceQuery := `
			INSERT INTO workspaces (
				id, name, description, status, owner,
				created_at, updated_at, organization_id
			) VALUES (
				gen_random_uuid(), $1, $2, 'Active', $3,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $4
			) RETURNING id`
		err = tx.QueryRow(createWorkspaceQuery,
			company,
			"Default Project for new user",
			userID,
			organizationID,
		).Scan(&workspaceID)
		if err != nil {
			log.Printf("Error creating workspace: %v", err)
			WriteInternalServerError(w, "Failed to create workspace")
			return
		}

		// Get Admin role ID
		var adminRoleID string
		err = tx.QueryRow(`SELECT id FROM roles WHERE LOWER(name) = 'admin' LIMIT 1`).Scan(&adminRoleID)
		if err == sql.ErrNoRows {
			WriteBadRequestError(w, "Admin role not found. Please create an 'Admin' role first.")
			return
		} else if err != nil {
			log.Printf("Error fetching Admin role: %v", err)
			WriteInternalServerError(w, "Failed to fetch Admin role")
			return
		}

		// Add user as Admin to workspace_members
		insertAdminQuery := `
			INSERT INTO workspace_members (
				id, user_id, role_id, workspace_id, status,
				created_at, created_by
			) VALUES (
				gen_random_uuid(), $1, $2, $3, 'Active',
				CURRENT_TIMESTAMP, $4
			)`
		_, err = tx.Exec(insertAdminQuery, userID, adminRoleID, workspaceID, userID)
		if err != nil {
			log.Printf("Error assigning user as Admin: %v", err)
			WriteInternalServerError(w, "Failed to assign workspace owner")
			return
		}
	}

	// Step 5: Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// Step 6: Return success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      userID,
		"message": "User created successfully",
	})
}

// HandleUpdateUser updates user details and uploads profile image to S3 if provided
func HandleUpdateUser(w http.ResponseWriter, r *http.Request, userId string) {
	w.Header().Set("Content-Type", "application/json")

	var user struct {
		FirstName         string `json:"first_name"`
		LastName          string `json:"last_name"`
		Email             string `json:"email"`
		Password          string `json:"password_hash"`
		State             string `json:"state"`
		Country           string `json:"country"`
		Address           string `json:"address"`
		TwoFactorEnabled  bool   `json:"two_factor_enabled"`
		ProfileImage      string `json:"profile_Image"`
		DOB               string `json:"dob"`
		Phone             string `json:"phone"`
		SubscribeToMailer bool   `json:"subscribe_to_mailer"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	var uploadedImageURL string

	// 🔹 Upload image if provided
	if user.ProfileImage != "" && strings.HasPrefix(user.ProfileImage, "data:image/") {
		url, err := UploadBase64ToS3(user.ProfileImage, userId, bucket, region)
		if err != nil {
			log.Printf("Error uploading image to S3: %v", err)
			WriteInternalServerError(w, "Failed to upload profile image")
			return
		}
		uploadedImageURL = url
	}

	// Update user record
	query := `
		UPDATE users
		SET
			first_name = COALESCE(NULLIF($1, ''), first_name),
			last_name = COALESCE(NULLIF($2, ''), last_name),
			email = COALESCE(NULLIF($3, ''), email),
			password_hash = COALESCE(NULLIF($4, ''), password_hash),
			state = COALESCE(NULLIF($5, ''), state),
			country = COALESCE(NULLIF($6, ''), country),
			address = COALESCE(NULLIF($7, ''), address),
			two_factor_enabled = $8,
			profile_image = COALESCE(NULLIF($9, ''), profile_image),
			updated_at = NOW(),
			dob = COALESCE(NULLIF($11, ''), dob),
			phone = COALESCE(NULLIF($12, ''), phone),
			subscribe_to_mailer = COALESCE($13, subscribe_to_mailer)
		WHERE id = $10
		RETURNING id
	`

	var id string
	err := DB.QueryRow(query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.State,
		user.Country,
		user.Address,
		user.TwoFactorEnabled,
		uploadedImageURL, // use the S3 URL if uploaded
		userId,
		user.DOB,
		user.Phone,
		user.SubscribeToMailer,
	).Scan(&id)

	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "User not found")
		return
	} else if err != nil {
		log.Printf("Error updating user: %v", err)
		WriteInternalServerError(w, "Failed to update user")
		return
	}

	response := map[string]interface{}{
		"id":            id,
		"profile_image": uploadedImageURL,
		"message":       "User updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// HandleUnSubscribeUser unsubscribes a user from the mailing list
func HandleUnSubscribeUser(w http.ResponseWriter, r *http.Request, userId string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		UPDATE users
		SET subscribe_to_mailer = false
		WHERE id = $1
		RETURNING id
	`

	var id string
	err := DB.QueryRow(query, userId).Scan(&id)

	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "User not found")
		return
	} else if err != nil {
		log.Printf("Error unsubscribing user: %v", err)
		WriteInternalServerError(w, "Failed to unsubscribe user")
		return
	}

	response := map[string]interface{}{
		"id":      id,
		"message": "User unsubscribed successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// HandleGetUser retrieves a specific user by ID along with organizations and workspaces they have access to
func HandleGetUser(w http.ResponseWriter, userId string) {
	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching user with ID: %s", userId)

	// --- Fetch base user info ---
	userQuery := `
        SELECT
		u.id,
		u.email,
		u.first_name,
		u.last_name,
		COALESCE(o.name, '') AS company,
		COALESCE(u.state, '') AS state,
		COALESCE(u.country, '') AS country,
		COALESCE(u.address, '') AS address,
		u.two_factor_enabled,
		COALESCE(u.profile_image, '') AS profile_image,
		u.status,
		u.created_at,
		u.updated_at,
		COALESCE(u.dob, '') AS dob,
		COALESCE(u.phone, '') AS phone,
		u.subscribe_to_mailer
	FROM users u
	LEFT JOIN organizations o
		ON o.owner = u.email
	WHERE u.id = $1
	AND u.status = 'Active';`

	var (
		id, email, firstName, lastName, company, state, country, address, profileImageURL, status, dob, phone string
		twoFactorEnabled, subscribeToMailer                                                                   bool
		createdAt, updatedAt                                                                                  time.Time
	)

	err := DB.QueryRow(userQuery, userId).Scan(
		&id, &email, &firstName, &lastName,
		&company, &state, &country, &address,
		&twoFactorEnabled, &profileImageURL,
		&status, &createdAt, &updatedAt,
		&dob, &phone, &subscribeToMailer,
	)

	if err == sql.ErrNoRows {
		log.Printf("No user found with ID: %s", userId)
		WriteNotFoundError(w, "User not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving user details: %v", err)
		WriteInternalServerError(w, "Failed to retrieve user details")
		return
	}

	// --- Fetch organizations and workspaces user has access to ---
	accessQuery := `
		SELECT
			o.id AS organization_id,
			o.name AS organization_name,
			w.id AS workspace_id,
			w.name AS workspace_name,
			w.description AS workspace_description
		FROM workspace_members wm
		JOIN workspaces w ON wm.workspace_id = w.id
		JOIN organizations o ON w.organization_id = o.id
		WHERE wm.user_id = $1
		ORDER BY o.name, w.name`

	rows, err := DB.Query(accessQuery, userId)
	if err != nil {
		log.Printf("Error retrieving user's organization/workspace access: %v", err)
		WriteInternalServerError(w, "Failed to retrieve user's access")
		return
	}
	defer rows.Close()

	orgMap := make(map[string]map[string]interface{})
	for rows.Next() {
		var orgID, orgName, workspaceID, workspaceName, workspaceDesc string

		if err := rows.Scan(&orgID, &orgName, &workspaceID, &workspaceName, &workspaceDesc); err != nil {
			log.Printf("Error scanning user access row: %v", err)
			WriteInternalServerError(w, "Error processing user's access data")
			return
		}

		org, exists := orgMap[orgID]
		if !exists {
			org = map[string]interface{}{
				"id":         orgID,
				"name":       orgName,
				"workspaces": []map[string]interface{}{},
			}
			orgMap[orgID] = org
		}

		org["workspaces"] = append(org["workspaces"].([]map[string]interface{}), map[string]interface{}{
			"id":          workspaceID,
			"name":        workspaceName,
			"description": workspaceDesc,
		})
	}

	var organizations []map[string]interface{}
	for _, org := range orgMap {
		organizations = append(organizations, org)
	}

	presign := ""
	presignDownload := ""
	if profileImageURL != "" {
		presign, presignDownload, err = PreSignURL(bucket, profileImageURL, region)
		if err != nil {
			log.Printf("Error on pre sign URL: %v", err)
			WriteBadRequestError(w, "Error on pre sign URL")
			return
		}
	}

	// --- Build final user object ---
	userData := map[string]interface{}{
		"id":                     id,
		"email":                  email,
		"first_name":             firstName,
		"last_name":              lastName,
		"company":                company,
		"state":                  state,
		"country":                country,
		"address":                address,
		"two_factor_enabled":     twoFactorEnabled,
		"profile_image":          presign,
		"profile_image_download": presignDownload,
		"created_at":             createdAt,
		"updated_at":             updatedAt,
		"organizations":          organizations,
		"dob":                    dob,
		"phone":                  phone,
		"subscribe_to_mailer":    subscribeToMailer,
	}

	json.NewEncoder(w).Encode(userData)
}

// HandleCreateSupportTicket handles creating a support ticket with optional attachments
func HandleCreateSupportTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form data (allowing text + files)
	err := r.ParseMultipartForm(20 << 20) // 20 MB max
	if err != nil {
		WriteBadRequestError(w, "Invalid form data or file too large")
		return
	}

	// Extract ticket fields from form
	supportType := r.FormValue("supportType")
	userId := r.FormValue("userId")
	subject := r.FormValue("subject")
	description := r.FormValue("description")

	if supportType == "" || subject == "" || description == "" {
		WriteBadRequestError(w, "supportType, subject, and description are required")
		return
	}

	// --- 1️⃣ Insert Support Ticket ---
	var ticketID string
	insertQuery := `
		INSERT INTO support_tickets (user_id, support_type, subject, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'Open', NOW(), NOW())
		RETURNING id
	`
	err = DB.QueryRow(insertQuery, userId, supportType, subject, description).Scan(&ticketID)
	if err != nil {
		log.Printf("Error creating support ticket: %v", err)
		WriteInternalServerError(w, "Failed to create support ticket")
		return
	}

	attachments := []map[string]interface{}{}
	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	// --- 2️⃣ Handle file attachments (if any) ---
	files := r.MultipartForm.File["attachments"] // Input field name = attachments[]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			continue
		}
		defer file.Close()

		// Upload file to S3
		fileURL, uploadErr := UploadFileToS3(fileHeader.Filename, bucket, region, file)
		if uploadErr != nil {
			log.Printf("Error uploading to S3: %v", uploadErr)
			continue
		}

		// Save attachment record
		var attachID string
		attachQuery := `
			INSERT INTO support_ticket_attachments (ticket_id, file_name, file_url, created_at)
			VALUES ($1, $2, $3, NOW())
			RETURNING id
		`
		err = DB.QueryRow(attachQuery, ticketID, fileHeader.Filename, fileURL).Scan(&attachID)
		if err != nil {
			log.Printf("Error inserting attachment: %v", err)
			continue
		}

		attachments = append(attachments, map[string]interface{}{
			"id":        attachID,
			"file_name": fileHeader.Filename,
			"file_url":  fileURL,
		})
	}

	// --- 3️⃣ Return Response ---
	resp := map[string]interface{}{
		"id":           ticketID,
		"message":      "Support ticket created successfully",
		"attachments":  attachments,
		"support_type": supportType,
		"subject":      subject,
		"description":  description,
	}

	json.NewEncoder(w).Encode(resp)
}

// HandleGetAllSupportTickets retrieves all support tickets (with attachments)
func HandleGetAllSupportTickets(w http.ResponseWriter, r *http.Request, userId string) {
	w.Header().Set("Content-Type", "application/json")
	query := `
		SELECT
			st.id,
			st.support_type,
			st.subject,
			st.description,
			st.status,
			st.created_at,
			st.updated_at
		FROM support_tickets st
		WHERE st.user_id = $1
		ORDER BY st.created_at DESC
	`

	rows, err := DB.Query(query, userId)
	if err != nil {
		log.Printf("Error fetching support tickets: %v", err)
		WriteInternalServerError(w, "Failed to fetch support tickets")
		return
	}
	defer rows.Close()

	type Ticket struct {
		ID          string                   `json:"id"`
		SupportType string                   `json:"support_type"`
		Subject     string                   `json:"subject"`
		Description string                   `json:"description"`
		Status      string                   `json:"status"`
		CreatedAt   time.Time                `json:"created_at"`
		UpdatedAt   time.Time                `json:"updated_at"`
		Attachments []map[string]interface{} `json:"attachments"`
	}

	tickets := []Ticket{}

	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.SupportType, &t.Subject, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			log.Printf("Error scanning ticket row: %v", err)
			WriteInternalServerError(w, "Error processing support tickets")
			return
		}

		attachQuery := `
			SELECT id, file_name, file_url, created_at
			FROM support_ticket_attachments
			WHERE ticket_id = $1
			ORDER BY created_at
		`

		attachRows, err := DB.Query(attachQuery, t.ID)
		if err != nil {
			log.Printf("Error fetching attachments for ticket %s: %v", t.ID, err)
			WriteInternalServerError(w, "Failed to fetch attachments")
			return
		}

		for attachRows.Next() {
			var (
				aID, aName, aURL string
				aCreatedAt       time.Time
			)
			if err := attachRows.Scan(&aID, &aName, &aURL, &aCreatedAt); err != nil {
				log.Printf("Error scanning attachment row: %v", err)
				continue
			}

			bucket := configs.GetEnv("AWS_BUCKET")
			region := configs.GetEnv("AWS_REGION")
			presign, presignDownload, err := PreSignURL(bucket, aURL, region)

			if err != nil {
				log.Printf("Error on pre sign URL: %v", err)
				WriteBadRequestError(w, "Error on pre sign URL")
				return
			}

			t.Attachments = append(t.Attachments, map[string]interface{}{
				"id":                aID,
				"file_name":         aName,
				"file_url":          presign,
				"file_download_url": presignDownload,
				"created_at":        aCreatedAt,
			})
		}
		attachRows.Close()

		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating ticket rows: %v", err)
		WriteInternalServerError(w, "Error reading tickets")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"tickets": tickets,
	})
}

// HandleGetUserFavourite retrieves a specific User by ID
func HandleGetUserFavourite(w http.ResponseWriter, userId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching users favourite with ID: %s", userId)

	query := `
        SELECT id, email, COALESCE(favourite_avatars, '{}') AS favourite_avatars
        FROM users
        WHERE id = $1 and status = 'Active'`

	var (
		id, email         string
		favourite_avatars []string
	)

	err := DB.QueryRow(query, userId).Scan(
		&id, &email, pq.Array(&favourite_avatars),
	)

	if err == sql.ErrNoRows {
		log.Printf("No user found with ID: %s", userId)
		WriteNotFoundError(w, "User not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving user favourite details: %v", err)
		WriteInternalServerError(w, "Failed to retrieve user favourite details")
		return
	}

	AgentConfig := map[string]interface{}{
		"id":                id,
		"email":             email,
		"favourite_avatars": favourite_avatars,
	}

	json.NewEncoder(w).Encode(AgentConfig)
}

// HandleUpdateUserFavourite updates the favourite of the user
func HandleUpdateUserFavourite(w http.ResponseWriter, r *http.Request, userId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Users Favourite with ID: %s", userId)

	var favouriteConfig struct {
		Avatars []string `json:"favourite_avatars" db:"favourite_avatars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&favouriteConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE users
        SET favourite_avatars = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		pq.Array(favouriteConfig.Avatars),
		userId,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No User found with ID: %s", userId)
		WriteNotFoundError(w, "User not found")
		return
	} else if err != nil {
		log.Printf("Error updating Users Favourite: %v", err)
		WriteInternalServerError(w, "Failed to update Users Favourite")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Users Favourite updated successfully",
	})
}

// Extract text from a DOCX using gooxml by iterating runs.
func loadDocx(r io.Reader, filename, ext string) ([]schema.Document, error) {
	tmp, err := os.CreateTemp("", "upload-*.docx")
	if err != nil {
		return nil, err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if _, err := io.Copy(tmp, r); err != nil {
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}

	doc, err := document.Open(tmp.Name())
	if err != nil {
		return nil, err
	}

	var b strings.Builder

	// Body paragraphs
	for _, p := range doc.Paragraphs() {
		for _, run := range p.Runs() {
			b.WriteString(run.Text())
		}
		b.WriteByte('\n')
	}

	// Tables (rows -> cells -> paragraphs -> runs)
	for _, tbl := range doc.Tables() {
		for _, row := range tbl.Rows() {
			for _, cell := range row.Cells() {
				for _, p := range cell.Paragraphs() {
					for _, run := range p.Runs() {
						b.WriteString(run.Text())
					}
					b.WriteByte('\n')
				}
			}
		}
	}

	text := strings.TrimSpace(b.String())
	return []schema.Document{
		{
			PageContent: text,
			Metadata: map[string]interface{}{
				"filename": filename,
				"filetype": ext,
			},
		},
	}, nil
}

// HandleCreateKBConfig creates a new Knowledge base
func HandleCreateKBConfig(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	// Parse up to 10 MB of uploaded files
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	var kbInput KBInput
	kbInput.Name = r.FormValue("name")
	kbInput.Description = r.FormValue("description")
	kbInput.IsActive = true
	namespace := uuid.New()
	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	// Collect files from "input" field
	files := r.MultipartForm.File["input"]
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error opening file: %v", err))
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(fh.Filename))

		// --- Upload original file to S3 ---
		key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace.String(), fh.Filename)
		url, err := SaveFileStreamToS3(file, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error uploading file to S3: %v", err))
			return
		}

		// Reopen file reader for parsing (because it’s consumed once)
		file2, err := fh.Open()
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error reopening file: %v", err))
			return
		}
		defer file2.Close()

		var docs []schema.Document

		switch ext {
		case ".txt":
			docs, err = documentloaders.NewText(file2).Load(r.Context())
		case ".docx":
			docs, err = loadDocx(file2, fh.Filename, ext)
		case ".csv":
			docs, err = documentloaders.NewCSV(file2).Load(r.Context())
		case ".pdf":
			tmpFile, _ := os.CreateTemp("", "upload-*.pdf")
			defer os.Remove(tmpFile.Name())
			io.Copy(tmpFile, file2)
			tmpFile.Close()

			pdfFile, _ := os.Open(tmpFile.Name())
			defer pdfFile.Close()
			docs, err = ReadOCRPDFWithGenAI(r.Context(), pdfFile)
		default:
			content, _ := io.ReadAll(file2)
			docs = []schema.Document{
				{
					PageContent: string(content),
					Metadata: map[string]interface{}{
						"filename": fh.Filename,
						"filetype": ext,
						"url":      url,
					},
				},
			}
		}

		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error loading %s: %v", fh.Filename, err))
			return
		}

		// Add docs to input
		for _, d := range docs {
			kbInput.Input = append(kbInput.Input, KBInputDoc{
				Content:  d.PageContent,
				Filename: fh.Filename,
				Filetype: ext,
				URL:      url, // Store the uploaded file S3 URL
			})
		}
	}

	// --- 2. Manual text entries ---
	texts := r.MultipartForm.Value["text"]
	for _, t := range texts {
		if t != "" {
			// Save manual entry as guid.txt into S3
			filename := fmt.Sprintf("%s.txt", uuid.New().String())
			key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace.String(), filename)
			url, err := SaveStringToS3(t, bucket, key, region)
			if err != nil {
				WriteInternalServerError(w, "Failed to upload manual text to S3")
				return
			}

			kbInput.Input = append(kbInput.Input, KBInputDoc{
				Content:  t,
				Filename: filename,
				Filetype: "text",
				URL:      url,
			})
		}
	}

	// --- 3. URL entries (we don’t upload these, just store) ---
	urls := r.MultipartForm.Value["url"]
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		raw = strings.ReplaceAll(raw, "\n", "")
		raw = strings.ReplaceAll(raw, "\r", "")
		raw = strings.ReplaceAll(raw, "\t", "")
		raw = strings.TrimSpace(raw)

		var payload URLPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			log.Printf("json error: %v", err)
			continue
		}

		b64 := payload.Text
		b64 = strings.ReplaceAll(b64, "\n", "")
		b64 = strings.ReplaceAll(b64, "\r", "")
		b64 = strings.ReplaceAll(b64, " ", "")
		b64 = strings.ReplaceAll(b64, "\t", "")
		b64 = strings.ReplaceAll(b64, "-", "+")
		b64 = strings.ReplaceAll(b64, "_", "/")

		if m := len(b64) % 4; m != 0 {
			b64 += strings.Repeat("=", 4-m)
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Printf("base64 decode failed: %v", err)
			continue
		}

		// Save url data as guid.txt into S3
		filename := fmt.Sprintf("%s.txt", uuid.New().String())
		data := string(decoded)
		key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace.String(), filename)
		url, err := SaveStringToS3(data, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload manual text to S3")
			return
		}

		kbInput.Input = append(kbInput.Input, KBInputDoc{
			Content:  data,
			Filename: filename,
			WebURL:   payload.URL,
			Filetype: "url",
			URL:      url,
		})
	}
	// --- Convert KBInput.Input into []schema.Document ---
	var schemaInputs []schema.Document
	for _, in := range kbInput.Input {
		meta := map[string]any{
			"filename": in.Filename,
			"filetype": in.Filetype,
			"url":      in.URL,
			"web_url":  in.WebURL,
		}
		schemaInputs = append(schemaInputs, schema.Document{
			PageContent: in.Content,
			Metadata:    meta,
		})
	}

	// Call universal loader
	docs, err := langchain.LoadUniversal(schemaInputs...)
	if err != nil {
		WriteInternalServerError(w, "Failed to load documents")
		return
	}

	// Convert docs into Pinecone IntegratedRecords
	records := make([]*pinecone.IntegratedRecord, len(docs))
	for i, d := range docs {
		records[i] = &pinecone.IntegratedRecord{
			"id":       fmt.Sprintf("rec%d", i+1),
			"text":     d.PageContent,
			"filename": d.Metadata["filename"],
			"type":     d.Metadata["filetype"],
			"url":      d.Metadata["url"],
			"web_url":  d.Metadata["web_url"],
		}
	}

	querypk := `
        SELECT id, name, host
        FROM pinecone_indexes
        WHERE is_active = true`

	var (
		pkid, name, host string
	)

	errpk := DB.QueryRow(querypk).Scan(
		&pkid, &name, &host,
	)

	if errpk != nil {
		WriteInternalServerError(w, "Failed to retrive Pinecone index")
		return
	}

	indexed := vectors.CreateIndex(records, namespace.String(), host)
	if !indexed {
		WriteInternalServerError(w, "Failed to create knowledge base vector index")
		return
	}

	// Insert into DB
	query := `
	    INSERT INTO knowledge_base (
	        id, name, description, namespace, index, is_active, created_at, updated_at, created_by, no_of_rec
	    ) VALUES (
	        $7, $1, $2, $3, $4,
	        $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6, $8
	    ) RETURNING id`

	var id string
	err1 := DB.QueryRow(
		query,
		kbInput.Name,
		kbInput.Description,
		namespace.String(),
		host,
		kbInput.IsActive,
		apiKeyId,
		namespace,
		len(records),
	).Scan(&id)

	if err1 != nil {
		WriteInternalServerError(w, fmt.Sprintf("Failed to create knowledge base %s", err))
		return
	}

	unique := make(map[string]DocumentRecord)

	// Deduplicate by filename+filetype+url
	for _, d := range docs {
		name := fmt.Sprintf("%v", d.Metadata["filename"])
		ftype := fmt.Sprintf("%v", d.Metadata["filetype"])
		url := fmt.Sprintf("%v", d.Metadata["url"])
		weburl := fmt.Sprintf("%v", d.Metadata["web_url"])
		key := name + "|" + ftype + "|" + url

		if _, exists := unique[key]; !exists {
			unique[key] = DocumentRecord{
				Name:     name,
				FileType: ftype,
				URL:      url,
				WebURL:   weburl,
			}
		}
	}

	// Insert unique records into documents table
	for _, u := range unique {
		_, err := DB.Exec(`
			INSERT INTO public.documents
			(id, name, preview_url, is_active, created_at, updated_at, created_by, type, kb_id, web_url)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5, $6, $7, $8)
		`,
			uuid.New(),         // id
			u.Name,             // name
			u.URL,              // preview_url
			true,               // is_active
			apiKeyId,           // created_by
			u.FileType,         // type
			namespace.String(), // kb_id
			u.WebURL,
		)
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error saving documents %s", err))
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Knowledge Base created successfully",
	})
}

func ReadOCRPDFWithGenAI(
	ctx context.Context,
	pdfFile *os.File,
) ([]schema.Document, error) {

	// Read PDF bytes
	pdfBytes, err := io.ReadAll(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pdf: %w", err)
	}

	// Init GenAI client
	client, err := genai.NewClient(
		ctx,
		option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
	)
	if err != nil {
		return nil, fmt.Errorf("genai client error: %w", err)
	}
	defer client.Close()

	// Gemini Vision model (OCR capable)
	model := client.GenerativeModel("gemini-3-flash-preview")

	// Prompt + PDF
	resp, err := model.GenerateContent(
		ctx,
		genai.Text("Extract all readable text from this PDF. Preserve paragraphs."),
		genai.Blob{
			MIMEType: "application/pdf",
			Data:     pdfBytes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("genai generation error: %w", err)
	}

	// Collect text
	var fullText string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if t, ok := part.(genai.Text); ok {
				fullText += string(t) + "\n"
			}
		}
	}

	if fullText == "" {
		return nil, fmt.Errorf("no text extracted from pdf")
	}

	// Return LangChain documents
	docs := []schema.Document{
		{
			PageContent: fullText,
			Metadata: map[string]any{
				"source":   "genai-ocr",
				"filetype": "pdf",
				"filename": pdfFile.Name(),
			},
		},
	}

	return docs, nil
}

// HandleAddDocToKBConfig adds files, text, or URLs to an existing Knowledge Base
func HandleAddDocToKBConfig(w http.ResponseWriter, r *http.Request, kbID, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	// Parse up to 10 MB of uploaded files
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	// 1. Fetch KB metadata from DB (namespace + index host)
	var (
		namespace, host string
		noOfRec         int
	)

	query := `
		SELECT namespace, index, no_of_rec
		FROM knowledge_base
		WHERE id = $1 AND is_active = true
	`
	err = DB.QueryRow(query, kbID).Scan(&namespace, &host, &noOfRec)
	if err != nil {
		WriteNotFoundError(w, "Knowledge Base not found or inactive")
		return
	}

	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	var kbInput KBInput

	// --- 1. File uploads ---
	files := r.MultipartForm.File["input"]
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error opening file: %v", err))
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(fh.Filename))
		key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace, fh.Filename)
		url, err := SaveFileStreamToS3(file, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error uploading file to S3: %v", err))
			return
		}

		// Reopen reader for parsing
		file2, _ := fh.Open()
		defer file2.Close()

		var docs []schema.Document
		switch ext {
		case ".txt":
			docs, err = documentloaders.NewText(file2).Load(r.Context())
		case ".docx":
			docs, err = loadDocx(file2, fh.Filename, ext)
		case ".csv":
			docs, err = documentloaders.NewCSV(file2).Load(r.Context())
		case ".pdf":
			tmpFile, _ := os.CreateTemp("", "upload-*.pdf")
			defer os.Remove(tmpFile.Name())
			io.Copy(tmpFile, file2)
			tmpFile.Close()
			pdfFile, _ := os.Open(tmpFile.Name())
			defer pdfFile.Close()
			fi, _ := pdfFile.Stat()
			docs, err = documentloaders.NewPDF(pdfFile, fi.Size()).Load(r.Context())
		default:
			content, _ := io.ReadAll(file2)
			docs = []schema.Document{{
				PageContent: string(content),
				Metadata: map[string]any{
					"filename": fh.Filename,
					"filetype": ext,
					"url":      url,
				},
			}}
		}
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error loading %s: %v", fh.Filename, err))
			return
		}

		for _, d := range docs {
			kbInput.Input = append(kbInput.Input, KBInputDoc{
				Content:  d.PageContent,
				Filename: fh.Filename,
				Filetype: ext,
				URL:      url,
			})
		}
	}

	// --- 2. Manual text entries ---
	for _, t := range r.MultipartForm.Value["text"] {
		if t != "" {
			filename := fmt.Sprintf("%s.txt", uuid.New().String())
			key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace, filename)
			url, err := SaveStringToS3(t, bucket, key, region)
			if err != nil {
				WriteInternalServerError(w, "Failed to upload text to S3")
				return
			}
			kbInput.Input = append(kbInput.Input, KBInputDoc{
				Content:  t,
				Filename: filename,
				Filetype: "text",
				URL:      url,
			})
		}
	}

	// --- 3. URL entries ---
	urls := r.MultipartForm.Value["url"]
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		raw = strings.ReplaceAll(raw, "\n", "")
		raw = strings.ReplaceAll(raw, "\r", "")
		raw = strings.ReplaceAll(raw, "\t", "")
		raw = strings.TrimSpace(raw)

		var payload URLPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			log.Printf("json error: %v", err)
			continue
		}

		b64 := payload.Text
		b64 = strings.ReplaceAll(b64, "\n", "")
		b64 = strings.ReplaceAll(b64, "\r", "")
		b64 = strings.ReplaceAll(b64, " ", "")
		b64 = strings.ReplaceAll(b64, "\t", "")
		b64 = strings.ReplaceAll(b64, "-", "+")
		b64 = strings.ReplaceAll(b64, "_", "/")

		if m := len(b64) % 4; m != 0 {
			b64 += strings.Repeat("=", 4-m)
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Printf("base64 decode failed: %v", err)
			continue
		}

		// Save url data as guid.txt into S3
		filename := fmt.Sprintf("%s.txt", uuid.New().String())
		data := string(decoded)
		key := fmt.Sprintf("clawdface/uploads/kb/%s/%s", namespace, filename)
		url, err := SaveStringToS3(data, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload manual text to S3")
			return
		}

		kbInput.Input = append(kbInput.Input, KBInputDoc{
			Content:  data,
			Filename: filename,
			WebURL:   payload.URL,
			Filetype: "url",
			URL:      url,
		})
	}

	// --- Convert into schema.Documents ---
	var schemaInputs []schema.Document
	for _, in := range kbInput.Input {
		schemaInputs = append(schemaInputs, schema.Document{
			PageContent: in.Content,
			Metadata: map[string]any{
				"filename": in.Filename,
				"filetype": in.Filetype,
				"url":      in.URL,
				"web_url":  in.WebURL,
			},
		})
	}

	docs, err := langchain.LoadUniversal(schemaInputs...)
	if err != nil {
		WriteInternalServerError(w, "Failed to load documents")
		return
	}

	// --- Push to Pinecone ---
	records := make([]*pinecone.IntegratedRecord, len(docs))
	for i, d := range docs {
		records[i] = &pinecone.IntegratedRecord{
			"id":       fmt.Sprintf("rec%d", noOfRec+i+1),
			"text":     d.PageContent,
			"filename": d.Metadata["filename"],
			"type":     d.Metadata["filetype"],
			"url":      d.Metadata["url"],
			"web_url":  d.Metadata["web_url"],
		}
	}

	indexed := vectors.CreateIndex(records, namespace, host)
	if !indexed {
		WriteInternalServerError(w, "Failed to add records to Pinecone")
		return
	}

	// --- Update KB record count ---
	_, err = DB.Exec(`UPDATE knowledge_base SET no_of_rec = no_of_rec + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		len(records), kbID)
	if err != nil {
		WriteInternalServerError(w, "Failed to update KB record count")
		return
	}

	// --- Insert into documents table ---
	unique := make(map[string]DocumentRecord)
	for _, d := range docs {
		name := fmt.Sprintf("%v", d.Metadata["filename"])
		ftype := fmt.Sprintf("%v", d.Metadata["filetype"])
		url := fmt.Sprintf("%v", d.Metadata["url"])
		weburl := fmt.Sprintf("%v", d.Metadata["web_url"])
		key := name + "|" + ftype + "|" + url
		if _, exists := unique[key]; !exists {
			unique[key] = DocumentRecord{Name: name, FileType: ftype, URL: url, WebURL: weburl}
		}
	}
	for _, u := range unique {
		_, err := DB.Exec(`
			INSERT INTO public.documents
			(id, name, preview_url, is_active, created_at, updated_at, created_by, type, kb_id, web_url)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5, $6, $7, $8)
		`,
			uuid.New(), u.Name, u.URL, true, apiKeyId, u.FileType, kbID, u.WebURL,
		)
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error saving documents: %v", err))
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"kb_id":   kbID,
		"added":   len(records),
		"message": "Content added successfully to existing Knowledge Base",
	})
}

// HandleUpdateKBConfig updates an existing Knowledge Base
func HandleUpdateKBConfig(w http.ResponseWriter, r *http.Request, kbID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Knowledge Base config with ID: %s", kbID)

	var kbConfig struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&kbConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE knowledge_base
        SET name = $1,
            description = $2,
            is_active = $3,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		kbConfig.Name,
		kbConfig.Description,
		kbConfig.IsActive,
		kbID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No knowledge base found with ID: %s", kbID)
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		log.Printf("Error updating knowledge base config: %v", err)
		WriteInternalServerError(w, "Failed to update knowledge base")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Knowledge Base updated successfully",
	})
}

// HandleURLContent fetch the URL content
func HandleURLContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var URLContent struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&URLContent); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	res, err := langchain.LoadFromURL(URLContent.URL) //langchain.loadFromURL(URLContent.URL)

	if err != nil {
		log.Printf("Error on fetch URL content: %v", err)
		WriteBadRequestError(w, "Error on fetch URL content")
		return
	}

	// Convert res (likely []schema.Document) into JSON serializable objects
	var docs []map[string]interface{}
	for _, d := range res {
		docs = append(docs, map[string]interface{}{
			"content": d.PageContent,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"URLContent": docs,
	})
}

// HandleURLPreSign method gets the S3 Pre sign URL expire in 15 min
func HandleURLPreSign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var URLPreSign struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&URLPreSign); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")
	res, resDownload, err := PreSignURL(bucket, URLPreSign.URL, region)

	if err != nil {
		log.Printf("Error on pre sign URL: %v", err)
		WriteBadRequestError(w, "Error on pre sign URL")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"presign":          res,
		"presign_download": resDownload,
	})
}

// HandleGetKBConfig retrieves a specific Knowledge Base config and its related documents
func HandleGetKBConfig(w http.ResponseWriter, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching knowledge base and documents for KB ID: %s", kbID)

	// --- Step 1: Fetch Knowledge Base ---
	response, shouldReturn := GetKBFunction(kbID, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(response)
}

// HandleGetKBConfigAPI retrieves a specific Knowledge Base config and its related documents
func HandleGetKBConfigAPI(w http.ResponseWriter, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching knowledge base and documents for KB ID: %s", kbID)

	// --- Step 1: Fetch Knowledge Base ---
	kbs, shouldReturn := GetKBFunction(kbID, w)
	if shouldReturn {
		return
	}

	kb, ok := kbs["knowledge_base"].(map[string]interface{})
	if !ok {
		// handle error properly
		log.Println("invalid knowledge_base structure")
		return
	}

	kbConfigAPI := map[string]interface{}{
		"id":          kb["id"],
		"name":        kb["name"],
		"description": kb["description"],
		"no_of_rec":   kb["no_of_rec"],
		"is_active":   kb["is_active"],
		"created_at":  kb["created_at"],
		"updated_at":  kb["updated_at"],
		"documents":   kbs["documents"],
	}

	json.NewEncoder(w).Encode(kbConfigAPI)
}

func GetKBFunction(kbID string, w http.ResponseWriter) (map[string]interface{}, bool) {
	kbQuery := `
        SELECT id, name, description, namespace, index, no_of_rec, is_active, created_at, updated_at, created_by
        FROM knowledge_base
        WHERE id = $1
    `

	var (
		id, name, description, namespace, index string
		isActive                                bool
		noOfRec                                 int
		createdAt, updatedAt                    time.Time
		createdBy                               string
	)

	err := DB.QueryRow(kbQuery, kbID).Scan(
		&id, &name, &description, &namespace, &index,
		&noOfRec, &isActive, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error fetching knowledge base: %v", err)
		WriteInternalServerError(w, "Failed to fetch knowledge base")
		return nil, true
	}

	// --- Step 2: Fetch Documents for this KB ---
	docQuery := `
        SELECT id, name, preview_url, is_active, type, kb_id, COALESCE(web_url, '') as web_url
        FROM documents
        WHERE kb_id = $1
    `

	rows, err := DB.Query(docQuery, kbID)
	if err != nil {
		log.Printf("Error querying documents: %v", err)
		WriteInternalServerError(w, "Failed to retrieve documents")
		return nil, true
	}
	defer rows.Close()

	var documents []map[string]interface{}

	for rows.Next() {
		var (
			docID, docName, previewURL, docType, kbIDOut, webURL string
			docActive                                            bool
		)

		if err := rows.Scan(&docID, &docName, &previewURL, &docActive, &docType, &kbIDOut, &webURL); err != nil {
			log.Printf("Error scanning document row: %v", err)
			continue
		}

		documents = append(documents, map[string]interface{}{
			"id":          docID,
			"name":        docName,
			"preview_url": previewURL,
			"is_active":   docActive,
			"type":        docType,
			"kb_id":       kbIDOut,
			"web_url":     webURL,
		})
	}

	// --- Step 3: Combine into response ---
	response := map[string]interface{}{
		"knowledge_base": map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"namespace":   namespace,
			"index":       index,
			"no_of_rec":   noOfRec,
			"is_active":   isActive,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		},
		"documents": documents,
	}
	return response, false
}

// HandleGetAllKBConfigs retrieves all knowledge bases with their related documents
func HandleGetAllKBConfigs(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching all KBs with documents for apiKeyId: %s", apiKeyId)

	// --- Step 1: Fetch all KBs ---
	kbList, shouldReturn := GetAllKBFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(kbList)
}

// HandleGetAllKBConfigs retrieves all knowledge bases with their related documents
func HandleGetAllKBConfigsAPI(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching all KBs with documents for apiKeyId: %s", apiKeyId)

	// --- Step 1: Fetch all KBs ---
	kbList, shouldReturn := GetAllKBFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	var kbConfigsAPI []map[string]interface{}

	for _, kb := range kbList {

		kbConfigAPI := map[string]interface{}{
			"id":          kb["id"],
			"name":        kb["name"],
			"description": kb["description"],
			"no_of_rec":   kb["no_of_rec"],
			"is_active":   kb["is_active"],
			"created_at":  kb["created_at"],
			"updated_at":  kb["updated_at"],
			"documents":   kb["documents"],
		}

		kbConfigsAPI = append(kbConfigsAPI, kbConfigAPI)
	}
	json.NewEncoder(w).Encode(kbConfigsAPI)
}

func GetAllKBFunction(apiKeyId string, w http.ResponseWriter) ([]map[string]interface{}, bool) {
	kbQuery := `
        SELECT
			kb.id,
			kb.name,
			kb.description,
			kb.namespace,
			kb.index,
			kb.no_of_rec,
			kb.is_active,
			kb.created_at,
			kb.updated_at,
			kb.created_by
		FROM knowledge_base kb
		JOIN api_keys ak2
			ON kb.created_by = ak2.id          -- creator’s API key
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id   -- same workspace condition
		-- Workspace joins to match organization
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id        -- workspace of requesting key
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id        -- workspace of creator
			AND w2.organization_id = w1.organization_id   -- ✅ same organization
		WHERE ak1.id = $1
		ORDER BY kb.created_at DESC;
    `

	rows, err := DB.Query(kbQuery, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve knowledge bases")
		return nil, true
	}
	defer rows.Close()

	var kbList []map[string]interface{}

	for rows.Next() {
		var (
			id, name, description, namespace, index string
			noOfRec                                 int
			isActive                                bool
			createdAt, updatedAt                    time.Time
			createdBy                               string
		)

		if err := rows.Scan(
			&id, &name, &description, &namespace,
			&index, &noOfRec, &isActive, &createdAt, &updatedAt, &createdBy,
		); err != nil {
			WriteInternalServerError(w, "Error scanning knowledge bases")
			return nil, true
		}

		// --- Step 2: Fetch Documents for this KB ---
		docQuery := `
            SELECT
				d.id,
				d.name,
				d.preview_url,
				d.is_active,
				d.type,
				d.kb_id,
				COALESCE(d.web_url, '') AS web_url
			FROM documents d
			JOIN api_keys ak2
				ON d.created_by = ak2.id                      -- creator's API key
			JOIN api_keys ak1
				ON ak2.workspace_id = ak1.workspace_id        -- workspace match
			-- Workspace joins for organization filtering
			JOIN workspaces w1
				ON ak1.workspace_id = w1.id                   -- workspace of requesting key
			JOIN workspaces w2
				ON ak2.workspace_id = w2.id                   -- workspace of creator
				AND w2.organization_id = w1.organization_id   -- ✅ organization match
			WHERE ak1.id = $2
			AND d.kb_id = $1
			ORDER BY d.created_at DESC;
        `
		docRows, err := DB.Query(docQuery, id, apiKeyId)
		if err != nil {
			log.Printf("Error querying documents for kb_id %s: %v", id, err)
			continue
		}

		var documents []map[string]interface{}
		for docRows.Next() {
			var (
				docID, docName, previewURL, docType, kbIDOut, webURL string
				docActive                                            bool
			)

			if err := docRows.Scan(&docID, &docName, &previewURL, &docActive, &docType, &kbIDOut, &webURL); err != nil {
				log.Printf("Error scanning document row for kb_id %s: %v", id, err)
				continue
			}

			documents = append(documents, map[string]interface{}{
				"id":          docID,
				"name":        docName,
				"preview_url": previewURL,
				"is_active":   docActive,
				"type":        docType,
				"kb_id":       kbIDOut,
				"web_url":     webURL,
			})
		}
		docRows.Close()

		// --- Step 3: Combine KB + docs ---
		kbList = append(kbList, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"namespace":   namespace,
			"index":       index,
			"no_of_rec":   noOfRec,
			"is_active":   isActive,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"documents":   documents,
		})
	}
	return kbList, false
}

// HandleDeleteKBConfig deletes a knowledge base safely using a transaction
func HandleDeleteKBConfig(w http.ResponseWriter, kbID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting knowledge base config with ID: %s", kbID)

	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// ---------------------------
	// STEP 1: Fetch KB index (hostId)
	// ---------------------------
	var hostId string
	querysel := `
		SELECT index
		FROM knowledge_base
		WHERE id = $1
	`
	err = tx.QueryRow(querysel, kbID).Scan(&hostId)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		WriteInternalServerError(w, "Failed to fetch KB")
		return
	}

	// ---------------------------
	// STEP 2: Count documents (to decide external cleanup)
	// ---------------------------
	var docCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM documents WHERE kb_id = $1`, kbID).Scan(&docCount)
	if err != nil {
		WriteInternalServerError(w, "Failed to check KB documents")
		return
	}

	// ---------------------------
	// STEP 3: Delete documents
	// ---------------------------
	_, err = tx.Exec(`DELETE FROM documents WHERE kb_id = $1`, kbID)
	if err != nil {
		WriteInternalServerError(w, "Failed to delete KB documents")
		return
	}

	// ---------------------------
	// STEP 4: Delete KB record
	// ---------------------------
	var deletedKBID string
	err = tx.QueryRow(`DELETE FROM knowledge_base WHERE id = $1 RETURNING id`, kbID).Scan(&deletedKBID)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		WriteInternalServerError(w, "Failed to delete KB")
		return
	}

	// ---------------------------
	// STEP 5: External cleanup (ONLY IF documents existed)
	// ---------------------------
	if docCount > 0 {
		bucket := configs.GetEnv("AWS_BUCKET")
		region := configs.GetEnv("AWS_REGION")

		if err := DeleteS3Namespace(bucket, kbID, region); err != nil {
			log.Printf("Error deleting S3 namespace: %v", err)
		}

		if err := vectors.DeletePineconeNamespace(hostId, kbID); err != nil {
			log.Printf("Error deleting Pinecone namespace: %v", err)
		}
	} else {
		log.Printf("Skipping S3 & Pinecone cleanup — No documents for KB %s", kbID)
	}

	// ---------------------------
	// STEP 6: Commit before doing external deletion
	// ---------------------------
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit KB delete")
		return
	}

	// ---------------------------
	// SUCCESS
	// ---------------------------
	json.NewEncoder(w).Encode(map[string]string{
		"id":      deletedKBID,
		"message": "Knowledge base deleted successfully",
	})
}

func CreateCollectionExternal(
	endpoint string,
	orgID string,
	collectionID string,
	description string,
	files []*multipart.FileHeader,
	texts []string,
	urls []string,
) (*CreateCollectionResponse, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// -------- REQUIRED FIELDS --------
	_ = writer.WriteField("organisation_id", orgID)
	_ = writer.WriteField("collection_id", collectionID)
	_ = writer.WriteField("description", description)

	// -------- FILES --------
	for _, fh := range files {

		file, err := fh.Open()
		if err != nil {
			return nil, err
		}

		part, err := writer.CreateFormFile("file", fh.Filename)
		if err != nil {
			file.Close()
			return nil, err
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			return nil, err
		}
	}

	// -------- TEXT --------
	for _, t := range texts {
		if strings.TrimSpace(t) != "" {
			_ = writer.WriteField("text", t)
		}
	}

	// -------- URL --------
	for _, u := range urls {
		if strings.TrimSpace(u) != "" {
			_ = writer.WriteField("url", u)
		}
	}

	writer.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ingestion failed: %s", string(respBody))
	}

	var result CreateCollectionResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func DeleteCollectionExternal(
	endpoint string,
	orgID string,
	collectionID string,
	documentID string,
) (bool, error) {

	// ---- Build JSON Payload ----
	payload := map[string]string{
		"organisation_id": orgID,
		"collection_id":   collectionID,
	}

	if strings.TrimSpace(documentID) != "" {
		payload["document_id"] = documentID
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	// ---- Create Request ----
	req, err := http.NewRequest(
		http.MethodDelete,
		endpoint,
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("delete collection failed: %s", string(respBody))
	}

	return true, nil
}

// HandleCreateKBConfig creates a new Knowledge base using RAG approach
func HandleCreateKBConfigRAG(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	var organizationId string
	query := `
		SELECT w.organization_id
		FROM api_keys ak
		JOIN workspaces w
			ON ak.workspace_id = w.id
		WHERE ak.id = $1;
	`
	err = DB.QueryRow(query, apiKeyId).Scan(&organizationId)
	if err != nil {
		WriteNotFoundError(w, "Organization not found for API key")
		return
	}

	var kbInput KBInput
	kbInput.Name = r.FormValue("name")
	kbInput.Description = r.FormValue("description")
	kbInput.IsActive = true

	namespace := uuid.New()
	var docs []DocumentRecord

	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	// =====================================================
	// ✅ CALL RAG INGESTION SERVICE
	// =====================================================
	files := r.MultipartForm.File["input"]
	texts := r.MultipartForm.Value["text"]
	urls := r.MultipartForm.Value["url"]

	// -------- TEXT INPUT --------
	for _, t := range texts {
		if t == "" {
			continue
		}
		id := uuid.New().String()
		filename := fmt.Sprintf("%s.txt", id)

		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace.String(), filename)
		url, err := SaveStringToS3(t, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload text to S3: "+err.Error())
			return
		}
		docs = append(docs, DocumentRecord{
			Name:     filename,
			FileType: "text",
			Content:  t,
			ID:       id,
			URL:      url,
		})
	}

	// -------- URL INPUT --------
	for _, raw := range urls {

		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		var payload URLPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			log.Printf("json error: %v", err)
			continue
		}

		b64 := payload.Text
		b64 = strings.ReplaceAll(b64, "-", "+")
		b64 = strings.ReplaceAll(b64, "_", "/")

		if m := len(b64) % 4; m != 0 {
			b64 += strings.Repeat("=", 4-m)
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Printf("base64 decode failed: %v", err)
			continue
		}

		id := uuid.New().String()
		filename := fmt.Sprintf("%s.txt", id)

		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace.String(), filename)
		url, err := SaveStringToS3(string(decoded), bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload URL text to S3: "+err.Error())
			return
		}

		docs = append(docs, DocumentRecord{
			Name:     filename,
			FileType: "url",
			Content:  string(decoded),
			WebURL:   payload.URL,
			ID:       id,
			URL:      url,
		})
	}

	// -------- FILE UPLOADS --------
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error opening file: %v", err))
			return
		}

		id := uuid.New().String()
		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace.String(), fh.Filename)

		url, err := SaveFileStreamToS3(file, bucket, key, region)

		// CLOSE IMMEDIATELY after upload
		file.Close()

		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error uploading file to S3: %v", err))
			return
		}

		docs = append(docs, DocumentRecord{
			Name:     fh.Filename,
			FileType: strings.ToLower(filepath.Ext(fh.Filename)),
			URL:      url,
			ID:       id,
		})
	}

	// =====================================================
	// ✅ START TRANSACTION
	// =====================================================
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start DB transaction")
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// -------- INSERT KB --------
	insertKBQuery := `
	INSERT INTO knowledge_base (
	    id, name, description, namespace, index,
	    is_active, created_at, updated_at, created_by, no_of_rec
	) VALUES (
	    $7, $1, $2, $3, $4,
	    $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6, $8
	) RETURNING id`

	var kbID string
	err = tx.QueryRow(
		insertKBQuery,
		kbInput.Name,
		kbInput.Description,
		namespace.String(),
		organizationId,
		kbInput.IsActive,
		apiKeyId,
		namespace,
		len(docs),
	).Scan(&kbID)

	if err != nil {
		WriteInternalServerError(w, "Failed to create knowledge base")
		return
	}

	sqs_url := configs.GetEnv("AWS_SQS_URL")
	service, err := NewSQSService(sqs_url)
	if err != nil {
		log.Fatal(err)
	}

	// -------- INSERT DOCUMENTS --------
	for _, u := range docs {
		_, err = tx.Exec(`
			INSERT INTO public.documents
			(id, name, preview_url, is_active, created_at, updated_at, created_by, type, kb_id, web_url)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5, $6, $7, $8)
		`,
			u.ID,
			u.Name,
			u.URL,
			true,
			apiKeyId,
			u.FileType,
			namespace.String(),
			u.WebURL,
		)

		if err != nil {
			WriteInternalServerError(w, "Error saving documents")
			return
		}

		// Send SQS message for async processing
		err = service.SendSQSMessage(organizationId, namespace.String(), u.ID, u.URL)
		if err != nil {
			log.Println("Failed to send SQS message:", err)
		}
	}

	// -------- COMMIT --------
	err = tx.Commit()
	if err != nil {
		WriteInternalServerError(w, "Transaction commit failed")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      kbID,
		"message": "Knowledge Base created successfully",
	})
}

func UpdateRAGDocumentStatus(documentID string, status string, message string) error {
	query := `
		UPDATE documents
		SET 
			status = $1,
			message = COALESCE($2, ''),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := DB.Exec(query, status, message, documentID)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no document found with id: %s", documentID)
	}

	return nil
}

type MessagePayload struct {
	FilePath string `json:"file_path"`
	OrgID    string `json:"org_id"`
	ColID    string `json:"col_id"`
	DocID    string `json:"doc_id"`
}

type SQSService struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSService(queueURL string) (*SQSService, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &SQSService{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

func (s *SQSService) SendSQSMessage(orgID, colID, docID, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := MessagePayload{
		FilePath: path,
		OrgID:    orgID,
		ColID:    colID,
		DocID:    docID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.queueURL,
		MessageBody: awsString(string(body)),
	})

	return err
}

func awsString(v string) *string {
	return &v
}

// HandleAddDocToKBConfig adds files, text, or URLs to an existing Knowledge Base
func HandleAddDocToKBConfigRAG(w http.ResponseWriter, r *http.Request, kbID, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM knowledge_base a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, kbID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving knowledge base: %v", err)
		WriteInternalServerError(w, "Failed to retrieve knowledge base")
		return
	}

	if !exists {
		log.Printf("No knowledge base found with ID: %s", kbID)
		WriteNotFoundError(w, "Knowledge Base not found")
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"kb_id":   kbID,
			"added":   0,
			"message": "No content added to existing Knowledge Base",
		})
		return
	}

	// ---------------------------------
	// Fetch KB metadata
	// ---------------------------------
	var namespace string
	var orgID string
	query := `
		SELECT namespace, index
		FROM knowledge_base
		WHERE id = $1 AND is_active = true
	`
	if err := DB.QueryRow(query, kbID).Scan(&namespace, &orgID); err != nil {
		WriteNotFoundError(w, "Knowledge Base not found or inactive")
		return
	}

	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	var docs []DocumentRecord

	// =====================================================
	// ✅ CALL RAG INGESTION SERVICE
	// =====================================================
	files := r.MultipartForm.File["input"]
	texts := r.MultipartForm.Value["text"]
	urls := r.MultipartForm.Value["url"]

	// -------- TEXT INPUT --------
	for _, t := range texts {
		if t == "" {
			continue
		}
		id := uuid.New().String()
		filename := fmt.Sprintf("%s.txt", id)

		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace, filename)
		url, err := SaveStringToS3(t, bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload text to S3: "+err.Error())
			return
		}
		docs = append(docs, DocumentRecord{
			Name:     filename,
			FileType: "text",
			Content:  t,
			ID:       id,
			URL:      url,
		})
	}

	// -------- URL INPUT --------
	for _, raw := range urls {

		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		var payload URLPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			log.Printf("json error: %v", err)
			continue
		}

		b64 := payload.Text
		b64 = strings.ReplaceAll(b64, "-", "+")
		b64 = strings.ReplaceAll(b64, "_", "/")

		if m := len(b64) % 4; m != 0 {
			b64 += strings.Repeat("=", 4-m)
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			log.Printf("base64 decode failed: %v", err)
			continue
		}

		id := uuid.New().String()
		filename := fmt.Sprintf("%s.txt", id)

		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace, filename)
		url, err := SaveStringToS3(string(decoded), bucket, key, region)
		if err != nil {
			WriteInternalServerError(w, "Failed to upload URL text to S3: "+err.Error())
			return
		}

		docs = append(docs, DocumentRecord{
			Name:     filename,
			FileType: "url",
			Content:  string(decoded),
			WebURL:   payload.URL,
			ID:       id,
			URL:      url,
		})
	}

	// -------- FILE UPLOADS --------
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error opening file: %v", err))
			return
		}

		id := uuid.New().String()
		key := fmt.Sprintf("clawdface/uploads/kbrag/%s/%s", namespace, fh.Filename)

		url, err := SaveFileStreamToS3(file, bucket, key, region)

		// CLOSE IMMEDIATELY after upload
		file.Close()

		if err != nil {
			WriteInternalServerError(w, fmt.Sprintf("Error uploading file to S3: %v", err))
			return
		}

		docs = append(docs, DocumentRecord{
			Name:     fh.Filename,
			FileType: strings.ToLower(filepath.Ext(fh.Filename)),
			URL:      url,
			ID:       id,
		})
	}

	// ---------------------------------
	// Prevent empty upload
	// ---------------------------------
	if len(docs) == 0 {
		json.NewEncoder(w).Encode(map[string]any{
			"kb_id":   kbID,
			"added":   len(docs),
			"message": "No content added to existing Knowledge Base",
		})
		return
	}

	// =====================================================
	// TRANSACTION
	// =====================================================
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start DB transaction")
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Update KB count
	_, err = tx.Exec(`
		UPDATE knowledge_base
		SET no_of_rec = no_of_rec + $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, len(docs), kbID)
	if err != nil {
		WriteInternalServerError(w, "Failed to update KB record count")
		return
	}

	sqs_url := configs.GetEnv("AWS_SQS_URL")
	service, err := NewSQSService(sqs_url)
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range docs {
		_, err = tx.Exec(`
			INSERT INTO documents
			(id, name, preview_url, is_active, created_at, updated_at, created_by, type, kb_id, web_url)
			VALUES ($1,$2,$3,true,NOW(),NOW(),$4,$5,$6,$7)
		`,
			d.ID,
			d.Name,
			d.URL,
			apiKeyId,
			d.FileType,
			kbID,
			d.WebURL,
		)
		if err != nil {
			WriteInternalServerError(w, "Error saving documents")
			return
		}

		// Send SQS message for async processing
		err = service.SendSQSMessage(orgID, namespace, d.ID, d.URL)
		if err != nil {
			log.Println("Failed to send SQS message:", err)
		}
	}

	if err = tx.Commit(); err != nil {
		WriteInternalServerError(w, "Transaction commit failed")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"kb_id":   kbID,
		"added":   len(docs),
		"message": "Content added successfully to existing Knowledge Base",
	})
}

// HandleUpdateKBConfig updates an existing Knowledge Base
func HandleUpdateKBConfigRAG(w http.ResponseWriter, r *http.Request, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Knowledge Base config with ID: %s", kbID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM knowledge_base a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, kbID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving knowledge base: %v", err)
		WriteInternalServerError(w, "Failed to retrieve knowledge base")
		return
	}

	if !exists {
		log.Printf("No knowledge base found with ID: %s", kbID)
		WriteNotFoundError(w, "Knowledge Base not found")
		return
	}

	var kbConfig struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&kbConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE knowledge_base
        SET name = $1,
            description = $2,
            is_active = $3,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
        RETURNING id`

	var id string
	err = DB.QueryRow(
		query,
		kbConfig.Name,
		kbConfig.Description,
		kbConfig.IsActive,
		kbID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No knowledge base found with ID: %s", kbID)
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		log.Printf("Error updating knowledge base config: %v", err)
		WriteInternalServerError(w, "Failed to update knowledge base")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Knowledge Base updated successfully",
	})
}

// HandleGetKBConfig retrieves a specific Knowledge Base config and its related documents
func HandleGetKBConfigRAG(w http.ResponseWriter, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching knowledge base and documents for KB ID: %s", kbID)

	// --- Step 1: Fetch Knowledge Base ---
	response, shouldReturn := GetKBFunctionRAG(kbID, w, apiKeyId)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(response)
}

// HandleGetKBConfigAPI retrieves a specific Knowledge Base config and its related documents
func HandleGetKBConfigAPIRAG(w http.ResponseWriter, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching knowledge base and documents for KB ID: %s", kbID)

	// --- Step 1: Fetch Knowledge Base ---
	kbs, shouldReturn := GetKBFunctionRAG(kbID, w, apiKeyId)
	if shouldReturn {
		return
	}

	kb, ok := kbs["knowledge_base"].(map[string]interface{})
	if !ok {
		// handle error properly
		log.Println("invalid knowledge_base structure")
		return
	}

	kbConfigAPI := map[string]interface{}{
		"id":          kb["id"],
		"name":        kb["name"],
		"description": kb["description"],
		"no_of_rec":   kb["no_of_rec"],
		"is_active":   kb["is_active"],
		"created_at":  kb["created_at"],
		"updated_at":  kb["updated_at"],
		"documents":   kbs["documents"],
	}

	json.NewEncoder(w).Encode(kbConfigAPI)
}

func GetKBFunctionRAG(kbID string, w http.ResponseWriter, apiKeyId string) (map[string]interface{}, bool) {
	kbQuery := `
       SELECT a.id, a.name, a.description, namespace, index, no_of_rec, a.is_active, a.created_at, a.updated_at, a.created_by
        FROM knowledge_base a
    JOIN api_keys ak2 ON a.created_by = ak2.id
    JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
    JOIN workspaces w1 ON ak1.workspace_id = w1.id
    JOIN workspaces w2 ON ak2.workspace_id = w2.id 
        AND w2.organization_id = w1.organization_id
    WHERE ak1.id = $2 AND a.id = $1
    `

	var (
		id, name, description, namespace, index string
		isActive                                bool
		noOfRec                                 int
		createdAt, updatedAt                    time.Time
		createdBy                               string
	)

	err := DB.QueryRow(kbQuery, kbID, apiKeyId).Scan(
		&id, &name, &description, &namespace, &index,
		&noOfRec, &isActive, &createdAt, &updatedAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error fetching knowledge base: %v", err)
		WriteInternalServerError(w, "Failed to fetch knowledge base")
		return nil, true
	}

	// --- Step 2: Fetch Documents for this KB ---
	docQuery := `
        SELECT id, name, preview_url, is_active, type, kb_id, COALESCE(web_url, '') as web_url, status, COALESCE(message, '') as message
        FROM documents
        WHERE kb_id = $1
    `

	rows, err := DB.Query(docQuery, kbID)
	if err != nil {
		log.Printf("Error querying documents: %v", err)
		WriteInternalServerError(w, "Failed to retrieve documents")
		return nil, true
	}
	defer rows.Close()

	var documents []map[string]interface{}

	for rows.Next() {
		var (
			docID, docName, previewURL, docType, kbIDOut, webURL, status, message string
			docActive                                                             bool
		)

		if err := rows.Scan(&docID, &docName, &previewURL, &docActive, &docType, &kbIDOut, &webURL, &status, &message); err != nil {
			log.Printf("Error scanning document row: %v", err)
			continue
		}

		documents = append(documents, map[string]interface{}{
			"id":          docID,
			"name":        docName,
			"preview_url": previewURL,
			"is_active":   docActive,
			"type":        docType,
			"kb_id":       kbIDOut,
			"web_url":     webURL,
			"status":      status,
			"message":     message,
		})
	}

	// --- Step 3: Combine into response ---
	response := map[string]interface{}{
		"knowledge_base": map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"namespace":   namespace,
			"index":       index,
			"no_of_rec":   noOfRec,
			"is_active":   isActive,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		},
		"documents": documents,
	}
	return response, false
}

// HandleGetAllKBConfigs retrieves all knowledge bases with their related documents
func HandleGetAllKBConfigsRAG(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching all KBs with documents for apiKeyId: %s", apiKeyId)

	// --- Step 1: Fetch all KBs ---
	kbList, shouldReturn := GetAllKBFunctionRAG(apiKeyId, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(kbList)
}

// HandleGetAllKBConfigs retrieves all knowledge bases with their related documents
func HandleGetAllKBConfigsAPIRAG(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching all KBs with documents for apiKeyId: %s", apiKeyId)

	// --- Step 1: Fetch all KBs ---
	kbList, shouldReturn := GetAllKBFunctionRAG(apiKeyId, w)
	if shouldReturn {
		return
	}

	var kbConfigsAPI []map[string]interface{}

	for _, kb := range kbList {

		kbConfigAPI := map[string]interface{}{
			"id":          kb["id"],
			"name":        kb["name"],
			"description": kb["description"],
			"no_of_rec":   kb["no_of_rec"],
			"is_active":   kb["is_active"],
			"created_at":  kb["created_at"],
			"updated_at":  kb["updated_at"],
			"documents":   kb["documents"],
		}

		kbConfigsAPI = append(kbConfigsAPI, kbConfigAPI)
	}
	json.NewEncoder(w).Encode(kbConfigsAPI)
}

func GetAllKBFunctionRAG(apiKeyId string, w http.ResponseWriter) ([]map[string]interface{}, bool) {
	kbQuery := `
        SELECT
			kb.id,
			kb.name,
			kb.description,
			kb.namespace,
			kb.index,
			kb.no_of_rec,
			kb.is_active,
			kb.created_at,
			kb.updated_at,
			kb.created_by
		FROM knowledge_base kb
		JOIN api_keys ak2
			ON kb.created_by = ak2.id          -- creator’s API key
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id   -- same workspace condition
		-- Workspace joins to match organization
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id        -- workspace of requesting key
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id        -- workspace of creator
			AND w2.organization_id = w1.organization_id   -- ✅ same organization
		WHERE ak1.id = $1
		ORDER BY kb.created_at DESC;
    `

	rows, err := DB.Query(kbQuery, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve knowledge bases")
		return nil, true
	}
	defer rows.Close()

	var kbList []map[string]interface{}

	for rows.Next() {
		var (
			id, name, description, namespace, index string
			noOfRec                                 int
			isActive                                bool
			createdAt, updatedAt                    time.Time
			createdBy                               string
		)

		if err := rows.Scan(
			&id, &name, &description, &namespace,
			&index, &noOfRec, &isActive, &createdAt, &updatedAt, &createdBy,
		); err != nil {
			WriteInternalServerError(w, "Error scanning knowledge bases")
			return nil, true
		}

		// --- Step 2: Fetch Documents for this KB ---
		docQuery := `
            SELECT
				d.id,
				d.name,
				d.preview_url,
				d.is_active,
				d.type,
				d.kb_id,
				COALESCE(d.web_url, '') AS web_url,
				d.status,
				COALESCE(d.message, '') AS message
			FROM documents d
			JOIN api_keys ak2
				ON d.created_by = ak2.id                      -- creator's API key
			JOIN api_keys ak1
				ON ak2.workspace_id = ak1.workspace_id        -- workspace match
			-- Workspace joins for organization filtering
			JOIN workspaces w1
				ON ak1.workspace_id = w1.id                   -- workspace of requesting key
			JOIN workspaces w2
				ON ak2.workspace_id = w2.id                   -- workspace of creator
				AND w2.organization_id = w1.organization_id   -- ✅ organization match
			WHERE ak1.id = $2
			AND d.kb_id = $1
			ORDER BY d.created_at DESC;
        `
		docRows, err := DB.Query(docQuery, id, apiKeyId)
		if err != nil {
			log.Printf("Error querying documents for kb_id %s: %v", id, err)
			continue
		}

		var documents []map[string]interface{}
		for docRows.Next() {
			var (
				// Added status and message here
				docID, docName, previewURL, docType, kbIDOut, webURL, status, message string
				docActive                                                             bool
			)

			// Added &status and &message to the Scan method
			if err := docRows.Scan(&docID, &docName, &previewURL, &docActive, &docType, &kbIDOut, &webURL, &status, &message); err != nil {
				log.Printf("Error scanning document row for kb_id %s: %v", id, err)
				continue
			}

			documents = append(documents, map[string]interface{}{
				"id":          docID,
				"name":        docName,
				"preview_url": previewURL,
				"is_active":   docActive,
				"type":        docType,
				"kb_id":       kbIDOut,
				"web_url":     webURL,
				"status":      status,  // Appended status
				"message":     message, // Appended message
			})
		}
		docRows.Close()

		// --- Step 3: Combine KB + docs ---
		kbList = append(kbList, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"namespace":   namespace,
			"index":       index,
			"no_of_rec":   noOfRec,
			"is_active":   isActive,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"documents":   documents,
		})
	}
	return kbList, false
}

// HandleDeleteKBConfig deletes a knowledge base safely using a transaction
func HandleDeleteKBConfigRAG(w http.ResponseWriter, kbID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting knowledge base config with ID: %s", kbID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM knowledge_base a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, kbID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving knowledge base: %v", err)
		WriteInternalServerError(w, "Failed to retrieve knowledge base")
		return
	}

	if !exists {
		log.Printf("No knowledge base found with ID: %s", kbID)
		WriteNotFoundError(w, "Knowledge Base not found")
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// ---------------------------
	// STEP 1: Fetch KB index (hostId)
	// ---------------------------
	var index string
	var namespace string
	querysel := `
		SELECT index, namespace
		FROM knowledge_base
		WHERE id = $1
	`
	err = tx.QueryRow(querysel, kbID).Scan(&index, &namespace)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		WriteInternalServerError(w, "Failed to fetch KB")
		return
	}

	// ---------------------------
	// STEP 2: Count documents (to decide external cleanup)
	// ---------------------------
	var docCount int
	err = tx.QueryRow(`SELECT COUNT(1) FROM documents WHERE kb_id = $1`, kbID).Scan(&docCount)
	if err != nil {
		WriteInternalServerError(w, "Failed to check KB documents")
		return
	}

	// ---------------------------
	// STEP 3: Delete documents
	// ---------------------------
	_, err = tx.Exec(`DELETE FROM documents WHERE kb_id = $1`, kbID)
	if err != nil {
		WriteInternalServerError(w, "Failed to delete KB documents")
		return
	}

	// ---------------------------
	// STEP 4: Delete KB record
	// ---------------------------
	var deletedKBID string
	err = tx.QueryRow(`DELETE FROM knowledge_base WHERE id = $1 RETURNING id`, kbID).Scan(&deletedKBID)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Knowledge base not found")
		return
	} else if err != nil {
		WriteInternalServerError(w, "Failed to delete KB")
		return
	}

	// ---------------------------
	// STEP 5: External cleanup (ONLY IF documents existed)
	// ---------------------------
	if docCount > 0 {
		bucket := configs.GetEnv("AWS_BUCKET")
		region := configs.GetEnv("AWS_REGION")

		if err := DeleteS3Namespace(bucket, kbID, region); err != nil {
			log.Printf("Error deleting S3 namespace: %v", err)
		}
		ingestionURL := configs.GetEnv("RAG_API_URL")

		//Delete the details from RAG service
		collectionResp, err := DeleteCollectionExternal(
			ingestionURL+"delete_collection",
			index,
			namespace,
			"",
		)
		if err != nil || !collectionResp {
			log.Printf("Error deleting collection from RAG service: %v", err)
		}
	} else {
		log.Printf("Skipping S3 & RAG cleanup — No documents for KB %s", kbID)
	}

	// ---------------------------
	// STEP 6: Commit before doing external deletion
	// ---------------------------
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit KB delete")
		return
	}

	// ---------------------------
	// SUCCESS
	// ---------------------------
	json.NewEncoder(w).Encode(map[string]string{
		"id":      deletedKBID,
		"message": "Knowledge base deleted successfully",
	})
}

// HandleDeleteDocToKB deletes a specific document from a knowledge base (with transaction)
func HandleDeleteDocToKBRAG(w http.ResponseWriter, docID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting document with ID: %s", docID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM documents a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, docID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving document: %v", err)
		WriteInternalServerError(w, "Failed to retrieve document")
		return
	}

	if !exists {
		log.Printf("No document found with ID: %s", docID)
		WriteNotFoundError(w, "Document not found")
		return
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// ---------------------------
	// STEP 1: Fetch document + KB details
	// ---------------------------
	query := `
        SELECT d.id, d.preview_url, d.kb_id, kb.index, kb.namespace
		FROM documents d
		JOIN knowledge_base kb
		ON d.kb_id = kb.id::text
		WHERE d.id = $1
	`

	var (
		documentID string
		fileName   string
		kbID       string
		index      string
		namespace  string
	)

	err = tx.QueryRow(query, docID).Scan(&documentID, &fileName, &kbID, &index, &namespace)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Document not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving document: %v", err)
		WriteInternalServerError(w, "Failed to retrieve document")
		return
	}

	// ---------------------------
	// STEP 2: Delete document from DB
	// ---------------------------
	_, err = tx.Exec(`DELETE FROM documents WHERE id = $1`, docID)
	if err != nil {
		log.Printf("Error deleting document from DB: %v", err)
		WriteInternalServerError(w, "Failed to delete document record")
		return
	}

	// ---------------------------
	// STEP 4: External cleanup AFTER COMMIT
	// ---------------------------
	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	// Delete from S3
	if err := DeleteS3Document(bucket, kbID, fileName, region); err != nil {
		log.Printf("Error deleting file from S3: %v", err)
	}
	ingestionURL := configs.GetEnv("RAG_API_URL")

	//Delete the details from RAG service
	collectionResp, err := DeleteCollectionExternal(
		ingestionURL+"delete_document",
		index,
		namespace,
		docID,
	)
	if err != nil || !collectionResp {
		log.Printf("Error deleting document from RAG service: %v", err)
	}

	// ---------------------------
	// STEP 3: Commit transaction
	// ---------------------------
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit document delete")
		return
	}

	// ---------------------------
	// SUCCESS RESPONSE
	// ---------------------------
	json.NewEncoder(w).Encode(map[string]string{
		"id":      documentID,
		"message": "Document deleted successfully",
	})
}

// HandleCreateMemoryCategory creates a new memory category
func HandleCreateMemoryCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
		CreatedBy   string `json:"created_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(payload.Name) == "" {
		WriteBadRequestError(w, "name is required")
		return
	}
	if strings.TrimSpace(payload.Description) == "" {
		WriteBadRequestError(w, "description is required")
		return
	}

	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	query := `
        INSERT INTO memory_category (
            id, name, description, is_active, created_at, updated_at, created_by
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULLIF($4, '')
        ) RETURNING id`

	var id string
	err := DB.QueryRow(query, payload.Name, payload.Description, isActive, payload.CreatedBy).Scan(&id)
	if err != nil {
		log.Printf("Error creating memory category: %v", err)
		WriteInternalServerError(w, "Failed to create memory category")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Memory category created successfully",
	})
}

// HandleUpdateMemoryCategory updates an existing memory category
func HandleUpdateMemoryCategory(w http.ResponseWriter, r *http.Request, categoryID string) {
	w.Header().Set("Content-Type", "application/json")
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	var isActive sql.NullBool
	if payload.IsActive != nil {
		isActive = sql.NullBool{Bool: *payload.IsActive, Valid: true}
	}

	query := `
        UPDATE memory_category
        SET name = COALESCE(NULLIF($1, ''), name),
            description = COALESCE(NULLIF($2, ''), description),
            is_active = COALESCE($3, is_active),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
        RETURNING id`

	var id string
	err := DB.QueryRow(query, payload.Name, payload.Description, isActive, categoryID).Scan(&id)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Memory category not found")
		return
	} else if err != nil {
		log.Printf("Error updating memory category: %v", err)
		WriteInternalServerError(w, "Failed to update memory category")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Memory category updated successfully",
	})
}

// HandleGetMemoryCategory retrieves a memory category by ID
func HandleGetMemoryCategory(w http.ResponseWriter, categoryID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching memory category with ID: %s", categoryID)

	query := `
        SELECT id, name, description, is_active, created_at, updated_at, created_by
        FROM memory_category
        WHERE id = $1`

	var (
		id, name, description, createdBy string
		isActive                         bool
		createdAt, updatedAt             time.Time
	)

	err := DB.QueryRow(query, categoryID).Scan(&id, &name, &description, &isActive, &createdAt, &updatedAt, &createdBy)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Memory category not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving memory category: %v", err)
		WriteInternalServerError(w, "Failed to retrieve memory category")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"is_active":   isActive,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"created_by":  createdBy,
	})
}

// HandleGetAllMemoryCategoriesByUser retrieves all memory categories for a user's organizations
func HandleGetAllMemoryCategoriesByUser(w http.ResponseWriter, userID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching memory categories for user ID: %s", userID)

	query := `
        SELECT DISTINCT mc.id, mc.name, mc.description, mc.is_active, mc.created_at, mc.updated_at, mc.created_by
        FROM memory_category mc
        JOIN workspace_members wm_created ON wm_created.user_id = mc.created_by
        JOIN workspaces w_created ON w_created.id = wm_created.workspace_id
        JOIN organizations o ON o.id = w_created.organization_id
        JOIN workspaces w_req ON w_req.organization_id = o.id
        JOIN workspace_members wm_req ON wm_req.workspace_id = w_req.id
        WHERE wm_req.user_id = $1 AND mc.is_active = true
        ORDER BY mc.created_at DESC`

	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error retrieving memory categories: %v", err)
		WriteInternalServerError(w, "Failed to retrieve memory categories")
		return
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var (
			id, name, description, createdBy string
			isActive                         bool
			createdAt, updatedAt             time.Time
		)
		if err := rows.Scan(&id, &name, &description, &isActive, &createdAt, &updatedAt, &createdBy); err != nil {
			log.Printf("Error scanning memory category: %v", err)
			WriteInternalServerError(w, "Failed to retrieve memory categories")
			return
		}
		categories = append(categories, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"is_active":   isActive,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"created_by":  createdBy,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating memory categories: %v", err)
		WriteInternalServerError(w, "Failed to retrieve memory categories")
		return
	}

	json.NewEncoder(w).Encode(categories)
}

// HandleDeleteMemoryCategory deletes a memory category
func HandleDeleteMemoryCategory(w http.ResponseWriter, categoryID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting memory category with ID: %s", categoryID)

	var id string
	err := DB.QueryRow(`
		UPDATE memory_category
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id`, categoryID).Scan(&id)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Memory category not found")
		return
	} else if err != nil {
		log.Printf("Error deleting memory category: %v", err)
		WriteInternalServerError(w, "Failed to delete memory category")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Memory category deleted successfully",
	})
}

// HandleDeleteDocToKB deletes a specific document from a knowledge base (with transaction)
func HandleDeleteDocToKB(w http.ResponseWriter, docID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting document with ID: %s", docID)

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// ---------------------------
	// STEP 1: Fetch document + KB details
	// ---------------------------
	query := `
        SELECT d.id, d.preview_url, d.kb_id, kb.index
		FROM documents d
		JOIN knowledge_base kb
		ON d.kb_id = kb.id::text
		WHERE d.id = $1
	`

	var (
		documentID string
		fileName   string
		kbID       string
		hostID     string
	)

	err = tx.QueryRow(query, docID).Scan(&documentID, &fileName, &kbID, &hostID)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Document not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving document: %v", err)
		WriteInternalServerError(w, "Failed to retrieve document")
		return
	}

	// ---------------------------
	// STEP 2: Delete document from DB
	// ---------------------------
	_, err = tx.Exec(`DELETE FROM documents WHERE id = $1`, docID)
	if err != nil {
		log.Printf("Error deleting document from DB: %v", err)
		WriteInternalServerError(w, "Failed to delete document record")
		return
	}

	// ---------------------------
	// STEP 4: External cleanup AFTER COMMIT
	// ---------------------------
	bucket := configs.GetEnv("AWS_BUCKET")
	region := configs.GetEnv("AWS_REGION")

	// Delete from S3
	if err := DeleteS3Document(bucket, kbID, fileName, region); err != nil {
		log.Printf("Error deleting file from S3: %v", err)
	}

	// Delete vectors from Pinecone
	if err := vectors.DeletePineconeDocument(hostID, kbID, fileName); err != nil {
		log.Printf("Error deleting vectors from Pinecone: %v", err)
	}

	// ---------------------------
	// STEP 3: Commit transaction
	// ---------------------------
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit document delete")
		return
	}

	// ---------------------------
	// SUCCESS RESPONSE
	// ---------------------------
	json.NewEncoder(w).Encode(map[string]string{
		"id":      documentID,
		"message": "Document deleted successfully",
	})
}

// GetAvatarsJSON returns avatar JSON for the given avatar IDs
func GetAvatarsJSON(avatarIDs []string) (json.RawMessage, error) {
	rows, err := DB.Query(`
		SELECT
			avatar_key_id,
			avatar_name,
			api_config,
			default_prompt
		FROM public.avatars
		WHERE avatar_key_id = ANY($1)
	`, pq.Array(avatarIDs))
	if err != nil {
		return nil, fmt.Errorf("error querying avatars: %v", err)
	}
	defer rows.Close()

	var avatars []map[string]interface{}

	for rows.Next() {
		var (
			avatarKeyID   string
			avatarName    string
			apiConfigJSON json.RawMessage
			defaultPrompt string
		)

		if err := rows.Scan(&avatarKeyID, &avatarName, &apiConfigJSON, &defaultPrompt); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Parse api_config (stored JSON)
		var apiConfig map[string]interface{}
		if len(apiConfigJSON) > 0 {
			if err := json.Unmarshal(apiConfigJSON, &apiConfig); err != nil {
				log.Printf("⚠️ invalid api_config for %s: %v", avatarKeyID, err)
				apiConfig = map[string]interface{}{}
			}
		} else {
			apiConfig = map[string]interface{}{}
		}

		// Build avatar response JSON object
		avatarData := map[string]interface{}{
			"config":                       apiConfig["config"],
			"timeout":                      apiConfig["timeout"],
			"is_custom":                    apiConfig["is_custom"],
			"frame_rate":                   apiConfig["frame_rate"],
			"exit_message":                 apiConfig["exit_message"],
			"idle_timeout":                 apiConfig["idle_timeout"],
			"persona_name":                 avatarName,
			"avatar_key_id":                avatarKeyID,
			"persona_prompt":               defaultPrompt,
			"silence_padding":              apiConfig["silence_padding"],
			"welcome_message":              apiConfig["welcome_message"],
			"avatar_data_source":           apiConfig["avatar_data_source"],
			"audio_features_type":          apiConfig["audio_features_type"],
			"eye_mask_replacement":         apiConfig["eye_mask_replacement"],
			"interpolation_config":         apiConfig["interpolation_config"],
			"scene_context_engine":         apiConfig["scene_context_engine"],
			"warning_exit_message":         apiConfig["warning_exit_message"],
			"exit_heads_up_message":        apiConfig["exit_heads_up_message"],
			"scene_analyzer_prompt":        apiConfig["scene_analyzer_prompt"],
			"conversational_context":       apiConfig["conversational_context"],
			"is_face_enhancer_enabled":     apiConfig["is_face_enhancer_enabled"],
			"audio_features_window_length": apiConfig["audio_features_window_length"],
		}

		avatars = append(avatars, avatarData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	jsonData, err := json.MarshalIndent(avatars, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling avatars JSON: %v", err)
	}

	return jsonData, nil
}

// ---------------------------
// Reusable Core Function
// ---------------------------
func CreateAgentConfig(tx *sql.Tx, agentConfig AgentConfigInput, apiKeyId string) (string, error) {
	query := `
        INSERT INTO agents (
			id, agent_name, agent_system_prompt, tools, config, is_active, avatars,
			created_at, updated_at, created_by, is_public, record, callback_url, callback_events, email, type, add_ons, default_system_prompt
        ) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $7, $8, $9, $10, $11, $12, $13, $14, $15
        ) RETURNING id`

	var id string
	err := tx.QueryRow(
		query,
		agentConfig.AgentName,
		agentConfig.AgentSystemPrompt,
		"{}", // tools placeholder
		agentConfig.Config,
		agentConfig.IsActive,
		agentConfig.Avatars,
		apiKeyId,
		agentConfig.IsPublic,
		agentConfig.Record,
		agentConfig.CallbackURL,
		pq.Array(agentConfig.CallbackEvents),
		agentConfig.Email,
		agentConfig.Type,
		agentConfig.AddOns,
		agentConfig.DefaultSystemPrompt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert agent config: %v", err)
	}

	// Insert knowledge bases
	kbInsert := `INSERT INTO agents_kb (id, agent_id, knowledge_base_id, mode) VALUES (gen_random_uuid(), $1, $2, $3)`
	for _, kb := range agentConfig.KnowledgeBase {
		if _, err := tx.Exec(kbInsert, id, kb.ID, kb.Mode); err != nil {
			return "", fmt.Errorf("failed to insert KB: %v", err)
		}
	}

	// Insert tools
	toolInsert := `INSERT INTO agents_tool (id, agent_id, tool_id) VALUES (gen_random_uuid(), $1, $2)`
	for _, tool := range agentConfig.Tool {
		if _, err := tx.Exec(toolInsert, id, tool.ID); err != nil {
			return "", fmt.Errorf("failed to insert tool: %v", err)
		}
	}

	// Insert MCP
	mcpInsert := `INSERT INTO agents_mcp (id, agent_id, mcp_id) VALUES (gen_random_uuid(), $1, $2)`
	for _, mcp := range agentConfig.MCP {
		if _, err := tx.Exec(mcpInsert, id, mcp.ID); err != nil {
			return "", fmt.Errorf("failed to insert MCP: %v", err)
		}
	}

	// Insert integration
	integrationInsert := `INSERT INTO agents_integration (id, agent_id, auth_config_id) VALUES (gen_random_uuid(), $1, $2)`
	for _, integration := range agentConfig.Integration {
		if _, err := tx.Exec(integrationInsert, id, integration.ID); err != nil {
			return "", fmt.Errorf("failed to insert integration: %v", err)
		}
	}

	return id, nil
}

type AgentTimeout struct {
	Timeout         int `json:"timeout"`
	MaxCallDuration int `json:"maxCallDuration"`
}

func ValidateTimeout(apiKeyId string, config json.RawMessage) error {

	var timeout AgentTimeout

	err := json.Unmarshal(config, &timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	var maxSessionDuration int

	query := `
	SELECT cl.max_session_duration
			FROM credit_limits cl
			JOIN organizations o ON o.id = cl.organization_id
			JOIN workspaces w ON w.organization_id = o.id
			JOIN api_keys ak ON ak.workspace_id = w.id
			WHERE ak.id = $1
			LIMIT 1;
	`

	err1 := DB.QueryRow(query, apiKeyId).Scan(&maxSessionDuration)
	if err1 != nil {
		return fmt.Errorf("failed to fetch organization plan: %w", err1)
	}

	maxTimeout := maxSessionDuration * 60
	if timeout.Timeout > maxTimeout {
		return fmt.Errorf(
			"Requested max call duration (%v) exceeds plan limit (%v)",
			timeout.Timeout,
			maxTimeout,
		)
	} else if timeout.MaxCallDuration > maxTimeout {
		return fmt.Errorf(
			"Requested max call duration (%v) exceeds plan limit (%v)",
			timeout.MaxCallDuration,
			maxTimeout,
		)
	}

	return nil
}

func CreateAgentAndStartConversation(w http.ResponseWriter, r *http.Request) {
	apiKeyId := r.Context().Value("apiKeyId").(string)
	apiKey := r.Header.Get("X-API-Key")

	var agentConfig AgentConfigInput

	// ✅ Decode ONCE
	if err := json.NewDecoder(r.Body).Decode(&agentConfig); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	// ✅ Step 1: Prepare agent
	agentConfig.IsActive = false

	// ✅ Validate config
	if err := ValidateTimeout(apiKeyId, agentConfig.Config); err != nil {
		log.Printf("Validation error: %v", err)
		WriteInternalServerError(w, "Validation failed")
		return
	}

	// ✅ Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// ✅ Create agent
	agentId, err := CreateAgentConfig(tx, agentConfig, apiKeyId)
	if err != nil {
		log.Printf("Error creating agent: %v", err)
		WriteInternalServerError(w, "Error creating agent")
		return
	}

	// ✅ Commit agent creation BEFORE using it
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// ✅ Step 2: Use created agentId for conversation
	roomId, err := GenerateRoomId()
	if err != nil {
		http.Error(w, "Failed to generate room ID", http.StatusInternalServerError)
		return
	}

	user_email := agentConfig.Email
	first_name := ""
	if agentConfig.Email == "" {
		user_email = "review@trugen.ai"
		first_name = "TruGen Review"
	} else {
		//retrive user name
		userQuery := `
        SELECT
		u.id,
		u.email,
		u.first_name,
		u.last_name
	FROM users u
	WHERE u.email = $1`

		var (
			id, email, firstName, lastName string
		)

		err = DB.QueryRow(userQuery, agentConfig.Email).Scan(
			&id, &email, &firstName, &lastName,
		)

		if err == sql.ErrNoRows {
			WriteNotFoundError(w, "User not found")
			return
		} else if err != nil {
			log.Printf("Error retrieving user details: %v", err)
			WriteInternalServerError(w, "Failed to retrieve user details")
			return
		}
		user_email = email
		first_name = firstName + " " + lastName
	}

	conversationId, url, agent, err := HandleConversationCreation(
		apiKey,
		agentId, // 🔥 use newly created agent
		apiKeyId,
		agentConfig.Config,
		agentConfig.Config,
		first_name,
		user_email,
		roomId,
		"",
		false,
		"",
		"",
	)
	if err != nil {
		http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Token generation
	token := GetLiveKitJoinToken(TokenSourceRequest{
		RoomName:            roomId,
		ParticipantName:     first_name,
		ParticipantIdentity: user_email,
	})

	// ✅ Final response
	response := map[string]interface{}{
		"agentId":        agentId,
		"conversationId": conversationId,
		"url":            url,
		"livekitUrl":     configs.GetEnv("LIVEKIT_URL"),
		"token":          token,
		"avatar": map[string]interface{}{
			"id":    agent.AvatarID,
			"name":  agent.AvatarName,
			"image": agent.AvatarImageURL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var characters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func GenerateRoomId() (string, error) {
	randomString := func(length int) (string, error) {
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}

		for i := 0; i < length; i++ {
			bytes[i] = characters[int(bytes[i])%len(characters)]
		}

		return string(bytes), nil
	}

	part1, err := randomString(4)
	if err != nil {
		return "", err
	}
	part2, err := randomString(4)
	if err != nil {
		return "", err
	}
	part3, err := randomString(4)
	if err != nil {
		return "", err
	}
	part4, err := randomString(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s-%s", part1, part2, part3, part4), nil
}

// ---------------------------
// HandleCreateAgentConfig
// ---------------------------
func HandleCreateAgentConfig(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	var agentConfig AgentConfigInput

	if err := json.NewDecoder(r.Body).Decode(&agentConfig); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	err := ValidateTimeout(apiKeyId, agentConfig.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate error : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	id, err := CreateAgentConfig(tx, agentConfig, apiKeyId)
	if err != nil {
		log.Println("Error creating agent config:", err)
		WriteInternalServerError(w, "Error creating agent config")
		return
	}

	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Agent created successfully",
	})
}

var defaultAvatar = map[string]interface{}{
	"timeout":                      240,
	"avatar_key_id":                "665a1170",
	"avatar_data_source":           "avatar-inference-data/",
	"frame_rate":                   25,
	"silence_padding":              0.05,
	"is_face_enhancer_enabled":     false,
	"is_custom":                    false,
	"eye_mask_replacement":         false,
	"audio_features_type":          "silent_smooth",
	"audio_features_window_length": 5,
	"persona_name":                 "Sample AI Agent",
	"persona_prompt":               "You're a helpful ai agent. Answer all questions appropriately.",
	"conversational_context":       "Sample Conversational Context",

	"config": map[string]interface{}{
		"llm": map[string]interface{}{
			"model":          "openai/gpt-oss-120b",
			"provider":       "groq",
			"fallback_model": "gpt-4.1-nano",
			"use_nltk":       false,
		},
		"stt": map[string]interface{}{
			"model":                             "nova-3",
			"provider":                          "deepgram",
			"min_endpointing_delay":             0.3,
			"max_endpointing_delay":             0.4,
			"fallback_model":                    "nova-2-general",
			"allow_interm_results_interruption": true,
		},
		"tts": map[string]interface{}{
			"model_id":           "eleven_turbo_v2_5",
			"provider":           "elevenlabs",
			"voice_id":           "ZUrEGyu8GFMwnHbvLhv2",
			"language":           "a",
			"effects_profile_id": "small-bluetooth-speaker-class-device",
			"encoding":           "pcm_s16le",
			"gender":             "female",
			"fallback_voice_id":  "am_puck",
			"pitch":              0,
			"speaking_rate":      1,
			"stability":          0.5,
			"similarity_boost":   0.75,
			"sample_rate":        16000,
		},
	},

	"interpolation_config": map[string]interface{}{
		"exp":     2,
		"enabled": true,
	},

	"idle_timeout": map[string]interface{}{
		"timeout": 30,
		"filler_phrases": []interface{}{
			"Hey it's been a while since we last spoke, are we still connected?",
			"I notice we haven't talked for a bit, is everything okay?",
		},
	},

	"welcome_message": map[string]interface{}{
		"wait_time": 2,
		"messages": []interface{}{
			"Hi, how are you doing today?",
			"Hello, how can I help you?",
		},
	},

	"warning_exit_message": map[string]interface{}{
		"callout_before": 10,
		"messages": []interface{}{
			"We are almost at the end of our call, thank you for your time.",
			"Thank you for your time. We will see you next time.",
		},
	},

	"exit_message": map[string]interface{}{
		"max_call_duration": 300,
		"messages": []interface{}{
			"We are at the end of our call, thank you for your time.",
			"Thank you for your time today.",
		},
	},
}

func deepMerge(dst, src map[string]interface{}) {
	for k, v := range src {
		if existing, ok := dst[k]; ok {
			dstMap, ok1 := existing.(map[string]interface{})
			srcMap, ok2 := v.(map[string]interface{})
			if ok1 && ok2 {
				deepMerge(dstMap, srcMap)
			}
			continue
		}
		dst[k] = v
	}
}

func NormalizeAvatars(raw json.RawMessage) json.RawMessage {
	// If avatars missing or empty → inject default
	if len(raw) == 0 || string(raw) == "[]" || string(raw) == "[{}]" {
		b, _ := json.Marshal([]map[string]interface{}{defaultAvatar})
		return b
	}

	// Unmarshal array
	var avatars []map[string]interface{}
	if err := json.Unmarshal(raw, &avatars); err != nil || len(avatars) == 0 {
		b, _ := json.Marshal([]map[string]interface{}{defaultAvatar})
		return b
	}

	// Deep-merge defaults into each avatar
	for i := range avatars {
		deepMerge(avatars[i], defaultAvatar)
	}

	// Marshal back
	b, _ := json.Marshal(avatars)
	return b
}

// NormalizeConfig unwraps config if it was sent as a JSON-encoded string instead of an object.
func NormalizeConfig(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}

	// remove leading spaces
	raw = bytes.TrimSpace(raw)

	// If it starts with "", it is a JSON string containing JSON
	if len(raw) > 0 && raw[0] == '"' {
		var inner string
		if err := json.Unmarshal(raw, &inner); err == nil {
			return json.RawMessage(inner)
		}
	}
	return raw
}

// ---------------------------
// HandleCreateAgentAPI
// ---------------------------
func HandleCreateAgentAPI(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	var config AgentConfigInput

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	config.Config = NormalizeConfig(config.Config)
	config.Avatars = NormalizeAvatars(config.Avatars)

	config.Type = "etev"
	config.IsActive = true
	config.IsPublic = true
	config.AddOns = json.RawMessage([]byte("[]"))

	err := ValidateTimeout(apiKeyId, config.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate timeout : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	id, err := CreateAgentConfig(tx, config, apiKeyId)
	if err != nil {
		log.Println("Error creating agent from API:", err)
		WriteInternalServerError(w, "Error creating agent from API")
		return
	}

	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Agent created successfully",
	})
}

// ---------------------------
// HandleCreateAgentConfigByTemplate
// ---------------------------
func HandleCreateAgentConfigByTemplate(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	var tempConfig struct {
		AvatarIDs  []string        `json:"avatar_ids"`
		TemplateID string          `json:"template_id"`
		Email      string          `json:"email"`
		Type       string          `json:"type"`
		AddOns     json.RawMessage `json:"add_on"`
	}

	tempConfig.Type = "etev"
	tempConfig.AddOns = json.RawMessage([]byte("[]"))

	if err := json.NewDecoder(r.Body).Decode(&tempConfig); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM templates a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, tempConfig.TemplateID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving template: %v", err)
		WriteInternalServerError(w, "Failed to retrieve template")
		return
	}

	if !exists {
		log.Printf("No template found with ID: %s", tempConfig.TemplateID)
		WriteNotFoundError(w, "Template not found")
		return
	}

	avatarsJSON, err := GetAvatarsJSON(tempConfig.AvatarIDs)
	if err != nil {
		log.Printf("Error retrieving template: %v", err)
		WriteInternalServerError(w, "Failed to retrieve template")
		return
	}

	if !exists {
		log.Printf("No template found with ID: %s", tempConfig.TemplateID)
		WriteNotFoundError(w, "Template not found")
		return
	}

	avatarsJSON, err = GetAvatarsJSON(tempConfig.AvatarIDs)
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch avatars")
		return
	}

	templateConfig, err := GetTemplateConfigById(tempConfig.TemplateID, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch template config")
		return
	}

	err = ValidateTimeout(apiKeyId, templateConfig.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate timeout : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	// Populate input
	agentConfig := AgentConfigInput{
		AgentName:         templateConfig.AgentName,
		AgentSystemPrompt: templateConfig.AgentSystemPrompt,
		Config:            templateConfig.Config,
		IsActive:          templateConfig.IsActive,
		IsPublic:          true,
		Record:            templateConfig.Record,
		CallbackURL:       templateConfig.Callback_url,
		CallbackEvents:    templateConfig.Callback_events,
		KnowledgeBase:     templateConfig.KnowledgeBase,
		MCP:               templateConfig.MCP,
		Tool:              templateConfig.Tool,
		Avatars:           avatarsJSON,
		Email:             tempConfig.Email,
		Type:              tempConfig.Type,
		AddOns:            tempConfig.AddOns,
	}
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	id, err := CreateAgentConfig(tx, agentConfig, apiKeyId)
	if err != nil {
		log.Println("Error creating agent from template:", err)
		WriteInternalServerError(w, "Error creating agent from template")
		return
	}

	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Agent created successfully",
	})
}

// GetTemplateConfigById retrieves a specific Template by ID as AgentConfig
func GetTemplateConfigById(templateConfigID string, apiKeyId string) (*AgentConfig, error) {
	log.Printf("Fetching template config with ID: %s", templateConfigID)

	query := `
        SELECT id, template_name, template_system_prompt, config, is_active, record,
               COALESCE(callback_url, '') AS callback_url, COALESCE(callback_events, '{}') AS callback_events
        FROM templates
        WHERE id = $1`

	var (
		id, name, prompt string
		configJSON       []byte
		isActive, record bool
		callbackURL      string
		callbackEvents   []string
	)

	err := DB.QueryRow(query, templateConfigID).Scan(
		&id, &name, &prompt, &configJSON, &isActive, &record, &callbackURL, pq.Array(&callbackEvents),
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no template found with ID: %s", templateConfigID)
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving template config: %v", err)
	}

	// --- Fetch knowledge bases ---
	kbQuery := `
        SELECT kb.id, kb.name, tk.mode
        FROM templates_kb tk
        JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
        WHERE tk.template_id = $1`

	rows, err := DB.Query(kbQuery, templateConfigID)
	if err != nil {
		return nil, fmt.Errorf("error fetching knowledge bases: %v", err)
	}
	defer rows.Close()

	var knowledgeBases []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Mode string `json:"mode"`
	}
	for rows.Next() {
		var kb struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Mode string `json:"mode"`
		}
		if err := rows.Scan(&kb.ID, &kb.Name, &kb.Mode); err != nil {
			return nil, fmt.Errorf("error scanning knowledge base row: %v", err)
		}
		knowledgeBases = append(knowledgeBases, kb)
	}

	// --- Fetch tools ---
	toolQuery := `
        SELECT t.id, t.name
        FROM templates_tool tt
        JOIN tools t ON t.id = tt.tool_id
        WHERE tt.template_id = $1`

	rows1, err := DB.Query(toolQuery, templateConfigID)
	if err != nil {
		return nil, fmt.Errorf("error fetching tools: %v", err)
	}
	defer rows1.Close()

	var toolsList []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	for rows1.Next() {
		var tool struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := rows1.Scan(&tool.ID, &tool.Name); err != nil {
			return nil, fmt.Errorf("error scanning tool row: %v", err)
		}
		toolsList = append(toolsList, tool)
	}

	// --- Fetch MCPs ---
	mcpQuery := `
        SELECT t.id, t.name
        FROM templates_mcp tt
        JOIN mcps t ON t.id = tt.mcp_id
        WHERE tt.template_id = $1`

	rows2, err := DB.Query(mcpQuery, templateConfigID)
	if err != nil {
		return nil, fmt.Errorf("error fetching MCPs: %v", err)
	}
	defer rows2.Close()

	var mcps []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	for rows2.Next() {
		var mcp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := rows2.Scan(&mcp.ID, &mcp.Name); err != nil {
			return nil, fmt.Errorf("error scanning MCP row: %v", err)
		}
		mcps = append(mcps, mcp)
	}

	// --- Build AgentConfig struct ---
	agentConfig := &AgentConfig{
		AgentName:         name,
		AgentSystemPrompt: prompt,
		Config:            json.RawMessage(configJSON),
		IsActive:          isActive,
		KnowledgeBase:     knowledgeBases,
		Tool:              toolsList,
		MCP:               mcps,
		Record:            record,
		Callback_url:      callbackURL,
		Callback_events:   callbackEvents,
		// Avatars and IsPublic can be filled separately if needed
	}

	return agentConfig, nil
}

// HandleUpdateAgentConfig updates an existing Agent
func HandleUpdateAgentConfig(w http.ResponseWriter, r *http.Request, agentConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Agent config with ID: %s", agentConfigID)

	var agentConfig struct {
		AgentName           string          `json:"agent_name"`
		AgentSystemPrompt   string          `json:"agent_system_prompt"`
		DefaultSystemPrompt bool            `json:"default_system_prompt"`
		Config              json.RawMessage `json:"config"`
		IsActive            bool            `json:"is_active"`
		Avatars             json.RawMessage `json:"avatars"`
		IsPublic            bool            `json:"is_public"`
		KnowledgeBase       []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Mode string `json:"mode"`
		} `json:"knowledge_base"`
		MCP []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"mcp"`
		Tool []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tool"`
		Integration []struct {
			ID string `json:"id"`
		} `json:"integration"`
		Record          bool            `json:"record"`
		Callback_url    string          `json:"callback_url"`
		Callback_events []string        `json:"callback_events" db:"callback_events"`
		Email           string          `json:"email"`
		Type            string          `json:"type"`
		AddOns          json.RawMessage `json:"add_on"`
	}

	if err := json.NewDecoder(r.Body).Decode(&agentConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}
	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM agents a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, agentConfigID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving agent: %v", err)
		WriteInternalServerError(w, "Failed to retrieve agent")
		return
	}

	if !exists {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return
	}

	agentConfig.AddOns = json.RawMessage([]byte("[]"))
	agentConfig.Type = "etev"
	//agentConfig.IsActive = true
	agentConfig.IsPublic = true
	agentConfig.Config = NormalizeConfig(agentConfig.Config)
	agentConfig.Avatars = NormalizeAvatars(agentConfig.Avatars)

	err = ValidateTimeout(apiKeyId, agentConfig.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate timeout : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	// Start transaction
	tx, errtran := DB.Begin()
	if errtran != nil {
		log.Printf("Error starting transaction: %v", errtran)
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update agents
	query := `
        UPDATE agents
        SET agent_name = $1,
            agent_system_prompt = $2,
			default_system_prompt = $3,
			tools = $4,
			config = $5,
			is_active = $6,
			avatars = $7,
            updated_at = CURRENT_TIMESTAMP,
			is_public = $8,
			record = $10,
			callback_url = $11,
			callback_events = $12,
			email = $13,
			type = $14,
			add_ons = $15
		WHERE id = $9
        RETURNING id`

	var id string
	err = tx.QueryRow(
		query,
		agentConfig.AgentName,
		agentConfig.AgentSystemPrompt,
		agentConfig.DefaultSystemPrompt,
		"{}",
		agentConfig.Config,
		agentConfig.IsActive,
		agentConfig.Avatars,
		agentConfig.IsPublic,
		agentConfigID,
		agentConfig.Record,
		agentConfig.Callback_url,
		pq.Array(agentConfig.Callback_events),
		agentConfig.Email,
		agentConfig.Type,
		agentConfig.AddOns,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return
	} else if err != nil {
		log.Printf("Error updating agent config: %v", err)
		WriteInternalServerError(w, "Failed to update agent")
		return
	}

	// Delete existing knowledge base links
	_, err = tx.Exec(`DELETE FROM agents_kb WHERE agent_id = $1`, agentConfigID)
	if err != nil {
		log.Printf("Error deleting old agent_kb entries: %v", err)
		WriteInternalServerError(w, "Failed to update knowledge base")
		return
	}
	// Insert new knowledge base links
	kbInsert := `
        INSERT INTO agents_kb (id, agent_id, knowledge_base_id, mode)
        VALUES (gen_random_uuid(), $1, $2, $3)`
	for _, kb := range agentConfig.KnowledgeBase {
		_, err := tx.Exec(kbInsert, agentConfigID, kb.ID, kb.Mode)
		if err != nil {
			log.Printf("Error inserting agents_kb: %v", err)
			WriteInternalServerError(w, "Failed to insert knowledge base")
			return
		}
	}
	// Delete existing tool links
	_, err = tx.Exec(`DELETE FROM agents_tool WHERE agent_id = $1`, agentConfigID)
	if err != nil {
		log.Printf("Error deleting old agent_tool entries: %v", err)
		WriteInternalServerError(w, "Failed to update Tool")
		return
	}
	// Insert new tool links
	toolInsert := `
        INSERT INTO agents_tool (id, agent_id, tool_id)
        VALUES (gen_random_uuid(), $1, $2)`
	for _, tool := range agentConfig.Tool {
		_, err := tx.Exec(toolInsert, agentConfigID, tool.ID)
		if err != nil {
			log.Printf("Error inserting agents_kb: %v", err)
			WriteInternalServerError(w, "Failed to add Tool")
			return
		}
	}
	// Delete existing mcp links
	_, err = tx.Exec(`DELETE FROM agents_mcp WHERE agent_id = $1`, agentConfigID)
	if err != nil {
		log.Printf("Error deleting old agent_mcp entries: %v", err)
		WriteInternalServerError(w, "Failed to update MCP")
		return
	}
	// Insert new MCP links
	mcpInsert := `
        INSERT INTO agents_mcp (id, agent_id, mcp_id)
        VALUES (gen_random_uuid(), $1, $2)`
	for _, mcp := range agentConfig.MCP {
		_, err := tx.Exec(mcpInsert, agentConfigID, mcp.ID)
		if err != nil {
			log.Printf("Error inserting agents_mcp: %v", err)
			WriteInternalServerError(w, "Failed to add MCP")
			return
		}
	}
	// Delete existing integrations
	_, err = tx.Exec(`DELETE FROM agents_integration WHERE agent_id = $1`, agentConfigID)
	if err != nil {
		log.Printf("Error deleting old agent_integrations entries: %v", err)
		WriteInternalServerError(w, "Failed to update integrations")
		return
	}
	// Insert new Integrations
	integrationInsert := `
        INSERT INTO agents_integration (id, agent_id, auth_config_id)
        VALUES (gen_random_uuid(), $1, $2)`
	for _, integration := range agentConfig.Integration {
		_, err := tx.Exec(integrationInsert, agentConfigID, integration.ID)
		if err != nil {
			log.Printf("Error inserting agents_integration: %v", err)
			WriteInternalServerError(w, "Failed to add integration")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// Success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Agent updated successfully",
	})
}

// HandleGetAvatarDetailsByAgent retrieves Avatar details by Agent ID
func HandleGetAvatarDetailsByAgent(w http.ResponseWriter, agentID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching agent config with ID: %s", agentID)

	query := `
        SELECT avatars
        FROM agents
        WHERE id = $1`

	var (
		avatar string
	)

	err := DB.QueryRow(query, agentID).Scan(
		&avatar,
	)
	//log.Printf("Avatar config: %s", avatar)
	var avatars []Avatar

	// Unmarshal JSON string into slice of Avatar
	err1 := json.Unmarshal([]byte(avatar), &avatars)
	if err1 != nil {
		log.Printf("failed to unmarshal avatars array: %v", err)
		return
	}

	if err == sql.ErrNoRows {
		log.Printf("No agent found with ID: %s", agentID)
		WriteNotFoundError(w, "Agent not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving agent config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve agent")
		return
	}
	//log.Printf("Avatar config: %v", &avatars[0].AvatarID)

	for i := range avatars {
		var profilePic, ImageURL, gender, name string
		queryAvatars := `SELECT display_picture, image_url, gender, avatar_name FROM avatars WHERE avatar_key_id = $1`
		err = DB.QueryRow(queryAvatars, avatars[i].AvatarID).Scan(&profilePic, &ImageURL, &gender, &name)
		if err == sql.ErrNoRows {
			WriteNotFoundError(w, "Avatar not found")
			return
		} else if err != nil {
			WriteInternalServerError(w, "Failed to fetch avatar profile pic")
			return
		}
		avatars[i].AvatarProfilePic = profilePic
		avatars[i].Gender = gender
		avatars[i].AvatarName = name
	}

	AgentConfig := map[string]interface{}{
		"avatars": avatars,
	}

	json.NewEncoder(w).Encode(AgentConfig)
}

// HandleGetAgentStatus retrieves a specific Agent by ID
func HandleGetAgentStatus(w http.ResponseWriter, agentConfigID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching agent config with ID: %s", agentConfigID)

	query := `
        SELECT id, is_public
        FROM agents
        WHERE id = $1`

	var (
		id, is_public string
	)

	err := DB.QueryRow(query, agentConfigID).Scan(
		&id, &is_public,
	)

	if err == sql.ErrNoRows {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving agent config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve agent")
		return
	}

	AgentConfig := map[string]interface{}{
		"id":        id,
		"is_public": is_public,
	}

	json.NewEncoder(w).Encode(AgentConfig)
}

// HandleGetAgentConfig retrieves a specific Agent by ID
func HandleGetAgentConfig(w http.ResponseWriter, agentConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching agent config with ID: %s", agentConfigID)

	AgentConfig, shouldReturn := getAgentFunction(agentConfigID, apiKeyId, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(AgentConfig)
}

// HandleGetAgentConfig retrieves a specific Agent by ID
func HandleGetAgentConfigAPI(w http.ResponseWriter, agentConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching agent config with ID: %s", agentConfigID)

	AgentConfig, shouldReturn := getAgentFunction(agentConfigID, apiKeyId, w)
	if shouldReturn {
		return
	}

	AgentConfigAPI := map[string]interface{}{
		"id":                    AgentConfig["id"],
		"agent_name":            AgentConfig["agent_name"],
		"agent_system_prompt":   AgentConfig["agent_system_prompt"],
		"default_system_prompt": AgentConfig["default_system_prompt"],
		"config":                AgentConfig["config"],
		"avatars":               AgentConfig["avatars"],
		"created_at":            AgentConfig["created_at"],
		"updated_at":            AgentConfig["updated_at"],
		"knowledge_base":        AgentConfig["knowledge_base"],
		"mcp":                   AgentConfig["mcp"],
		"tool":                  AgentConfig["tool"],
		"integration":           AgentConfig["integration"],
		"record":                AgentConfig["record"],
		"callback_url":          AgentConfig["callback_url"],
		"callback_events":       AgentConfig["callback_events"],
	}

	json.NewEncoder(w).Encode(AgentConfigAPI)
}

func getAgentFunction(agentConfigID string, apiKeyId string, w http.ResponseWriter) (map[string]interface{}, bool) {
	query := `
	SELECT a.id, agent_name, agent_system_prompt, default_system_prompt, tools, config, a.is_active, avatars, a.created_at, a.updated_at, a.created_by, a.is_public,  record, COALESCE(callback_url, '') AS callback_url, COALESCE(callback_events, '{}') AS callback_events, COALESCE(email, '') AS email, type, add_ons
	FROM agents a
		JOIN api_keys ak2
			ON a.created_by = ak2.id                       
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id         
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id                    
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id                    
			AND w2.organization_id = w1.organization_id    
		WHERE ak1.id = $2 and a.id = $1`

	var (
		id, name, prompt, atype         string
		isActive, isPublic, record      bool
		defaultSystemPrompt             bool
		createdAt, updatedAt            time.Time
		createdBy, callback_url, email  string
		callback_events                 []string
		add_ons, tools, config, avatars json.RawMessage
	)

	err := DB.QueryRow(query, agentConfigID, apiKeyId).Scan(
		&id, &name, &prompt, &defaultSystemPrompt, &tools,
		&config, &isActive, &avatars, &createdAt, &updatedAt, &createdBy, &isPublic, &record, &callback_url, pq.Array(&callback_events), &email, &atype, &add_ons,
	)

	if err == sql.ErrNoRows {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error retrieving agent config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve agent")
		return nil, true
	}

	// Fetch linked knowledge bases
	kbQuery := `
        SELECT kb.id, kb.name, kb.description, kb.namespace, kb.index, tk.mode
        FROM agents_kb tk
        JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
        WHERE tk.agent_id = $1`
	rows, err := DB.Query(kbQuery, agentConfigID)
	if err != nil {
		log.Printf("Error fetching knowledge bases: %v", err)
		WriteInternalServerError(w, "Failed to fetch linked knowledge bases")
		return nil, true
	}
	defer rows.Close()
	var knowledgeBases []map[string]string
	for rows.Next() {
		var kbID, kbName, kbDescription, kbNamespace, kbIndex, kbMode string
		if err := rows.Scan(&kbID, &kbName, &kbDescription, &kbNamespace, &kbIndex, &kbMode); err != nil {
			log.Printf("Error scanning knowledge base row: %v", err)
			WriteInternalServerError(w, "Failed to parse knowledge base")
			return nil, true
		}
		knowledgeBases = append(knowledgeBases, map[string]string{
			"id":          kbID,
			"name":        kbName,
			"description": kbDescription,
			"namespace":   kbNamespace,
			"index":       kbIndex,
			"mode":        kbMode,
		})
	}
	// Fetch linked tools
	toolQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.request_config, t.event_messages
        FROM agents_tool tt
        JOIN tools t ON t.id = tt.tool_id
        WHERE tt.agent_id = $1`
	rows1, err1 := DB.Query(toolQuery, agentConfigID)
	if err1 != nil {
		log.Printf("Error fetching tools: %v", err1)
		WriteInternalServerError(w, "Failed to fetch linked tools")
		return nil, true
	}
	defer rows1.Close()
	var tool []map[string]string
	for rows1.Next() {
		var toolID, toolName, toolDescription, tooltype, arguments, request_config, event_messages string
		if err := rows1.Scan(&toolID, &toolName, &toolDescription, &tooltype, &arguments, &request_config, &event_messages); err != nil {
			log.Printf("Error scanning tools row: %v", err)
			WriteInternalServerError(w, "Failed to parse tools")
			return nil, true
		}
		tool = append(tool, map[string]string{
			"id":             toolID,
			"type":           tooltype,
			"schema":         arguments,
			"request_config": request_config,
			"event_messages": event_messages,
		})
	}
	// Fetch linked MCPs
	mcpQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.cache_tools_list, t.event_messages
        FROM agents_mcp tt
        JOIN mcps t ON t.id = tt.mcp_id
        WHERE tt.agent_id = $1`
	rows2, err2 := DB.Query(mcpQuery, agentConfigID)
	if err2 != nil {
		log.Printf("Error fetching mcp: %v", err2)
		WriteInternalServerError(w, "Failed to fetch linked mcp")
		return nil, true
	}
	defer rows2.Close()
	var mcp []map[string]string
	for rows2.Next() {
		var mcpID, mcpName, mcpDescription, mcptype, arguments, cache_tools_list, event_messages string
		if err := rows2.Scan(&mcpID, &mcpName, &mcpDescription, &mcptype, &arguments, &cache_tools_list, &event_messages); err != nil {
			log.Printf("Error scanning mcp row: %v", err)
			WriteInternalServerError(w, "Failed to parse MCP")
			return nil, true
		}
		mcp = append(mcp, map[string]string{
			"id":             mcpID,
			"name":           mcpName,
			"description":    mcpDescription,
			"type":           mcptype,
			"request_config": arguments,
			//"cache_tools_list": cache_tools_list,
			"event_messages": event_messages,
		})
	}

	// Fetch linked Integrations
	integrationQuery := `
        SELECT t.auth_config_id, t.user_id, t.toolkit_slug, t.account_id, t.status
        FROM agents_integration tt
        JOIN integrations_config t ON t.auth_config_id = tt.auth_config_id
        WHERE tt.agent_id = $1`
	rows3, err3 := DB.Query(integrationQuery, agentConfigID)
	if err3 != nil {
		log.Printf("Error fetching integration: %v", err3)
		WriteInternalServerError(w, "Failed to fetch linked integration")
		return nil, true
	}
	defer rows3.Close()
	var integration []map[string]string
	for rows3.Next() {
		var auth_config_id, user_id, toolkit_slug, account_id, status string
		if err := rows3.Scan(&auth_config_id, &user_id, &toolkit_slug, &account_id, &status); err != nil {
			log.Printf("Error scanning integration row: %v", err)
			WriteInternalServerError(w, "Failed to parse Integration")
			return nil, true
		}
		integration = append(integration, map[string]string{
			"id":         auth_config_id,
			"slug":       toolkit_slug,
			"user_id":    user_id,
			"account_id": account_id,
			"status":     status,
		})
	}

	AgentConfig := map[string]interface{}{
		"id":                    id,
		"agent_name":            name,
		"agent_system_prompt":   prompt,
		"default_system_prompt": defaultSystemPrompt,
		"tools":                 tools,
		"config":                config,
		"is_active":             isActive,
		"avatars":               avatars,
		"created_at":            createdAt,
		"updated_at":            updatedAt,
		"is_public":             isPublic,
		"knowledge_base":        knowledgeBases,
		"tool":                  tool,
		"mcp":                   mcp,
		"integration":           integration,
		"record":                record,
		"callback_url":          callback_url,
		"callback_events":       callback_events,
		"email":                 email,
		"type":                  atype,
		"add_on":                add_ons,
	}
	return AgentConfig, false
}

// HandleGetAllAgentConfigsAPI retrieves all Agents
func HandleGetAllAgentConfigsAPI(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")

	agentConfigs, shouldReturn := GetAllAgentFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	var agentConfigsAPI []map[string]interface{}

	for _, agent := range agentConfigs {

		AgentConfigAPI := map[string]interface{}{
			"id":                    agent["id"],
			"agent_name":            agent["agent_name"],
			"agent_system_prompt":   agent["agent_system_prompt"],
			"default_system_prompt": agent["default_system_prompt"],
			"config":                agent["config"],
			"avatars":               agent["avatars"],
			"created_at":            agent["created_at"],
			"updated_at":            agent["updated_at"],
			"knowledge_base":        agent["knowledge_base"],
			"tool":                  agent["tool"],
			"mcp":                   agent["mcp"],
			"integration":           agent["integration"],
			"record":                agent["record"],
			"callback_url":          agent["callback_url"],
			"callback_events":       agent["callback_events"],
		}

		agentConfigsAPI = append(agentConfigsAPI, AgentConfigAPI)
	}

	json.NewEncoder(w).Encode(agentConfigsAPI)
}

// HandleGetAllAgentConfigs retrieves all Agents
func HandleGetAllAgentConfigs(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	agentConfigs, shouldReturn := GetAllAgentFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(agentConfigs)
}

func GetAllAgentFunction(apiKeyId string, w http.ResponseWriter) ([]map[string]interface{}, bool) {
	query := `
        SELECT
			a.id,
			a.agent_name,
			a.agent_system_prompt,
			a.default_system_prompt,
			a.tools,
			a.config,
			a.is_active,
			a.avatars,
			a.created_at,
			a.updated_at,
			a.created_by,
			a.is_public,
			a.record,
			COALESCE(a.callback_url, '') AS callback_url,
			COALESCE(a.callback_events, '{}') AS callback_events,
			COALESCE(a.email, '') AS email,
			a.type,
			a.add_ons
		FROM agents a
		JOIN api_keys ak2
			ON a.created_by = ak2.id                       -- creator API key
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id         -- workspace match
		-- Workspace → Organization filtering
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id                    -- requester workspace
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id                    -- creator workspace
			AND w2.organization_id = w1.organization_id    -- ✅ SAME org
		WHERE ak1.id = $1 and a.is_active = true
		ORDER BY a.created_at DESC;
		`

	rows, err := DB.Query(query, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve agents")
		return nil, true
	}
	defer rows.Close()

	var agentConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, name, prompt, callback_url, atype string
			isActive, isPublic, record            bool
			defaultSystemPrompt                   bool
			createdAt, updatedAt                  time.Time
			createdBy, email                      string
			callback_events                       []string
			add_ons, tools, config, avatars       json.RawMessage
		)

		if err := rows.Scan(
			&id, &name, &prompt, &defaultSystemPrompt, &tools,
			&config, &isActive, &avatars, &createdAt, &updatedAt, &createdBy, &isPublic, &record, &callback_url, pq.Array(&callback_events), &email, &atype, &add_ons,
		); err != nil {
			WriteInternalServerError(w, "Error scanning Agents")
			return nil, true
		}

		// Fetch knowledge bases for this template
		kbQuery := `
            SELECT kb.id, kb.name, kb.description, kb.namespace, kb.index, tk.mode
            FROM agents_kb tk
            JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
            WHERE tk.agent_id = $1`
		kbRows, err := DB.Query(kbQuery, id)
		if err != nil {
			WriteInternalServerError(w, "Failed to fetch linked knowledge bases")
			return nil, true
		}
		var knowledgeBases []map[string]string
		for kbRows.Next() {
			var kbID, kbName, kbDescription, kbNamespace, kbIndex, kbMode string
			if err := kbRows.Scan(&kbID, &kbName, &kbDescription, &kbNamespace, &kbIndex, &kbMode); err != nil {
				WriteInternalServerError(w, "Failed to parse knowledge base")
				return nil, true
			}
			knowledgeBases = append(knowledgeBases, map[string]string{
				"id":          kbID,
				"name":        kbName,
				"description": kbDescription,
				"namespace":   kbNamespace,
				"index":       kbIndex,
				"mode":        kbMode,
			})
		}
		kbRows.Close()
		// Fetch linked tools
		toolQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.request_config, t.event_messages
        FROM agents_tool tt
        JOIN tools t ON t.id = tt.tool_id
        WHERE tt.agent_id = $1`
		rows1, err1 := DB.Query(toolQuery, id)
		if err1 != nil {
			log.Printf("Error fetching tools: %v", err1)
			WriteInternalServerError(w, "Failed to fetch linked tools")
			return nil, true
		}
		var tool []map[string]string
		for rows1.Next() {
			var toolID, toolName, toolDescription, tooltype, arguments, request_config, event_messages string
			if err := rows1.Scan(&toolID, &toolName, &toolDescription, &tooltype, &arguments, &request_config, &event_messages); err != nil {
				log.Printf("Error scanning tools row: %v", err)
				WriteInternalServerError(w, "Failed to parse tools")
				return nil, true
			}
			tool = append(tool, map[string]string{
				"id":             toolID,
				"type":           tooltype,
				"schema":         arguments,
				"request_config": request_config,
				"event_messages": event_messages,
			})
		}
		defer rows1.Close()
		// Fetch linked MCPs
		mcpQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.cache_tools_list, t.event_messages
        FROM agents_mcp tt
        JOIN mcps t ON t.id = tt.mcp_id
        WHERE tt.agent_id = $1`
		rows2, err2 := DB.Query(mcpQuery, id)
		if err2 != nil {
			log.Printf("Error fetching mcp: %v", err2)
			WriteInternalServerError(w, "Failed to fetch linked mcp")
			return nil, true
		}
		var mcp []map[string]string
		for rows2.Next() {
			var mcpID, mcpName, mcpDescription, mcptype, arguments, cache_tools_list, event_messages string
			if err := rows2.Scan(&mcpID, &mcpName, &mcpDescription, &mcptype, &arguments, &cache_tools_list, &event_messages); err != nil {
				log.Printf("Error scanning mcp row: %v", err)
				WriteInternalServerError(w, "Failed to parse MCP")
				return nil, true
			}
			mcp = append(mcp, map[string]string{
				"id":             mcpID,
				"name":           mcpName,
				"description":    mcpDescription,
				"type":           mcptype,
				"request_config": arguments,
				//"cache_tools_list": cache_tools_list,
				"event_messages": event_messages,
			})
		}
		defer rows2.Close()

		// Fetch linked Integrations
		integrationQuery := `
         SELECT t.auth_config_id, t.user_id, t.toolkit_slug, t.account_id, t.status
        FROM agents_integration tt
        JOIN integrations_config t ON t.auth_config_id = tt.auth_config_id
        WHERE tt.agent_id = $1`
		rows3, err3 := DB.Query(integrationQuery, id)
		if err3 != nil {
			log.Printf("Error fetching integration: %v", err3)
			WriteInternalServerError(w, "Failed to fetch linked integration")
			return nil, true
		}
		var integration []map[string]string
		for rows3.Next() {
			var auth_config_id, user_id, toolkit_slug, account_id, status string
			if err := rows3.Scan(&auth_config_id, &user_id, &toolkit_slug, &account_id, &status); err != nil {
				log.Printf("Error scanning integration row: %v", err)
				WriteInternalServerError(w, "Failed to parse Integration")
				return nil, true
			}
			integration = append(integration, map[string]string{
				"id":         auth_config_id,
				"slug":       toolkit_slug,
				"user_id":    user_id,
				"account_id": account_id,
				"status":     status,
			})
		}
		defer rows3.Close()

		agentConfigs = append(agentConfigs, map[string]interface{}{
			"id":                    id,
			"agent_name":            name,
			"agent_system_prompt":   prompt,
			"default_system_prompt": defaultSystemPrompt,
			"tools":                 tools,
			"config":                config,
			"is_active":             isActive,
			"avatars":               avatars,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
			"is_public":             isPublic,
			"knowledge_base":        knowledgeBases,
			"tool":                  tool,
			"mcp":                   mcp,
			"integration":           integration,
			"record":                record,
			"callback_url":          callback_url,
			"callback_events":       callback_events,
			"email":                 email,
			"type":                  atype,
			"add_on":                add_ons,
		})
	}
	return agentConfigs, false
}

// HandleDeleteAgentConfig deletes an Agent
func HandleDeleteAgentConfig(w http.ResponseWriter, agentConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting agent config with ID: %s", agentConfigID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM agents a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, agentConfigID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving agent: %v", err)
		WriteInternalServerError(w, "Failed to retrieve agent")
		return
	}

	if !exists {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Soft delete agent (update is_active = false)
	query := `
		UPDATE agents
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`

	var id string
	err = tx.QueryRow(query, agentConfigID).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No agent found with ID: %s", agentConfigID)
		WriteNotFoundError(w, "Agent not found")
		return
	} else if err != nil {
		log.Printf("Error soft deleting agent config: %v", err)
		WriteInternalServerError(w, "Failed to delete agent")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}
	// Success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Agent deleted successfully",
	})
}

// handle ScriptToVideo
func HandleScriptToVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.Header.Get("X-API-Key")

	// Parse and Validate
	requestPayload := ScriptToVideoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		WriteBadRequestError(w, "Invalid JSON payload")
		return
	}
	if err := requestPayload.Validate(); err != nil {
		WriteBadRequestError(w, err.Error())
		return
	}

	// generates video generationId, converstionId, calculates time based on text
	generationID := uuid.New()
	conversationID := uuid.New()
	estimatedSeconds := TextToTime(requestPayload.Script)
	estimatedMinutes := int(estimatedSeconds/60) + 1 // ceiling minutes for credit calculation

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Validate credits
	var status string

	err = tx.QueryRow(`
		SELECT o_status, o_api_key_id
		FROM validate_s2v_request($1, $2, $3)
	`, apiKey, generationID, int(estimatedMinutes)).Scan(&status, &conversationID)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Credit validation failed: %v", err))
		return
	}

	switch status {
	case "NO_API_KEY_FOUND":
		WriteBadRequestError(w, "Invalid API key")
		return
	case "NOT_ENOUGH_SESSION_LEFT":
		WriteBadRequestError(w, "Concurrent session limit reached")
		return
	case "NOT_ENOUGH_CREDIT_LEFT":
		WriteBadRequestError(w, "Not enough credits")
		return
	case "REQUEST_APPROVED":
	default:
		WriteBadRequestError(w, status)
		return
	}

	// Insert to DB
	_, err = tx.Exec(`
		INSERT INTO script_to_video (generation_id, avatar_id, voice_id, provider_name, model_name, script, callback_url, status, created_at, modified_at, s3_url, conversation_id, estimated_time, video_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW(), '', $8, $9, '')`,
		generationID, requestPayload.AvatarID, requestPayload.VoiceID, requestPayload.ProviderName, requestPayload.ModelName, requestPayload.Script, requestPayload.CallbackURL, conversationID, estimatedSeconds)
	if err != nil {
		WriteInternalServerError(w, "Failed to save request")
		return
	}

	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}
	committed = true

	delay := time.Duration(estimatedMinutes+5) * time.Minute
	ScheduleOneTimeResetCreditJob(delay, generationID.String())

	infraPayload := map[string]interface{}{
		"input": map[string]interface{}{
			"avatar_id":    requestPayload.AvatarID,
			"callback_url": requestPayload.CallbackURL,
			"input_text":   requestPayload.Script,
			"job_id":       generationID,
			"tts_config": map[string]interface{}{
				"provider": requestPayload.ProviderName,
				"model":    requestPayload.ModelName,
				"voice_id": requestPayload.VoiceID,
			},
		},
	}

	_, err = infra.PostJob(
		configs.GetEnv("S2V_INFRA_POST_JOB_URL"),
		infraPayload,
		configs.GetEnv("S2V_INFRA_AUTH_BEARER"),
	)
	if err != nil {
		DB.Exec(`UPDATE script_to_video SET status = 'failed', modified_at = NOW() WHERE generation_id = $1`, generationID)
		log.Printf("infra.PostJob failed for generation %s: %v", generationID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"generation_id": generationID,
			"status":        "failed",
		})
		return
	}

	DB.Exec(`UPDATE script_to_video SET status = 'processing', modified_at = NOW() WHERE generation_id = $1`, generationID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"generation_id":      generationID,
		"estimated_duration": estimatedSeconds,
		"status":             "processing",
	})
}

// update usage - Called by infra
func HandleUpdateScriptToVideoUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request_payload struct {
		GenerationID   string  `json:"process_id"`
		Status         string  `json:"status"`
		ActualDuration float64 `json:"video_duration"`
		S3URL          string  `json:"video_url"`
		VideoName      string  `json:"video_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request_payload); err != nil {
		WriteBadRequestError(w, "Invalid JSON")
		return
	}

	if request_payload.GenerationID == "" || (request_payload.Status != "completed" && request_payload.Status != "failed") {
		WriteBadRequestError(w, "Invalid generation_id or status")
		return
	}

	// Strip presign query params so we store a stable base S3 URL
	if request_payload.S3URL != "" {
		if parsedURL, parseErr := url.Parse(request_payload.S3URL); parseErr == nil {
			parsedURL.RawQuery = ""
			request_payload.S3URL = parsedURL.String()
		}
	}

	actualDuration := request_payload.ActualDuration
	if request_payload.Status == "failed" {
		actualDuration = 0
	}

	result, err := DB.Exec(`
		UPDATE script_to_video
		SET status = $1, s3_url = $2, actual_time = $3, modified_at = NOW(), video_name = $5
		WHERE generation_id = $4 AND status NOT IN ('completed', 'failed')`,
		request_payload.Status, request_payload.S3URL, actualDuration, request_payload.GenerationID, request_payload.VideoName)
	if err != nil {
		log.Printf("script_to_video update failed for generation %s: %v", request_payload.GenerationID, err)
		WriteInternalServerError(w, "Update failed")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		WriteNotFoundError(w, "Record not found or already finalized")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// get genstatus
func HandleGetGenStatus(w http.ResponseWriter, r *http.Request, generationId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing X-API-Key"})
		return
	}

	var (
		generation_id, avatarID, voiceID, providerName, modelName, script, callbackURL, s3URL, status string
		createdAt, modifiedAt                                                                         time.Time
		estimatedTime, actualTime                                                                     sql.NullFloat64
		videoName                                                                                     sql.NullString
	)

	err := DB.QueryRow(`
		SELECT stv.generation_id, stv.avatar_id, stv.voice_id, stv.provider_name, stv.model_name, stv.script, stv.callback_url, stv.status, stv.created_at, stv.modified_at, stv.s3_url, stv.estimated_time, stv.actual_time, stv.video_name
		FROM script_to_video stv
		JOIN api_keys ak ON ak.id = stv.conversation_id
		WHERE stv.generation_id = $1 AND ak.key_hash = $2`, generationId, apiKey).
		Scan(&generation_id, &avatarID, &voiceID, &providerName, &modelName, &script, &callbackURL, &status, &createdAt, &modifiedAt, &s3URL, &estimatedTime, &actualTime, &videoName)

	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Generation ID not found")
		return
	}
	if err != nil {
		WriteInternalServerError(w, "Failed to fetch record")
		return
	}

	// Generate a fresh presigned URL when the job is completed; empty string otherwise
	freshVideoURL := ""
	if status == "completed" && s3URL != "" {
		region := configs.GetEnv("AWS_REGION")
		videoBucket := configs.GetEnv("AWS_BUCKET_ADDITIONAL")
		presignedURL, presignErr := GenerateVideoPresignedURL(s3URL, region, videoBucket)
		if presignErr != nil {
			log.Printf("Failed to generate presigned video URL for generation %s: %v", generationId, presignErr)
			WriteInternalServerError(w, "Video ready but failed to generate download URL")
			return
		}
		freshVideoURL = presignedURL
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"generation_id":  generation_id,
		"avatar_id":      avatarID,
		"voice_id":       voiceID,
		"provider_name":  providerName,
		"model_name":     modelName,
		"script":         script,
		"callback_url":   callbackURL,
		"status":         status,
		"created_at":     createdAt,
		"modified_at":    modifiedAt,
		"video_url":      freshVideoURL,
		"video_name":     videoName.String,
		"estimated_time": estimatedTime.Float64,
		"actual_time":    actualTime.Float64,
	})
}

// TextToTime returns estimated duration in seconds based on word count at 143 WPM
func TextToTime(script string) float64 {
	wordCount := len(strings.Fields(script))
	if wordCount == 0 {
		return 0
	}
	seconds := (float64(wordCount) / 143.0) * 60
	return float64(int(seconds*100)) / 100
}

func (r *ScriptToVideoRequest) Validate() error {
	if r.AvatarID == "" {
		return errors.New("agent_id is required")
	}
	if r.VoiceID == "" {
		return errors.New("voice_id is required")
	}
	if r.Script == "" {
		return errors.New("script is required")
	}
	if r.CallbackURL == "" {
		return errors.New("callback_url is required")
	}
	return nil
}

// HandleWorkspaceUserInvitation creates a new user invitation for a workspace
func HandleWorkspaceUserInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var invite struct {
		UserEmail   string `json:"user_email"`
		RoleID      string `json:"role_id"`
		WorkspaceID string `json:"workspace_id"`
		CreatedBy   string `json:"created_by"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&invite); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	// Basic validation
	if invite.UserEmail == "" || invite.RoleID == "" || invite.WorkspaceID == "" {
		WriteBadRequestError(w, "Missing required fields: user_email, role_id, workspace_id")
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 1️⃣ Validate workspace exists
	var workspaceName string
	err = tx.QueryRow(`
    SELECT name
    FROM workspaces
    WHERE id = $1
	`, invite.WorkspaceID).Scan(&workspaceName)

	if err == sql.ErrNoRows {
		WriteBadRequestError(w, "Workspace not found")
		return
	}
	if err != nil {
		log.Printf("Error fetching workspace name: %v", err)
		WriteInternalServerError(w, "Error validating workspace")
		return
	}
	var exists bool
	// 2️⃣ Validate role exists
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, invite.RoleID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking role: %v", err)
		WriteInternalServerError(w, "Error validating role")
		return
	}
	if !exists {
		WriteBadRequestError(w, "Role not found")
		return
	}
	var invitationID string
	var userId string
	// Validate user exists
	err = tx.QueryRow(`SELECT id FROM users WHERE email = $1`, invite.UserEmail).Scan(&userId)
	if err == sql.ErrNoRows {

		// 3️⃣ Check for existing pending invitation
		err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM invitations
			WHERE user_email = $1 AND workspace_id = $2
			AND status = 'Pending' AND expiry > NOW()
		)
	`, invite.UserEmail, invite.WorkspaceID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking existing invitations: %v", err)
			WriteInternalServerError(w, "Failed to verify existing invitations")
			return
		}
		if exists {
			WriteBadRequestError(w, "User already has a pending invitation for this workspace")
			return
		}

		// 4️⃣ Create invitation
		insertQuery := `
		INSERT INTO invitations (
			id, user_email, role_id, workspace_id, status, created_at, created_by, expiry
		) VALUES (
			gen_random_uuid(), $1, $2, $3, 'Pending', CURRENT_TIMESTAMP, $4, NOW() + INTERVAL '10 days'
		) RETURNING id`
		err = tx.QueryRow(insertQuery, invite.UserEmail, invite.RoleID, invite.WorkspaceID, invite.CreatedBy).Scan(&invitationID)
		if err != nil {
			log.Printf("Error inserting invitation: %v", err)
			WriteInternalServerError(w, "Failed to create invitation")
			return
		}

		var inviterName string
		err = tx.QueryRow(`
			SELECT first_name
			FROM users
			WHERE id = $1
			`, invite.CreatedBy).Scan(&inviterName)

		if err == sql.ErrNoRows {
			WriteBadRequestError(w, "Inviter not found")
			return
		}
		if err != nil {
			log.Printf("Error fetching Inviter name: %v", err)
			WriteInternalServerError(w, "Error validating User")
			return
		}

		//Send Email
		//emailHtml := GetEmailTemplateHTML("invitation")
		emailTemplate, err := GetEmailTemplateByName("EMAIL_TEMPLATE_INVITATION")
		if err != nil {
			log.Println("Error fetching email template:", err)
			return
		}
		emailHtml := emailTemplate.EmailContent

		subject := inviterName + " Invited you to Join Project"
		appURL := os.Getenv("STRIPE_PAYMENT_SUCCESS_URL")

		//Replace the required details to the templates
		appURL = strings.ReplaceAll(appURL, "checkout/success", "sign-up/"+invitationID)
		emailHtml = strings.ReplaceAll(emailHtml, "Krishna", invite.UserEmail)
		emailHtml = strings.ReplaceAll(emailHtml, "John Doe", inviterName)
		emailHtml = strings.ReplaceAll(emailHtml, "TRUVIZ INC", workspaceName)
		emailHtml = strings.ReplaceAll(emailHtml, "InvitationURL", appURL)
		err = SendEmail(invite.UserEmail, subject, emailHtml)
		if err != nil {
			log.Println("Email error:", err)
		}

	} else if err != nil {
		log.Printf("Error checking user: %v", err)
		WriteInternalServerError(w, "Error validating user")
		return
	}

	if userId != "" {
		insertMemberQuery := `
			INSERT INTO workspace_members (
				id, user_id, role_id, workspace_id, status,
				created_at, created_by
			) VALUES (
				gen_random_uuid(), $1, $2, $3, 'Active',
				CURRENT_TIMESTAMP, $4
			)RETURNING id`
		err = tx.QueryRow(insertMemberQuery, userId, invite.RoleID, invite.WorkspaceID, invite.CreatedBy).Scan(&invitationID)
		if err != nil {
			log.Printf("Error adding user to workspace_members: %v", err)
			WriteInternalServerError(w, "Failed to add user to workspace")
			return
		}

	}

	// 5️⃣ Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to save invitation")
		return
	}

	// 6️⃣ Return success
	json.NewEncoder(w).Encode(map[string]string{
		"id":      invitationID,
		"message": "Invitation created successfully",
	})
}

// HandleGetAllWorkspaceInvitations retrieves all invitations for a given workspace
func HandleGetAllWorkspaceInvitations(w http.ResponseWriter, workspaceID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT
			id,
			user_email,
			role_id,
			workspace_id,
			status,
			created_at,
			created_by,
			expiry
		FROM invitations
		WHERE workspace_id = $1
		ORDER BY created_at DESC`

	rows, err := DB.Query(query, workspaceID)
	if err != nil {
		log.Printf("Error retrieving invitations: %v", err)
		WriteInternalServerError(w, "Failed to retrieve invitations")
		return
	}
	defer rows.Close()

	var invitations []map[string]interface{}

	for rows.Next() {
		var (
			id, userEmail, roleID, workspaceIDStr, status string
			createdBy                                     sql.NullString
			createdAt, expiry                             time.Time
		)

		if err := rows.Scan(
			&id, &userEmail, &roleID, &workspaceIDStr, &status, &createdAt, &createdBy, &expiry,
		); err != nil {
			log.Printf("Error scanning invitation row: %v", err)
			WriteInternalServerError(w, "Error reading invitation record")
			return
		}

		invitation := map[string]interface{}{
			"id":           id,
			"user_email":   userEmail,
			"role_id":      roleID,
			"workspace_id": workspaceIDStr,
			"status":       status,
			"created_at":   createdAt,
			"expiry":       expiry,
		}

		if createdBy.Valid {
			invitation["created_by"] = createdBy.String
		} else {
			invitation["created_by"] = nil
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error reading invitation rows: %v", err)
		WriteInternalServerError(w, "Error processing invitations")
		return
	}

	json.NewEncoder(w).Encode(invitations)
}

// HandleGetWorkspaceInvitationByID retrieves a single invitation by its ID
func HandleGetWorkspaceInvitationByID(w http.ResponseWriter, invitationID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT
		i.id,
		i.user_email,
		i.role_id,
		i.workspace_id,
		w.name AS workspace_name,
		i.status,
		i.created_at,
		i.created_by,
		i.expiry,
		w.organization_id,
		o.name AS organization_name
	FROM invitations i
	LEFT JOIN workspaces w ON i.workspace_id = w.id
	LEFT JOIN organizations o ON w.organization_id = o.id
	WHERE i.id = $1
	LIMIT 1;`

	row := DB.QueryRow(query, invitationID)

	var (
		id, userEmail, roleID, workspaceID, workspaceName, status string
		organizationID, organizationName                          sql.NullString
		createdBy                                                 sql.NullString
		createdAt, expiry                                         time.Time
	)

	if err := row.Scan(
		&id,
		&userEmail,
		&roleID,
		&workspaceID,
		&workspaceName,
		&status,
		&createdAt,
		&createdBy,
		&expiry,
		&organizationID,
		&organizationName,
	); err != nil {
		if err == sql.ErrNoRows {
			WriteNotFoundError(w, "Invitation not found")
			return
		}
		log.Printf("Error retrieving invitation: %v", err)
		WriteInternalServerError(w, "Failed to retrieve invitation")
		return
	}

	invitation := map[string]interface{}{
		"id":                id,
		"user_email":        userEmail,
		"role_id":           roleID,
		"workspace_id":      workspaceID,
		"workspace_name":    workspaceName,
		"status":            status,
		"created_at":        createdAt,
		"expiry":            expiry,
		"organization_id":   nil,
		"organization_name": nil,
	}

	if organizationID.Valid {
		invitation["organization_id"] = organizationID.String
	}
	if organizationID.Valid {
		invitation["organization_name"] = organizationName.String
	}

	if createdBy.Valid {
		invitation["created_by"] = createdBy.String
	} else {
		invitation["created_by"] = nil
	}

	json.NewEncoder(w).Encode(invitation)
}

// HandleListWorkspaceMembers lists all members for a given workspace
func HandleListWorkspaceMembers(w http.ResponseWriter, r *http.Request, workspaceID string) {
	w.Header().Set("Content-Type", "application/json")

	if workspaceID == "" {
		WriteBadRequestError(w, "workspace_id is required")
		return
	}

	query := `
        SELECT wm.id, wm.user_id, u.email, u.first_name, u.last_name,
               wm.role_id, r.name AS role_name, wm.status, wm.created_at, wm.updated_at
        FROM workspace_members wm
        JOIN users u ON wm.user_id = u.id
        JOIN roles r ON wm.role_id = r.id
        WHERE wm.workspace_id = $1
        ORDER BY wm.created_at DESC
    `

	rows, err := DB.Query(query, workspaceID)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Failed to list workspace members: %v", err))
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var (
			id, userID, email, firstName, lastName, roleID, roleName, status string
			createdAt, updatedAt                                             time.Time
		)

		if err := rows.Scan(&id, &userID, &email, &firstName, &lastName, &roleID, &roleName, &status, &createdAt, &updatedAt); err != nil {
			WriteInternalServerError(w, "Error scanning workspace member row")
			return
		}

		members = append(members, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"role_id":    roleID,
			"role_name":  roleName,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	json.NewEncoder(w).Encode(members)
}

// HandleUpdateWorkspaceMember updates a member's role or status
func HandleUpdateWorkspaceMember(w http.ResponseWriter, r *http.Request, memberID string) {
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		RoleID    string `json:"role_id,omitempty"`
		Status    string `json:"status,omitempty"`
		UpdatedBy string `json:"updated_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	if memberID == "" {
		WriteBadRequestError(w, "member_id is required")
		return
	}

	query := `
		UPDATE workspace_members
		SET
			role_id = CASE
				WHEN $2 <> '' THEN $2::uuid
				ELSE role_id
			END,
			status = CASE
				WHEN $3 <> '' THEN $3
				ELSE status
			END,
			updated_by = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id
	`

	var id string
	err := DB.QueryRow(query, memberID, input.RoleID, input.Status, input.UpdatedBy).Scan(&id)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Failed to update workspace member: %v", err))
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Project member updated successfully",
	})
}

// HandleDeleteWorkspaceMember removes a member from a workspace
func HandleDeleteWorkspaceMember(w http.ResponseWriter, r *http.Request, memberID string) {
	w.Header().Set("Content-Type", "application/json")

	if memberID == "" {
		WriteBadRequestError(w, "member_id is required")
		return
	}

	res, err := DB.Exec(`DELETE FROM workspace_members WHERE id = $1`, memberID)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Failed to delete workspace member: %v", err))
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		WriteBadRequestError(w, "No member found with the given ID")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project member deleted successfully",
	})
}

// HandleCreateIntegrationConfig creates a new integration config
func HandleCreateIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload struct {
		UserID      string `json:"user_id"`
		ToolkitSlug string `json:"toolkit_slug"`
		AccountID   string `json:"account_id"`
		Status      string `json:"status"`
		CreatedBy   string `json:"created_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	if payload.UserID == "" || payload.ToolkitSlug == "" || payload.AccountID == "" || payload.Status == "" {
		WriteBadRequestError(w, "Missing required fields: user_id, auth_config_id, toolkit_slug, account_id, status")
		return
	}

	query := `
		INSERT INTO integrations_config (
			user_id, auth_config_id, toolkit_slug, account_id, status,
			created_at, updated_at, created_by
		) VALUES (
			$1, gen_random_uuid(), $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5
		) RETURNING auth_config_id`

	var authConfigID string
	err := DB.QueryRow(
		query,
		payload.UserID,
		payload.ToolkitSlug,
		payload.AccountID,
		payload.Status,
		payload.CreatedBy,
	).Scan(&authConfigID)
	if err != nil {
		log.Printf("Error creating integration config: %v", err)
		WriteInternalServerError(w, "Failed to create integration config")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"auth_config_id": authConfigID,
		"message":        "Integration config created successfully",
	})
}

// HandleGetIntegrationConfig retrieves a specific integration config by auth_config_id
func HandleGetIntegrationConfig(w http.ResponseWriter, authConfigID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT user_id, auth_config_id, toolkit_slug, account_id, status, created_at, updated_at, created_by
		FROM integrations_config
		WHERE auth_config_id = $1`

	var (
		userID, authID, toolkitSlug, accountID, status, createdBy string
		createdAt, updatedAt                                      time.Time
	)

	err := DB.QueryRow(query, authConfigID).Scan(
		&userID,
		&authID,
		&toolkitSlug,
		&accountID,
		&status,
		&createdAt,
		&updatedAt,
		&createdBy,
	)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Integration config not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving integration config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve integration config")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":        userID,
		"auth_config_id": authID,
		"toolkit_slug":   toolkitSlug,
		"account_id":     accountID,
		"status":         status,
		"created_at":     createdAt,
		"updated_at":     updatedAt,
		"created_by":     createdBy,
	})
}

// HandleGetIntegrationConfigsByUser lists integration configs for a user
func HandleGetIntegrationConfigsByUser(w http.ResponseWriter, userID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT user_id, auth_config_id, toolkit_slug, account_id, status, created_at, updated_at, created_by
		FROM integrations_config
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error retrieving integration configs: %v", err)
		WriteInternalServerError(w, "Failed to retrieve integration configs")
		return
	}
	defer rows.Close()

	var configs []map[string]interface{}
	for rows.Next() {
		var (
			uid, authID, toolkitSlug, accountID, status, createdBy string
			createdAt, updatedAt                                   time.Time
		)
		if err := rows.Scan(&uid, &authID, &toolkitSlug, &accountID, &status, &createdAt, &updatedAt, &createdBy); err != nil {
			log.Printf("Error scanning integration config: %v", err)
			WriteInternalServerError(w, "Failed to retrieve integration configs")
			return
		}
		configs = append(configs, map[string]interface{}{
			"user_id":        uid,
			"auth_config_id": authID,
			"toolkit_slug":   toolkitSlug,
			"account_id":     accountID,
			"status":         status,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
			"created_by":     createdBy,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating integration configs: %v", err)
		WriteInternalServerError(w, "Failed to retrieve integration configs")
		return
	}

	json.NewEncoder(w).Encode(configs)
}

// HandleUpdateIntegrationConfig updates an integration config by auth_config_id
func HandleUpdateIntegrationConfig(w http.ResponseWriter, r *http.Request, authConfigID string) {
	w.Header().Set("Content-Type", "application/json")

	var payload struct {
		UserID      string `json:"user_id"`
		ToolkitSlug string `json:"toolkit_slug"`
		AccountID   string `json:"account_id"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
		UPDATE integrations_config
		SET user_id = $1,
			toolkit_slug = $2,
			account_id = COALESCE(NULLIF($3, ''), account_id),
			status = COALESCE(NULLIF($4, 'Pending'), status),
			updated_at = CURRENT_TIMESTAMP
		WHERE auth_config_id = $5
		RETURNING auth_config_id`

	var updatedID string
	err := DB.QueryRow(
		query,
		payload.UserID,
		payload.ToolkitSlug,
		payload.AccountID,
		payload.Status,
		authConfigID,
	).Scan(&updatedID)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Integration config not found")
		return
	} else if err != nil {
		log.Printf("Error updating integration config: %v", err)
		WriteInternalServerError(w, "Failed to update integration config")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"auth_config_id": updatedID,
		"message":        "Integration config updated successfully",
	})
}

// HandleDeleteIntegrationConfig deletes an integration config by auth_config_id
func HandleDeleteIntegrationConfig(w http.ResponseWriter, authConfigID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `DELETE FROM integrations_config WHERE auth_config_id = $1 RETURNING auth_config_id`
	var deletedID string
	err := DB.QueryRow(query, authConfigID).Scan(&deletedID)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Integration config not found")
		return
	} else if err != nil {
		log.Printf("Error deleting integration config: %v", err)
		WriteInternalServerError(w, "Failed to delete integration config")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"auth_config_id": deletedID,
		"message":        "Integration config deleted successfully",
	})
}

// HandleCreateWorkspace creates a new workspace and adds the owner as Admin
func HandleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var workspace struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Organization string `json:"organization_id"`
		Status       string `json:"status"`
		Owner        string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&workspace); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteInternalServerError(w, "Transaction start failed")
		return
	}
	defer tx.Rollback()

	// --- Step 1: Insert the workspace ---
	var workspaceID string
	createWorkspaceQuery := `
		INSERT INTO workspaces (id, name, description, status, owner, created_at, updated_at, organization_id)
		VALUES (gen_random_uuid(), $1, $2, COALESCE(NULLIF($3, ''), 'Active'), $4, NOW(), NOW(), $5)
		RETURNING id`
	err = tx.QueryRow(createWorkspaceQuery,
		workspace.Name,
		workspace.Description,
		workspace.Status,
		workspace.Owner,
		workspace.Organization,
	).Scan(&workspaceID)
	if err != nil {
		log.Printf("Error inserting workspace: %v", err)
		WriteInternalServerError(w, "Failed to create workspace")
		return
	}

	// --- Step 2: Get Admin role ID ---
	var adminRoleID string
	getRoleQuery := `SELECT id FROM roles WHERE LOWER(name) = 'admin' LIMIT 1`
	err = tx.QueryRow(getRoleQuery).Scan(&adminRoleID)
	if err != nil {
		log.Printf("Error fetching admin role ID: %v", err)
		WriteInternalServerError(w, "Failed to fetch admin role")
		return
	}

	// --- Step 3: Add owner as Admin in workspace_members ---
	addMemberQuery := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role_id, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 'Active', NOW(), NOW())`
	_, err = tx.Exec(addMemberQuery, workspaceID, workspace.Owner, adminRoleID)
	if err != nil {
		log.Printf("Error adding workspace owner to members: %v", err)
		WriteInternalServerError(w, "Workspace created, but failed to add owner as admin")
		return
	}

	// --- Step 4: Commit transaction ---
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to finalize workspace creation")
		return
	}

	// --- Step 5: Respond success ---
	json.NewEncoder(w).Encode(map[string]string{
		"id":      workspaceID,
		"message": "Project created successfully and owner added as admin",
	})
}

// HandleUpdateWorkspace updates an existing workspace
func HandleUpdateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	w.Header().Set("Content-Type", "application/json")

	var workspace struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Organization string `json:"organization_id"`
		Status       string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&workspace); err != nil {
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
		UPDATE workspaces
		SET name = $1,
			description = $2,
			status = COALESCE(NULLIF($3, ''), status),
			updated_at = NOW(),
			organization_id = $5
		WHERE id = $4
		RETURNING id`

	var id string
	err := DB.QueryRow(query, workspace.Name, workspace.Description, workspace.Status, workspaceID, workspace.Organization).Scan(&id)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Workspace not found")
		return
	} else if err != nil {
		log.Printf("Error updating workspace: %v", err)
		WriteInternalServerError(w, "Failed to update workspace")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Project updated successfully",
	})
}

// HandleGetWorkspace retrieves a single workspace by ID
func HandleGetWorkspace(w http.ResponseWriter, workspaceID string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT id, name, description, status, owner, created_at, updated_at, organization_id
		FROM workspaces
		WHERE id = $1`

	var (
		id, name, description, status, owner, organization string
		createdAt, updatedAt                               time.Time
	)

	err := DB.QueryRow(query, workspaceID).Scan(
		&id, &name, &description, &status, &owner, &createdAt, &updatedAt, &organization,
	)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Workspace not found")
		return
	} else if err != nil {
		log.Printf("Error fetching workspace: %v", err)
		WriteInternalServerError(w, "Failed to fetch workspace")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"name":         name,
		"description":  description,
		"status":       status,
		"owner":        owner,
		"created_at":   createdAt,
		"updated_at":   updatedAt,
		"organization": organization,
	})
}

// HandleGetAllWorkspaces retrieves all workspaces for an owner
func HandleGetAllWorkspaces(w http.ResponseWriter, owner string) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT
			wm.workspace_id,
			w.name AS workspace_name,
			w.description AS workspace_description,
			w.status as workspace_status,
			w.owner AS workspace_owner,
			w.created_at as workspace_created_at,
			w.organization_id,
			o.name AS organization_name,
			wm.role_id,
			r.name AS role_name,
			wm.user_id
		FROM workspace_members wm
		JOIN workspaces w ON wm.workspace_id = w.id
		JOIN organizations o ON w.organization_id = o.id
		JOIN roles r ON wm.role_id = r.id
		WHERE wm.user_id = $1
		ORDER BY o.name, w.name;
		`

	rows, err := DB.Query(query, owner)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve workspaces")
		return
	}
	defer rows.Close()

	var workspaces []map[string]interface{}

	for rows.Next() {
		var (
			id, name, description, status, owner, organizationid, organizationname, roleid, rolename, userid string
			createdAt                                                                                        time.Time
		)

		if err := rows.Scan(&id, &name, &description, &status, &owner, &createdAt, &organizationid, &organizationname, &roleid, &rolename, &userid); err != nil {
			WriteInternalServerError(w, "Error scanning workspace row")
			return
		}

		workspaces = append(workspaces, map[string]interface{}{
			"id":                id,
			"name":              name,
			"description":       description,
			"status":            status,
			"owner":             owner,
			"created_at":        createdAt,
			"user_id":           userid,
			"organization":      organizationid,
			"organization_name": organizationname,
			"role_id":           roleid,
			"role_name":         rolename,
		})
	}

	json.NewEncoder(w).Encode(workspaces)
}

// HandleDeleteWorkspace deletes a workspace by ID
func HandleDeleteWorkspace(w http.ResponseWriter, workspaceID string) {
	w.Header().Set("Content-Type", "application/json")

	var id string
	err := DB.QueryRow(`DELETE FROM workspaces WHERE id = $1 RETURNING id`, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		WriteNotFoundError(w, "Workspace not found")
		return
	} else if err != nil {
		log.Printf("Error deleting workspace: %v", err)
		WriteInternalServerError(w, "Failed to delete workspace")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Project deleted successfully",
	})
}

// HandleGetAllActiveRoles retrieves all roles with status = 'Active'
func HandleGetAllActiveRoles(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT id, name, description, status, created_at
		FROM roles
		WHERE status = 'Active'
		ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve active roles")
		return
	}
	defer rows.Close()

	var roles []map[string]interface{}

	for rows.Next() {
		var (
			id, name, description, status string
			createdAt                     time.Time
		)

		if err := rows.Scan(&id, &name, &description, &status, &createdAt); err != nil {
			WriteInternalServerError(w, "Error scanning active role row")
			return
		}

		roles = append(roles, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"status":      status,
			"created_at":  createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		WriteInternalServerError(w, "Error reading active role rows")
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// HandleCreateTemplateConfig creates a template and links it with knowledge bases
func HandleCreateTemplateConfig(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	var templateConfig struct {
		TemplateName         string          `json:"template_name"`
		TemplateSystemPrompt string          `json:"template_system_prompt"`
		Config               json.RawMessage `json:"config"`
		KnowledgeBase        []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Mode string `json:"mode"`
		} `json:"knowledge_base"`
		MCP []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"mcp"`
		Tool []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tool"`
		IsActive        bool     `json:"is_active"`
		Record          bool     `json:"record"`
		Callback_url    string   `json:"callback_url"`
		Callback_events []string `json:"callback_events" db:"callback_events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&templateConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	err := ValidateTimeout(apiKeyId, templateConfig.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate error : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	templateConfig.IsActive = true

	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback() // rollback on any error

	// Insert into templates table
	query := `
        INSERT INTO templates (
            id, template_name, template_system_prompt,
            tools, config, is_active, created_at, updated_at, created_by, record, callback_url, callback_events
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, $4,
            $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6, $7, $8, $9
        ) RETURNING id`

	var templateID string
	err = tx.QueryRow(
		query,
		templateConfig.TemplateName,
		templateConfig.TemplateSystemPrompt,
		"{}",
		templateConfig.Config,
		templateConfig.IsActive,
		apiKeyId,
		templateConfig.Record,
		templateConfig.Callback_url,
		pq.Array(templateConfig.Callback_events),
	).Scan(&templateID)
	if err != nil {
		log.Printf("Error creating template config: %v", err)
		WriteInternalServerError(w, "Failed to create template")
		return
	}

	// Insert into templates_kb for each knowledge_base
	kbInsert := `
        INSERT INTO templates_kb (
            id, template_id, knowledge_base_id, mode
        ) VALUES (
            gen_random_uuid(), $1, $2, $3
        )`

	for _, kb := range templateConfig.KnowledgeBase {
		_, err := tx.Exec(kbInsert, templateID, kb.ID, kb.Mode)
		if err != nil {
			log.Printf("Error inserting into templates_kb: %v", err)
			WriteInternalServerError(w, "Failed to save template knowledge base")
			return
		}
	}

	// Insert into templates_tool for each tool
	toolInsert := `
        INSERT INTO templates_tool (
            id, template_id, tool_id
        ) VALUES (
            gen_random_uuid(), $1, $2
        )`

	for _, tool := range templateConfig.Tool {
		_, err := tx.Exec(toolInsert, templateID, tool.ID)
		if err != nil {
			log.Printf("Error inserting into templates_tool: %v", err)
			WriteInternalServerError(w, "Failed to save template tool")
			return
		}
	}

	// Insert into templates_mcp for each mcp
	mcpInsert := `
        INSERT INTO templates_mcp (
            id, template_id, mcp_id
        ) VALUES (
            gen_random_uuid(), $1, $2
        )`

	for _, mcp := range templateConfig.MCP {
		_, err := tx.Exec(mcpInsert, templateID, mcp.ID)
		if err != nil {
			log.Printf("Error inserting into templates_mcp: %v", err)
			WriteInternalServerError(w, "Failed to save template mcp")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// Success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      templateID,
		"message": "Template created successfully",
	})
}

// HandleUpdateTemplateConfig updates an existing Template
func HandleUpdateTemplateConfig(w http.ResponseWriter, r *http.Request, templateConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating template config with ID: %s", templateConfigID)

	var templateConfig struct {
		TemplateName         string          `json:"template_name"`
		TemplateSystemPrompt string          `json:"template_system_prompt"`
		Config               json.RawMessage `json:"config"`
		KnowledgeBase        []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Mode string `json:"mode"`
		} `json:"knowledge_base"`
		MCP []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"mcp"`
		Tool []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tool"`
		IsActive        bool     `json:"is_active"`
		Record          bool     `json:"record"`
		Callback_url    string   `json:"callback_url"`
		Callback_events []string `json:"callback_events" db:"callback_events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&templateConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM templates a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, templateConfigID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving template: %v", err)
		WriteInternalServerError(w, "Failed to retrieve template")
		return
	}

	if !exists {
		log.Printf("No template found with ID: %s", templateConfigID)
		WriteNotFoundError(w, "Template not found")
		return
	}

	err = ValidateTimeout(apiKeyId, templateConfig.Config)
	if err != nil {
		WriteInternalServerError(w, fmt.Sprintf("Validate error : %v", err))
		log.Printf("Timeout validation error: %v", err)
		return
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update template
	query := `
        UPDATE templates
        SET template_name = $1,
            template_system_prompt = $2,
            tools = $3,
            config = $4,
            is_active = $5,
            updated_at = CURRENT_TIMESTAMP,
			record = $7,
			callback_url = $8,
			callback_events = $9
        WHERE id = $6
        RETURNING id`

	var id string
	err = tx.QueryRow(
		query,
		templateConfig.TemplateName,
		templateConfig.TemplateSystemPrompt,
		"{}",
		templateConfig.Config,
		templateConfig.IsActive,
		templateConfigID,
		templateConfig.Record,
		templateConfig.Callback_url,
		pq.Array(templateConfig.Callback_events),
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No template found with ID: %s", templateConfigID)
		WriteNotFoundError(w, "Template not found")
		return
	} else if err != nil {
		log.Printf("Error updating template config: %v", err)
		WriteInternalServerError(w, "Failed to update template")
		return
	}

	// Delete existing knowledge base links
	_, err = tx.Exec(`DELETE FROM templates_kb WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting old template_kb entries: %v", err)
		WriteInternalServerError(w, "Failed to update knowledge base")
		return
	}

	// Insert new knowledge base links
	kbInsert := `
        INSERT INTO templates_kb (id, template_id, knowledge_base_id, mode)
        VALUES (gen_random_uuid(), $1, $2, $3)`

	for _, kb := range templateConfig.KnowledgeBase {
		_, err := tx.Exec(kbInsert, templateConfigID, kb.ID, kb.Mode)
		if err != nil {
			log.Printf("Error inserting templates_kb: %v", err)
			WriteInternalServerError(w, "Failed to insert knowledge base")
			return
		}
	}

	// Delete existing tool links
	_, err = tx.Exec(`DELETE FROM templates_tool WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting old template_tool entries: %v", err)
		WriteInternalServerError(w, "Failed to update Tool")
		return
	}

	// Insert new tool links
	toolInsert := `
        INSERT INTO templates_tool (id, template_id, tool_id)
        VALUES (gen_random_uuid(), $1, $2)`

	for _, tool := range templateConfig.Tool {
		_, err := tx.Exec(toolInsert, templateConfigID, tool.ID)
		if err != nil {
			log.Printf("Error inserting templates_kb: %v", err)
			WriteInternalServerError(w, "Failed to insert Tool")
			return
		}
	}

	// Delete existing mcp links
	_, err = tx.Exec(`DELETE FROM templates_mcp WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting old template_mcp entries: %v", err)
		WriteInternalServerError(w, "Failed to update MCP")
		return
	}

	// Insert new MCP links
	mcpInsert := `
        INSERT INTO templates_mcp (id, template_id, mcp_id)
        VALUES (gen_random_uuid(), $1, $2)`

	for _, mcp := range templateConfig.MCP {
		_, err := tx.Exec(mcpInsert, templateConfigID, mcp.ID)
		if err != nil {
			log.Printf("Error inserting templates_mcp: %v", err)
			WriteInternalServerError(w, "Failed to insert MCP")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// Success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Template updated successfully",
	})
}

// HandleGetTemplateConfig retrieves a specific Template by ID
func HandleGetTemplateConfig(w http.ResponseWriter, templateConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching template config with ID: %s", templateConfigID)

	// Fetch template
	templateConfig, shouldReturn := GetTemplateFunction(templateConfigID, w, apiKeyId)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(templateConfig)
}

// HandleGetTemplateConfigAPI retrieves a specific Template by ID
func HandleGetTemplateConfigAPI(w http.ResponseWriter, templateConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching template config with ID: %s", templateConfigID)

	// Fetch template
	templateConfig, shouldReturn := GetTemplateFunction(templateConfigID, w, apiKeyId)
	if shouldReturn {
		return
	}

	tempConfigAPI := map[string]interface{}{
		"id":                     templateConfig["id"],
		"template_name":          templateConfig["template_name"],
		"template_system_prompt": templateConfig["template_system_prompt"],
		"config":                 templateConfig["config"],
		"created_at":             templateConfig["created_at"],
		"updated_at":             templateConfig["updated_at"],
		"knowledge_base":         templateConfig["knowledge_base"],
		"record":                 templateConfig["record"],
	}

	json.NewEncoder(w).Encode(tempConfigAPI)
}

func GetTemplateFunction(templateConfigID string, w http.ResponseWriter, apiKeyId string) (map[string]interface{}, bool) {
	query := `
        SELECT a.id, template_name, template_system_prompt, tools, config, a.is_active, a.created_at, a.updated_at, a.created_by, record, COALESCE(callback_url, '') AS callback_url, COALESCE(callback_events, '{}') AS callback_events
        FROM templates a
    JOIN api_keys ak2 ON a.created_by = ak2.id
    JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
    JOIN workspaces w1 ON ak1.workspace_id = w1.id
    JOIN workspaces w2 ON ak2.workspace_id = w2.id 
        AND w2.organization_id = w1.organization_id
    WHERE ak1.id = $2 AND a.id = $1`

	var (
		id, name, prompt, tools, config, callback_url string
		isActive, record                              bool
		createdAt, updatedAt                          time.Time
		createdBy                                     string
		callback_events                               []string
	)

	err := DB.QueryRow(query, templateConfigID, apiKeyId).Scan(
		&id, &name, &prompt, &tools,
		&config, &isActive, &createdAt, &updatedAt, &createdBy, &record, &callback_url, pq.Array(&callback_events),
	)

	if err == sql.ErrNoRows {
		log.Printf("No template found with ID: %s", templateConfigID)
		WriteNotFoundError(w, "Template not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error retrieving template config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve template")
		return nil, true
	}

	// Fetch linked knowledge bases
	kbQuery := `
        SELECT kb.id, kb.name, kb.description, kb.namespace, kb.index, tk.mode
        FROM templates_kb tk
        JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
        WHERE tk.template_id = $1`

	rows, err := DB.Query(kbQuery, templateConfigID)
	if err != nil {
		log.Printf("Error fetching knowledge bases: %v", err)
		WriteInternalServerError(w, "Failed to fetch linked knowledge bases")
		return nil, true
	}
	defer rows.Close()

	var knowledgeBases []map[string]string
	for rows.Next() {
		var kbID, kbName, kbDescription, kbNamespace, kbIndex, kbMode string
		if err := rows.Scan(&kbID, &kbName, &kbDescription, &kbNamespace, &kbIndex, &kbMode); err != nil {
			log.Printf("Error scanning knowledge base row: %v", err)
			WriteInternalServerError(w, "Failed to parse knowledge base")
			return nil, true
		}
		knowledgeBases = append(knowledgeBases, map[string]string{
			"id":          kbID,
			"name":        kbName,
			"description": kbDescription,
			"namespace":   kbNamespace,
			"index":       kbIndex,
			"mode":        kbMode,
		})
	}

	// Fetch linked tools
	toolQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.request_config, t.event_messages
        FROM templates_tool tt
        JOIN tools t ON t.id = tt.tool_id
        WHERE tt.template_id = $1`

	rows1, err1 := DB.Query(toolQuery, templateConfigID)
	if err1 != nil {
		log.Printf("Error fetching tools: %v", err1)
		WriteInternalServerError(w, "Failed to fetch linked tools")
		return nil, true
	}
	defer rows1.Close()

	var tool []map[string]string
	for rows1.Next() {
		var toolID, toolName, toolDescription, tooltype, arguments, request_config, event_messages string
		if err := rows1.Scan(&toolID, &toolName, &toolDescription, &tooltype, &arguments, &request_config, &event_messages); err != nil {
			log.Printf("Error scanning tools row: %v", err)
			WriteInternalServerError(w, "Failed to parse tools")
			return nil, true
		}
		tool = append(tool, map[string]string{
			"id":             toolID,
			"type":           tooltype,
			"schema":         arguments,
			"request_config": request_config,
			"event_messages": event_messages,
		})
	}

	// Fetch linked MCPs
	mcpQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.cache_tools_list, t.event_messages
        FROM templates_mcp tt
        JOIN mcps t ON t.id = tt.mcp_id
        WHERE tt.template_id = $1`

	rows2, err2 := DB.Query(mcpQuery, templateConfigID)
	if err2 != nil {
		log.Printf("Error fetching mcp: %v", err2)
		WriteInternalServerError(w, "Failed to fetch linked mcp")
		return nil, true
	}
	defer rows2.Close()

	var mcp []map[string]string
	for rows2.Next() {
		var mcpID, mcpName, mcpDescription, mcptype, arguments, cache_tools_list, event_messages string
		if err := rows2.Scan(&mcpID, &mcpName, &mcpDescription, &mcptype, &arguments, &cache_tools_list, &event_messages); err != nil {
			log.Printf("Error scanning mcp row: %v", err)
			WriteInternalServerError(w, "Failed to parse MCP")
			return nil, true
		}
		mcp = append(mcp, map[string]string{
			"id":             mcpID,
			"name":           mcpName,
			"description":    mcpDescription,
			"type":           mcptype,
			"request_config": arguments,
			//"cache_tools_list": cache_tools_list,
			"event_messages": event_messages,
		})
	}

	templateConfig := map[string]interface{}{
		"id":                     id,
		"template_name":          name,
		"template_system_prompt": prompt,
		"tools":                  tools,
		"config":                 config,
		"is_active":              isActive,
		"created_at":             createdAt,
		"updated_at":             updatedAt,
		"knowledge_base":         knowledgeBases,
		"tool":                   tool,
		"mcp":                    mcp,
		"record":                 record,
		"callback_url":           callback_url,
		"callback_events":        callback_events,
	}
	return templateConfig, false
}

// HandleGetAllTemplateConfigs retrieves all Templates
func HandleGetAllTemplateConfigs(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	templateConfigs, shouldReturn := GetAllTemplateFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(templateConfigs)
}

// HandleGetAllTemplateConfigs retrieves all Templates
func HandleGetAllTemplateConfigsAPI(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	templateConfigs, shouldReturn := GetAllTemplateFunction(apiKeyId, w)
	if shouldReturn {
		return
	}

	var templateConfigsAPI []map[string]interface{}

	for _, temp := range templateConfigs {

		tempConfigAPI := map[string]interface{}{
			"id":                     temp["id"],
			"template_name":          temp["template_name"],
			"template_system_prompt": temp["template_system_prompt"],
			"config":                 temp["config"],
			"created_at":             temp["created_at"],
			"updated_at":             temp["updated_at"],
			"knowledge_base":         temp["knowledge_base"],
			"record":                 temp["record"],
		}

		templateConfigsAPI = append(templateConfigsAPI, tempConfigAPI)
	}

	json.NewEncoder(w).Encode(templateConfigsAPI)
}

func GetAllTemplateFunction(apiKeyId string, w http.ResponseWriter) ([]map[string]interface{}, bool) {
	query := `
        SELECT
			t.id,
			t.template_name,
			t.template_system_prompt,
			t.tools,
			t.config,
			t.is_active,
			t.created_at,
			t.updated_at,
			t.created_by,
			t.record,
			COALESCE(t.callback_url, '') AS callback_url,
			COALESCE(t.callback_events, '{}') AS callback_events
		FROM templates t
		JOIN api_keys ak2
			ON t.created_by = ak2.id                    -- creator API key
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id      -- workspace match
		-- Workspace → organization validation
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id                 -- requester workspace
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id                 -- creator workspace
			AND w2.organization_id = w1.organization_id -- ✅ SAME org
		WHERE ak1.id = $1
		ORDER BY t.created_at DESC;`

	rows, err := DB.Query(query, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve templates")
		return nil, true
	}
	defer rows.Close()

	var templateConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, name, prompt, tools, config, callback_url string
			isActive, record                              bool
			createdAt, updatedAt                          time.Time
			createdBy                                     string
			callback_events                               []string
		)

		if err := rows.Scan(
			&id, &name, &prompt, &tools,
			&config, &isActive, &createdAt, &updatedAt, &createdBy, &record, &callback_url, pq.Array(&callback_events),
		); err != nil {
			WriteInternalServerError(w, "Error scanning Templates")
			return nil, true
		}

		// Fetch knowledge bases for this template
		kbQuery := `
            SELECT kb.id, kb.name, kb.description, kb.namespace, kb.index, tk.mode
            FROM templates_kb tk
            JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
            WHERE tk.template_id = $1`

		kbRows, err := DB.Query(kbQuery, id)
		if err != nil {
			WriteInternalServerError(w, "Failed to fetch linked knowledge bases")
			return nil, true
		}
		var knowledgeBases []map[string]string
		for kbRows.Next() {
			var kbID, kbName, kbDescription, kbNamespace, kbIndex, kbMode string
			if err := kbRows.Scan(&kbID, &kbName, &kbDescription, &kbNamespace, &kbIndex, &kbMode); err != nil {
				WriteInternalServerError(w, "Failed to parse knowledge base")
				return nil, true
			}
			knowledgeBases = append(knowledgeBases, map[string]string{
				"id":          kbID,
				"name":        kbName,
				"description": kbDescription,
				"namespace":   kbNamespace,
				"index":       kbIndex,
				"mode":        kbMode,
			})
		}
		kbRows.Close()

		// Fetch linked tools
		toolQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.request_config, t.event_messages
        FROM templates_tool tt
        JOIN tools t ON t.id = tt.tool_id
        WHERE tt.template_id = $1`

		rows1, err1 := DB.Query(toolQuery, id)
		if err1 != nil {
			log.Printf("Error fetching tools: %v", err1)
			WriteInternalServerError(w, "Failed to fetch linked tools")
			return nil, true
		}

		var tool []map[string]string
		for rows1.Next() {
			var toolID, toolName, toolDescription, tooltype, arguments, request_config, event_messages string
			if err := rows1.Scan(&toolID, &toolName, &toolDescription, &tooltype, &arguments, &request_config, &event_messages); err != nil {
				log.Printf("Error scanning tools row: %v", err)
				WriteInternalServerError(w, "Failed to parse tools")
				return nil, true
			}
			tool = append(tool, map[string]string{
				"id":             toolID,
				"type":           tooltype,
				"schema":         arguments,
				"request_config": request_config,
				"event_messages": event_messages,
			})
		}
		defer rows1.Close()
		// Fetch linked MCPs
		mcpQuery := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.cache_tools_list, t.event_messages
        FROM templates_mcp tt
        JOIN mcps t ON t.id = tt.mcp_id
        WHERE tt.template_id = $1`

		rows2, err2 := DB.Query(mcpQuery, id)
		if err2 != nil {
			log.Printf("Error fetching mcp: %v", err2)
			WriteInternalServerError(w, "Failed to fetch linked mcp")
			return nil, true
		}

		var mcp []map[string]string
		for rows2.Next() {
			var mcpID, mcpName, mcpDescription, mcptype, arguments, cache_tools_list, event_messages string
			if err := rows2.Scan(&mcpID, &mcpName, &mcpDescription, &mcptype, &arguments, &cache_tools_list, &event_messages); err != nil {
				log.Printf("Error scanning mcp row: %v", err)
				WriteInternalServerError(w, "Failed to parse MCP")
				return nil, true
			}
			mcp = append(mcp, map[string]string{
				"id":             mcpID,
				"name":           mcpName,
				"description":    mcpDescription,
				"type":           mcptype,
				"request_config": arguments,
				//"cache_tools_list": cache_tools_list,
				"event_messages": event_messages,
			})
		}
		defer rows2.Close()

		templateConfigs = append(templateConfigs, map[string]interface{}{
			"id":                     id,
			"template_name":          name,
			"template_system_prompt": prompt,
			"tools":                  tools,
			"config":                 config,
			"is_active":              isActive,
			"created_at":             createdAt,
			"updated_at":             updatedAt,
			"knowledge_base":         knowledgeBases,
			"tool":                   tool,
			"mcp":                    mcp,
			"record":                 record,
			"callback_url":           callback_url,
			"callback_events":        callback_events,
		})
	}
	return templateConfigs, false
}

// HandleDeleteTemplateConfig deletes a Template and related data
func HandleDeleteTemplateConfig(w http.ResponseWriter, templateConfigID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting template config with ID: %s", templateConfigID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM templates a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, templateConfigID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving template: %v", err)
		WriteInternalServerError(w, "Failed to retrieve template")
		return
	}

	if !exists {
		log.Printf("No template found with ID: %s", templateConfigID)
		WriteNotFoundError(w, "Template not found")
		return
	}

	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		WriteInternalServerError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Delete mappings first (if applicable)
	_, err = tx.Exec(`DELETE FROM templates_kb WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting template KB mappings: %v", err)
		WriteInternalServerError(w, "Failed to delete template KB")
		return
	}

	// Delete mappings first (if applicable)
	_, err = tx.Exec(`DELETE FROM templates_tool WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting template tool mappings: %v", err)
		WriteInternalServerError(w, "Failed to delete template tool")
		return
	}

	// Delete mappings first (if applicable)
	_, err = tx.Exec(`DELETE FROM templates_mcp WHERE template_id = $1`, templateConfigID)
	if err != nil {
		log.Printf("Error deleting template mcp mappings: %v", err)
		WriteInternalServerError(w, "Failed to delete template mcp")
		return
	}

	// Delete template itself
	var id string
	err = tx.QueryRow(`DELETE FROM templates WHERE id = $1 RETURNING id`, templateConfigID).Scan(&id)
	if err == sql.ErrNoRows {
		log.Printf("No template found with ID: %s", templateConfigID)
		WriteNotFoundError(w, "Template not found")
		return
	} else if err != nil {
		log.Printf("Error deleting template config: %v", err)
		WriteInternalServerError(w, "Failed to delete template")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		WriteInternalServerError(w, "Failed to commit transaction")
		return
	}

	// Success response
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Template deleted successfully",
	})
}

// HandleCreateToolConfig creates a new Tool
func HandleCreateToolConfig(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	var toolConfig struct {
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Type          string          `json:"type"`
		Arguments     json.RawMessage `json:"schema"`
		RequestConfig json.RawMessage `json:"request_config"`
		EventMessages json.RawMessage `json:"event_messages"`
		IsActive      bool            `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&toolConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        INSERT INTO tools (
            id, name,
            description, type, arguments, request_config, event_messages, is_active,
            created_at, updated_at, created_by
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, $4,
            $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $8
        ) RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		toolConfig.Name,
		toolConfig.Description,
		toolConfig.Type,
		toolConfig.Arguments,
		toolConfig.RequestConfig,
		toolConfig.EventMessages,
		true,
		apiKeyId,
	).Scan(&id)

	if err != nil {
		log.Printf("Error creating tools config: %v", err)
		WriteInternalServerError(w, "Failed to create tools")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Tools created successfully",
	})
}

// HandleUpdateTemplateConfig updates an existing Template
func HandleUpdateToolConfig(w http.ResponseWriter, r *http.Request, toolID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating tool config with ID: %s", toolID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM tools a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, toolID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving tool: %v", err)
		WriteInternalServerError(w, "Failed to retrieve tool")
		return
	}

	if !exists {
		log.Printf("No tool found with ID: %s", toolID)
		WriteNotFoundError(w, "Tool not found")
		return
	}

	var toolConfig struct {
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		Type          string          `json:"type"`
		Arguments     json.RawMessage `json:"schema"`
		RequestConfig json.RawMessage `json:"request_config"`
		EventMessages json.RawMessage `json:"event_messages"`
		IsActive      bool            `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&toolConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE tools
        SET name = $1,
            description = $2,
            type = $3,
            arguments = $4,
			request_config = $5,
			event_messages = $6,
            is_active = $7,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $8
        RETURNING id`

	var id string
	err = DB.QueryRow(
		query,
		toolConfig.Name,
		toolConfig.Description,
		toolConfig.Type,
		toolConfig.Arguments,
		toolConfig.RequestConfig,
		toolConfig.EventMessages,
		true,
		toolID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No tool found with ID: %s", toolID)
		WriteNotFoundError(w, "Tool not found")
		return
	} else if err != nil {
		log.Printf("Error updating tool config: %v", err)
		WriteInternalServerError(w, "Failed to update tool")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Tool updated successfully",
	})
}

// HandleGetToolConfig retrieves a specific Tool by ID
func HandleGetToolConfig(w http.ResponseWriter, toolID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching tool config with ID: %s", toolID)

	query := `
        SELECT a.id, a.name, a.description, type, arguments, request_config, event_messages, a.is_active, a.created_at, a.updated_at, a.created_by
        FROM tools a
    JOIN api_keys ak2 ON a.created_by = ak2.id
    JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
    JOIN workspaces w1 ON ak1.workspace_id = w1.id
    JOIN workspaces w2 ON ak2.workspace_id = w2.id 
        AND w2.organization_id = w1.organization_id
    WHERE ak1.id = $2 AND a.id = $1`

	var (
		id, name, description, tooltype, arguments, request_config, event_messages string
		isActive                                                                   bool
		createdAt, updatedAt                                                       time.Time
		createdBy                                                                  string
	)

	err := DB.QueryRow(query, toolID, apiKeyId).Scan(
		&id, &name, &description, &tooltype,
		&arguments, &request_config, &event_messages, &isActive, &createdAt, &updatedAt, &createdBy,
	)

	if err == sql.ErrNoRows {
		log.Printf("No tool found with ID: %s", toolID)
		WriteNotFoundError(w, "Tool not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving tool config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve tool")
		return
	}

	toolConfig := map[string]interface{}{
		"id":             id,
		"type":           tooltype,
		"schema":         arguments,
		"request_config": request_config,
		"event_messages": event_messages,
		"created_at":     createdAt,
		"updated_at":     updatedAt,
	}

	json.NewEncoder(w).Encode(toolConfig)
}

// HandleGetAllToolConfigs retrieves all Tools
func HandleGetAllToolConfigs(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	query := `
        SELECT t.id, t.name, t.description, t.type, t.arguments, t.request_config, t.event_messages, t.is_active, t.created_at, t.updated_at, t.created_by
        FROM tools t
		JOIN api_keys ak2
			ON t.created_by = ak2.id
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1
		ORDER BY t.created_at DESC;`

	rows, err := DB.Query(query, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve tools")
		return
	}
	defer rows.Close()

	var toolConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, name, description, tooltype, arguments, request_config, event_messages string
			isActive                                                                   bool
			createdAt, updatedAt                                                       time.Time
			createdBy                                                                  string
		)

		if err := rows.Scan(
			&id, &name, &description, &tooltype,
			&arguments, &request_config, &event_messages, &isActive, &createdAt, &updatedAt, &createdBy,
		); err != nil {
			WriteInternalServerError(w, "Error scanning Tools")
			return
		}

		toolConfigs = append(toolConfigs, map[string]interface{}{
			"id":             id,
			"type":           tooltype,
			"schema":         arguments,
			"request_config": request_config,
			"event_messages": event_messages,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		})
	}

	json.NewEncoder(w).Encode(toolConfigs)
}

// HandleDeleteToolConfig deletes Tool
func HandleDeleteToolConfig(w http.ResponseWriter, toolID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting tool config with ID: %s", toolID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM tools a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, toolID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving tool: %v", err)
		WriteInternalServerError(w, "Failed to retrieve tool")
		return
	}

	if !exists {
		log.Printf("No tool found with ID: %s", toolID)
		WriteNotFoundError(w, "Tool not found")
		return
	}

	query := `DELETE FROM tools WHERE id = $1 RETURNING id`

	var id string
	err = DB.QueryRow(query, toolID).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No tool found with ID: %s", toolID)
		WriteNotFoundError(w, "Tool not found")
		return
	} else if err != nil {
		log.Printf("Error deleting tool config: %v", err)
		WriteInternalServerError(w, "Failed to delete tool")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Tool deleted successfully",
	})
}

// HandleCreateMCPConfig creates a new MCP
func HandleCreateMCPConfig(w http.ResponseWriter, r *http.Request, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	var mcpConfig struct {
		Name             string          `json:"name"`
		Description      string          `json:"description"`
		Type             string          `json:"type"`
		Arguments        json.RawMessage `json:"request_config"`
		Cache_tools_list bool            `json:"cache_tools_list"`
		EventMessages    json.RawMessage `json:"event_messages"`
		IsActive         bool            `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mcpConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        INSERT INTO mcps (
            id, name,
            description, type, arguments, cache_tools_list, event_messages, is_active,
            created_at, updated_at, created_by
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, $4,
            $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $8
        ) RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		mcpConfig.Name,
		mcpConfig.Description,
		mcpConfig.Type,
		mcpConfig.Arguments,
		mcpConfig.Cache_tools_list,
		"{}", //mcpConfig.EventMessages,
		true,
		apiKeyId,
	).Scan(&id)

	if err != nil {
		log.Printf("Error creating mcps config: %v", err)
		WriteInternalServerError(w, "Failed to create mcp")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "MCP created successfully",
	})
}

// HandleUpdateMCPConfig updates an existing MCP
func HandleUpdateMCPConfig(w http.ResponseWriter, r *http.Request, mcpID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating MCP config with ID: %s", mcpID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM mcps a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, mcpID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving mcp: %v", err)
		WriteInternalServerError(w, "Failed to retrieve mcp")
		return
	}

	if !exists {
		log.Printf("No mcp found with ID: %s", mcpID)
		WriteNotFoundError(w, "MCP not found")
		return
	}

	var mcpConfig struct {
		Name             string          `json:"name"`
		Description      string          `json:"description"`
		Type             string          `json:"type"`
		Arguments        json.RawMessage `json:"request_config"`
		Cache_tools_list bool            `json:"cache_tools_list"`
		EventMessages    json.RawMessage `json:"event_messages"`
		IsActive         bool            `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mcpConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE mcps
        SET name = $1,
            description = $2,
            type = $3,
            arguments = $4,
			cache_tools_list = $5,
			event_messages = $6,
            is_active = $7,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $8
        RETURNING id`

	var id string
	err = DB.QueryRow(
		query,
		mcpConfig.Name,
		mcpConfig.Description,
		mcpConfig.Type,
		mcpConfig.Arguments,
		mcpConfig.Cache_tools_list,
		"{}", //mcpConfig.EventMessages,
		true,
		mcpID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No MCP found with ID: %s", mcpID)
		WriteNotFoundError(w, "MCP not found")
		return
	} else if err != nil {
		log.Printf("Error updating mcp config: %v", err)
		WriteInternalServerError(w, "Failed to update mcp")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "MCP updated successfully",
	})
}

// HandleGetMCPConfig retrieves a specific MCP by ID
func HandleGetMCPConfig(w http.ResponseWriter, mcpID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching mcp config with ID: %s", mcpID)

	query := `
        SELECT a.id, a.name, a.description, type, arguments, cache_tools_list, event_messages, a.is_active, a.created_at, a.updated_at, a.created_by
        FROM mcps a
    JOIN api_keys ak2 ON a.created_by = ak2.id
    JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
    JOIN workspaces w1 ON ak1.workspace_id = w1.id
    JOIN workspaces w2 ON ak2.workspace_id = w2.id 
        AND w2.organization_id = w1.organization_id
    WHERE ak1.id = $2 AND a.id = $1`

	var (
		id, name, description, tooltype, arguments, event_messages string
		isActive, cache_tools_list                                 bool
		createdAt, updatedAt                                       time.Time
		createdBy                                                  string
	)

	err := DB.QueryRow(query, mcpID, apiKeyId).Scan(
		&id, &name, &description, &tooltype,
		&arguments, &cache_tools_list, &event_messages, &isActive, &createdAt, &updatedAt, &createdBy,
	)

	if err == sql.ErrNoRows {
		log.Printf("No mcp found with ID: %s", mcpID)
		WriteNotFoundError(w, "MCP not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving mcp config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve mcp")
		return
	}

	templateConfig := map[string]interface{}{
		"id":             id,
		"name":           name,
		"description":    description,
		"type":           tooltype,
		"request_config": arguments,
		"created_at":     createdAt,
		"updated_at":     updatedAt,
	}

	json.NewEncoder(w).Encode(templateConfig)
}

// HandleGetAllMCPConfigs retrieves all MCPs
func HandleGetAllMCPConfigs(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	query := `
        SELECT m.id, m.name, m.description, m.type, m.arguments, m.cache_tools_list, m.event_messages, m.is_active, m.created_at, m.updated_at, m.created_by
        FROM mcps m
		JOIN api_keys ak2
			ON m.created_by = ak2.id
		JOIN api_keys ak1
			ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1
			ON ak1.workspace_id = w1.id
		JOIN workspaces w2
			ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1
		ORDER BY m.created_at DESC;`

	rows, err := DB.Query(query, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve MCPs")
		return
	}
	defer rows.Close()

	var mcpConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, name, description, tooltype, arguments, event_messages string
			isActive, cache_tools_list                                 bool
			createdAt, updatedAt                                       time.Time
			createdBy                                                  string
		)

		if err := rows.Scan(
			&id, &name, &description, &tooltype,
			&arguments, &cache_tools_list, &event_messages, &isActive, &createdAt, &updatedAt, &createdBy,
		); err != nil {
			WriteInternalServerError(w, "Error scanning MCPs")
			return
		}

		mcpConfigs = append(mcpConfigs, map[string]interface{}{
			"id":             id,
			"name":           name,
			"description":    description,
			"type":           tooltype,
			"request_config": arguments,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		})
	}

	json.NewEncoder(w).Encode(mcpConfigs)
}

// HandleDeleteMCPConfig deletes MCP
func HandleDeleteMCPConfig(w http.ResponseWriter, mcpID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Deleting MCP config with ID: %s", mcpID)

	querycheck := `
			SELECT EXISTS(
				SELECT 1
				FROM mcps a
				JOIN api_keys ak2 ON a.created_by = ak2.id
				JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
				JOIN workspaces w1 ON ak1.workspace_id = w1.id
				JOIN workspaces w2 ON ak2.workspace_id = w2.id 
					AND w2.organization_id = w1.organization_id
				WHERE ak1.id = $2 AND a.id = $1
			)`

	var exists bool

	err := DB.QueryRow(querycheck, mcpID, apiKeyId).Scan(&exists)
	if err != nil {
		log.Printf("Error retrieving mcp: %v", err)
		WriteInternalServerError(w, "Failed to retrieve mcp")
		return
	}

	if !exists {
		log.Printf("No mcp found with ID: %s", mcpID)
		WriteNotFoundError(w, "MCP not found")
		return
	}

	query := `DELETE FROM mcps WHERE id = $1 RETURNING id`

	var id string
	err = DB.QueryRow(query, mcpID).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No MCP found with ID: %s", mcpID)
		WriteNotFoundError(w, "MCP not found")
		return
	} else if err != nil {
		log.Printf("Error deleting mcp config: %v", err)
		WriteInternalServerError(w, "Failed to delete MCP")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "MCP deleted successfully",
	})
}

// HandleGetAllProviderConfigs retrieves all Providers
func HandleGetAllProviderConfigs(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	providerConfigs, shouldReturn := GetAllProviderFunction(w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(providerConfigs)
}

// HandleGetAllProviderConfigsAPI retrieves all Providers
func HandleGetAllProviderConfigsAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	providerConfigs, shouldReturn := GetAllProviderFunction(w)
	if shouldReturn {
		return
	}

	var providerConfigsAPI []map[string]interface{}

	for _, provider := range providerConfigs {

		providerConfigAPI := map[string]interface{}{
			"display_name":      provider["display_name"],
			"provider":          provider["provider"],
			"provider_name":     provider["provider_name"],
			"provider_type":     provider["provider_type"],
			"model":             provider["model"],
			"additional_fields": provider["additional_fields"],
			"description":       provider["description"],
		}

		providerConfigsAPI = append(providerConfigsAPI, providerConfigAPI)
	}

	json.NewEncoder(w).Encode(providerConfigsAPI)
}

func GetAllProviderFunction(w http.ResponseWriter) ([]map[string]interface{}, bool) {
	query := `
        SELECT id, provider, type, model, provider_name, display_name, display_picture, is_active, additional_fields, created_at, updated_at, description
        FROM providers where is_active = true`

	rows, err := DB.Query(query)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve providers")
		return nil, true
	}
	defer rows.Close()

	var providerConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, provider, provider_type, model, provider_name, display_name, display_picture, description string
			is_active                                                                                     bool
			created_at, updated_at                                                                        time.Time
			additional_fields                                                                             json.RawMessage
		)

		if err := rows.Scan(
			&id, &provider, &provider_type, &model,
			&provider_name, &display_name, &display_picture, &is_active, &additional_fields,
			&created_at, &updated_at, &description,
		); err != nil {
			WriteInternalServerError(w, "Error scanning Providers")
			return nil, true
		}

		providerConfigs = append(providerConfigs, map[string]interface{}{
			"id":                id,
			"provider":          provider,
			"provider_type":     provider_type,
			"model":             model,
			"provider_name":     provider_name,
			"display_name":      display_name,
			"display_picture":   display_picture,
			"additional_fields": additional_fields,
			"is_active":         is_active,
			"created_at":        created_at,
			"updated_at":        updated_at,
			"description":       description,
		})
	}
	return providerConfigs, false
}

// HandleGetConversation retrieves a specific Conversation by ID
func HandleGetConversation(w http.ResponseWriter, conversationID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching conversation with ID: %s", conversationID)

	conversationConfigs, shouldReturn := GetConversationFunction(conversationID, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(conversationConfigs)
}

// HandleGetConversation retrieves a specific Conversation by ID
func HandleGetConversationAPI(w http.ResponseWriter, conversationID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching conversation with ID: %s", conversationID)

	conversationConfigs, shouldReturn := GetConversationFunction(conversationID, w)
	if shouldReturn {
		return
	}

	convConfigAPI := map[string]interface{}{
		"id":            conversationConfigs["id"],
		"agent_id":      conversationConfigs["agent_id"],
		"agent_name":    conversationConfigs["agent_name"],
		"user_id":       conversationConfigs["user_id"],
		"user_name":     conversationConfigs["user_name"],
		"avatar_id":     conversationConfigs["avatar_id"],
		"context":       conversationConfigs["context"],
		"is_recorded":   conversationConfigs["is_recorded"],
		"recording_url": conversationConfigs["recording_url"],
		"status":        conversationConfigs["status"],
		"transcript":    conversationConfigs["transcript"],
		"created_at":    conversationConfigs["created_at"],
		"updated_at":    conversationConfigs["updated_at"],
		"duration":      conversationConfigs["duration"],
	}

	json.NewEncoder(w).Encode(convConfigAPI)
}

// HandleGetConversationsByAgentAPI retrieves all conversations for an agent ID
func HandleGetConversationsByAgentAPI(w http.ResponseWriter, agentID string, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching conversations for agent ID: %s", agentID)

	rows, err := DB.Query(`
		SELECT id
		FROM conversations
		WHERE agent_id = $1
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		log.Printf("Error fetching conversations for agent %s: %v", agentID, err)
		WriteInternalServerError(w, "Failed to fetch conversations")
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var conversationID string
		if err := rows.Scan(&conversationID); err != nil {
			log.Printf("Error scanning conversation id: %v", err)
			WriteInternalServerError(w, "Failed to scan conversations")
			return
		}

		conversationConfigs, shouldReturn := GetConversationFunctionNoS3(conversationID, w)
		if shouldReturn {
			return
		}

		convConfigAPI := map[string]interface{}{
			"id":            conversationConfigs["id"],
			"agent_id":      conversationConfigs["agent_id"],
			"agent_name":    conversationConfigs["agent_name"],
			"user_id":       conversationConfigs["user_id"],
			"user_name":     conversationConfigs["user_name"],
			"avatar_id":     conversationConfigs["avatar_id"],
			"context":       conversationConfigs["context"],
			"is_recorded":   conversationConfigs["is_recorded"],
			"recording_url": conversationConfigs["recording_url"],
			"status":        conversationConfigs["status"],
			"created_at":    conversationConfigs["created_at"],
			"updated_at":    conversationConfigs["updated_at"],
			"duration":      conversationConfigs["duration"],
		}
		results = append(results, convConfigAPI)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating conversations: %v", err)
		WriteInternalServerError(w, "Failed to iterate conversations")
		return
	}

	json.NewEncoder(w).Encode(results)
}

func GetConversationFunction(conversationID string, w http.ResponseWriter) (map[string]interface{}, bool) {
	query := `
		SELECT c.id, c.agent_id, a.agent_name, c.context, c.meta_data, c.join_link, c.status, c.is_active, c.created_at, c.updated_at, c.job_id, f.feedback, f.rating, a.record,
		COALESCE(c.snippets, '{}') as snippets, COALESCE(c.user_name, '') as user_name, COALESCE(c.user_id, '') as user_id, COALESCE(c.chat_history, '{}') as chat_history, COALESCE(c.summary, '') as summary, a.type, COALESCE(u.usage_json, '{}') AS usage_json, c.avatar_key_id, COALESCE(c.message, ''),
		COALESCE(cl.usage_duration, 0) AS duration
		FROM conversations c
		JOIN agents a ON a.id = c.agent_id
		LEFT JOIN usage_metrics u ON u.conversation_id = c.id
		LEFT JOIN feedback f ON f.conversation_id = c.id
		LEFT JOIN LATERAL (
			SELECT usage_duration
			FROM credit_limit_logs
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) cl ON true
		WHERE c.id = $1`

	var (
		id, agent_id, agent_name, context, meta_data, join_link, status, summary, atype, message string
		createdAt, updatedAt                                                                     time.Time
		is_active, is_recorded                                                                   bool
		job_id, feedback                                                                         sql.NullString
		rating                                                                                   sql.NullInt16
		user_name, user_id, usage_json, avatar_id                                                string
		snippets, chat_history                                                                   json.RawMessage
		durationSeconds                                                                          float64
	)

	err := DB.QueryRow(query, conversationID).Scan(
		&id, &agent_id, &agent_name, &context, &meta_data, &join_link, &status, &is_active, &createdAt, &updatedAt, &job_id, &feedback, &rating, &is_recorded, &snippets, &user_name, &user_id, &chat_history, &summary, &atype, &usage_json, &avatar_id, &message, &durationSeconds,
	)

	if err == sql.ErrNoRows {
		log.Printf("No conversation found with ID: %s", conversationID)
		WriteNotFoundError(w, "Conversation not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error retrieving conversation config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve conversation")
		return nil, true
	}
	bucket := configs.GetEnv("AWS_BUCKET_ADDITIONAL")
	region := configs.GetEnv("AWS_REGION")
	key := fmt.Sprintf("egress/%s/room_recording.mp4", id)
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	signurl, signurl_download, err := PreSignURL(bucket, url, region)
	if err != nil {
		log.Printf("Error PreSign: %v", err)
		signurl = ""
		signurl_download = ""
	}

	isAvailable := CheckS3FileAvailability(signurl_download)
	if !isAvailable {
		signurl = "" // reset URL if file doesn’t exist or is inaccessible
		signurl_download = ""
		log.Printf("PreSign1: %v", signurl_download)
	}

	url1 := fmt.Sprintf("egress/%s/transcript.json", id)
	url2 := fmt.Sprintf("egress/%s/transcript.jsonl", id)

	var entries []Transcript
	var tran_json string
	if ok, _ := S3ObjectExists(bucket, url1, region); ok {

		data, err := ReadS3ObjectBytes(bucket, url1, region)
		if err != nil {
			log.Fatalf("failed to read %s: %v", url1, err)
		}

		tran_json = string(data)
		entries, err = BuildTranscripts(data, conversationID)
		if err != nil {
			log.Fatalf("failed to build transcripts: %v", err)
		}

	} else if ok, _ := S3ObjectExists(bucket, url2, region); ok {

		entries, err = ReadS3JSONLines[Transcript](bucket, url2, region)
		if err != nil {
			log.Fatalf("failed to read jsonl transcripts: %v", err)
		}

	} else {
		log.Printf("❌ No transcript found for %s", id)
	}

	trans := entries

	conversationConfigs := map[string]interface{}{
		"id":                     id,
		"agent_id":               agent_id,
		"agent_name":             agent_name,
		"context":                context,
		"meta_data":              json.RawMessage(meta_data),
		"join_link":              join_link,
		"status":                 status,
		"is_active":              is_active,
		"created_at":             createdAt,
		"updated_at":             updatedAt,
		"job_id":                 job_id,
		"feedback":               feedback,
		"rating":                 rating,
		"transcript":             trans,
		"tran_json":              tran_json,
		"recording_url":          signurl,
		"recording_url_download": signurl_download,
		"is_recorded":            is_recorded,
		"snippets":               snippets,
		"user_name":              user_name,
		"user_id":                user_id,
		"chat_history":           chat_history,
		"summary":                summary,
		"type":                   atype,
		"usage_json":             usage_json,
		"duration":               durationSeconds,
		"avatar_id":              avatar_id,
		"message":                message,
	}
	return conversationConfigs, false
}

func GetConversationFunctionNoS3(conversationID string, w http.ResponseWriter) (map[string]interface{}, bool) {
	query := `
		SELECT c.id, c.agent_id, a.agent_name, c.context, c.meta_data, c.join_link, c.status, c.is_active, c.created_at, c.updated_at, c.job_id, f.feedback, f.rating, a.record,
		COALESCE(c.snippets, '{}') as snippets, COALESCE(c.user_name, '') as user_name, COALESCE(c.user_id, '') as user_id, COALESCE(c.chat_history, '{}') as chat_history, COALESCE(c.summary, '') as summary, a.type, COALESCE(u.usage_json, '{}') AS usage_json, c.avatar_key_id,
		COALESCE(cl.usage_duration, 0) AS duration
		FROM conversations c
		JOIN agents a ON a.id = c.agent_id
		LEFT JOIN usage_metrics u ON u.conversation_id = c.id
		LEFT JOIN feedback f ON f.conversation_id = c.id
		LEFT JOIN LATERAL (
			SELECT usage_duration
			FROM credit_limit_logs
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) cl ON true
		WHERE c.id = $1`

	var (
		id, agent_id, agent_name, context, meta_data, join_link, status, summary, atype string
		createdAt, updatedAt                                                            time.Time
		is_active, is_recorded                                                          bool
		job_id, feedback                                                                sql.NullString
		rating                                                                          sql.NullInt16
		user_name, user_id, usage_json, avatar_id                                       string
		snippets, chat_history                                                          json.RawMessage
		durationSeconds                                                                 float64
	)

	err := DB.QueryRow(query, conversationID).Scan(
		&id, &agent_id, &agent_name, &context, &meta_data, &join_link, &status, &is_active, &createdAt, &updatedAt, &job_id, &feedback, &rating, &is_recorded, &snippets, &user_name, &user_id, &chat_history, &summary, &atype, &usage_json, &avatar_id, &durationSeconds,
	)

	if err == sql.ErrNoRows {
		log.Printf("No conversation found with ID: %s", conversationID)
		WriteNotFoundError(w, "Conversation not found")
		return nil, true
	} else if err != nil {
		log.Printf("Error retrieving conversation config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve conversation")
		return nil, true
	}
	bucket := configs.GetEnv("AWS_BUCKET_ADDITIONAL")
	region := configs.GetEnv("AWS_REGION")
	key := fmt.Sprintf("egress/%s/room_recording.mp4", id)
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	signurl, signurl_download, err := PreSignURL(bucket, url, region)
	if err != nil {
		log.Printf("Error PreSign: %v", err)
		signurl = ""
		signurl_download = ""
	}

	conversationConfigs := map[string]interface{}{
		"id":                     id,
		"agent_id":               agent_id,
		"agent_name":             agent_name,
		"context":                context,
		"meta_data":              json.RawMessage(meta_data),
		"join_link":              join_link,
		"status":                 status,
		"is_active":              is_active,
		"created_at":             createdAt,
		"updated_at":             updatedAt,
		"job_id":                 job_id,
		"feedback":               feedback,
		"rating":                 rating,
		"recording_url":          signurl,
		"recording_url_download": signurl_download,
		"is_recorded":            is_recorded,
		"snippets":               snippets,
		"user_name":              user_name,
		"user_id":                user_id,
		"chat_history":           chat_history,
		"summary":                summary,
		"type":                   atype,
		"usage_json":             usage_json,
		"duration":               durationSeconds,
		"avatar_id":              avatar_id,
	}
	return conversationConfigs, false
}

func BuildTranscripts(
	data []byte,
	conversationID string,
) ([]Transcript, error) {

	var input InputFile
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	var transcripts []Transcript

	for _, item := range input.Items {

		if item.Type != "message" &&
			item.Type != "function_call" &&
			item.Type != "function_call_output" {
			continue
		}

		var extra json.RawMessage
		if item.Extra != nil {
			b, _ := json.Marshal(item.Extra)
			extra = b
		}

		t := Transcript{
			Conversation_id: conversationID,
			Type:            item.Type,
			Name:            item.Name,
			CallID:          item.CallID,
			Extra:           extra,
		}

		// Timestamp
		if v, ok := item.Metrics["started_speaking_at"].(float64); ok {
			t.Message_timestamp = int32(v)
			t.Timestamp = time.Unix(int64(v), 0).UTC().Format(time.RFC3339)
		} else {
			t.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		switch item.Type {

		case "message":
			t.Role = item.Role
			t.Content = strings.Join(item.Content, " ")

		case "function_call":
			t.Role = "assistant"
			t.Arguments = item.Arguments

		case "function_call_output":
			t.Role = "tool"
			t.Output = item.Output
			t.IsError = item.IsError
		}

		transcripts = append(transcripts, t)
	}

	return transcripts, nil
}

// HandleGetConversationStatus retrieves a specific Conversation by ID
func HandleGetConversationStatus(w http.ResponseWriter, conversationID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Fetching conversation with ID: %s", conversationID)

	query := `
        SELECT id, status
        FROM conversations
        WHERE id = $1`

	var (
		id, status string
	)

	err := DB.QueryRow(query, conversationID).Scan(
		&id, &status,
	)

	if err == sql.ErrNoRows {
		log.Printf("No conversation found with ID: %s", conversationID)
		WriteNotFoundError(w, "Conversation not found")
		return
	} else if err != nil {
		log.Printf("Error retrieving conversation config: %v", err)
		WriteInternalServerError(w, "Failed to retrieve conversation")
		return
	}

	conversationConfigs := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	json.NewEncoder(w).Encode(conversationConfigs)
}

// HandleGetAllConversations retrieves all Conversation
func HandleGetAllConversations(w http.ResponseWriter, apiKeyId string) {
	w.Header().Set("Content-Type", "application/json")
	query := `
       SELECT
			c.id,
			c.agent_id,
			a.agent_name,
			c.context,
			c.meta_data,
			c.join_link,
			c.status,
			c.is_active,
			c.created_at,
			c.updated_at,
			c.job_id,
			f.feedback,
			f.rating,
			COALESCE(u.status, '') AS usage_status,
			COALESCE(u.usage_json, '{}') AS usage_json,
			COALESCE(c.snippets, '{}') AS snippets,
			COALESCE(c.user_name, '') AS user_name,
			COALESCE(c.user_id, '') AS user_id,
			COALESCE(c.chat_history, '{}') AS chat_history,
			a.type,
			c.avatar_key_id
		FROM conversations c
		JOIN agents a ON a.id = c.agent_id
		LEFT JOIN usage_metrics u ON u.conversation_id = c.id
		LEFT JOIN feedback f ON f.conversation_id = c.id
		JOIN api_keys ak2 ON c.created_by = ak2.id
		JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1 ON ak1.workspace_id = w1.id
		JOIN workspaces w2 ON ak2.workspace_id = w2.id AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1
		ORDER BY c.created_at DESC;`

	rows, err := DB.Query(query, apiKeyId)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve conversations")
		return
	}
	defer rows.Close()

	var conversationConfigs []map[string]interface{}
	for rows.Next() {
		var (
			id, agent_id, agent_name, context, meta_data, join_link, status, usage_status, usage_json, atype string
			createdAt, updatedAt                                                                             time.Time
			is_active                                                                                        bool
			job_id, feedback                                                                                 sql.NullString
			rating                                                                                           sql.NullInt16
			user_name, user_id, avatar_id                                                                    string
			snippets, chat_history                                                                           json.RawMessage
		)

		if err := rows.Scan(
			&id, &agent_id, &agent_name, &context, &meta_data, &join_link, &status, &is_active, &createdAt, &updatedAt, &job_id, &feedback, &rating, &usage_status, &usage_json, &snippets, &user_name, &user_id, &chat_history, &atype, &avatar_id,
		); err != nil {
			log.Printf("Error: %v", err)
			WriteInternalServerError(w, "Error scanning conversations")
			return
		}

		conversationConfigs = append(conversationConfigs, map[string]interface{}{
			"id":           id,
			"agent_id":     agent_id,
			"agent_name":   agent_name,
			"context":      context,
			"meta_data":    json.RawMessage(meta_data),
			"join_link":    join_link,
			"status":       status,
			"is_active":    is_active,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
			"job_id":       job_id,
			"feedback":     feedback,
			"rating":       rating,
			"usage_status": usage_status,
			"usage_json":   usage_json,
			"snippets":     snippets,
			"user_name":    user_name,
			"user_id":      user_id,
			"chat_history": chat_history,
			"type":         atype,
			"avatar_id":    avatar_id,
		})
	}

	json.NewEncoder(w).Encode(conversationConfigs)
}

func VoicetoVideoSessionCreation(
	apiKey string,
	avatarId string,
	conversationId string,
) (string, error) {
	//SINGLE-PHASE DB TRANSACTION
	tx, err := DB.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Validate user request
	var status string
	var userUUID uuid.UUID

	err = tx.QueryRow(`
        SELECT o_status, o_api_key_id
        FROM validate_user_request($1, $2)
    `, apiKey, conversationId).Scan(&status, &userUUID)
	if err != nil {
		return "", fmt.Errorf("validate_user_request failed: %v", err)
	}

	switch status {
	case "NO_API_KEY_FOUND":
		return "", fmt.Errorf("invalid API key")
	case "NOT_ENOUGH_SESSION_LEFT":
		return "", fmt.Errorf("concurrent session limit reached")
	case "NOT_ENOUGH_CREDIT_LEFT":
		return "", fmt.Errorf("not enough credits")
	case "REQUEST_APPROVED":
		// continue
	default:
		return "", fmt.Errorf("unknown validation status: %s", status)
	}

	var maxCallDuration int
	err = tx.QueryRow(`
        SELECT cl.max_session_duration
			FROM credit_limits cl
			JOIN organizations o ON o.id = cl.organization_id
			JOIN workspaces w ON w.organization_id = o.id
			JOIN api_keys ak ON ak.workspace_id = w.id
			WHERE ak.key_hash = $1
			LIMIT 1;
    `, apiKey).Scan(&maxCallDuration)
	if err != nil {
		return "", fmt.Errorf("credit_limits load failed: %v", err)
	}

	delay := time.Duration(maxCallDuration+2) * time.Minute
	ScheduleOneTimeResetCreditJob(delay, conversationId)

	// Insert conversation (job_id NULL)
	_, err = tx.Exec(`
        INSERT INTO conversations (
            id, agent_id, context, join_link, status, is_active,
            created_at, updated_at, created_by, meta_data, job_id,
            user_name, user_id, type, avatar_key_id
        )
        VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
        )
    `,
		conversationId,
		"00000000-0000-0000-0000-000000000000", // Dummy agent ID for VoicetoVideo
		"{}",
		"",
		"Initializing",
		true,
		time.Now(),
		time.Now(),
		apiKey,
		"[]",
		nil,
		"",
		"",
		"vtva",
		avatarId,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create conversation: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit failed: %v", err)
	}
	committed = true

	// Update job_id quickly
	_, err = DB.Exec(`
        UPDATE conversations
        SET job_id = $1, updated_at = NOW()
        WHERE id = $2
    `, "vtva job", conversationId)
	if err != nil {
		return "", fmt.Errorf("failed to update job_id: %v", err)
	}

	return conversationId, nil
}

func GetLiveKitJoinToken(body TokenSourceRequest) string {
	at := auth.NewAccessToken(os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))

	// If this room doesn't exist, it'll be automatically created when
	// the first participant joins
	roomName := body.RoomName
	if roomName == "" {
		roomName = "quickstart-room"
	}
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant)

	if body.RoomConfig != nil {
		at.SetRoomConfig(body.RoomConfig)
	}

	// Participant related fields.
	// `participantIdentity` will be available as LocalParticipant.identity
	// within the livekit-client SDK
	if body.ParticipantIdentity != "" {
		at.SetIdentity(body.ParticipantIdentity)
	} else {
		at.SetIdentity("quickstart-identity")
	}
	if body.ParticipantName != "" {
		at.SetName(body.ParticipantName)
	} else {
		at.SetName("quickstart-username")
	}
	if len(body.ParticipantMetadata) > 0 {
		at.SetMetadata(body.ParticipantMetadata)
	}
	if len(body.ParticipantAttributes) > 0 {
		at.SetAttributes(body.ParticipantAttributes)
	}

	token, _ := at.ToJWT()
	return token
}

func GetLiveKitListToken(body TokenSourceRequest) (string, error) {
	at := auth.NewAccessToken(
		os.Getenv("LIVEKIT_API_KEY"),
		os.Getenv("LIVEKIT_API_SECRET"),
	)

	// ✅ Grant permissions
	grant := &auth.VideoGrant{
		RoomList: true, // --list equivalent
	}

	// Optional room join
	if body.RoomName != "" {
		grant.Room = body.RoomName
	}

	at.AddGrant(grant)

	// Optional room config
	if body.RoomConfig != nil {
		at.SetRoomConfig(body.RoomConfig)
	}

	// Participant identity
	if body.ParticipantIdentity != "" {
		at.SetIdentity(body.ParticipantIdentity)
	} else {
		at.SetIdentity("anonymous")
	}

	// Participant name
	if body.ParticipantName != "" {
		at.SetName(body.ParticipantName)
	}

	// Metadata & attributes
	if len(body.ParticipantMetadata) > 0 {
		at.SetMetadata(body.ParticipantMetadata)
	}
	if len(body.ParticipantAttributes) > 0 {
		at.SetAttributes(body.ParticipantAttributes)
	}

	token, err := at.ToJWT()
	if err != nil {
		return "", err
	}

	return token, nil
}

func HandleConversationCreation(
	apiKey string,
	agentId string,
	apiKeyId string,
	context json.RawMessage,
	metaData json.RawMessage,
	uname string,
	uid string,
	roomId string,
	meetingURL string,
	externalMeeting bool,
	avatarId string,
	mode string,
) (string, string, Agent, error) {

	agentUUID, err := uuid.Parse(agentId)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("invalid agent ID: %v", err)
	}

	conversationId := uuid.New().String()

	// ---------------------------------------------------
	//               SINGLE-PHASE DB TRANSACTION
	// ---------------------------------------------------
	tx, err := DB.Begin()
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("failed to start transaction: %v", err)
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// If missing API key, resolve it
	if apiKey == "" && apiKeyId == "" {
		err = tx.QueryRow(`
            SELECT agents.created_by, api_keys.key_hash
            FROM agents
            JOIN api_keys ON agents.created_by = api_keys.id
            WHERE agents.id = $1
        `, agentUUID).Scan(&apiKeyId, &apiKey)
		if err != nil {
			return "", "", Agent{}, fmt.Errorf("failed to fetch agent owner key: %v", err)
		}
	}

	addon := ""
	if externalMeeting {
		addon = "em"
	}

	// Validate user request
	var status string
	var userUUID uuid.UUID

	err = tx.QueryRow(`
        SELECT o_status, o_api_key_id
        FROM validate_user_request($1, $2, $3, $4, $5)
    `, apiKey, agentUUID, conversationId, addon, mode).Scan(&status, &userUUID)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("validate_user_request failed: %v", err)
	}

	switch status {
	case "NO_API_KEY_FOUND":
		return "", "", Agent{}, fmt.Errorf("invalid API key")
	case "AGENT_NOT_FOUND":
		return "", "", Agent{}, fmt.Errorf("agent not found")
	case "NOT_ENOUGH_SESSION_LEFT":
		return "", "", Agent{}, fmt.Errorf("concurrent session limit reached")
	case "NOT_ENOUGH_CREDIT_LEFT":
		return "", "", Agent{}, fmt.Errorf("not enough credits")
	case "REQUEST_APPROVED":
		// continue
	default:
		return "", "", Agent{}, fmt.Errorf("unknown validation status: %s", status)
	}

	// Safe marshal
	jsonContext, err := json.Marshal(context)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("failed to marshal context: %v", err)
	}

	// Load agent
	var metadata, avatars, agentName, callbackURL, atype string
	var recordRoom, defaultSystemPrompt bool
	var callbackEvents []string

	err = tx.QueryRow(`
        SELECT config::text, avatars::text, agent_name, record,
               COALESCE(callback_url, ''),
               COALESCE(callback_events, '{}'), type,
			   default_system_prompt
        FROM agents
        WHERE id = $1 AND created_by = $2
    `, agentUUID, apiKeyId).Scan(
		&metadata,
		&avatars,
		&agentName,
		&recordRoom,
		&callbackURL,
		pq.Array(&callbackEvents),
		&atype,
		&defaultSystemPrompt,
	)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("agent load failed: %v", err)
	}

	// Parse metadata JSON
	var agentConfig map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &agentConfig); err != nil {
		return "", "", Agent{}, fmt.Errorf("agent metadata JSON invalid: %v", err)
	}
	// Select random agent
	selectedAgent, selectedCount, agentErr := getRandomAgent(avatars, avatarId)
	if agentErr != nil {
		return "", "", Agent{}, fmt.Errorf("getRandomAgent failed: %v, %v", agentErr, selectedCount)
	}

	selectedAgent.Mode = atype
	selectedAgent.AgentId = agentName

	// Load avatar metadata
	var defaultPrompt, avatarName, visualConfig, imageURL string
	err = tx.QueryRow(`
        SELECT default_prompt, avatar_name, visual_config, image_url
        FROM avatars
        WHERE avatar_key_id = $1
    `, selectedAgent.AvatarID).Scan(&defaultPrompt, &avatarName, &visualConfig, &imageURL)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("avatar config load failed: %v", err)
	}

	var vConfig types.VisualConfig

	if jsonErr1 := json.Unmarshal([]byte(visualConfig), &vConfig); jsonErr1 != nil {
		return "", "", Agent{}, fmt.Errorf("visual Config params invalid: %v", jsonErr1)
	}
	selectedAgent.AvatarImageURL = imageURL
	selectedAgent.Avatar = vConfig
	selectedAgent.PersonaName = agentName
	if defaultSystemPrompt {
		selectedAgent.PersonaPrompt = selectedAgent.PersonaPrompt + "\n\n" + defaultPrompt
	}
	selectedAgent.PersonaPrompt = strings.ReplaceAll(
		selectedAgent.PersonaPrompt,
		"<AVATAR_NAME>",
		agentName,
	)

	var conContext ConversationalContext

	if jsonErr2 := json.Unmarshal([]byte(context), &conContext); jsonErr2 != nil {
		return "", "", Agent{}, fmt.Errorf("context Config params invalid: %v", jsonErr2)
	}

	if conContext.Text != "" {
		selectedAgent.PersonaPrompt = selectedAgent.PersonaPrompt + "\n\n" + "# Conversationl Context : \n" + string(jsonContext)
	}

	if conContext.WakePhrase != "" {
		selectedAgent.Config.STT.WakePhrase = conContext.WakePhrase
	}
	if externalMeeting {
		selectedAgent.Config.STT.Provider = "trugen-ext-meetings"
	}

	selectedAgent.AvatarName = avatarName
	selectedAgent.RecordRoom = recordRoom
	selectedAgent.SessionKey = "session-" + roomId
	selectedAgent.AvatarId = selectedAgent.AvatarID
	selectedAgent.ConversationId = conversationId
	if externalMeeting {
		selectedAgent.ConnectionType = "email_dispatch"
	} else {
		selectedAgent.ConnectionType = "website"
	}

	if callbackURL != "" {
		selectedAgent.Callback = &Callback{
			URL:            callbackURL,
			EventsToListen: callbackEvents,
		}
	}

	if meetingURL != "" {
		selectedAgent.Communication = &Communication{
			MeetingURL: meetingURL,
			Provider:   "recall",
		}
	}

	//Need to remove this after front end changes
	selectedAgent.IdleTimeout.Messages = selectedAgent.IdleTimeout.FillerPhrases
	selectedAgent.Config.VAD.Provider = "silero"
	// Knowledge bases
	kbRows, err := tx.Query(`
        SELECT kb.name, kb.description, kb.namespace, kb.index, COALESCE(tk.mode, 'agentic rag') AS mode
        FROM agents_kb tk
        JOIN knowledge_base kb ON kb.id = tk.knowledge_base_id
        WHERE tk.agent_id = $1
    `, agentUUID)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("knowledge base load failed: %v", err)
	}
	defer kbRows.Close()

	var kbList []KnowledgeBase
	for kbRows.Next() {
		kb := NewKnowledgeBase()
		if err := kbRows.Scan(&kb.Name, &kb.Description, &kb.Namespace, &kb.IndexHost, &kb.Mode); err != nil {
			return "", "", Agent{}, fmt.Errorf("failed scanning KB: %v", err)
		}
		if kb.Mode == "" { // if mode is empty, set default
			kb.Mode = "agentic rag"
		}
		kbList = append(kbList, kb)
	}
	selectedAgent.KnowledgeBase = kbList

	// Memories
	if memoryConfig, ok := agentConfig["memory"]; ok {
		memoryBytes, err := json.Marshal(memoryConfig)
		if err == nil {
			var memory Memory
			if err := json.Unmarshal(memoryBytes, &memory); err == nil {
				selectedAgent.Memory = memory
			}
		}
	}

	selectedAgent.GatewayToken = agentConfig["gateway_token"].(string)
	selectedAgent.OpenClawURL = agentConfig["openclaw_url"].(string)

	// MCP
	mcpRows, mcperr := tx.Query(`
        SELECT m.type, m.name, m.description, m.arguments, m.cache_tools_list, m.event_messages
        FROM agents_mcp am
        JOIN mcps m ON m.id = am.mcp_id
        WHERE am.agent_id = $1
    `, agentUUID)
	if mcperr != nil {
		return "", "", Agent{}, fmt.Errorf("MCP load failed: %v", mcperr)
	}
	defer mcpRows.Close()

	var mcpList []MCP
	for mcpRows.Next() {
		var (
			mcpType, mcpName, mcpDesc string
			argsJSON, evtJSON         json.RawMessage
			cacheTools                bool
		)
		if err := mcpRows.Scan(&mcpType, &mcpName, &mcpDesc, &argsJSON, &cacheTools, &evtJSON); err != nil {
			return "", "", Agent{}, fmt.Errorf("MCP scan failed: %v", err)
		}

		var requestConfig RequestConfigMCP
		if jsonErr := json.Unmarshal(argsJSON, &requestConfig); jsonErr != nil {
			return "", "", Agent{}, fmt.Errorf("MCP params invalid: %v", jsonErr)
		}

		mcpList = append(mcpList, MCP{
			Type:          mcpType,
			Name:          mcpName,
			Description:   mcpDesc,
			RequestConfig: requestConfig,
			EventMessages: evtJSON,
		})
	}

	// Tools
	toolRows, toolerr := tx.Query(`
        SELECT m.type, m.name, m.description, m.arguments, m.request_config, m.event_messages
        FROM agents_tool am
        JOIN tools m ON m.id = am.tool_id
        WHERE am.agent_id = $1
    `, agentUUID)
	if toolerr != nil {
		return "", "", Agent{}, fmt.Errorf("Tool load failed: %v", toolerr)
	}
	defer toolRows.Close()

	var toolList_C []Tool
	var toolList_H []Tool
	for toolRows.Next() {
		var (
			toolType, toolName, toolDesc string
			argsJSON, evtJSON, reqJSON   json.RawMessage
		)
		if err := toolRows.Scan(&toolType, &toolName, &toolDesc, &argsJSON, &reqJSON, &evtJSON); err != nil {
			return "", "", Agent{}, fmt.Errorf("Tool scan failed: %v", err)
		}

		var requestConfig RequestConfigTool
		if jsonErr := json.Unmarshal(reqJSON, &requestConfig); jsonErr != nil {
			return "", "", Agent{}, fmt.Errorf("Tool params invalid: %v", jsonErr)
		}

		switch toolType {
		case "tool.client":
			toolList_C = append(toolList_C, Tool{
				Type:          toolType,
				Schema:        argsJSON,
				RequestConfig: requestConfig,
				EventMessages: evtJSON,
			})
		case "tool.api":
			toolList_H = append(toolList_H, Tool{
				Type:          toolType,
				Schema:        argsJSON,
				RequestConfig: requestConfig,
				EventMessages: evtJSON,
			})
		}
	}

	var actions = Actions{MCPServers: mcpList, ClientTools: toolList_C, HTTPTools: toolList_H}

	selectedAgent.Actions = actions

	var maxCallDuration int
	err = tx.QueryRow(`
        SELECT cl.max_session_duration
			FROM credit_limits cl
			JOIN organizations o ON o.id = cl.organization_id
			JOIN workspaces w ON w.organization_id = o.id
			JOIN api_keys ak ON ak.workspace_id = w.id
			WHERE ak.key_hash = $1
			LIMIT 1;
    `, apiKey).Scan(&maxCallDuration)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("credit_limits load failed: %v", err)
	}

	if selectedAgent.ExitMessage.MaxCallDuration > 0 {
		maxCallDuration = selectedAgent.ExitMessage.MaxCallDuration / 60 // Convert seconds to minutes
	}

	delay := time.Duration(maxCallDuration+5) * time.Minute
	ScheduleOneTimeResetCreditJob(delay, conversationId)

	if mode != "" {
		selectedAgent.Mode = mode
		atype = mode
	}

	// Insert conversation (job_id NULL)
	_, err = tx.Exec(`
        INSERT INTO conversations (
            id, agent_id, context, join_link, status, is_active,
            created_at, updated_at, created_by, meta_data, job_id,
            user_name, user_id, type, avatar_key_id
        )
        VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
        )
    `,
		conversationId,
		agentId,
		jsonContext,
		roomId,
		"Initializing",
		true,
		time.Now(),
		time.Now(),
		apiKeyId,
		string(metaData),
		nil,
		uname,
		uid,
		atype,
		selectedAgent.Avatar.ID,
	)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("failed to create conversation: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", "", Agent{}, fmt.Errorf("commit failed: %v", err)
	}
	committed = true

	// Prepare payload
	payload := map[string]interface{}{
		"input": map[string]interface{}{
			"record_room":         recordRoom,
			"conversation_id":     conversationId,
			"room_id":             roomId,
			"timeout":             200,
			"user_id":             uid,
			"user_name":           uname,
			"callback":            callbackURL,
			"max_message_history": 20,
			"agent":               selectedAgent,
		},
	}

	jsonStr, _ := json.Marshal(payload) //only for dev
	str := string(jsonStr)
	log.Printf("Payload to Infra: %s", str)

	// agentNameCon := configs.GetEnv("LIVEKIT_AGENTNAME")

	// disErr := dispatchAgent(roomId, agentNameCon, str)
	// if disErr != nil {
	// 	delay = 10 * time.Second
	// 	ScheduleOneTimeResetCreditJob(delay, conversationId)
	// 	return "", "", Agent{}, fmt.Errorf("invalid dispatch response: %v", err)
	// }

	// Update job_id
	_, err = DB.Exec(`
        UPDATE conversations
        SET job_id = $1, updated_at = NOW()
        WHERE id = $2
    `, roomId, conversationId)
	if err != nil {
		return "", "", Agent{}, fmt.Errorf("failed to update job_id: %v", err)
	}

	return conversationId, roomId, selectedAgent, nil
}

// Handle external meeting addons
func HandleJoinExternalMeeting(w http.ResponseWriter, r *http.Request, roomId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.Header.Get("X-API-Key")
	apiKeyId := r.Context().Value("apiKeyId").(string)

	// Parse Join External Meetings request payload
	request_payload := JoinExternalMeetingsRequest{}
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	json.Unmarshal(body, &request_payload)
	if err = request_payload.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	ctxObj := map[string]interface{}{
		"text": request_payload.ConversationalContext,
	}
	if request_payload.WakePhrase != "" {
		ctxObj["wake_phrase"] = request_payload.WakePhrase
	}

	ctxBytes, err := json.Marshal(ctxObj)
	if err != nil {
		return
	}

	ConversationalContext := json.RawMessage(ctxBytes)

	conversationId, createdRoomID, agent, err := HandleConversationCreation(apiKey, request_payload.AgentID, apiKeyId, ConversationalContext, json.RawMessage("{}"), request_payload.UserName, request_payload.UserID, roomId, request_payload.MeetingURL, true, "", "")
	if err != nil {
		http.Error(w, "Error creating conversation: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Conversation created with ID: %s, %s, %s", conversationId, createdRoomID, agent.AvatarID)
		return
	}

	var agentName string
	err = DB.QueryRow(`
        SELECT agent_name
        FROM agents
        WHERE id = $1
    `, request_payload.AgentID).Scan(
		&agentName,
	)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	var agentEmail string
	err = DB.QueryRow(`
        SELECT email
        FROM agents
        WHERE id = $1
    `, request_payload.AgentID).Scan(
		&agentEmail,
	)
	if err != nil {
		log.Printf("[HandleJoinExternalMeeting] ERROR: could not get agent email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	agentDisplayName, videoUrl, success, err := PostAvatarRequest(agentEmail, roomId, request_payload.MeetingURL, "Admin")
	if err != nil || !success {
		log.Printf("[HandleJoinExternalMeeting] ERROR: PostAvatarRequest failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to get avatar session: %v", err)))
		return
	}

	isPosted, err := PostRecallRequest(request_payload.MeetingURL, agentDisplayName, roomId, videoUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.WriteHeader(http.StatusOK)
	statusMessage := "Agent request successfully created."
	if !isPosted {
		statusMessage = "Unable to create agent request."
	}
	w.Write([]byte(statusMessage))
}

func buildRecallRequestPayload(meetingUrl string, displayName string, lkRoomID string, videoUrl string) ([]byte, string, error) {
	avatarStream := os.Getenv("AVATAR_VIDEO_STREAM")
	recallWS := os.Getenv("RECALL_VIDEO_WS_URL")
	if recallWS == "" {
		return nil, "", fmt.Errorf("RECALL_VIDEO_WS_URL is not set")
	}

	cameraURL := videoUrl
	if cameraURL == "" {
		cleanAvatarStream := strings.TrimRight(avatarStream, "/")
		cameraURL = fmt.Sprintf("%s/%s?agent=%s", cleanAvatarStream, lkRoomID, url.QueryEscape(displayName))
	}
	wsURL := fmt.Sprintf("%s/%s", recallWS, lkRoomID)

	realtimeEndpoints := []map[string]interface{}{
		{
			"type": "websocket",
			"url":  wsURL,
			"events": []string{
				"participant_events.join",
				"participant_events.leave",
				"participant_events.speech_on",
				"participant_events.speech_off",
				"participant_events.screenshare_on",
				"participant_events.screenshare_off",
				"participant_events.webcam_on",
				"participant_events.webcam_off",
				"transcript.data",
				"transcript.partial_data",
				"video_separate_png.data",
			},
		},
	}

	payload := map[string]interface{}{
		"meeting_url": meetingUrl,
		"bot_name":    displayName,
		"output_media": map[string]interface{}{
			"camera": map[string]interface{}{
				"kind": "webpage",
				"config": map[string]interface{}{
					"url": cameraURL,
				},
			},
		},
		"recording_config": map[string]interface{}{
			"video_mixed_layout": "gallery_view_v2",
			"include_bot_in_recording": map[string]interface{}{
				"audio": true,
			},
			"video_separate_png": map[string]interface{}{
				"include_screenshare": true,
				"include_webcam":      true,
			},
			"transcript": map[string]interface{}{
				"provider": map[string]interface{}{
					"deepgram_streaming": map[string]interface{}{
						"model":        "nova-3",
						"language":     "en-US",
						"smart_format": true,
						"endpointing":  200,
						"numerals":     true,
						"keyterm":      []string{"Lisa", "TruGen"},
					},
				},
				"diarization": map[string]interface{}{
					"use_separate_streams_when_available": true,
				},
			},
			"realtime_endpoints": realtimeEndpoints,
		},
		"variant": map[string]string{
			"zoom":            "web_gpu",
			"google_meet":     "web_gpu",
			"microsoft_teams": "web_gpu",
			"webex":           "web_gpu",
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal recall payload: %w", err)
	}
	return jsonBody, wsURL, nil
}

func postRecallRequestWithLabel(label string, meetingUrl string, displayName string, lkRoomID string, videoUrl string) (bool, error) {
	recallURL := os.Getenv("RECALL_API_URL")
	recallToken := os.Getenv("RECALL_API_TOKEN")

	log.Printf("[%s] meetingUrl=%s displayName=%s lkRoomID=%s", label, meetingUrl, displayName, lkRoomID)

	if recallURL == "" {
		return false, fmt.Errorf("RECALL_API_URL is not set")
	}
	if recallToken == "" {
		return false, fmt.Errorf("RECALL_API_TOKEN is not set")
	}

	jsonBody, wsURL, err := buildRecallRequestPayload(meetingUrl, displayName, lkRoomID, videoUrl)
	if err != nil {
		return false, err
	}

	log.Printf("[%s] Recall WS endpoint - url=%s", label, wsURL)
	log.Printf("[%s] Request Body:\n%s", label, string(jsonBody))

	req, err := http.NewRequest("POST", recallURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create recall request: %w", err)
	}
	req.Header.Set("Authorization", recallToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("recall API request failed: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read recall response: %w", err)
	}

	log.Printf("[%s] Response status=%d body=%s", label, res.StatusCode, string(respBody))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return false, fmt.Errorf("recall API returned status %d: %s", res.StatusCode, string(respBody))
	}

	log.Printf("[%s] SUCCESS bot created", label)
	return true, nil
}

// Method to Post request to Recall.AI
func PostRecallRequest(meetingUrl string, displayName string, lkRoomID string, videoUrl string) (bool, error) {
	return postRecallRequestWithLabel("PostRecallRequest", meetingUrl, displayName, lkRoomID, videoUrl)
}

// PostRecallRequestV2 is kept for scheduled-job compatibility and uses the same payload as direct joins.
func PostRecallRequestV2(meetingUrl string, displayName string, lkRoomID string, videoUrl string) (bool, error) {
	return postRecallRequestWithLabel("PostRecallRequestV2", meetingUrl, displayName, lkRoomID, videoUrl)
}

// HandleRecallTrigger is a lightweight endpoint called by the Next.js start-agent route
// when an external meeting is detected. It sends a Recall.AI bot into the meeting
// using the LiveKit room that was already created and dispatched by the frontend.
// It does NOT create a new conversation or room — those already exist.
func HandleRecallTrigger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		MeetingURL     string `json:"meeting_url"`
		RoomName       string `json:"room_name"`
		DisplayName    string `json:"display_name"`
		ConversationID string `json:"conversation_id"`
		VideoURL       string `json:"video_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.MeetingURL == "" {
		http.Error(w, "meeting_url is required", http.StatusBadRequest)
		return
	}
	if req.RoomName == "" {
		http.Error(w, "room_name is required", http.StatusBadRequest)
		return
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = "ClawdFace"
	}

	log.Printf("[recall-trigger] Sending Recall.AI bot → meeting=%s room=%s convId=%s",
		req.MeetingURL, req.RoomName, req.ConversationID)

	ok, err := PostRecallRequest(req.MeetingURL, displayName, req.RoomName, req.VideoURL)
	if err != nil {
		log.Printf("[recall-trigger] ERROR: %v", err)
		http.Error(w, "Failed to send Recall bot: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Recall bot request failed", http.StatusInternalServerError)
		return
	}

	log.Printf("[recall-trigger] ✓ Recall bot dispatched → meeting=%s room=%s", req.MeetingURL, req.RoomName)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Recall bot dispatched into meeting",
		"room":    req.RoomName,
	})
}

// HandleCheckAgentEmailUniqueness checks if an email is already assigned to any agent.
// Query params: email (required).
func HandleCheckAgentEmailUniqueness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email query parameter is required", http.StatusBadRequest)
		return
	}

	var exists bool
	err := DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM agents WHERE email = $1)`,
		email,
	).Scan(&exists)

	if err != nil {
		log.Printf("Error checking agent email uniqueness: %v", err)
		http.Error(w, "Failed to check email uniqueness", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"unique": !exists})
}

func dispatchAgent(roomName, agentName, metaData string) error {

	apiKey := configs.GetEnv("LIVEKIT_API_KEY")
	apiSecret := configs.GetEnv("LIVEKIT_API_SECRET")
	liveURL := configs.GetEnv("LIVEKIT_URL")
	// Initialize the client
	client := lksdk.NewAgentDispatchServiceClient(liveURL, apiKey, apiSecret)

	// Create the dispatch request
	_, err := client.CreateDispatch(context.Background(), &livekit.CreateAgentDispatchRequest{
		Room:      roomName,
		AgentName: agentName,
		Metadata:  metaData,
	})
	if err != nil {
		// Handle error
		log.Printf("dispatchAgent error: %v", err)
	}
	return err
}

// HandleGetAgentByEmail retrieves a specific agent ID by email
func HandleGetAgentByEmail(email string) (string, error) {

	log.Printf("Fetching Agent with email ID: %s", email)

	query := `
        SELECT id
        FROM agents
        WHERE email = $1
		ORDER BY created_at desc
		LIMIT 1`

	var (
		id string
	)

	err := DB.QueryRow(query, email).Scan(
		&id,
	)

	if err == sql.ErrNoRows {
		log.Printf("No agent found with email: %s", email)
		return "", fmt.Errorf("agent not found with email: %s", email)
	} else if err != nil {
		log.Printf("Error retrieving Agent: %v", err)
		return "", fmt.Errorf("failed to retrieve Agent: %v", err)
	}

	return id, nil
}

// HandleUpdateConversationSnippet updates an existing conversation with snippet
func HandleUpdateConversationSnippet(w http.ResponseWriter, r *http.Request, convID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Conversation with ID: %s", convID)

	var convConfig struct {
		Snippets json.RawMessage `json:"snippets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&convConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE conversations
        SET snippets = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		convConfig.Snippets,
		convID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No Conversation found with ID: %s", convID)
		WriteNotFoundError(w, "Conversation not found")
		return
	} else if err != nil {
		log.Printf("Error updating Conversation: %v", err)
		WriteInternalServerError(w, "Failed to update Conversation")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Conversation updated successfully",
	})
}

// HandleUpdateConversationSnippet updates an existing conversation with snippet
func HandleUpdateConversationChat(w http.ResponseWriter, r *http.Request, convID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Conversation with ID: %s", convID)

	var convConfig struct {
		Chats json.RawMessage `json:"chats"`
	}

	if err := json.NewDecoder(r.Body).Decode(&convConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE conversations
        SET chat_history = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		convConfig.Chats,
		convID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No Conversation found with ID: %s", convID)
		WriteNotFoundError(w, "Conversation not found")
		return
	} else if err != nil {
		log.Printf("Error updating Conversation: %v", err)
		WriteInternalServerError(w, "Failed to update Conversation")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Conversation updated successfully",
	})
}

// HandleDeleteConversationChat deletes Conversation Chat History
func HandleDeleteConversationChat(w http.ResponseWriter, convID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Resetting Conversation Chat with ID: %s", convID)

	query := `
        UPDATE conversations
        SET chat_history = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		"{}",
		convID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No Conversation found with ID: %s", convID)
		WriteNotFoundError(w, "Conversation not found")
		return
	} else if err != nil {
		log.Printf("Error updating Conversation chat: %v", err)
		WriteInternalServerError(w, "Failed to update Conversation chat")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Conversation Chat reset successfully",
	})
}

func NewKnowledgeBase() KnowledgeBase {
	return KnowledgeBase{
		Enabled:       true,
		RetrievalTopK: 1,
	}
}

// HandleUpdateConversationSummary updates an existing conversation with summary
func HandleUpdateConversationSummary(w http.ResponseWriter, r *http.Request, convID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Updating Conversation with ID: %s", convID)

	var convConfig struct {
		Summary string `json:"summary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&convConfig); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	query := `
        UPDATE conversations
        SET summary = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		convConfig.Summary,
		convID,
	).Scan(&id)

	if err == sql.ErrNoRows {
		log.Printf("No Conversation found with ID: %s", convID)
		WriteNotFoundError(w, "Conversation not found")
		return
	} else if err != nil {
		log.Printf("Error updating Conversation: %v", err)
		WriteInternalServerError(w, "Failed to update Conversation")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Conversation updated successfully",
	})
}

func getRandomAvatar(avatarsJSON string) (Avatar, error) {
	var avatars []Avatar

	// Unmarshal JSON string into slice of Avatar
	err := json.Unmarshal([]byte(avatarsJSON), &avatars)
	if err != nil {
		return Avatar{}, fmt.Errorf("failed to unmarshal avatars array: %v", err)
	}

	// Check for empty input
	if len(avatars) == 0 {
		return Avatar{}, fmt.Errorf("no avatars available to choose from")
	}

	// Select a random avatar
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(avatars))

	// Return the selected avatar
	return avatars[randomIndex], nil
}

func getRandomAgent(avatarsJSON string, avatarId string) (Agent, int, error) {
	var agents []Agent

	// Unmarshal JSON string into slice of Avatar
	err := json.Unmarshal([]byte(avatarsJSON), &agents)
	if err != nil {
		return Agent{}, 0, fmt.Errorf("failed to unmarshal agent array: %v", err)
	}

	// Check for empty input
	if len(agents) == 0 {
		return Agent{}, 0, fmt.Errorf("no agents available to choose from")
	}

	if avatarId != "" {
		for _, agent := range agents {
			if agent.AvatarID == avatarId {
				return agent, len(agents), nil
			}
		}
	}

	// Select a random avatar
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(agents))

	// Return the selected avatar
	return agents[randomIndex], len(agents), nil
}

func HandleGetUsageMetricsByKeyHash(w http.ResponseWriter, apiKey string) {
	w.Header().Set("Content-Type", "application/json")
	query := `
		SELECT um.id, um.conversation_id, um.status, um.usage_json, um.measured_at, um.request_id
		FROM usage_metrics um
		JOIN conversations c ON um.conversation_id = c.id
		JOIN api_keys ak ON c.created_by = ak.id
		WHERE ak.key_hash = $1
		ORDER BY um.measured_at DESC
	`

	rows, err := DB.Query(query, apiKey)
	if err != nil {
		WriteInternalServerError(w, "Failed to retrieve usage metrics")
		return
	}
	defer rows.Close()

	var metrics []map[string]interface{}
	for rows.Next() {
		var (
			id, conversationID, status, usageJSON, requestID string
			measuredAt                                       time.Time
		)

		if err := rows.Scan(
			&id,
			&conversationID,
			&status,
			&usageJSON,
			&measuredAt,
			&requestID,
		); err != nil {
			log.Printf("Error: %v", err)
			WriteInternalServerError(w, "Error scanning usage metrics")
			return
		}

		metrics = append(metrics, map[string]interface{}{
			"id":              id,
			"conversation_id": conversationID,
			"status":          status,
			"usage_json":      json.RawMessage(usageJSON),
			"measured_at":     measuredAt,
			"request_id":      requestID,
		})
	}

	json.NewEncoder(w).Encode(metrics)
}

// HandleCreateFeedback creates a new feedback entry
func HandleCreateFeedback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var feedback struct {
		ConversationID string `json:"conversation_id"`
		Feedback       string `json:"feedback"`
		Rating         int    `json:"rating"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		log.Printf("Error decoding request body: %v", err)
		WriteBadRequestError(w, "Invalid request body")
		return
	}

	// Validation
	if feedback.ConversationID == "" {
		WriteBadRequestError(w, "conversation_id is required")
		return
	}
	if feedback.Rating < 0 || feedback.Rating > 5 { // assuming 1–5 scale
		WriteBadRequestError(w, "rating must be between 0 and 5")
		return
	}

	query := `
        INSERT INTO feedback (
            id, conversation_id, feedback, rating, created_at
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, CURRENT_TIMESTAMP
        ) RETURNING id`

	var id string
	err := DB.QueryRow(
		query,
		feedback.ConversationID,
		feedback.Feedback,
		feedback.Rating,
	).Scan(&id)

	if err != nil {
		log.Printf("Error creating feedback: %v", err)
		WriteInternalServerError(w, "Failed to create feedback")
		return
	}

	// Respond
	json.NewEncoder(w).Encode(map[string]string{
		"id":      id,
		"message": "Feedback submitted successfully",
	})
}

func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url := configs.GetEnv("API_URL")
	token := configs.GetEnv("API_TOKEN")
	statusHealth, errHealth := infra.GetHealth(url, token)
	if errHealth != nil {
		fmt.Printf("Health check failed: %v\n", errHealth)
		WriteInternalServerError(w, "Health check failed")
		return
	}

	// Respond with the full struct as JSON
	json.NewEncoder(w).Encode(statusHealth)
}

func GetJobStatus(w http.ResponseWriter, r *http.Request, jobID string) {
	w.Header().Set("Content-Type", "application/json")
	url := configs.GetEnv("API_URL")
	token := configs.GetEnv("API_TOKEN")
	status, err := infra.GetJobStatus(url, token, jobID)
	if err != nil {
		fmt.Printf("Job Status check failed: %v\n", err)
		WriteInternalServerError(w, "Job Status check failed")
		return
	}

	// Respond with the full struct as JSON
	json.NewEncoder(w).Encode(status)
}

func HandleUsageCreation(conversationId string, status string, jobId string, usage json.RawMessage, sessionId string, participantId string, message string) error {
	log.Printf("Creating usage for conversation ID: %s with job id: %s", conversationId, jobId)

	var agentid string
	// Fetch conversation
	err := DB.QueryRow(`
			SELECT agent_id
			FROM conversations
			WHERE id = $1 and status != 'ENDED'
		`, conversationId).Scan(&agentid)
	if err != nil {
		log.Printf("Error fetching conversation: %v", err)
		return fmt.Errorf("error fetching conversation: %v", err)
	}

	querySel := `
    SELECT conversation_id
    FROM usage_metrics
    WHERE conversation_id = $1
`
	var existingID string
	errSel := DB.QueryRow(querySel, conversationId).Scan(&existingID)

	switch errSel {
	case sql.ErrNoRows:
		// Insert
		_, err = DB.Exec(`
        INSERT INTO usage_metrics (
            request_id,
            conversation_id,
            status,
            usage_json
        ) VALUES ($1, $2, $3, $4)
    `, jobId, conversationId, status, usage)
		if err != nil {
			log.Printf("Error creating usage data: %v", err)
			return fmt.Errorf("error creating usage data: %v", err)
		}
	case nil:
		// Update
		_, err = DB.Exec(`
        UPDATE usage_metrics
        SET request_id = $1,
            status = $2,
            usage_json = $3
        WHERE conversation_id = $4
    `, jobId, status, usage, conversationId)
		if err != nil {
			log.Printf("Error updating usage data: %v", err)
			return fmt.Errorf("error updating usage data: %v", err)
		}
	default:
		// Unexpected error
		log.Printf("Error checking usage data: %v", errSel)
		return fmt.Errorf("error checking usage data: %v", errSel)
	}

	querylog := `
    UPDATE credit_limit_logs
    SET session_id     = COALESCE($1, ''),
        participant_id = COALESCE($2, ''),
        updated_at     = CURRENT_TIMESTAMP
    WHERE conversation_id = $3`
	resultlog, errlog := DB.Exec(querylog, sessionId, participantId, conversationId)

	if errlog != nil {
		log.Printf("Error updating credit_limit_logs: %v", resultlog)
		return fmt.Errorf("error updating credit_limit_logs: %v", errlog)
	}

	query := `
    UPDATE conversations
    SET status = $1,
		message = COALESCE($2, ''),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = $3`

	result, err := DB.Exec(query, status, message, conversationId)
	if err != nil {
		log.Printf("Error updating conversation: %v", err)
		return fmt.Errorf("error updating conversation: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("error checking updated rows: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("No Conversation found with ID: %s", conversationId)
		return fmt.Errorf("no conversation data found")
	}

	return nil
}

func HandleUsageUpdate(conversationId string, status string, message string) error {
	log.Printf("Updating usage for conversation ID: %s with status: %s and message: %s", conversationId, status, message)

	var agentid string
	// Fetch conversation
	err := DB.QueryRow(`
			SELECT agent_id
			FROM conversations
			WHERE id = $1 and status != 'ENDED'
		`, conversationId).Scan(&agentid)
	if err != nil {
		log.Printf("Error fetching conversation: %v", err)
		return fmt.Errorf("error fetching conversation: %v", err)
	}

	query := `
    UPDATE conversations
    SET status = $1,
		message = COALESCE($2, ''),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = $3`

	result, err := DB.Exec(query, status, message, conversationId)
	if err != nil {
		log.Printf("Error updating conversation: %v", err)
		return fmt.Errorf("error updating conversation: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("error checking updated rows: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("No Conversation found with ID: %s", conversationId)
		return fmt.Errorf("no conversation data found")
	}

	return nil
}

// Cron job runs every 15 min to update region
func UpdateRegionofConversations() {
	log.Println("Starting Region Update job...")

	updateRegion()

	log.Println("Region Update job completed.")
}

// -------------------------------------------------------------------
// Reset active subscriptions that have expired
// -------------------------------------------------------------------
func updateRegion() {
	query := `
		SELECT conversation_id, COALESCE(session_id, ''), COALESCE(participant_id, '')
		FROM credit_limit_logs
		WHERE region IS NULL
		AND session_id IS NOT NULL
		AND updated_at >= NOW() - INTERVAL '30 minutes';
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Println("❌ Failed to query :", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var sessionID string
		var participantID string

		if err := rows.Scan(&id, &sessionID, &participantID); err != nil {
			log.Println("❌ Failed to scan row:", err)
			continue
		}

		err := getAndUpdateRegion(id, sessionID, participantID)
		if err != nil {
			log.Printf("❌ Update failed for session=%s id=%s: %v\n", sessionID, id, err)
			continue
		}
	}
}

func getAndUpdateRegion(id string, sessionID string, participantID string) error {

	token, errtoken := GetLiveKitListToken(TokenSourceRequest{})
	if errtoken != nil {
		return fmt.Errorf("failed to get token: %w", errtoken)
	}
	ctx := context.Background()
	region, err1 := GetLiveKitSession(ctx, sessionID, token)

	if err1 != nil {
		return fmt.Errorf("failed to get region: %w", err1)
	}
	regionStr, locationStr, err2 := GetRegionByParticipantID(region, participantID)
	if err2 != nil {
		log.Printf("❌ failed to get region by participant session=%s id=%s: %v\n", sessionID, regionStr, err2)
		return fmt.Errorf("failed to get region by participant identity: %w", err2)
	}

	_, err := DB.Exec(`
		UPDATE credit_limit_logs
		SET
			region = $1,
			location = $2,
			updated_at = NOW()
		WHERE conversation_id = $3
	`, regionStr, locationStr, id)

	return err
}

func GetRegionByParticipantID(
	jsonData []byte,
	participantID string,
) (string, string, error) {

	var resp struct {
		Participants []struct {
			Region   string `json:"region"`
			Location string `json:"location"`
			Sessions []struct {
				ParticipantID string `json:"participantId"`
			} `json:"sessions"`
		} `json:"participants"`
	}

	if err := json.Unmarshal(jsonData, &resp); err != nil {
		return "", "", err
	}

	for _, p := range resp.Participants {
		for _, s := range p.Sessions {
			if s.ParticipantID == participantID {
				return p.Region, p.Location, nil
			}
		}
	}

	return "", "", fmt.Errorf("participantId %s not found", participantID)
}

func GetLiveKitSession(
	ctx context.Context,
	sessionID string,
	bearerToken string,
) ([]byte, error) {
	url := fmt.Sprintf(
		configs.GetEnv("LIVEKIT_SESSION_URL"),
		sessionID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"livekit api error: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	return io.ReadAll(resp.Body)
}

func ScheduleOneTimeResetCreditJob(delay time.Duration, conversationID string) {
	time.AfterFunc(delay, func() {
		err := ResetCredit(conversationID)
		if err != nil {
			log.Printf("Error %v", err)
			return
		}
		log.Printf("Reset completed for conversation %s", conversationID)
	})
}

func ResetCredit(conversationID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Fetch status + temp credit + org
	var status string
	var orgID string
	var tempCredit int

	err = tx.QueryRow(`
		SELECT cll.status, cll.organization_id, cll.temp_credit
		FROM credit_limit_logs cll
		WHERE cll.conversation_id = $1
		FOR UPDATE
	`, conversationID).Scan(&status, &orgID, &tempCredit)

	if err != nil {
		return fmt.Errorf("failed to fetch credit log: %w", err)
	}

	// Not initiated? Nothing to fix
	if status != "Initiated" {
		return tx.Commit() // No-op
	}

	// 2a. Update log status
	_, err = tx.Exec(`
		UPDATE credit_limit_logs
		SET status = 'Terminated', temp_credit = 0, updated_at = NOW()
		WHERE conversation_id = $1
	`, conversationID)

	if err != nil {
		return fmt.Errorf("failed to update credit log: %w", err)
	}

	// 2b. Update log status
	_, err = tx.Exec(`
		UPDATE conversations
		SET status = 'TERMINATED', updated_at = NOW()
		WHERE id = $1
	`, conversationID)

	if err != nil {
		return fmt.Errorf("failed to update credit log: %w", err)
	}

	// If no temp credit, nothing else to do
	if tempCredit <= 0 {
		return tx.Commit()
	}

	// 3. Add temp back to balance, set temp=0
	_, err = tx.Exec(`
		UPDATE credit_limits
		SET
			balance_credit = balance_credit + $1,
			updated_at = NOW()
		WHERE organization_id = $2
	`, tempCredit, orgID)

	if err != nil {
		return fmt.Errorf("failed to update credit limits: %w", err)
	}

	return tx.Commit()
}

// WriteInternalServerError sends a standardized JSON error response
func WriteInternalServerError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// WriteBadRequestError sends a standardized JSON error response
func WriteBadRequestError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// WriteNotFoundError sends a standardized JSON error response
func WriteNotFoundError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
func GetOrganizationStatistics(w http.ResponseWriter, r *http.Request, orgId string) {
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if startDateStr == "" || endDateStr == "" {
		WriteBadRequestError(w, "startDate and endDate are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		WriteBadRequestError(w, "Invalid startDate format, expected YYYY-MM-DD")
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		WriteBadRequestError(w, "Invalid endDate format, expected YYYY-MM-DD")
		return
	}

	days := int(endDate.Sub(startDate).Hours() / 24)
	var groupBy string
	if days <= 30 {
		groupBy = "day"
	} else if days <= 90 {
		groupBy = "week"
	} else {
		groupBy = "month"
	}

	log.Printf("orgId: %s, startDate: %s, endDate: %s, days: %d, groupBy: %s", orgId, startDateStr, endDateStr, days, groupBy)

	// Fetch all conversations for concurrency calculation
	type convData struct {
		ID       string
		Start    time.Time
		Duration float64
		Region   string
	}
	convQuery := `
		SELECT conversation_id, created_at, usage_duration, COALESCE(region, '')
		FROM credit_limit_logs
		WHERE organization_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at
	`
	endDatePlusOne := endDate.AddDate(0, 0, 1)
	rows, err := DB.Query(convQuery, orgId, startDate, endDatePlusOne)
	if err != nil {
		log.Printf("Error fetching conversations: %v", err)
		WriteInternalServerError(w, "Failed to fetch conversations")
		return
	}
	var conversations []convData
	for rows.Next() {
		var conv convData
		err := rows.Scan(&conv.ID, &conv.Start, &conv.Duration, &conv.Region)
		if err != nil {
			log.Printf("Error scanning conversation: %v", err)
			rows.Close()
			WriteInternalServerError(w, "Failed to scan conversation")
			return
		}
		conversations = append(conversations, conv)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating conversations: %v", err)
		WriteInternalServerError(w, "Failed to iterate conversations")
		return
	}

	var stats StatsResponse

	// Function to calculate peak concurrency and concurrent count for a slice of conversations
	calculateConcurrency := func(convs []convData) (int, int) {
		if len(convs) == 0 {
			return 0, 0
		}
		// Sort by start time
		sort.Slice(convs, func(i, j int) bool {
			return convs[i].Start.Before(convs[j].Start)
		})
		maxConc := 0
		concurrentCount := 0
		type activeItem struct {
			end time.Time
			idx int
		}
		active := make([]activeItem, 0)
		isConcurrent := make([]bool, len(convs))
		for i, conv := range convs {
			endTime := conv.Start.Add(time.Duration(conv.Duration) * time.Second)
			// Remove ended (end <= start is not overlapping)
			j := 0
			for _, a := range active {
				if a.end.After(conv.Start) {
					active[j] = a
					j++
				}
			}
			active = active[:j]
			if len(active) > 0 {
				isConcurrent[i] = true
				for _, a := range active {
					isConcurrent[a.idx] = true
				}
			}
			active = append(active, activeItem{end: endTime, idx: i})
			if len(active) > maxConc {
				maxConc = len(active)
			}
		}
		for _, v := range isConcurrent {
			if v {
				concurrentCount++
			}
		}
		return maxConc, concurrentCount
	}

	// Calculate overall peak concurrency and concurrent count
	stats.PeakConcurrency, stats.Concurrent = calculateConcurrency(conversations)

	// Query for totals (without peak_concurrency now)
	totalsQuery := `
		SELECT
			COUNT(DISTINCT conversation_id) AS total_sessions,
			COALESCE(SUM(usage_duration), 0.0) AS total_seconds,
			COUNT(DISTINCT region) AS total_countries
		FROM credit_limit_logs
		WHERE organization_id = $1 AND created_at::date BETWEEN $2 AND $3
	`
	err = DB.QueryRow(totalsQuery, orgId, startDateStr, endDateStr).Scan(&stats.TotalSessions, &stats.TotalSeconds, &stats.TotalCountries)
	if err != nil {
		log.Printf("Error fetching totals: %v", err)
		WriteInternalServerError(w, "Failed to fetch statistics")
		return
	}

	// Generate periods based on groupBy
	var periods []time.Time
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		periods = append(periods, current)
		if groupBy == "day" {
			current = current.AddDate(0, 0, 1)
		} else if groupBy == "week" {
			current = current.AddDate(0, 0, 7)
		} else { // month
			current = current.AddDate(0, 1, 0)
		}
	}

	// Build graphView
	var graphView []GraphData
	for _, period := range periods {
		var periodEnd time.Time
		if groupBy == "day" {
			periodEnd = period.AddDate(0, 0, 1)
		} else if groupBy == "week" {
			periodEnd = period.AddDate(0, 0, 7)
		} else {
			periodEnd = period.AddDate(0, 1, 0)
		}

		// Filter conversations in this period
		var periodConvs []convData
		sessionSet := make(map[string]bool)
		countrySet := make(map[string]bool)
		var totalSeconds float64
		for _, conv := range conversations {
			if conv.Start.After(periodEnd) || conv.Start.Equal(periodEnd) {
				continue
			}
			if conv.Start.Before(period) {
				continue
			}
			periodConvs = append(periodConvs, conv)
			sessionSet[conv.ID] = true
			if conv.Region != "" {
				countrySet[conv.Region] = true
			}
			totalSeconds += conv.Duration
		}

		_, periodConcurrent := calculateConcurrency(periodConvs)

		var gd GraphData
		gd.Sessions = len(sessionSet)
		gd.Seconds = totalSeconds
		gd.Concurrent = periodConcurrent
		if len(countrySet) > 0 {
			gd.Countries = make([]string, 0, len(countrySet))
			for country := range countrySet {
				gd.Countries = append(gd.Countries, country)
			}
			sort.Strings(gd.Countries)
		}

		if groupBy == "day" {
			gd.Label = period.Format("2006-01-02")
		} else if groupBy == "week" {
			endWeek := period.AddDate(0, 0, 6)
			gd.Label = period.Format("2006-01-02") + " - " + endWeek.Format("2006-01-02")
		} else {
			gd.Label = period.Format("2006-01")
		}

		graphView = append(graphView, gd)
	}

	stats.GraphView = graphView

	log.Printf("Returning stats: %+v", stats)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetGlobalMetrics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startDateStr := q.Get("startDate")
	endDateStr := q.Get("endDate")

	if startDateStr == "" || endDateStr == "" {
		WriteBadRequestError(w, "startDate and endDate are required")
		return
	}

	startDate, err := time.Parse(dateLayoutYYYYMMDD, startDateStr)
	if err != nil {
		WriteBadRequestError(w, "Invalid startDate format, expected YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse(dateLayoutYYYYMMDD, endDateStr)
	if err != nil {
		WriteBadRequestError(w, "Invalid endDate format, expected YYYY-MM-DD")
		return
	}

	if endDate.Before(startDate) {
		WriteBadRequestError(w, "endDate must be greater than or equal to startDate")
		return
	}

	userID := strings.TrimSpace(q.Get("userId"))
	orgID := strings.TrimSpace(q.Get("orgId"))
	agentID := strings.TrimSpace(q.Get("agentId"))
	status := strings.TrimSpace(q.Get("status"))
	region := strings.TrimSpace(q.Get("region"))
	convType := strings.TrimSpace(q.Get("type"))
	groupByParam := strings.ToLower(strings.TrimSpace(q.Get("groupBy")))

	var minSeconds *float64
	if val := strings.TrimSpace(q.Get("minSeconds")); val != "" {
		parsed, parseErr := strconv.ParseFloat(val, 64)
		if parseErr != nil {
			WriteBadRequestError(w, "minSeconds must be a valid number")
			return
		}
		minSeconds = &parsed
	}

	var maxSeconds *float64
	if val := strings.TrimSpace(q.Get("maxSeconds")); val != "" {
		parsed, parseErr := strconv.ParseFloat(val, 64)
		if parseErr != nil {
			WriteBadRequestError(w, "maxSeconds must be a valid number")
			return
		}
		maxSeconds = &parsed
	}

	if minSeconds != nil && maxSeconds != nil && *maxSeconds < *minSeconds {
		WriteBadRequestError(w, "maxSeconds must be greater than or equal to minSeconds")
		return
	}

	days := int(endDate.Sub(startDate).Hours() / 24)
	groupBy := "month"
	if days <= 30 {
		groupBy = "day"
	} else if days <= 90 {
		groupBy = "week"
	}
	if groupByParam == "day" || groupByParam == "week" || groupByParam == "month" {
		groupBy = groupByParam
	}

	type convData struct {
		ConversationID   string
		OrganizationID   string
		OrganizationName string
		Start            time.Time
		Duration         float64
		Region           string
		Location         string
		UserID           string
		UserName         string
		AgentID          string
		AgentName        string
		Status           string
		Type             string
		AvatarID         string
	}

	endDatePlusOne := endDate.AddDate(0, 0, 1)
	baseQuery := `
		SELECT
			cll.conversation_id,
			cll.organization_id,
			COALESCE(o.name, ''),
			cll.created_at,
			cll.usage_duration,
			COALESCE(cll.region, ''),
			COALESCE(cll.location, ''),
			COALESCE(c.user_id, ''),
			COALESCE(c.user_name, ''),
			COALESCE(c.agent_id::text, ''),
			COALESCE(a.agent_name, ''),
			COALESCE(c.status, ''),
			COALESCE(c.type, ''),
			COALESCE(c.avatar_key_id, '')
		FROM credit_limit_logs cll
		LEFT JOIN conversations c ON c.id = cll.conversation_id
		LEFT JOIN agents a ON a.id = c.agent_id
		LEFT JOIN organizations o ON o.id = cll.organization_id
		WHERE cll.created_at >= $1 AND cll.created_at < $2`

	args := []interface{}{startDate, endDatePlusOne}
	argPos := 3

	if orgID != "" {
		baseQuery += fmt.Sprintf(" AND cll.organization_id = $%d", argPos)
		args = append(args, orgID)
		argPos++
	}
	if userID != "" {
		baseQuery += fmt.Sprintf(" AND COALESCE(c.user_id, '') = $%d", argPos)
		args = append(args, userID)
		argPos++
	}
	if agentID != "" {
		baseQuery += fmt.Sprintf(" AND COALESCE(c.agent_id::text, '') = $%d", argPos)
		args = append(args, agentID)
		argPos++
	}
	if status != "" {
		baseQuery += fmt.Sprintf(" AND COALESCE(c.status, '') = $%d", argPos)
		args = append(args, status)
		argPos++
	}
	if region != "" {
		baseQuery += fmt.Sprintf(" AND COALESCE(cll.region, '') = $%d", argPos)
		args = append(args, region)
		argPos++
	}
	if convType != "" {
		baseQuery += fmt.Sprintf(" AND COALESCE(c.type, '') = $%d", argPos)
		args = append(args, convType)
		argPos++
	}
	if minSeconds != nil {
		baseQuery += fmt.Sprintf(" AND cll.usage_duration >= $%d", argPos)
		args = append(args, *minSeconds)
		argPos++
	}
	if maxSeconds != nil {
		baseQuery += fmt.Sprintf(" AND cll.usage_duration <= $%d", argPos)
		args = append(args, *maxSeconds)
		argPos++
	}

	baseQuery += " ORDER BY cll.created_at"

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		log.Printf("Error fetching global metrics conversations: %v", err)
		WriteInternalServerError(w, "Failed to fetch global metrics")
		return
	}
	defer rows.Close()

	var conversations []convData
	for rows.Next() {
		var conv convData
		if scanErr := rows.Scan(
			&conv.ConversationID,
			&conv.OrganizationID,
			&conv.OrganizationName,
			&conv.Start,
			&conv.Duration,
			&conv.Region,
			&conv.Location,
			&conv.UserID,
			&conv.UserName,
			&conv.AgentID,
			&conv.AgentName,
			&conv.Status,
			&conv.Type,
			&conv.AvatarID,
		); scanErr != nil {
			log.Printf("Error scanning global metrics conversation: %v", scanErr)
			WriteInternalServerError(w, "Failed to scan global metrics")
			return
		}
		conversations = append(conversations, conv)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating global metrics conversations: %v", err)
		WriteInternalServerError(w, "Failed to iterate global metrics")
		return
	}

	calculateConcurrency := func(convs []convData) (int, int) {
		if len(convs) == 0 {
			return 0, 0
		}

		sorted := make([]convData, len(convs))
		copy(sorted, convs)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Start.Before(sorted[j].Start)
		})

		maxConc := 0
		concurrentCount := 0
		type activeItem struct {
			end time.Time
			idx int
		}
		active := make([]activeItem, 0)
		isConcurrent := make([]bool, len(sorted))

		for i, conv := range sorted {
			endTime := conv.Start.Add(time.Duration(conv.Duration) * time.Second)
			j := 0
			for _, a := range active {
				if a.end.After(conv.Start) {
					active[j] = a
					j++
				}
			}
			active = active[:j]

			if len(active) > 0 {
				isConcurrent[i] = true
				for _, a := range active {
					isConcurrent[a.idx] = true
				}
			}

			active = append(active, activeItem{end: endTime, idx: i})
			if len(active) > maxConc {
				maxConc = len(active)
			}
		}

		for _, v := range isConcurrent {
			if v {
				concurrentCount++
			}
		}

		return maxConc, concurrentCount
	}

	var periods []time.Time
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		periods = append(periods, current)
		if groupBy == "day" {
			current = current.AddDate(0, 0, 1)
		} else if groupBy == "week" {
			current = current.AddDate(0, 0, 7)
		} else {
			current = current.AddDate(0, 1, 0)
		}
	}

	uniqueSessionSet := make(map[string]bool)
	countrySet := make(map[string]bool)
	orgSet := make(map[string]bool)
	orgNameByID := make(map[string]string)
	userSet := make(map[string]bool)
	var totalSeconds float64

	for _, conv := range conversations {
		uniqueSessionSet[conv.ConversationID] = true
		totalSeconds += conv.Duration
		if conv.Region != "" {
			countrySet[conv.Region] = true
		}
		if conv.OrganizationID != "" {
			orgSet[conv.OrganizationID] = true
			if _, exists := orgNameByID[conv.OrganizationID]; !exists {
				orgNameByID[conv.OrganizationID] = conv.OrganizationName
			} else if orgNameByID[conv.OrganizationID] == "" && conv.OrganizationName != "" {
				orgNameByID[conv.OrganizationID] = conv.OrganizationName
			}
		}
		if conv.UserID != "" {
			userSet[conv.UserID] = true
		}
	}

	countries := make([]string, 0, len(countrySet))
	for country := range countrySet {
		countries = append(countries, country)
	}
	sort.Strings(countries)

	orgIDs := make([]string, 0, len(orgSet))
	for id := range orgSet {
		orgIDs = append(orgIDs, id)
	}
	sort.Strings(orgIDs)

	orgNames := make([]string, 0, len(orgIDs))
	for _, id := range orgIDs {
		orgNames = append(orgNames, orgNameByID[id])
	}

	peakConcurrency, concurrent := calculateConcurrency(conversations)

	graphView := make([]GlobalGraphData, 0, len(periods))
	for _, period := range periods {
		var periodEnd time.Time
		if groupBy == "day" {
			periodEnd = period.AddDate(0, 0, 1)
		} else if groupBy == "week" {
			periodEnd = period.AddDate(0, 0, 7)
		} else {
			periodEnd = period.AddDate(0, 1, 0)
		}

		var periodConvs []convData
		sessionSet := make(map[string]bool)
		periodCountrySet := make(map[string]bool)
		periodOrgNameByID := make(map[string]string)
		detailsMap := make(map[string]*GlobalConversationDetail)
		var periodSeconds float64

		for _, conv := range conversations {
			if conv.Start.Before(period) || !conv.Start.Before(periodEnd) {
				continue
			}

			periodConvs = append(periodConvs, conv)
			sessionSet[conv.ConversationID] = true
			periodSeconds += conv.Duration

			if conv.Region != "" {
				periodCountrySet[conv.Region] = true
			}
			if conv.OrganizationID != "" {
				if _, exists := periodOrgNameByID[conv.OrganizationID]; !exists {
					periodOrgNameByID[conv.OrganizationID] = conv.OrganizationName
				} else if periodOrgNameByID[conv.OrganizationID] == "" && conv.OrganizationName != "" {
					periodOrgNameByID[conv.OrganizationID] = conv.OrganizationName
				}
			}

			if existing, ok := detailsMap[conv.ConversationID]; ok {
				existing.Seconds += conv.Duration
				if conv.Start.Before(existing.StartedAt) {
					existing.StartedAt = conv.Start
				}
				if existing.Region == "" && conv.Region != "" {
					existing.Region = conv.Region
				}
				if existing.Location == "" && conv.Location != "" {
					existing.Location = conv.Location
				}
				continue
			}

			detailsMap[conv.ConversationID] = &GlobalConversationDetail{
				ConversationID:   conv.ConversationID,
				OrganizationID:   conv.OrganizationID,
				OrganizationName: conv.OrganizationName,
				UserID:           conv.UserID,
				UserName:         conv.UserName,
				AgentID:          conv.AgentID,
				AgentName:        conv.AgentName,
				Status:           conv.Status,
				Type:             conv.Type,
				AvatarID:         conv.AvatarID,
				Region:           conv.Region,
				Location:         conv.Location,
				StartedAt:        conv.Start,
				Seconds:          conv.Duration,
			}
		}

		_, periodConcurrent := calculateConcurrency(periodConvs)

		gd := GlobalGraphData{
			Sessions:   len(sessionSet),
			Seconds:    periodSeconds,
			Concurrent: periodConcurrent,
		}

		if len(periodCountrySet) > 0 {
			gd.Countries = make([]string, 0, len(periodCountrySet))
			for country := range periodCountrySet {
				gd.Countries = append(gd.Countries, country)
			}
			sort.Strings(gd.Countries)
		}

		if len(periodOrgNameByID) > 0 {
			gd.Organizations = make([]string, 0, len(periodOrgNameByID))
			for id := range periodOrgNameByID {
				gd.Organizations = append(gd.Organizations, id)
			}
			sort.Strings(gd.Organizations)

			gd.OrganizationNames = make([]string, 0, len(gd.Organizations))
			for _, id := range gd.Organizations {
				gd.OrganizationNames = append(gd.OrganizationNames, periodOrgNameByID[id])
			}
		}

		if len(detailsMap) > 0 {
			gd.Conversations = make([]GlobalConversationDetail, 0, len(detailsMap))
			for _, detail := range detailsMap {
				gd.Conversations = append(gd.Conversations, *detail)
			}
			sort.Slice(gd.Conversations, func(i, j int) bool {
				if gd.Conversations[i].StartedAt.Equal(gd.Conversations[j].StartedAt) {
					return gd.Conversations[i].ConversationID < gd.Conversations[j].ConversationID
				}
				return gd.Conversations[i].StartedAt.Before(gd.Conversations[j].StartedAt)
			})
		}

		if groupBy == "day" {
			gd.Label = period.Format("2006-01-02")
		} else if groupBy == "week" {
			endWeek := period.AddDate(0, 0, 6)
			gd.Label = period.Format("2006-01-02") + " - " + endWeek.Format("2006-01-02")
		} else {
			gd.Label = period.Format("2006-01")
		}

		graphView = append(graphView, gd)
	}

	appliedFilters := map[string]string{}
	if userID != "" {
		appliedFilters["userId"] = userID
	}
	if orgID != "" {
		appliedFilters["orgId"] = orgID
	}
	if agentID != "" {
		appliedFilters["agentId"] = agentID
	}
	if status != "" {
		appliedFilters["status"] = status
	}
	if region != "" {
		appliedFilters["region"] = region
	}
	if convType != "" {
		appliedFilters["type"] = convType
	}
	if minSeconds != nil {
		appliedFilters["minSeconds"] = fmt.Sprintf("%v", *minSeconds)
	}
	if maxSeconds != nil {
		appliedFilters["maxSeconds"] = fmt.Sprintf("%v", *maxSeconds)
	}

	resp := GlobalMetricsResponse{
		StartDate:          startDateStr,
		EndDate:            endDateStr,
		GroupBy:            groupBy,
		AppliedFilters:     appliedFilters,
		Countries:          countries,
		OrgIDs:             orgIDs,
		OrgNames:           orgNames,
		TotalSessions:      len(uniqueSessionSet),
		TotalSeconds:       totalSeconds,
		TotalCountries:     len(countrySet),
		TotalOrganizations: len(orgSet),
		TotalUsers:         len(userSet),
		PeakConcurrency:    peakConcurrency,
		Concurrent:         concurrent,
		GraphView:          graphView,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandlePresentationGetUploadURL handles POST /v1/presentations/signed-url
func HandlePresentationGetUploadURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	contentTypes := map[string]string{
		"pdf":  "application/pdf",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"png":  "image/png",
		"jpeg": "image/jpeg",
	}
	const maxFileSize = 100 * 1024 * 1024 // 100 MB in bytes
	const expiry = 15 * time.Minute

	var req struct {
		Name  string `json:"name"`
		Files []struct {
			Filename      string `json:"filename"`
			ContentType   string `json:"content_type"`
			ContentLength int64  `json:"content_length"`
		} `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = "default"
	}
	if len(req.Files) == 0 {
		http.Error(w, `{"error":"files must not be empty"}`, http.StatusBadRequest)
		return
	}

	// Validate all files upfront before touching the DB
	for i, f := range req.Files {
		if f.Filename == "" {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: filename is required"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentType == "" {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: content_type is required"}`, i), http.StatusBadRequest)
			return
		}
		if _, ok := contentTypes[f.ContentType]; !ok {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: file type not allowed. Accepted: pdf, ppt, pptx, png, jpeg"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentLength <= 0 {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: content_length must be a positive number"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentLength > maxFileSize {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: file size exceeds 100MB limit"}`, i), http.StatusBadRequest)
			return
		}
	}

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	bucket := configs.GetEnv("AWS_BUCKET_PRESENTATIONS")
	region := configs.GetEnv("AWS_REGION")

	// Start transaction — presentation + all resources created together or not at all
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var presentationId string
	err = tx.QueryRow(`
		SELECT id FROM presentations 
		WHERE LOWER(name) = LOWER($1) AND created_by = $2 
		ORDER BY created_at DESC LIMIT 1`, req.Name, apiKeyId).Scan(&presentationId)

	if err != nil {
		if err == sql.ErrNoRows {
			presentationId = uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO presentations (id, name, status, created_by, created_at, updated_at)
				VALUES ($1, $2, 'pending', $3, NOW(), NOW())`,
				presentationId, req.Name, apiKeyId,
			)
			if err != nil {
				log.Printf("Error inserting presentation: %v", err)
				http.Error(w, `{"error":"failed to create presentation"}`, http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Error querying presentation: %v", err)
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	type resourceResult struct {
		ResourceId string            `json:"resource_id"`
		Filename   string            `json:"filename"`
		URL        string            `json:"url"`
		Headers    map[string]string `json:"headers"`
		Key        string            `json:"key"`
		ExpiresIn  int               `json:"expires_in"`
	}

	var resources []resourceResult

	for _, f := range req.Files {
		mimeType := contentTypes[f.ContentType]
		resourceId := uuid.New().String()
		key := fmt.Sprintf("clawdface/uploads/input/%s/%s/%s", presentationId, resourceId, f.Filename)
		s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)

		signedURL, err := GeneratePresignedUploadURL(bucket, region, key, mimeType, f.ContentLength, expiry)
		if err != nil {
			log.Printf("Error generating presigned URL for %s: %v", f.Filename, err)
			http.Error(w, `{"error":"failed to generate upload URL"}`, http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO presentation_resources (id, presentation_id, name, content_type, content_length, s3_key, s3_url, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW())`,
			resourceId, presentationId, f.Filename, mimeType, f.ContentLength, key, s3URL,
		)
		if err != nil {
			log.Printf("Error inserting presentation resource: %v", err)
			http.Error(w, `{"error":"failed to create resource"}`, http.StatusInternalServerError)
			return
		}

		resources = append(resources, resourceResult{
			ResourceId: resourceId,
			Filename:   f.Filename,
			URL:        signedURL,
			Headers: map[string]string{
				"Content-Type":   mimeType,
				"Content-Length": fmt.Sprintf("%d", f.ContentLength),
			},
			Key:       key,
			ExpiresIn: int(expiry.Seconds()),
		})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, `{"error":"failed to save presentation"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"presentation_id": presentationId,
		"resources":       resources,
	})
}

// HandleUpdatePresentationResourceStatus handles POST /v1/presentations/resources/{resourceId}/status
// Internal endpoint — called by Lambda (file landed) and ingestion pipeline (processing done).
func HandleUpdatePresentationResourceStatus(w http.ResponseWriter, r *http.Request, resourceId string) {
	w.Header().Set("Content-Type", "application/json")

	validStatuses := map[string]bool{
		"processing": true,
		"completed":  true,
		"failed":     true,
	}

	var req struct {
		Status      string `json:"status"`
		Message     string `json:"message"`
		IngestionId string `json:"ingestion_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		http.Error(w, `{"error":"status is required"}`, http.StatusBadRequest)
		return
	}
	if !validStatuses[req.Status] {
		http.Error(w, `{"error":"invalid status. Accepted: processing, completed, failed"}`, http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(`
		UPDATE presentation_resources
		SET status = $1, message = NULLIF($2, ''), ingestion_id = NULLIF($3, ''), updated_at = NOW()
		WHERE id = $4`,
		req.Status, req.Message, req.IngestionId, resourceId,
	)
	if err != nil {
		log.Printf("Error updating resource status: %v", err)
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"resource not found"}`, http.StatusNotFound)
		return
	}

	// Check if all resources for this presentation are terminal (completed or failed)
	var presentationId string
	var pending, processing int
	var anyFailed bool
	err = DB.QueryRow(`
		SELECT
			pr.presentation_id,
			COUNT(*) FILTER (WHERE pr.status = 'pending') AS pending,
			COUNT(*) FILTER (WHERE pr.status = 'processing') AS processing,
			BOOL_OR(pr.status = 'failed') AS any_failed
		FROM presentation_resources pr
		WHERE pr.presentation_id = (SELECT presentation_id FROM presentation_resources WHERE id = $1)
		GROUP BY pr.presentation_id`,
		resourceId,
	).Scan(&presentationId, &pending, &processing, &anyFailed)
	if err == nil && pending == 0 && processing == 0 {
		finalStatus := "completed"
		if anyFailed {
			finalStatus = "failed"
		}
		DB.Exec(`UPDATE presentations SET status = $1, updated_at = NOW() WHERE id = $2`, finalStatus, presentationId)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"resource_id": resourceId,
		"status":      req.Status,
	})
}

// HandleAddPresentationResources handles POST /v1/presentations/{presentationId}/resources
// Adds more files to an existing presentation — same signed-url flow, scoped to existing presentation.
func HandleAddPresentationResources(w http.ResponseWriter, r *http.Request, presentationId string) {
	w.Header().Set("Content-Type", "application/json")

	contentTypes := map[string]string{
		"pdf":  "application/pdf",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"png":  "image/png",
		"jpeg": "image/jpeg",
	}
	const maxFileSize = 100 * 1024 * 1024
	const expiry = 15 * time.Minute

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Verify presentation exists and belongs to the same org
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM presentations p
		JOIN api_keys ak2 ON ak2.id = p.created_by
		JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1 ON ak1.workspace_id = w1.id
		JOIN workspaces w2 ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1 AND p.id = $2`,
		apiKeyId, presentationId,
	).Scan(&count)
	if err != nil || count == 0 {
		http.Error(w, `{"error":"presentation not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		Files []struct {
			Filename      string `json:"filename"`
			ContentType   string `json:"content_type"`
			ContentLength int64  `json:"content_length"`
		} `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if len(req.Files) == 0 {
		http.Error(w, `{"error":"files must not be empty"}`, http.StatusBadRequest)
		return
	}

	for i, f := range req.Files {
		if f.Filename == "" {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: filename is required"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentType == "" {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: content_type is required"}`, i), http.StatusBadRequest)
			return
		}
		if _, ok := contentTypes[f.ContentType]; !ok {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: file type not allowed. Accepted: pdf, ppt, pptx, png, jpeg"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentLength <= 0 {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: content_length must be a positive number"}`, i), http.StatusBadRequest)
			return
		}
		if f.ContentLength > maxFileSize {
			http.Error(w, fmt.Sprintf(`{"error":"files[%d]: file size exceeds 100MB limit"}`, i), http.StatusBadRequest)
			return
		}
	}

	bucket := configs.GetEnv("AWS_BUCKET_PRESENTATIONS")
	region := configs.GetEnv("AWS_REGION")

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	type resourceResult struct {
		ResourceId string            `json:"resource_id"`
		Filename   string            `json:"filename"`
		URL        string            `json:"url"`
		Headers    map[string]string `json:"headers"`
		Key        string            `json:"key"`
		ExpiresIn  int               `json:"expires_in"`
	}

	var resources []resourceResult

	for _, f := range req.Files {
		mimeType := contentTypes[f.ContentType]
		resourceId := uuid.New().String()
		key := fmt.Sprintf("clawdface/uploads/input/%s/%s/%s", presentationId, resourceId, f.Filename)
		s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)

		signedURL, err := GeneratePresignedUploadURL(bucket, region, key, mimeType, f.ContentLength, expiry)
		if err != nil {
			log.Printf("Error generating presigned URL for %s: %v", f.Filename, err)
			http.Error(w, `{"error":"failed to generate upload URL"}`, http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO presentation_resources (id, presentation_id, name, content_type, content_length, s3_key, s3_url, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', NOW(), NOW())`,
			resourceId, presentationId, f.Filename, mimeType, f.ContentLength, key, s3URL,
		)
		if err != nil {
			log.Printf("Error inserting resource: %v", err)
			http.Error(w, `{"error":"failed to create resource"}`, http.StatusInternalServerError)
			return
		}

		resources = append(resources, resourceResult{
			ResourceId: resourceId,
			Filename:   f.Filename,
			URL:        signedURL,
			Headers: map[string]string{
				"Content-Type":   mimeType,
				"Content-Length": fmt.Sprintf("%d", f.ContentLength),
			},
			Key:       key,
			ExpiresIn: int(expiry.Seconds()),
		})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, `{"error":"failed to save resources"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"presentation_id": presentationId,
		"resources":       resources,
	})
}

// HandleUpdatePresentation handles PUT /v1/presentations/{presentationId}
func HandleUpdatePresentation(w http.ResponseWriter, r *http.Request, presentationId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(`
		UPDATE presentations
		SET name = $1, updated_at = NOW()
		WHERE id = $2 AND created_by = $3`,
		req.Name, presentationId, apiKeyId,
	)
	if err != nil {
		log.Printf("Error updating presentation: %v", err)
		http.Error(w, `{"error":"failed to update presentation"}`, http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"presentation not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   presentationId,
		"name": req.Name,
	})
}

// HandleDeletePresentation handles DELETE /v1/presentations/{presentationId}
func HandleDeletePresentation(w http.ResponseWriter, r *http.Request, presentationId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify ownership before deleting
	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM presentations WHERE id = $1 AND created_by = $2`, presentationId, apiKeyId).Scan(&count)
	if err != nil || count == 0 {
		http.Error(w, `{"error":"presentation not found"}`, http.StatusNotFound)
		return
	}

	// Delete resources, agent links, then the presentation
	if _, err = tx.Exec(`DELETE FROM presentation_resources WHERE presentation_id = $1`, presentationId); err != nil {
		log.Printf("Error deleting resources: %v", err)
		http.Error(w, `{"error":"failed to delete presentation"}`, http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec(`DELETE FROM agents_presentations WHERE presentation_id = $1`, presentationId); err != nil {
		log.Printf("Error deleting agent links: %v", err)
		http.Error(w, `{"error":"failed to delete presentation"}`, http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec(`DELETE FROM presentations WHERE id = $1`, presentationId); err != nil {
		log.Printf("Error deleting presentation: %v", err)
		http.Error(w, `{"error":"failed to delete presentation"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing delete: %v", err)
		http.Error(w, `{"error":"failed to delete presentation"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      presentationId,
		"deleted": true,
	})
}

// HandleGetAllTagNames handles GET /v1/presentations/tags
// Returns all distinct presentation tags scoped to the same organization as the requesting API key
func HandleGetAllTagNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	rows, err := DB.Query(`
		SELECT DISTINCT p.name
		FROM presentations p
		JOIN api_keys ak2 ON ak2.id = p.created_by
		JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1 ON ak1.workspace_id = w1.id
		JOIN workspaces w2 ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1 AND p.name IS NOT NULL AND p.name != ''
		ORDER BY p.name ASC`,
		apiKeyId,
	)
	if err != nil {
		log.Printf("Error fetching tag names: %v", err)
		http.Error(w, `{"error":"failed to fetch tags"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tags = append(tags, name)
	}

	if tags == nil {
		tags = []string{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags": tags,
	})
}

// HandleGetAllPresentations handles GET /v1/presentations
// Returns all presentations scoped to the same organization as the requesting API key,
// with each presentation's resources embedded.
func HandleGetAllPresentations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	rows, err := DB.Query(`
		SELECT
			p.id, p.name, p.status, p.created_at, p.updated_at,
			pr.id, pr.name, pr.content_type, pr.content_length, pr.s3_key, pr.s3_url,
			pr.status, COALESCE(pr.message, ''), COALESCE(pr.ingestion_id, ''),
			pr.created_at, pr.updated_at
		FROM presentations p
		JOIN api_keys ak2 ON ak2.id = p.created_by
		JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1 ON ak1.workspace_id = w1.id
		JOIN workspaces w2 ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		LEFT JOIN presentation_resources pr ON pr.presentation_id = p.id
		WHERE ak1.id = $1
		ORDER BY p.created_at DESC, pr.created_at ASC`,
		apiKeyId,
	)
	if err != nil {
		log.Printf("Error fetching all presentations: %v", err)
		http.Error(w, `{"error":"failed to fetch presentations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Use ordered slice to preserve presentation order, map for grouping
	var order []string
	presentationMap := map[string]map[string]interface{}{}

	for rows.Next() {
		var pID, pName, pStatus string
		var pCreatedAt, pUpdatedAt time.Time
		var rID, rName, rContentType, rS3Key, rS3URL, rStatus, rMessage, rIngestionId sql.NullString
		var rContentLength sql.NullInt64
		var rCreatedAt, rUpdatedAt sql.NullTime

		if err := rows.Scan(
			&pID, &pName, &pStatus, &pCreatedAt, &pUpdatedAt,
			&rID, &rName, &rContentType, &rContentLength, &rS3Key, &rS3URL,
			&rStatus, &rMessage, &rIngestionId,
			&rCreatedAt, &rUpdatedAt,
		); err != nil {
			continue
		}

		if _, exists := presentationMap[pID]; !exists {
			order = append(order, pID)
			presentationMap[pID] = map[string]interface{}{
				"id":         pID,
				"name":       pName,
				"status":     pStatus,
				"created_at": pCreatedAt,
				"updated_at": pUpdatedAt,
				"resources":  []map[string]interface{}{},
			}
		}

		if rID.Valid {
			resource := map[string]interface{}{
				"id":             rID.String,
				"name":           rName.String,
				"content_type":   rContentType.String,
				"content_length": rContentLength.Int64,
				"s3_key":         rS3Key.String,
				"s3_url":         rS3URL.String,
				"status":         rStatus.String,
				"message":        rMessage.String,
				"ingestion_id":   rIngestionId.String,
				"created_at":     rCreatedAt.Time,
				"updated_at":     rUpdatedAt.Time,
			}
			presentationMap[pID]["resources"] = append(
				presentationMap[pID]["resources"].([]map[string]interface{}),
				resource,
			)
		}
	}

	presentations := make([]map[string]interface{}, 0, len(order))
	for _, id := range order {
		presentations = append(presentations, presentationMap[id])
	}

	json.NewEncoder(w).Encode(presentations)
}

// HandleGetPresentationsByAgent handles GET /v1/presentations/{agentId}
func HandleGetPresentationsByAgent(w http.ResponseWriter, r *http.Request, agentId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	rows, err := DB.Query(`
		SELECT p.id, p.name, p.status, p.created_at, p.updated_at
		FROM presentations p
		JOIN agents_presentations ap ON ap.presentation_id = p.id
		WHERE ap.agent_id = $1 AND p.created_by = $2
		ORDER BY p.created_at DESC`,
		agentId, apiKeyId,
	)
	if err != nil {
		log.Printf("Error fetching presentations: %v", err)
		http.Error(w, `{"error":"failed to fetch presentations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var presentations []map[string]interface{}
	for rows.Next() {
		var id, name, status string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &status, &createdAt, &updatedAt); err != nil {
			continue
		}
		presentations = append(presentations, map[string]interface{}{
			"id":         id,
			"name":       name,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	if presentations == nil {
		presentations = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(presentations)
}

// HandleGetPresentationsByOrg handles GET /v1/presentations/org/{orgId}
func HandleGetPresentationsByOrg(w http.ResponseWriter, r *http.Request, orgId string) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := DB.Query(`
		SELECT p.id, p.name, p.status, p.created_at, p.updated_at
		FROM presentations p
		JOIN api_keys ak ON ak.id = p.created_by
		JOIN workspaces w ON w.id = ak.workspace_id
		WHERE w.organization_id = $1
		ORDER BY p.created_at DESC`,
		orgId,
	)
	if err != nil {
		log.Printf("Error fetching presentations by org: %v", err)
		http.Error(w, `{"error":"failed to fetch presentations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var presentations []map[string]interface{}
	for rows.Next() {
		var id, name, status string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &status, &createdAt, &updatedAt); err != nil {
			continue
		}
		presentations = append(presentations, map[string]interface{}{
			"id":         id,
			"name":       name,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	if presentations == nil {
		presentations = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(presentations)
}

// HandleGetPresentation handles GET /v1/presentations/{presentationId}
func HandleGetPresentation(w http.ResponseWriter, r *http.Request, presentationId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var id, name, status string
	var createdAt, updatedAt time.Time
	err := DB.QueryRow(`
		SELECT id, name, status, created_at, updated_at
		FROM presentations
		WHERE id = $1 AND created_by = $2`,
		presentationId, apiKeyId,
	).Scan(&id, &name, &status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"presentation not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching presentation: %v", err)
		http.Error(w, `{"error":"failed to fetch presentation"}`, http.StatusInternalServerError)
		return
	}

	rows, err := DB.Query(`
		SELECT id, name, content_type, content_length, s3_key, s3_url, status, COALESCE(message, ''), COALESCE(ingestion_id, ''), created_at, updated_at
		FROM presentation_resources
		WHERE presentation_id = $1
		ORDER BY created_at ASC`,
		presentationId,
	)
	if err != nil {
		log.Printf("Error fetching resources: %v", err)
		http.Error(w, `{"error":"failed to fetch resources"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}
	for rows.Next() {
		var rid, rname, contentType, s3Key, s3URL, rstatus, message, ingestionId string
		var contentLength int64
		var rCreatedAt, rUpdatedAt time.Time
		if err := rows.Scan(&rid, &rname, &contentType, &contentLength, &s3Key, &s3URL, &rstatus, &message, &ingestionId, &rCreatedAt, &rUpdatedAt); err != nil {
			continue
		}
		resources = append(resources, map[string]interface{}{
			"id":             rid,
			"name":           rname,
			"content_type":   contentType,
			"content_length": contentLength,
			"s3_key":         s3Key,
			"s3_url":         s3URL,
			"status":         rstatus,
			"message":        message,
			"ingestion_id":   ingestionId,
			"created_at":     rCreatedAt,
			"updated_at":     rUpdatedAt,
		})
	}
	if resources == nil {
		resources = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"name":       name,
		"status":     status,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"resources":  resources,
	})
}

// HandleGetPresentationResources handles GET /v1/presentations/{presentationId}/resources
func HandleGetPresentationResources(w http.ResponseWriter, r *http.Request, presentationId string) {
	w.Header().Set("Content-Type", "application/json")

	apiKeyId, _ := r.Context().Value("apiKeyId").(string)
	if apiKeyId == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Verify the presentation belongs to the same org as the requesting key
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM presentations p
		JOIN api_keys ak2 ON ak2.id = p.created_by
		JOIN api_keys ak1 ON ak2.workspace_id = ak1.workspace_id
		JOIN workspaces w1 ON ak1.workspace_id = w1.id
		JOIN workspaces w2 ON ak2.workspace_id = w2.id
			AND w2.organization_id = w1.organization_id
		WHERE ak1.id = $1 AND p.id = $2`,
		apiKeyId, presentationId,
	).Scan(&count)
	if err != nil || count == 0 {
		http.Error(w, `{"error":"presentation not found"}`, http.StatusNotFound)
		return
	}

	rows, err := DB.Query(`
		SELECT id, name, content_type, content_length, s3_key, s3_url, status, COALESCE(message, ''), COALESCE(ingestion_id, ''), created_at, updated_at
		FROM presentation_resources
		WHERE presentation_id = $1
		ORDER BY created_at ASC`,
		presentationId,
	)
	if err != nil {
		log.Printf("Error fetching presentation resources: %v", err)
		http.Error(w, `{"error":"failed to fetch resources"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}
	for rows.Next() {
		var id, name, contentType, s3Key, s3URL, status, message, ingestionId string
		var contentLength int64
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &contentType, &contentLength, &s3Key, &s3URL, &status, &message, &ingestionId, &createdAt, &updatedAt); err != nil {
			continue
		}
		resources = append(resources, map[string]interface{}{
			"id":             id,
			"name":           name,
			"content_type":   contentType,
			"content_length": contentLength,
			"s3_key":         s3Key,
			"s3_url":         s3URL,
			"status":         status,
			"message":        message,
			"ingestion_id":   ingestionId,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		})
	}
	if resources == nil {
		resources = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(resources)
}

// HandleGetPresentationResourceStatus handles GET /v1/presentations/resources/{resourceId}/status
func HandleGetPresentationResourceStatus(w http.ResponseWriter, r *http.Request, resourceId string) {
	w.Header().Set("Content-Type", "application/json")

	var status string
	var message sql.NullString
	var updatedAt time.Time
	err := DB.QueryRow(`
		SELECT status, message, updated_at
		FROM presentation_resources
		WHERE id = $1`,
		resourceId,
	).Scan(&status, &message, &updatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"resource not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching resource status: %v", err)
		http.Error(w, `{"error":"failed to fetch status"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"resource_id": resourceId,
		"status":      status,
		"message":     message.String,
		"updated_at":  updatedAt,
	})
}

func GetGlobalMetricsConversationByID(w http.ResponseWriter, r *http.Request, conversationID string) {
	w.Header().Set("Content-Type", "application/json")

	if strings.TrimSpace(conversationID) == "" {
		WriteBadRequestError(w, "conversationId is required")
		return
	}

	conversationConfigs, shouldReturn := GetConversationFunction(conversationID, w)
	if shouldReturn {
		return
	}

	json.NewEncoder(w).Encode(conversationConfigs)
}

// getConversationRoomID fetches the LiveKit room ID (stored in join_link) for a given conversation.
func getConversationRoomID(conversationID string) (string, error) {
	var roomID string
	err := DB.QueryRow(`SELECT join_link FROM conversations WHERE id = $1`, conversationID).Scan(&roomID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("conversation not found: %s", conversationID)
	}
	return roomID, err
}

// newRoomServiceClient returns a configured LiveKit RoomServiceClient.
func newRoomServiceClient() *lksdk.RoomServiceClient {
	return lksdk.NewRoomServiceClient(
		configs.GetEnv("LIVEKIT_URL"),
		configs.GetEnv("LIVEKIT_API_KEY"),
		configs.GetEnv("LIVEKIT_API_SECRET"),
	)
}

// getAgentParticipantIdentity lists participants in the given LiveKit room and returns
// the identity of the agent participant. It identifies the agent by checking for the
// "lk.agent_name" attribute (set automatically by the LiveKit agents framework) or by
// participant Kind == AGENT as a fallback.
func getAgentParticipantIdentity(roomID string) (string, error) {
	client := newRoomServiceClient()

	resp, err := client.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: roomID,
	})
	if err != nil {
		return "", fmt.Errorf("ListParticipants failed: %w", err)
	}

	log.Printf("getAgentParticipantIdentity: room=%s participants=%d", roomID, len(resp.Participants))

	// Prefer participant with lk.agent_name attribute (matches frontend logic)
	for _, p := range resp.Participants {
		if _, ok := p.Attributes["lk.agent_name"]; ok {
			log.Printf("getAgentParticipantIdentity: found agent by attribute — identity=%s", p.Identity)
			return p.Identity, nil
		}
	}

	// Fallback: participant Kind == AGENT
	for _, p := range resp.Participants {
		if p.Kind == livekit.ParticipantInfo_AGENT {
			log.Printf("getAgentParticipantIdentity: found agent by kind — identity=%s", p.Identity)
			return p.Identity, nil
		}
	}

	return "", fmt.Errorf("no agent participant found in room %s", roomID)
}

// EnsureSchema creates the scheduled_jobs table and runs any pending migrations.
func EnsureSchema() error {
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS scheduled_jobs (
		id TEXT PRIMARY KEY,
		name TEXT,
		meeting_url TEXT,
		agent_email_id TEXT,
		cron TEXT,
		expiry TEXT,
		status TEXT,
		created_at TEXT,
		uid TEXT,
		start_time TEXT,
		trace_id TEXT
	);`)
	if err != nil {
		return err
	}
	migrations := []string{
		`ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS trace_id TEXT`,
		`ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS start_time TEXT`,
		`ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS uid TEXT`,
	}
	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			return err
		}
	}
	return nil
}

// GetAllJobsFromDB returns all scheduled jobs, optionally filtered by agent email IDs.
func GetAllJobsFromDB(agentEmailIDs []string) ([]ScheduledJob, error) {
	sqlStmt := `
		SELECT id, name, cron, agent_email_id, meeting_url, expiry, status, COALESCE(uid,''), COALESCE(start_time,''), COALESCE(created_at,''), COALESCE(trace_id,'')
		FROM scheduled_jobs
	`
	var args []interface{}
	if len(agentEmailIDs) > 0 {
		placeholders := make([]string, len(agentEmailIDs))
		for i, id := range agentEmailIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, id)
		}
		sqlStmt += " WHERE agent_email_id IN (" + strings.Join(placeholders, ", ") + ")"
	}
	res, err := DB.Query(sqlStmt, args...)
	if err != nil {
		return []ScheduledJob{}, err
	}
	defer res.Close()
	var jobs []ScheduledJob
	for res.Next() {
		var job ScheduledJob
		if err := res.Scan(&job.ID, &job.Name, &job.Cron, &job.AgentEmailID, &job.MeetingURL, &job.Expiry, &job.Status, &job.UID, &job.StartTime, &job.CreatedAt, &job.TraceID); err != nil {
			return []ScheduledJob{}, err
		}
		jobs = append(jobs, job)
	}
	if err = res.Err(); err != nil {
		return []ScheduledJob{}, err
	}
	if jobs == nil {
		jobs = []ScheduledJob{}
	}
	return jobs, nil
}

// AddJobToDB inserts a new scheduled job record.
func AddJobToDB(ctx context.Context, job ScheduledJob) error {
	sqlStmt := `
	INSERT INTO scheduled_jobs (id, name, cron, agent_email_id, meeting_url, expiry, status, uid, start_time, created_at, trace_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	_, err := DB.Exec(sqlStmt, job.ID, job.Name, job.Cron, job.AgentEmailID, job.MeetingURL, job.Expiry, job.Status, job.UID, job.StartTime, time.Now().UTC().Format(time.RFC3339), job.TraceID)
	return err
}

// GetJobFromDBByID fetches a single job by ID.
func GetJobFromDBByID(jobID string) (ScheduledJob, error) {
	sqlStmt := `SELECT id, name, cron, agent_email_id, meeting_url, expiry FROM scheduled_jobs WHERE id = $1;`
	row := DB.QueryRow(sqlStmt, jobID)
	var job ScheduledJob
	err := row.Scan(&job.ID, &job.Name, &job.Cron, &job.AgentEmailID, &job.MeetingURL, &job.Expiry)
	if err != nil {
		if err == sql.ErrNoRows {
			return ScheduledJob{}, nil
		}
		return ScheduledJob{}, err
	}
	return job, nil
}

// DeleteJobFromDBByID removes a job record by ID.
func DeleteJobFromDBByID(jobID string) error {
	_, err := DB.Exec(`DELETE FROM scheduled_jobs WHERE id = $1;`, jobID)
	return err
}

// UpdateJobStatusInDB updates the status field of a job.
func UpdateJobStatusInDB(jobID string, status string) error {
	_, err := DB.Exec(`UPDATE scheduled_jobs SET status = $1 WHERE id = $2;`, status, jobID)
	return err
}

// UpdateJobStartTimeInDB updates the start_time field of a job (used to reflect next run time for recurring jobs).
func UpdateJobStartTimeInDB(jobID string, startTime string) error {
	_, err := DB.Exec(`UPDATE scheduled_jobs SET start_time = $1 WHERE id = $2;`, startTime, jobID)
	return err
}

// performAgentRPC joins the LiveKit room as a short-lived server participant and sends
// an RPC request to the agent via WebRTC (PerformRpc). The room join is synchronous so
// we know the connection is established before we dispatch. PerformRpc itself is run in
// a goroutine — the response path (agent → server) is unreliable from a local dev
// machine behind NAT, but the request path (server → agent) succeeds consistently.
// We return as soon as the request is dispatched rather than waiting for the ACK.
func performAgentRPC(roomID, agentIdentity, method, payload string) error {
	lkURL := configs.GetEnv("LIVEKIT_URL")
	apiKey := configs.GetEnv("LIVEKIT_API_KEY")
	apiSecret := configs.GetEnv("LIVEKIT_API_SECRET")

	serverIdentity := "server-rpc-" + uuid.New().String()[:8]
	at := auth.NewAccessToken(apiKey, apiSecret)
	at.SetIdentity(serverIdentity)
	at.SetName("Server")
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomID,
	}
	grant.SetCanPublishData(true)
	at.AddGrant(grant)

	token, err := at.ToJWT()
	if err != nil {
		return fmt.Errorf("token generation failed: %w", err)
	}

	room, err := lksdk.ConnectToRoomWithToken(lkURL, token, &lksdk.RoomCallback{}, lksdk.WithAutoSubscribe(false))
	if err != nil {
		return fmt.Errorf("room connect failed: %w", err)
	}

	// PerformRpc blocks until the agent sends back an ACK + response. On a local dev
	// machine the return data channel is often unreliable (NAT / DTLS issues), so the
	// call times out even though the agent receives and processes the request. Running
	// it in a goroutine lets us return 200 immediately; the background goroutine logs
	// any response error and disconnects the room when done.
	go func() {
		defer room.Disconnect()
		_, err := room.LocalParticipant.PerformRpc(lksdk.PerformRpcParams{
			DestinationIdentity: agentIdentity,
			Method:              method,
			Payload:             payload,
		})
		if err != nil {
			log.Printf("performAgentRPC: response not received for method=%s agent=%s (request was delivered): %v", method, agentIdentity, err)
		}
	}()

	log.Printf("performAgentRPC: dispatched method=%s → agent=%s room=%s", method, agentIdentity, roomID)
	return nil
}

// HandleAgentSpeak locates the agent participant in the LiveKit room for the given
// conversation and calls the agent's "speak" RPC method with the provided text.
// PUT /conversation/{conversationId}/speak
func HandleAgentSpeak(w http.ResponseWriter, r *http.Request, conversationID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("HandleAgentSpeak: conversationID=%s", conversationID)

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		WriteBadRequestError(w, "Invalid request body: 'text' field is required")
		return
	}

	roomID, err := getConversationRoomID(conversationID)
	if err != nil {
		log.Printf("HandleAgentSpeak: room lookup failed: %v", err)
		WriteNotFoundError(w, "Conversation not found")
		return
	}

	agentIdentity, err := getAgentParticipantIdentity(roomID)
	if err != nil {
		log.Printf("HandleAgentSpeak: agent not found room=%s err=%v", roomID, err)
		WriteInternalServerError(w, "Agent participant not found in room")
		return
	}

	rpcPayload, _ := json.Marshal(map[string]string{"text": body.Text})

	if err := performAgentRPC(roomID, agentIdentity, "speak", string(rpcPayload)); err != nil {
		log.Printf("HandleAgentSpeak: RPC failed room=%s agent=%s err=%v", roomID, agentIdentity, err)
		WriteInternalServerError(w, "Failed to send speak command to agent")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"conversation_id": conversationID,
		"agent_identity":  agentIdentity,
		"message":         "Speak command sent successfully",
	})
}

// HandleEndConversation gracefully ends a conversation by calling the agent's
// "end_session" RPC method, letting the agent shut down cleanly.
// DELETE /conversation/{conversationId}
func HandleEndConversation(w http.ResponseWriter, conversationID string) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("HandleEndConversation: conversationID=%s", conversationID)

	roomID, err := getConversationRoomID(conversationID)
	if err != nil {
		log.Printf("HandleEndConversation: room lookup failed: %v", err)
		WriteNotFoundError(w, "Conversation not found")
		return
	}

	agentIdentity, err := getAgentParticipantIdentity(roomID)
	if err != nil {
		log.Printf("HandleEndConversation: agent not found room=%s err=%v", roomID, err)
		WriteInternalServerError(w, "Agent participant not found in room")
		return
	}

	if err := performAgentRPC(roomID, agentIdentity, "end_session", ""); err != nil {
		log.Printf("HandleEndConversation: RPC failed room=%s agent=%s err=%v", roomID, agentIdentity, err)
		WriteInternalServerError(w, "Failed to end conversation")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"conversation_id": conversationID,
		"agent_identity":  agentIdentity,
		"message":         "Conversation ended successfully",
	})
}
