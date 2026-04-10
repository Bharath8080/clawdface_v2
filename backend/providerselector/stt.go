package providerselector

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// --- Deepgram ---

type sttDeepgram struct {
	name   string
	apiKey string
	model  string
}

func newSTTDeepgram(name, apiKey, model string) *sttDeepgram {
	if model == "" {
		model = "nova-2"
	}
	return &sttDeepgram{name: name, apiKey: apiKey, model: model}
}

func (d *sttDeepgram) Name() string            { return d.name }
func (d *sttDeepgram) Type() ProviderType      { return STT }
func (d *sttDeepgram) Meta() map[string]string { return map[string]string{"model": d.model} }

func (d *sttDeepgram) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.deepgram.com/v1/projects", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Token "+d.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

// --- OpenAI STT ---

type sttOpenAI struct {
	name   string
	apiKey string
	model  string
}

func newSTTOpenAI(name, apiKey, model string) *sttOpenAI {
	if model == "" {
		model = "whisper-1"
	}
	return &sttOpenAI{name: name, apiKey: apiKey, model: model}
}

func (o *sttOpenAI) Name() string            { return o.name }
func (o *sttOpenAI) Type() ProviderType      { return STT }
func (o *sttOpenAI) Meta() map[string]string { return map[string]string{"model": o.model} }

func (o *sttOpenAI) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

// --- AssemblyAI ---

type sttAssemblyAI struct {
	name   string
	apiKey string
	model  string
}

func newSTTAssemblyAI(name, apiKey, model string) *sttAssemblyAI {
	if model == "" {
		model = "best"
	}
	return &sttAssemblyAI{name: name, apiKey: apiKey, model: model}
}

func (a *sttAssemblyAI) Name() string            { return a.name }
func (a *sttAssemblyAI) Type() ProviderType      { return STT }
func (a *sttAssemblyAI) Meta() map[string]string { return map[string]string{"model": a.model} }

func (a *sttAssemblyAI) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.assemblyai.com/v2/account", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", a.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

// --- Groq ---

type sttGroq struct {
	name   string
	apiKey string
	model  string
}

func newSTTGroq(name, apiKey, model string) *sttGroq {
	if model == "" {
		model = "whisper-large-v3-turbo"
	}
	return &sttGroq{name: name, apiKey: apiKey, model: model}
}

func (g *sttGroq) Name() string            { return g.name }
func (g *sttGroq) Type() ProviderType      { return STT }
func (g *sttGroq) Meta() map[string]string { return map[string]string{"model": g.model} }

func (g *sttGroq) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.groq.com/openai/v1/models", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

// --- ElevenLabs STT ---

type sttElevenLabs struct {
	name   string
	apiKey string
	model  string
}

func newSTTElevenLabs(name, apiKey, model string) *sttElevenLabs {
	if model == "" {
		model = "scribe_v1"
	}
	return &sttElevenLabs{name: name, apiKey: apiKey, model: model}
}

func (e *sttElevenLabs) Name() string            { return e.name }
func (e *sttElevenLabs) Type() ProviderType      { return STT }
func (e *sttElevenLabs) Meta() map[string]string { return map[string]string{"model": e.model} }

func (e *sttElevenLabs) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.elevenlabs.io/v1/speech-to-text", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("xi-api-key", e.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}
