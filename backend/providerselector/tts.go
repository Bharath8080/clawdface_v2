package providerselector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- ElevenLabs TTS ---

type ttsElevenLabs struct {
	name    string
	apiKey  string
	model   string
	voiceID string
}

func newTTSElevenLabs(name, apiKey, model, voiceID string) *ttsElevenLabs {
	if model == "" {
		model = "eleven_turbo_v2_5"
	}
	return &ttsElevenLabs{name: name, apiKey: apiKey, model: model, voiceID: voiceID}
}

func (e *ttsElevenLabs) Name() string       { return e.name }
func (e *ttsElevenLabs) Type() ProviderType { return TTS }
func (e *ttsElevenLabs) Meta() map[string]string {
	m := map[string]string{"model": e.model}
	if e.voiceID != "" {
		m["voice_id"] = e.voiceID
	}
	return m
}

func (e *ttsElevenLabs) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	vid := e.voiceID
	if vid == "" {
		vid = "21m00Tcm4TlvDq8ikWAM"
	}
	body := fmt.Sprintf(`{"text":"hi","model_id":%q}`, e.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.elevenlabs.io/v1/text-to-speech/"+vid+"/stream", strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("xi-api-key", e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var buf [1]byte
	if _, err := resp.Body.Read(buf[:]); err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	ttfb := time.Since(start)

	io.Copy(io.Discard, resp.Body)
	total := time.Since(start)

	return CheckResult{Latency: total, TTFB: ttfb, TotalTime: total, Available: true}, nil
}

// --- Cartesia ---

type ttsCartesia struct {
	name    string
	apiKey  string
	model   string
	voiceID string
}

func newTTSCartesia(name, apiKey, model, voiceID string) *ttsCartesia {
	if model == "" {
		model = "sonic-2"
	}
	return &ttsCartesia{name: name, apiKey: apiKey, model: model, voiceID: voiceID}
}

func (c *ttsCartesia) Name() string       { return c.name }
func (c *ttsCartesia) Type() ProviderType { return TTS }
func (c *ttsCartesia) Meta() map[string]string {
	m := map[string]string{"model": c.model}
	if c.voiceID != "" {
		m["voice_id"] = c.voiceID
	}
	return m
}

func (c *ttsCartesia) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()

	if c.voiceID != "" {
		body := fmt.Sprintf(
			`{"model_id":%q,"transcript":"hi","voice":{"mode":"id","id":%q},"output_format":{"container":"raw","encoding":"pcm_f32le","sample_rate":8000}}`,
			c.model, c.voiceID,
		)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			"https://api.cartesia.ai/tts/bytes", strings.NewReader(body))
		if err != nil {
			return CheckResult{}, err
		}
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Cartesia-Version", "2024-06-10")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return CheckResult{Latency: time.Since(start), Available: false}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			io.Copy(io.Discard, resp.Body)
			return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		var buf [1]byte
		if _, err := resp.Body.Read(buf[:]); err != nil {
			return CheckResult{Latency: time.Since(start), Available: false}, err
		}
		ttfb := time.Since(start)

		io.Copy(io.Discard, resp.Body)
		total := time.Since(start)

		return CheckResult{Latency: total, TTFB: ttfb, TotalTime: total, Available: true}, nil
	}

	// Fallback: auth check via voice listing.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.cartesia.ai/voices", nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Cartesia-Version", "2024-06-10")

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

// --- OpenAI TTS ---

type ttsOpenAI struct {
	name    string
	apiKey  string
	model   string
	voiceID string
}

func newTTSOpenAI(name, apiKey, model, voiceID string) *ttsOpenAI {
	if model == "" {
		model = "tts-1"
	}
	if voiceID == "" {
		voiceID = "alloy"
	}
	return &ttsOpenAI{name: name, apiKey: apiKey, model: model, voiceID: voiceID}
}

func (o *ttsOpenAI) Name() string       { return o.name }
func (o *ttsOpenAI) Type() ProviderType { return TTS }
func (o *ttsOpenAI) Meta() map[string]string {
	return map[string]string{"model": o.model, "voice_id": o.voiceID}
}

func (o *ttsOpenAI) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	body := fmt.Sprintf(`{"model":%q,"input":"hi","voice":%q}`, o.model, o.voiceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/audio/speech", strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var buf [1]byte
	if _, err := resp.Body.Read(buf[:]); err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	ttfb := time.Since(start)

	io.Copy(io.Discard, resp.Body)
	total := time.Since(start)

	return CheckResult{Latency: total, TTFB: ttfb, TotalTime: total, Available: true}, nil
}

// --- Deepgram TTS ---

type ttsDeepgram struct {
	name    string
	apiKey  string
	model   string
	voiceID string
}

func newTTSDeepgram(name, apiKey, model, voiceID string) *ttsDeepgram {
	if model == "" {
		model = "aura-2-thalia-en"
	}
	return &ttsDeepgram{name: name, apiKey: apiKey, model: model, voiceID: voiceID}
}

func (d *ttsDeepgram) Name() string       { return d.name }
func (d *ttsDeepgram) Type() ProviderType { return TTS }
func (d *ttsDeepgram) Meta() map[string]string {
	m := map[string]string{"model": d.model}
	if d.voiceID != "" {
		m["voice_id"] = d.voiceID
	}
	return m
}

func (d *ttsDeepgram) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	body := `{"text":"hi"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepgram.com/v1/speak?model="+d.model, strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Token "+d.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var buf [1]byte
	if _, err := resp.Body.Read(buf[:]); err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	ttfb := time.Since(start)

	io.Copy(io.Discard, resp.Body)
	total := time.Since(start)

	return CheckResult{Latency: total, TTFB: ttfb, TotalTime: total, Available: true}, nil
}

// --- LMNT ---

type ttsLMNT struct {
	name    string
	apiKey  string
	model   string
	voiceID string
}

func newTTSLMNT(name, apiKey, model, voiceID string) *ttsLMNT {
	if model == "" {
		model = "aurora"
	}
	return &ttsLMNT{name: name, apiKey: apiKey, model: model, voiceID: voiceID}
}

func (l *ttsLMNT) Name() string       { return l.name }
func (l *ttsLMNT) Type() ProviderType { return TTS }
func (l *ttsLMNT) Meta() map[string]string {
	m := map[string]string{"model": l.model}
	if l.voiceID != "" {
		m["voice_id"] = l.voiceID
	}
	return m
}

func (l *ttsLMNT) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	voice := l.voiceID
	if voice == "" {
		voice = "lily"
	}
	body := fmt.Sprintf(`{"text":"hi","voice":%q,"format":"mp3"}`, voice)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.lmnt.com/v1/ai/speech/stream", strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("X-API-Key", l.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return CheckResult{Latency: time.Since(start), Available: false}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var buf [1]byte
	if _, err := resp.Body.Read(buf[:]); err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	ttfb := time.Since(start)

	io.Copy(io.Discard, resp.Body)
	total := time.Since(start)

	return CheckResult{Latency: total, TTFB: ttfb, TotalTime: total, Available: true}, nil
}
