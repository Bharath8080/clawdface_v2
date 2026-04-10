package providerselector

import (
	"context"
	"time"
)

// ProviderType identifies the category of an AI provider.
type ProviderType string

const (
	STT   ProviderType = "stt"
	LLM   ProviderType = "llm"
	TTS   ProviderType = "tts"
	Infra ProviderType = "infra"
)

// Provider is the interface every monitored provider must implement.
type Provider interface {
	Name() string
	Type() ProviderType
	Check(ctx context.Context) (CheckResult, error)
	Meta() map[string]string
}

// CheckResult holds the outcome of a single health check against a provider.
type CheckResult struct {
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency_ns"`
	TTFB      time.Duration `json:"ttfb_ns,omitempty"`
	TotalTime time.Duration `json:"total_time_ns,omitempty"`
	Available bool          `json:"available"`
	Error     string        `json:"error,omitempty"`
}
