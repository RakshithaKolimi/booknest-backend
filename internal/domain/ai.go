package domain

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Intent string

const IntentSemanticSearch Intent = "semantic_search"
const IntentRecommendation Intent = "recommendation"
const IntentChat Intent = "chat"
const IntentGetBook Intent = "get_book"
const IntentGetBooksByCategory Intent = "get_books_by_category"
const IntentSummary Intent = "summary"
const IntentGetBookDetails Intent = "get_book_details"
const IntentNotRelated Intent = "not_related"

type AIController interface {
	RegisterRoutes(r gin.IRouter)
}

type AIService interface {
	Chat(ctx context.Context, input AIChatRequest, userID string) (*AIChatResponse, error)
	Generate(ctx context.Context, prompt string) (string, error)
	Embed(ctx context.Context, inputs []string) ([][]float64, error)
}

type AIChatRequest struct {
	Message   string `json:"message"`
	Prompt    string `json:"prompt,omitempty"`
	SessionID string `json:"session_id"`
}

type AIChatResponse struct {
	SessionID  string `json:"session_id,omitempty"`
	Message    string `json:"message"`
	References []Book `json:"references,omitempty"`
}

type AIIntentToolCall struct {
	Tool     string `json:"tool"`
	Query    string `json:"query,omitempty"`
	Category string `json:"category,omitempty"`
	BookName string `json:"book_name,omitempty"`
}
