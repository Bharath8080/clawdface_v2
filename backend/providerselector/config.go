package providerselector

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CheckInterval time.Duration   `yaml:"check_interval"`
	Ranking       RankingConfig   `yaml:"ranking"`
	Providers     ProvidersConfig `yaml:"providers"`
}

type RankingConfig struct {
	LatencyWeight      float64 `yaml:"latency_weight"`
	AvailabilityWeight float64 `yaml:"availability_weight"`
	WindowSize         int     `yaml:"window_size"`
}

type ProvidersConfig struct {
	STT   []ProviderConfig `yaml:"stt"`
	LLM   []ProviderConfig `yaml:"llm"`
	TTS   []ProviderConfig `yaml:"tts"`
	Infra []ProviderConfig `yaml:"infra"`
}

type ProviderConfig struct {
	Name          string `yaml:"name"`
	ProviderType  string `yaml:"provider_type"`
	BaseURL       string `yaml:"base_url"`
	Model         string `yaml:"model"`
	VoiceID       string `yaml:"voice_id"`
	Enabled       bool   `yaml:"enabled"`
	APIKey        string `yaml:"api_key"`
	SecretKey     string `yaml:"secret_key"`
	Region        string `yaml:"region"`
	MaxConcurrent int    `yaml:"max_concurrent"` // infra only; 0 = unlimited
	ProjectID     string `yaml:"project_id"`     // infra/cerebrium: project identifier
	AppID         string `yaml:"app_id"`         // infra/cerebrium: app/deployment identifier
}

var envVarRe = regexp.MustCompile(`\$\{([^}]+)\}`)

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

// LoadConfig reads a YAML config file at path, substituting ${VAR} references
// with environment variable values. Returns defaults if file is missing.
func LoadConfig(path string) (*Config, error) {
	loadDotEnv(".env")
	loadDotEnv(".env.local")

	cfg := &Config{
		CheckInterval: 60 * time.Second,
		Ranking: RankingConfig{
			LatencyWeight:      0.3,
			AvailabilityWeight: 0.7,
			WindowSize:         100,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	expanded := envVarRe.ReplaceAllFunc(data, func(match []byte) []byte {
		key := string(match[2 : len(match)-1])
		return []byte(os.Getenv(key))
	})

	if err := yaml.Unmarshal(expanded, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
