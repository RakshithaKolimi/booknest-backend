package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"booknest/internal/domain"
	"booknest/internal/http/routes"
	aipkg "booknest/internal/pkg/ai"
	"booknest/internal/service/ai_service"
)

type aiController struct {
	service domain.AIService
}

func NewAIController(service ...domain.AIService) *aiController {
	controller := &aiController{}
	if len(service) > 0 {
		controller.service = service[0]
	}
	return controller
}

func (c *aiController) RegisterRoutes(r gin.IRouter) {
	RegisterAIRoutes(r, c)
}

func RegisterAIRoutes(r gin.IRouter, controller *aiController) {
	r.GET(routes.AIHealthRoute, controller.Health)
	r.POST(routes.AIChatRoute, controller.Chat)
}

func (controller *aiController) Health(c *gin.Context) {
	status := aipkg.HealthFromEnv()
	if status.Status != "ok" {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}
	c.JSON(http.StatusOK, status)
}

// Chat godoc
// @Summary      Chat with AI
// @Description  Sends a message to the configured AI provider
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.AIChatRequest  true  "AI chat request"
// @Success      200  {object}  domain.AIChatResponse
// @Failure      400  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /ai/chat [post]
func (controller *aiController) Chat(c *gin.Context) {
	if controller.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": ai_service.ErrProviderUnavailable.Error()})
		return
	}

	var input domain.AIChatRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sanitizeInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := controller.service.Chat(c, input)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "message is required" {
			status = http.StatusBadRequest
		}
		if errors.Is(err, ai_service.ErrProviderUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
