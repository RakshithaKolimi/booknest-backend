package domain

import (
	"context"

	"github.com/gin-gonic/gin"
)

type AIController interface {
	RegisterRoutes(r gin.IRouter)
}

type AIService interface {
	Chat(ctx context.Context, input AIChatRequest) (*AIChatResponse, error)
	Embed(ctx context.Context, inputs []string) ([][]float64, error)
}

type AIChatRequest struct {
	Message string `json:"message"`
	Prompt  string `json:"prompt,omitempty"`
}

type AIChatResponse struct {
	Message string `json:"message"`
}
