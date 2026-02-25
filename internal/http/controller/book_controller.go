package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
	"booknest/internal/middleware"
)

type bookController struct {
	service domain.BookService
}

func NewBookController(service domain.BookService) domain.BookController {
	return &bookController{service: service}
}

func (c *bookController) RegisterRoutes(r gin.IRouter) {
	public := r.Group("/books")
	{
		public.POST("/filter", c.filterBooks)
		public.GET("/:id", c.getBook)
		public.GET("", c.listBooks)
	}

	admin := r.Group("/books")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.RequireAdmin())
	{
		admin.POST("", c.createBook)
		admin.PUT("/:id", c.updateBook)
		admin.DELETE("/:id", c.deleteBook)
	}
}

// createBook godoc
// @Summary      Create book
// @Description  Creates a new book (admin only)
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.BookInput  true  "Book input"
// @Success      201  {object}  domain.Book
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /books [post]
func (c *bookController) createBook(ctx *gin.Context) {
	var input domain.BookInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.service.CreateBook(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, book)
}

// getBook godoc
// @Summary      Get book by ID
// @Description  Fetches a single book by its ID
// @Tags         Books
// @Produce      json
// @Param        id  path  string  true  "Book ID"
// @Success      200  {object}  domain.Book
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /books/{id} [get]
func (c *bookController) getBook(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := c.service.GetBook(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	ctx.JSON(http.StatusOK, book)
}

// listBooks godoc
// @Summary      List books
// @Description  Returns a list of books
// @Tags         Books
// @Produce      json
// @Success      200  {array}  domain.Book
// @Failure      500  {object}  map[string]string
// @Router       /books [get]
func (c *bookController) listBooks(ctx *gin.Context) {
	limit := 10
	offset := 0

	books, err := c.service.ListBooks(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, books)
}

// filterBooks godoc
// @Summary      Filter books
// @Description  Filters books by criteria and pagination
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        search  query  string  false  "Search by name, author, or ISBN"
// @Param        limit   query  int     false  "Result limit"
// @Param        offset  query  int     false  "Result offset"
// @Param        payload  body  domain.BookFilter  false  "Book filter payload"
// @Success      200  {object}  domain.BookSearchResult
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /books/filter [post]
func (c *bookController) filterBooks(ctx *gin.Context) {
	var filter domain.BookFilter

	if v := ctx.Query("search"); v != "" {
		filter.Search = &v
	}

	if err := ctx.ShouldBindJSON(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := uint64(10)
	offset := uint64(0)

	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.ParseUint(v, 10, 64)
	}

	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.ParseUint(v, 10, 64)
	}

	result, err := c.service.FilterByCriteria(
		ctx,
		filter,
		domain.QueryOptions{Limit: limit, Offset: offset},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *bookController) updateBook(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input domain.BookInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := c.service.UpdateBook(ctx, id, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, book)
}

func (c *bookController) deleteBook(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.DeleteBook(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}
