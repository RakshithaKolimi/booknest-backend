package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	ProviderOpenAI = "openai"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type HealthStatus struct {
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Model    string `json:"model,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Config struct {
	Provider string
	OpenAI   OpenAIConfig
}

func (cfg Config) providerName() string {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		return ProviderOpenAI
	}
	return provider
}

func NewProviderFromEnv() (Provider, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewProvider(cfg)
}

func NewProvider(cfg Config) (Provider, error) {
	switch cfg.providerName() {
	case ProviderOpenAI:
		return NewOpenAIProvider(cfg.OpenAI)
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q", cfg.Provider)
	}
}

func HealthFromEnv() HealthStatus {
	provider := Config{Provider: os.Getenv("AI_PROVIDER")}.providerName()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return HealthStatus{
			Status:   "unavailable",
			Provider: provider,
			Error:    err.Error(),
		}
	}

	return cfg.Health()
}

func (cfg Config) Health() HealthStatus {
	if err := cfg.Validate(); err != nil {
		return HealthStatus{
			Status:   "unavailable",
			Provider: cfg.providerName(),
			Error:    err.Error(),
		}
	}

	status := HealthStatus{
		Status:   "ok",
		Provider: cfg.providerName(),
	}

	if status.Provider == ProviderOpenAI {
		status.Model = cfg.OpenAI.withDefaults().Model
	}

	return status
}
