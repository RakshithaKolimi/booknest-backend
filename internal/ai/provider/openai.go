package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

const (
	DefaultOpenAIBaseURL = "https://api.openai.com/v1"
	DefaultOpenAIModel   = "gpt-5.4-nano"
)

type OpenAIConfig struct {
	APIKey  string
	Model   string
}

func (cfg OpenAIConfig) withDefaults() OpenAIConfig {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)

	if cfg.Model == "" {
		cfg.Model = DefaultOpenAIModel
	}

	return cfg
}

func (cfg OpenAIConfig) Validate() error {
	cfg = cfg.withDefaults()

	if err := requireValue("OPENAI_API_KEY", cfg.APIKey); err != nil {
		return err
	}

	return nil
}

type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  openai.Client
}

func NewOpenAIProvider(cfg OpenAIConfig) (*OpenAIProvider, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	provider := OpenAIProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		client:  openai.NewClient(option.WithAPIKey(cfg.APIKey)),
	}

	return &provider, nil
}

func (p *OpenAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Basic validation
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt is required")
	}

	// Send request to OpenAI API
	res, err := p.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:           p.model,
		Input:           responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Temperature:     openai.Float(0.7),
		MaxOutputTokens: openai.Int(50),
	})
	if err != nil {
		return "", fmt.Errorf("send OpenAI request: %w", err)
	}

	// Extract and return the generated text
	return res.OutputText(), nil
}