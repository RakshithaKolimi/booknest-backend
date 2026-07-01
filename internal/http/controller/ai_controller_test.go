package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
	"booknest/internal/service/ai_service"
)

type mockAIService struct {
	chatFunc     func(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error)
	generateFunc func(ctx context.Context, prompt string) (string, error)
	embedFunc    func(ctx context.Context, inputs []string) ([][]float64, error)
}

func (m *mockAIService) Chat(ctx context.Context, request domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, request, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAIService) Generate(ctx context.Context, prompt string) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt)
	}
	return "", errors.New("not implemented")
}

func (m *mockAIService) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, inputs)
	}
	return nil, errors.New("not implemented")
}

func TestAIControllerChatSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	book := domain.Book{ID: uuid.New(), Name: "Dune"}
	service := &mockAIService{
		chatFunc: func(ctx context.Context, input domain.AIChatRequest, gotUserID string) (*domain.AIChatResponse, error) {
			if gotUserID != userID.String() {
				t.Fatalf("expected user ID %s, got %s", userID, gotUserID)
			}
			if input.Message != "find Dune" {
				t.Fatalf("expected sanitized message, got %q", input.Message)
			}
			if input.SessionID != "abc-123" {
				t.Fatalf("expected session ID from request, got %q", input.SessionID)
			}
			return &domain.AIChatResponse{
				SessionID:  "abc-123",
				Message:    "Here is the book I found.",
				References: []domain.Book{book},
			}, nil
		},
	}
	controller := NewAIController(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID.String())
	c.Request = httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"session_id":"abc-123","message":"  find   Dune  "}`))
	c.Request.Header.Set("Content-Type", "application/json")

	controller.Chat(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var response domain.AIChatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Message != "Here is the book I found." {
		t.Fatalf("unexpected response message: %q", response.Message)
	}
	if response.SessionID != "abc-123" {
		t.Fatalf("unexpected session ID: %q", response.SessionID)
	}
	if len(response.References) != 1 || response.References[0].ID != book.ID {
		t.Fatalf("unexpected references: %+v", response.References)
	}
}

func TestAIControllerChatValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name  string
		setup func(*gin.Context)
		body  string
		want  int
	}{
		{
			name: "missing user ID",
			body: `{"message":"hello"}`,
			want: http.StatusUnauthorized,
		},
		{
			name: "bad JSON",
			setup: func(c *gin.Context) {
				c.Set("user_id", uuid.New().String())
			},
			body: `{`,
			want: http.StatusBadRequest,
		},
		{
			name: "sanitizer rejects oversized message",
			setup: func(c *gin.Context) {
				c.Set("user_id", uuid.New().String())
			},
			body: `{"message":"` + strings.Repeat("a", defaultMaxInputLength+1) + `"}`,
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewAIController(&mockAIService{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.setup != nil {
				tt.setup(c)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.Chat(c)

			if w.Code != tt.want {
				t.Fatalf("expected %d, got %d: %s", tt.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestAIControllerChatServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		controller *aiController
	}{
		{
			name:       "missing service",
			controller: NewAIController(),
		},
		{
			name: "provider unavailable error",
			controller: NewAIController(&mockAIService{
				chatFunc: func(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
					return nil, ai_service.ErrProviderUnavailable
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_id", uuid.New().String())
			c.Request = httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"message":"hello"}`))
			c.Request.Header.Set("Content-Type", "application/json")

			tt.controller.Chat(c)

			if w.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAIControllerChatServiceErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "message required", err: errors.New("message is required"), want: http.StatusBadRequest},
		{name: "invalid session ID", err: ai_service.ErrInvalidSessionID, want: http.StatusBadRequest},
		{name: "forbidden session", err: ai_service.ErrForbiddenSession, want: http.StatusForbidden},
		{name: "unknown service error", err: errors.New("provider failed"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewAIController(&mockAIService{
				chatFunc: func(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
					return nil, tt.err
				},
			})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_id", uuid.New().String())
			c.Request = httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"message":"hello"}`))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.Chat(c)

			if w.Code != tt.want {
				t.Fatalf("expected %d, got %d: %s", tt.want, w.Code, w.Body.String())
			}
		})
	}
}
