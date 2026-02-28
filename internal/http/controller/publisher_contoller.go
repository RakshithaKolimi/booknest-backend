package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

type publisherController struct {
	service domain.PublisherService
}

// NewPublisherController creates a new publisher controller instance
func NewPublisherController(service domain.PublisherService) domain.PublisherController {
	return &publisherController{service: service}
}

// RegisterRoutes registers all publisher routes
func (c *publisherController) RegisterRoutes(r gin.IRouter) {
	RegisterPublisherRoutes(r, getJWTConfig(), c)
}

func (c *publisherController) List(ctx *gin.Context) {
	limit := 50
	offset := 0
	search := strings.TrimSpace(ctx.Query("search"))

	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}
	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	publishers, err := c.service.List(ctx, limit, offset, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, publishers)
}

// GetByID godoc
// @Summary      Get publisher
// @Description  Fetch publisher by ID
// @Tags         Publishers
// @Produce      json
// @Param        id   path  string  true  "Publisher ID"
// @Success      200  {object}  domain.Publisher
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /publishers/{id} [get]
func (c *publisherController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher id"})
		return
	}

	publisher, err := c.service.FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "publisher not found"})
		return
	}

	ctx.JSON(http.StatusOK, publisher)
}

// Create godoc
// @Summary      Create publisher
// @Description  Create a new publisher
// @Tags         Publishers
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.PublisherInput  true  "Publisher input"
// @Success      201  {object}  domain.Publisher
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /publishers [post]
func (c *publisherController) Create(ctx *gin.Context) {
	var input domain.PublisherInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	publisher, err := c.service.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, publisher)
}

// Update godoc
// @Summary      Update publisher
// @Description  Update publisher details
// @Tags         Publishers
// @Accept       json
// @Produce      json
// @Param        id       path  string                 true  "Publisher ID"
// @Param        payload  body  domain.PublisherInput  true  "Publisher input"
// @Success      200  {object}  domain.Publisher
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /publishers/{id} [put]
func (c *publisherController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher id"})
		return
	}

	var input domain.PublisherInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	publisher, err := c.service.Update(ctx, id, input)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "publisher not found"})
		return
	}

	ctx.JSON(http.StatusOK, publisher)
}

// SetActive godoc
// @Summary      Activate or deactivate publisher
// @Description  Enable or disable a publisher
// @Tags         Publishers
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Publisher ID"
// @Param        payload  body  map[string]bool  true  "Active flag"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /publishers/{id}/status [patch]
func (c *publisherController) SetActive(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher id"})
		return
	}

	var input struct {
		Active bool `json:"active" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.SetActive(ctx, id, input.Active); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Publisher status updated successfully",
	})
}

// Delete godoc
// @Summary      Delete publisher
// @Description  Soft delete publisher
// @Tags         Publishers
// @Produce      json
// @Param        id   path  string  true  "Publisher ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /publishers/{id} [delete]
func (c *publisherController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher id"})
		return
	}

	if err := c.service.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Publisher deleted successfully",
	})
}
