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
	DefaultOpenAIBaseURL        = "https://api.openai.com/v1"
	DefaultOpenAIModel          = "gpt-5.4-nano"
	DefaultOpenAIEmbeddingModel = "text-embedding-3-small"
)

type OpenAIConfig struct {
	APIKey         string
	Model          string
	EmbeddingModel string
}

func (cfg OpenAIConfig) withDefaults() OpenAIConfig {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.EmbeddingModel = strings.TrimSpace(cfg.EmbeddingModel)

	if cfg.Model == "" {
		cfg.Model = DefaultOpenAIModel
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = DefaultOpenAIEmbeddingModel
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
	apiKey         string
	model          string
	embeddingModel string
	baseURL        string
	client         openai.Client
}

func NewOpenAIProvider(cfg OpenAIConfig) (*OpenAIProvider, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	provider := OpenAIProvider{
		apiKey:         cfg.APIKey,
		model:          cfg.Model,
		embeddingModel: cfg.EmbeddingModel,
		client:         openai.NewClient(option.WithAPIKey(cfg.APIKey)),
	}

	return &provider, nil
}

func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, temperature float64) (string, error) {
	// Basic validation
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt is required")
	}

	// Send request to OpenAI API
	res, err := p.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:           p.model,
		Input:           responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Temperature:     openai.Float(temperature),
		MaxOutputTokens: openai.Int(200),
	})
	if err != nil {
		return "", fmt.Errorf("send OpenAI request: %w", err)
	}

	// Extract and return the generated text
	return res.OutputText(), nil
}

func (p *OpenAIProvider) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	if len(inputs) == 0 {
		return nil, errors.New("inputs are required")
	}

	trimmed := make([]string, 0, len(inputs))
	for _, input := range inputs {
		value := strings.TrimSpace(input)
		if value == "" {
			return nil, errors.New("inputs must not contain empty strings")
		}
		trimmed = append(trimmed, value)
	}

	res, err := p.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: p.embeddingModel,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: trimmed,
		},
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	})
	if err != nil {
		return nil, fmt.Errorf("send OpenAI embedding request: %w", err)
	}

	vectors := make([][]float64, len(res.Data))
	for _, embedding := range res.Data {
		if embedding.Index < 0 || int(embedding.Index) >= len(vectors) {
			return nil, fmt.Errorf("OpenAI embedding returned out-of-range index %d", embedding.Index)
		}
		vectors[int(embedding.Index)] = embedding.Embedding
	}

	for i, vector := range vectors {
		if len(vector) == 0 {
			return nil, fmt.Errorf("OpenAI embedding response missing vector at index %d", i)
		}
	}

	return vectors, nil
}
