package ai_service

import (
	"context"
	"errors"
	"testing"

	"booknest/internal/domain"
)

type mockProvider struct {
	prompt string
	reply  string
	err    error
}

func (m *mockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	m.prompt = prompt
	if m.err != nil {
		return "", m.err
	}
	return m.reply, nil
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
