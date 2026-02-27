package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

type authorController struct {
	service domain.AuthorService
}

func NewAuthorController(service domain.AuthorService) domain.AuthorController {
	return &authorController{service: service}
}

func (c *authorController) RegisterRoutes(r gin.IRouter) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET(routes.AuthorsRoute, c.List)
		protected.GET(routes.AuthorByIDRoute, c.GetByID)
	}

	admin := r.Group("")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.RequireAdmin())
	{
		admin.POST(routes.AuthorsRoute, c.Create)
		admin.PUT(routes.AuthorByIDRoute, c.Update)
		admin.DELETE(routes.AuthorByIDRoute, c.Delete)
	}
}

// Create godoc
// @Summary      Create author
// @Description  Creates a new author (admin only)
// @Tags         Authors
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.AuthorInput  true  "Author input"
// @Success      201  {object}  domain.Author
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /authors [post]
func (c *authorController) Create(ctx *gin.Context) {
	var input domain.AuthorInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid input"})
		return
	}

	author, err := c.service.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, author)
}

// List godoc
// @Summary      List authors
// @Description  Lists authors
// @Tags         Authors
// @Produce      json
// @Param        limit   query  int  false  "Result limit"
// @Param        offset  query  int  false  "Result offset"
// @Param        search  query  string  false  "Search by author name"
// @Success      200  {array}  domain.Author
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /authors [get]
func (c *authorController) List(ctx *gin.Context) {
	limit := 50
	offset := 0
	search := strings.TrimSpace(ctx.Query("search"))

	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}
	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	authors, err := c.service.List(ctx, limit, offset, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, authors)
}

// GetByID godoc
// @Summary      Get author by ID
// @Description  Fetches a single author by its ID
// @Tags         Authors
// @Produce      json
// @Param        id  path  string  true  "Author ID"
// @Success      200  {object}  domain.Author
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /authors/{id} [get]
func (c *authorController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id"})
		return
	}

	author, err := c.service.FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "author not found"})
		return
	}

	ctx.JSON(http.StatusOK, author)
}

func (c *authorController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id"})
		return
	}

	var input domain.AuthorInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	author, err := c.service.Update(ctx, id, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, author)
}

func (c *authorController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id"})
		return
	}

	if err := c.service.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Author deleted successfully"})
}
