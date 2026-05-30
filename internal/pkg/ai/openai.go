package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultOpenAIBaseURL = "https://api.openai.com/v1"
	DefaultOpenAIModel   = "gpt-5.4-nano"
)

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

func (cfg OpenAIConfig) withDefaults() OpenAIConfig {
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")

	if cfg.Model == "" {
		cfg.Model = DefaultOpenAIModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultOpenAIBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultHTTPTimeout
	}

	return cfg
}

func (cfg OpenAIConfig) Validate() error {
	cfg = cfg.withDefaults()

	if err := requireValue("OPENAI_API_KEY", cfg.APIKey); err != nil {
		return err
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return fmt.Errorf("invalid OPENAI_BASE_URL: %w", err)
	}

	return nil
}

type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAIProvider(cfg OpenAIConfig) (*OpenAIProvider, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &OpenAIProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: cfg.BaseURL,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (p *OpenAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt is required")
	}

	payload := openAIResponseRequest{
		Model: p.model,
		Input: prompt,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/responses",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create OpenAI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send OpenAI request: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read OpenAI response: %w", err)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", openAIStatusError(res.StatusCode, resBody)
	}

	var response openAIResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return "", fmt.Errorf("decode OpenAI response: %w", err)
	}

	text := response.Text()
	if text == "" {
		return "", errors.New("OpenAI response did not include text output")
	}

	return text, nil
}

type openAIResponseRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openAIResponse struct {
	OutputText string             `json:"output_text"`
	Output     []openAIOutputItem `json:"output"`
}

func (r openAIResponse) Text() string {
	if strings.TrimSpace(r.OutputText) != "" {
		return r.OutputText
	}

	var parts []string
	for _, item := range r.Output {
		for _, content := range item.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				parts = append(parts, content.Text)
			}
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n"))
}

type openAIOutputItem struct {
	Type    string                `json:"type"`
	Content []openAIOutputContent `json:"content"`
}

type openAIOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func openAIStatusError(statusCode int, body []byte) error {
	var errorResponse openAIErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		message := strings.TrimSpace(errorResponse.Error.Message)
		if message != "" {
			return fmt.Errorf("OpenAI request failed with status %d: %s", statusCode, message)
		}
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Errorf("OpenAI request failed with status %d", statusCode)
	}
	return fmt.Errorf("OpenAI request failed with status %d: %s", statusCode, bodyText)
}
