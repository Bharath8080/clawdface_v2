package providerselector

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- OpenAI-compatible (default LLM) ---

type llmOpenAICompat struct {
	name    string
	baseURL string
	apiKey  string
	model   string
}

func newLLMOpenAICompat(name, baseURL, apiKey, model string) *llmOpenAICompat {
	return &llmOpenAICompat{
		name:    name,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
	}
}

func (o *llmOpenAICompat) Name() string       { return o.name }
func (o *llmOpenAICompat) Type() ProviderType { return LLM }
func (o *llmOpenAICompat) Meta() map[string]string {
	m := map[string]string{"model": o.model}
	if o.baseURL != "https://api.openai.com" {
		m["base_url"] = o.baseURL
	}
	return m
}

func (o *llmOpenAICompat) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	body := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"hi"}],"max_tokens":1}`, o.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/v1/chat/completions", strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
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

// --- Anthropic ---

type llmAnthropic struct {
	name   string
	apiKey string
	model  string
}

func newLLMAnthropic(name, apiKey, model string) *llmAnthropic {
	return &llmAnthropic{name: name, apiKey: apiKey, model: model}
}

func (a *llmAnthropic) Name() string            { return a.name }
func (a *llmAnthropic) Type() ProviderType      { return LLM }
func (a *llmAnthropic) Meta() map[string]string { return map[string]string{"model": a.model} }

func (a *llmAnthropic) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	body := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"hi"}],"max_tokens":1}`, a.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
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

// --- Google ---

type llmGoogle struct {
	name   string
	apiKey string
	model  string
}

func newLLMGoogle(name, apiKey, model string) *llmGoogle {
	return &llmGoogle{name: name, apiKey: apiKey, model: model}
}

func (g *llmGoogle) Name() string            { return g.name }
func (g *llmGoogle) Type() ProviderType      { return LLM }
func (g *llmGoogle) Meta() map[string]string { return map[string]string{"model": g.model} }

func (g *llmGoogle) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model, g.apiKey,
	)
	body := `{"contents":[{"parts":[{"text":"hi"}]}],"generationConfig":{"maxOutputTokens":1}}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
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

// --- Ollama ---

type llmOllama struct {
	name    string
	baseURL string
}

func newLLMOllama(name, baseURL string) *llmOllama {
	return &llmOllama{name: name, baseURL: strings.TrimRight(baseURL, "/")}
}

func (o *llmOllama) Name() string            { return o.name }
func (o *llmOllama) Type() ProviderType      { return LLM }
func (o *llmOllama) Meta() map[string]string { return map[string]string{"base_url": o.baseURL} }

func (o *llmOllama) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+"/api/tags", nil)
	if err != nil {
		return CheckResult{}, err
	}
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

// --- AWS Bedrock ---

type llmBedrock struct {
	name      string
	accessKey string
	secretKey string
	region    string
	model     string
}

func newLLMBedrock(name, accessKey, secretKey, region, model string) *llmBedrock {
	if region == "" {
		region = "us-east-1"
	}
	return &llmBedrock{
		name:      name,
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
		model:     model,
	}
}

func (b *llmBedrock) Name() string       { return b.name }
func (b *llmBedrock) Type() ProviderType { return LLM }
func (b *llmBedrock) Meta() map[string]string {
	return map[string]string{"model": b.model, "region": b.region}
}

func (b *llmBedrock) Check(ctx context.Context) (CheckResult, error) {
	start := time.Now()

	host := fmt.Sprintf("bedrock-runtime.%s.amazonaws.com", b.region)
	urlPath := "/model/" + url.PathEscape(b.model) + "/converse"
	canonicalURI := "/model/" + bedrockURIEncode(b.model) + "/converse"
	endpoint := "https://" + host + urlPath
	reqBody := `{"messages":[{"role":"user","content":[{"text":"hi"}]}],"inferenceConfig":{"maxTokens":1}}`

	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	dateTimeStr := now.Format("20060102T150405Z")

	bodyHash := bedrockSHA256Hex([]byte(reqBody))
	canonicalHeaders := "content-type:application/json\nhost:" + host + "\nx-amz-date:" + dateTimeStr + "\n"
	signedHeaders := "content-type;host;x-amz-date"
	canonicalRequest := "POST\n" + canonicalURI + "\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + bodyHash

	credentialScope := dateStr + "/" + b.region + "/bedrock/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + dateTimeStr + "\n" + credentialScope + "\n" + bedrockSHA256Hex([]byte(canonicalRequest))

	signingKey := bedrockHMAC(
		bedrockHMAC(
			bedrockHMAC(
				bedrockHMAC([]byte("AWS4"+b.secretKey), []byte(dateStr)),
				[]byte(b.region),
			),
			[]byte("bedrock"),
		),
		[]byte("aws4_request"),
	)
	signature := hex.EncodeToString(bedrockHMAC(signingKey, []byte(stringToSign)))
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		b.accessKey, credentialScope, signedHeaders, signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(reqBody))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Date", dateTimeStr)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CheckResult{Latency: time.Since(start), Available: false}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return CheckResult{Latency: time.Since(start), Available: false},
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}
	return CheckResult{Latency: time.Since(start), Available: true}, nil
}

func bedrockURIEncode(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			sb.WriteByte(c)
		} else {
			fmt.Fprintf(&sb, "%%%02X", c)
		}
	}
	return sb.String()
}

func bedrockSHA256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func bedrockHMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
