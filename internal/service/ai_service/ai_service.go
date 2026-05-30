package ai_service

import (
	"context"
	"errors"
	"strings"

	aiprovider "booknest/internal/ai/provider"
	"booknest/internal/domain"
)

var ErrProviderUnavailable = errors.New("AI provider is unavailable")

type aiService struct {
	provider aiprovider.Provider
}

func NewAIService(provider aiprovider.Provider) domain.AIService {
	return &aiService{provider: provider}
}

func (s *aiService) Chat(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error) {
	message := strings.TrimSpace(input.Message)
	if message == "" {
		message = strings.TrimSpace(input.Prompt)
	}
	if message == "" {
		return nil, errors.New("message is required")
	}
	if s.provider == nil {
		return nil, ErrProviderUnavailable
	}

	reply, err := s.provider.Generate(ctx, message)
	if err != nil {
		return nil, err
	}

	return &domain.AIChatResponse{Message: reply}, nil
}
