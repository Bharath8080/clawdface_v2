package providerselector

import (
	"fmt"
	"strings"
)

// BuildProviders constructs Provider instances from the config, skipping disabled entries.
func BuildProviders(cfg *Config) ([]Provider, error) {
	var providers []Provider

	for _, pc := range cfg.Providers.STT {
		if !pc.Enabled {
			continue
		}
		p, err := buildSTT(pc)
		if err != nil {
			return nil, fmt.Errorf("stt/%s: %w", pc.Name, err)
		}
		providers = append(providers, p)
	}

	for _, pc := range cfg.Providers.LLM {
		if !pc.Enabled {
			continue
		}
		p, err := buildLLM(pc)
		if err != nil {
			return nil, fmt.Errorf("llm/%s: %w", pc.Name, err)
		}
		providers = append(providers, p)
	}

	for _, pc := range cfg.Providers.TTS {
		if !pc.Enabled {
			continue
		}
		p, err := buildTTS(pc)
		if err != nil {
			return nil, fmt.Errorf("tts/%s: %w", pc.Name, err)
		}
		providers = append(providers, p)
	}

	for _, pc := range cfg.Providers.Infra {
		if !pc.Enabled {
			continue
		}
		p, err := buildInfra(pc)
		if err != nil {
			return nil, fmt.Errorf("infra/%s: %w", pc.Name, err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}

func providerTypeName(pc ProviderConfig) string {
	if pc.ProviderType != "" {
		return strings.ToLower(pc.ProviderType)
	}
	return strings.ToLower(pc.Name)
}

func buildSTT(pc ProviderConfig) (Provider, error) {
	switch providerTypeName(pc) {
	case "deepgram":
		return newSTTDeepgram(pc.Name, pc.APIKey, pc.Model), nil
	case "openai":
		return newSTTOpenAI(pc.Name, pc.APIKey, pc.Model), nil
	case "assemblyai":
		return newSTTAssemblyAI(pc.Name, pc.APIKey, pc.Model), nil
	case "groq":
		return newSTTGroq(pc.Name, pc.APIKey, pc.Model), nil
	case "elevenlabs":
		return newSTTElevenLabs(pc.Name, pc.APIKey, pc.Model), nil
	default:
		return nil, fmt.Errorf("unknown provider_type: %s", providerTypeName(pc))
	}
}

func buildLLM(pc ProviderConfig) (Provider, error) {
	switch strings.ToLower(pc.ProviderType) {
	case "openai_compat", "openai-compat", "":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		return newLLMOpenAICompat(pc.Name, baseURL, pc.APIKey, pc.Model), nil
	case "anthropic":
		return newLLMAnthropic(pc.Name, pc.APIKey, pc.Model), nil
	case "google":
		return newLLMGoogle(pc.Name, pc.APIKey, pc.Model), nil
	case "ollama":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return newLLMOllama(pc.Name, baseURL), nil
	case "bedrock":
		return newLLMBedrock(pc.Name, pc.APIKey, pc.SecretKey, pc.Region, pc.Model), nil
	default:
		return nil, fmt.Errorf("unknown llm provider_type: %s", pc.ProviderType)
	}
}

func buildTTS(pc ProviderConfig) (Provider, error) {
	switch providerTypeName(pc) {
	case "elevenlabs":
		return newTTSElevenLabs(pc.Name, pc.APIKey, pc.Model, pc.VoiceID), nil
	case "cartesia":
		return newTTSCartesia(pc.Name, pc.APIKey, pc.Model, pc.VoiceID), nil
	case "openai":
		return newTTSOpenAI(pc.Name, pc.APIKey, pc.Model, pc.VoiceID), nil
	case "deepgram":
		return newTTSDeepgram(pc.Name, pc.APIKey, pc.Model, pc.VoiceID), nil
	case "lmnt":
		return newTTSLMNT(pc.Name, pc.APIKey, pc.Model, pc.VoiceID), nil
	default:
		return nil, fmt.Errorf("unknown provider_type: %s", providerTypeName(pc))
	}
}

// BuildInfraConcurrencyManager creates a ConcurrencyManager seeded with all
// enabled infra providers and their max_concurrent limits from the config.
func BuildInfraConcurrencyManager(cfg *Config) *ConcurrencyManager {
	limits := make(map[string]int)
	for _, pc := range cfg.Providers.Infra {
		if pc.Enabled {
			limits[pc.Name] = pc.MaxConcurrent
		}
	}
	return NewConcurrencyManager(limits)
}

func buildInfra(pc ProviderConfig) (Provider, error) {
	switch providerTypeName(pc) {
	case "cerebrium":
		return newInfraCerebrium(pc.Name, pc.BaseURL, pc.ProjectID, pc.AppID, pc.APIKey), nil
	case "runpod":
		return newInfraRunpod(pc.Name, pc.BaseURL, pc.APIKey), nil
	default:
		return nil, fmt.Errorf("unknown provider_type: %s", providerTypeName(pc))
	}
}

// BuildInfraActiveRunCounters returns a map of provider name → ActiveRunCounter
// for all enabled infra providers that support live run-count queries (Cerebrium
// and Runpod). Providers with no project_id/app_id configured are included but
// will return an error when queried.
func BuildInfraActiveRunCounters(cfg *Config) map[string]ActiveRunCounter {
	counters := make(map[string]ActiveRunCounter)
	for _, pc := range cfg.Providers.Infra {
		if !pc.Enabled {
			continue
		}
		switch strings.ToLower(providerTypeName(pc)) {
		case "cerebrium":
			counters[pc.Name] = newInfraCerebrium(pc.Name, pc.BaseURL, pc.ProjectID, pc.AppID, pc.APIKey)
		case "runpod":
			counters[pc.Name] = newInfraRunpod(pc.Name, pc.BaseURL, pc.APIKey)
		}
	}
	return counters
}
