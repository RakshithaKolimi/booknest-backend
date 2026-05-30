package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"booknest/internal/domain"
	"booknest/internal/service/ai_service"
)

type mockAIService struct {
	response *domain.AIChatResponse
	err      error
	input    domain.AIChatRequest
}

func (m *mockAIService) Chat(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error) {
	m.input = input
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestAIChat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &mockAIService{response: &domain.AIChatResponse{Message: "Try The Hobbit."}}

	engine := gin.New()
	NewAIController(service).RegisterRoutes(engine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"message":"recommend fantasy"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var got domain.AIChatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Message != "Try The Hobbit." {
		t.Fatalf("unexpected chat response: %+v", got)
	}
	if service.input.Message != "recommend fantasy" {
		t.Fatalf("expected request to reach service, got %+v", service.input)
	}
}

func TestAIChatValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &mockAIService{err: errors.New("message is required")}

	engine := gin.New()
	NewAIController(service).RegisterRoutes(engine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"message":""}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestAIChatProviderUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &mockAIService{err: ai_service.ErrProviderUnavailable}

	engine := gin.New()
	NewAIController(service).RegisterRoutes(engine)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai/chat", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %d", w.Code)
	}
}
