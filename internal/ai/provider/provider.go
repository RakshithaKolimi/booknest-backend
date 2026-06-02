package provider

import (
	"context"
	"fmt"
	"strings"
)

const ProviderOpenAI = "openai"

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Embed(ctx context.Context, inputs []string) ([][]float64, error)
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

// NewProviderFromEnv creates a new AI provider instance based on environment variables.
func NewProviderFromEnv() (Provider, error) {
	// Load and validate config from environment variables
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	// Create provider instance based on config
	return NewProvider(cfg)
}

// NewProvider creates a new AI provider instance based on the given config.
func NewProvider(cfg Config) (Provider, error) {
	switch cfg.providerName() {
	case ProviderOpenAI:
		return NewOpenAIProvider(cfg.OpenAI)
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q", cfg.Provider)
	}
}
