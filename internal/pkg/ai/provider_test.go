package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Run("loads OpenAI config", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("AI_PROVIDER", " openai ")
		t.Setenv("OPENAI_API_KEY", " test-key ")
		t.Setenv("OPENAI_CHAT_MODEL", " test-model ")
		t.Setenv("OPENAI_BASE_URL", " https://example.test/v1/ ")
		t.Setenv("OPENAI_HTTP_TIMEOUT_SECONDS", "5")

		cfg, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.Provider != ProviderOpenAI {
			t.Fatalf("expected provider %q, got %q", ProviderOpenAI, cfg.Provider)
		}
		if cfg.OpenAI.APIKey != "test-key" {
			t.Fatalf("expected trimmed api key, got %q", cfg.OpenAI.APIKey)
		}
		if cfg.OpenAI.Model != "test-model" {
			t.Fatalf("expected configured model, got %q", cfg.OpenAI.Model)
		}
		if cfg.OpenAI.BaseURL != "https://example.test/v1" {
			t.Fatalf("expected trimmed base URL, got %q", cfg.OpenAI.BaseURL)
		}
		if cfg.OpenAI.Timeout != 5*time.Second {
			t.Fatalf("expected 5s timeout, got %s", cfg.OpenAI.Timeout)
		}
	})

	t.Run("uses defaults", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("OPENAI_API_KEY", "test-key")

		cfg, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.Provider != ProviderOpenAI {
			t.Fatalf("expected default provider %q, got %q", ProviderOpenAI, cfg.Provider)
		}
		if cfg.OpenAI.Model != "" {
			t.Fatalf("expected env config to leave default model to OpenAIConfig, got %q", cfg.OpenAI.Model)
		}
		if cfg.OpenAI.BaseURL != "" {
			t.Fatalf("expected env config to leave default base URL to OpenAIConfig, got %q", cfg.OpenAI.BaseURL)
		}
		if cfg.OpenAI.Timeout != DefaultHTTPTimeout {
			t.Fatalf("expected default timeout %s, got %s", DefaultHTTPTimeout, cfg.OpenAI.Timeout)
		}
	})

	t.Run("requires api key", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("AI_PROVIDER", "openai")

		_, err := LoadConfigFromEnv()
		if err == nil || err.Error() != "OPENAI_API_KEY is required" {
			t.Fatalf("expected api key error, got %v", err)
		}
	})

	t.Run("rejects unsupported provider", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("AI_PROVIDER", "local")
		t.Setenv("OPENAI_API_KEY", "test-key")

		_, err := LoadConfigFromEnv()
		if err == nil || !strings.Contains(err.Error(), `unsupported AI_PROVIDER "local"`) {
			t.Fatalf("expected provider error, got %v", err)
		}
	})

	t.Run("rejects invalid timeout", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("OPENAI_API_KEY", "test-key")
		t.Setenv("OPENAI_HTTP_TIMEOUT_SECONDS", "0")

		_, err := LoadConfigFromEnv()
		if err == nil || err.Error() != "OPENAI_HTTP_TIMEOUT_SECONDS must be greater than 0" {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})
}

func clearAIEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"AI_PROVIDER",
		"OPENAI_API_KEY",
		"OPENAI_CHAT_MODEL",
		"OPENAI_BASE_URL",
		"OPENAI_HTTP_TIMEOUT_SECONDS",
	} {
		t.Setenv(key, "")
	}
}

func TestNewProvider(t *testing.T) {
	provider, err := NewProvider(Config{
		Provider: ProviderOpenAI,
		OpenAI: OpenAIConfig{
			APIKey: "test-key",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := provider.(*OpenAIProvider); !ok {
		t.Fatalf("expected OpenAI provider, got %T", provider)
	}

	_, err = NewProvider(Config{Provider: "future-provider"})
	if err == nil || !strings.Contains(err.Error(), `unsupported AI_PROVIDER "future-provider"`) {
		t.Fatalf("expected unsupported provider error, got %v", err)
	}
}

func TestHealth(t *testing.T) {
	t.Run("reports configured provider", func(t *testing.T) {
		status := Config{
			Provider: ProviderOpenAI,
			OpenAI:   OpenAIConfig{APIKey: "test-key", Model: "test-model"},
		}.Health()

		if status.Status != "ok" || status.Provider != ProviderOpenAI || status.Model != "test-model" || status.Error != "" {
			t.Fatalf("unexpected health status: %+v", status)
		}
	})

	t.Run("reports config error", func(t *testing.T) {
		status := Config{Provider: ProviderOpenAI}.Health()

		if status.Status != "unavailable" || status.Provider != ProviderOpenAI || status.Error != "OPENAI_API_KEY is required" {
			t.Fatalf("unexpected health status: %+v", status)
		}
	})

	t.Run("reports unsupported env provider", func(t *testing.T) {
		clearAIEnv(t)
		t.Setenv("AI_PROVIDER", "local")

		status := HealthFromEnv()
		if status.Status != "unavailable" || status.Provider != "local" || !strings.Contains(status.Error, `unsupported AI_PROVIDER "local"`) {
			t.Fatalf("unexpected health status: %+v", status)
		}
	})
}

func TestOpenAIProviderGenerate(t *testing.T) {
	var receivedAuth string
	var receivedContentType string
	var receivedRequest openAIResponseRequest

	provider, err := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		Model:   "test-model",
		BaseURL: "https://example.test",
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	provider.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected /responses path, got %s", r.URL.Path)
		}
		receivedAuth = r.Header.Get("Authorization")
		receivedContentType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		body := `{
			"output": [{
				"type": "message",
				"content": [
					{"type": "output_text", "text": "First sentence."},
					{"type": "output_text", "text": "Second sentence."}
				]
			}]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	got, err := provider.Generate(context.Background(), "Recommend a book")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "First sentence.\nSecond sentence." {
		t.Fatalf("expected joined output text, got %q", got)
	}
	if receivedAuth != "Bearer test-key" {
		t.Fatalf("expected bearer auth, got %q", receivedAuth)
	}
	if receivedContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", receivedContentType)
	}
	if receivedRequest.Model != "test-model" || receivedRequest.Input != "Recommend a book" {
		t.Fatalf("unexpected request payload: %+v", receivedRequest)
	}
}

func TestOpenAIProviderGenerateUsesOutputText(t *testing.T) {
	provider, err := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: "https://example.test",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	provider.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"output_text":"A direct answer."}`)),
		}, nil
	})

	got, err := provider.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "A direct answer." {
		t.Fatalf("expected output_text, got %q", got)
	}
}

func TestOpenAIProviderGenerateReturnsProviderError(t *testing.T) {
	provider, err := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: "https://example.test",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	provider.client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid api key"}}`)),
		}, nil
	})

	_, err = provider.Generate(context.Background(), "prompt")
	if err == nil || !strings.Contains(err.Error(), "invalid api key") {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestOpenAIProviderGenerateValidatesPrompt(t *testing.T) {
	provider, err := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = provider.Generate(context.Background(), " ")
	if err == nil || err.Error() != "prompt is required" {
		t.Fatalf("expected prompt error, got %v", err)
	}
}
