package provider

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const DefaultHTTPTimeout = 30 * time.Second

// LoadConfigFromEnv loads the AI provider configuration from environment variables and validates it.
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Provider: strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER"))),
		OpenAI: OpenAIConfig{
			APIKey:         strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
			Model:          strings.TrimSpace(os.Getenv("OPENAI_CHAT_MODEL")),
			EmbeddingModel: strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_MODEL")),
		},
	}

	// Set default provider if not specified
	if cfg.Provider == "" {
		cfg.Provider = ProviderOpenAI
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	switch cfg.providerName() {
	case ProviderOpenAI:
		return cfg.OpenAI.Validate()
	default:
		return fmt.Errorf("unsupported AI_PROVIDER %q", cfg.Provider)
	}
}

func requireValue(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(name + " is required")
	}
	return nil
}
