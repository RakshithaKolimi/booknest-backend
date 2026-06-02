package ai_service

import (
	"context"
	"errors"
	"testing"

	"booknest/internal/domain"
)

type mockProvider struct {
	prompt  string
	reply   string
	err     error
	inputs  [][]string
	vectors [][]float64
}

func (m *mockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	m.prompt = prompt
	if m.err != nil {
		return "", m.err
	}
	return m.reply, nil
}

func (m *mockProvider) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	m.inputs = append(m.inputs, append([]string(nil), inputs...))
	if m.err != nil {
		return nil, m.err
	}
	if m.vectors != nil {
		return m.vectors, nil
	}
	return [][]float64{{1, 2, 3}}, nil
}

func TestAIServiceChat(t *testing.T) {
	provider := &mockProvider{reply: "A useful answer."}
	service := NewAIService(provider)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: " hello "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Message != "A useful answer." {
		t.Fatalf("unexpected response: %+v", response)
	}
	if provider.prompt != "hello" {
		t.Fatalf("expected trimmed prompt, got %q", provider.prompt)
	}
}

func TestAIServiceChatUsesPromptAlias(t *testing.T) {
	provider := &mockProvider{reply: "A useful answer."}
	service := NewAIService(provider)

	_, err := service.Chat(context.Background(), domain.AIChatRequest{Prompt: "legacy prompt"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.prompt != "legacy prompt" {
		t.Fatalf("expected prompt alias, got %q", provider.prompt)
	}
}

func TestAIServiceChatValidation(t *testing.T) {
	service := NewAIService(&mockProvider{})

	_, err := service.Chat(context.Background(), domain.AIChatRequest{})
	if err == nil || err.Error() != "message is required" {
		t.Fatalf("expected message validation error, got %v", err)
	}
}

func TestAIServiceChatProviderUnavailable(t *testing.T) {
	service := NewAIService(nil)

	_, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "hello"})
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("expected provider unavailable error, got %v", err)
	}
}

func TestAIServiceEmbed(t *testing.T) {
	provider := &mockProvider{vectors: [][]float64{{1, 2}, {3, 4}}}
	service := NewAIService(provider)

	vectors, err := service.Embed(context.Background(), []string{" title ", "description"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if len(provider.inputs) != 1 || provider.inputs[0][0] != " title " {
		t.Fatalf("expected provider to receive original inputs, got %+v", provider.inputs)
	}
}

func TestAIServiceEmbedValidation(t *testing.T) {
	service := NewAIService(&mockProvider{})

	_, err := service.Embed(context.Background(), []string{"valid", " "})
	if err == nil || err.Error() != "inputs must not contain empty strings" {
		t.Fatalf("expected validation error, got %v", err)
	}
}
