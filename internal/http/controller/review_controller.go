package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type reviewController struct {
	service domain.ReviewService
}

func NewReviewController(service domain.ReviewService) domain.ReviewController {
	return &reviewController{service: service}
}

func (c *reviewController) RegisterRoutes(r gin.IRouter) {
	RegisterReviewRoutes(r, getJWTConfig(), c)
}

// ListBookReviews godoc
// @Summary      List book reviews
// @Description  Returns all reviews for a specific book with summary stats
// @Tags         Reviews
// @Produce      json
// @Param        id  path  string  true  "Book ID"
// @Success      200  {object}  domain.ReviewListResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /books/{id}/reviews [get]
func (c *reviewController) ListBookReviews(ctx *gin.Context) {
	bookID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	result, err := c.service.ListBookReviews(ctx, bookID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// UpsertBookReview godoc
// @Summary      Create or update a review
// @Description  Creates or updates the current user's review for a specific book
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Book ID"
// @Param        payload  body  domain.ReviewInput  true  "Review input"
// @Success      201  {object}  domain.Review
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /books/{id}/reviews [post]
func (c *reviewController) UpsertBookReview(ctx *gin.Context) {
	bookID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	var input domain.ReviewInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sanitizeInput(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := c.service.UpsertBookReview(ctx, bookID, userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrReviewRequiresPurchase) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, review)
}
