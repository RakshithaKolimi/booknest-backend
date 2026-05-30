package ai

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultHTTPTimeout = 30 * time.Second

func LoadConfigFromEnv() (Config, error) {
	// Load and validate config from environment variables
	cfg := Config{
		Provider: strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER"))),
		OpenAI: OpenAIConfig{
			APIKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
			Model:   strings.TrimSpace(os.Getenv("OPENAI_CHAT_MODEL")),
			BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")), "/"),
			Timeout: DefaultHTTPTimeout,
		},
	}

	// Set default provider if not specified
	if cfg.Provider == "" {
		cfg.Provider = ProviderOpenAI
	}

	// Validate required fields based on provider
	if rawTimeout := strings.TrimSpace(os.Getenv("OPENAI_HTTP_TIMEOUT_SECONDS")); rawTimeout != "" {
		timeout, err := parseTimeoutSeconds("OPENAI_HTTP_TIMEOUT_SECONDS", rawTimeout)
		if err != nil {
			return Config{}, err
		}
		cfg.OpenAI.Timeout = timeout
	}

	// Validate the config
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

func parseTimeoutSeconds(envName string, raw string) (time.Duration, error) {
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", envName, err)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", envName)
	}
	return time.Duration(seconds) * time.Second, nil
}

func requireValue(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(name + " is required")
	}
	return nil
}
