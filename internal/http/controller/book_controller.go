package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

type bookController struct {
	service domain.BookService
}

func NewBookController(service domain.BookService) domain.BookController {
	return &bookController{service: service}
}

func (c *bookController) RegisterRoutes(r gin.IRouter) {
	RegisterBookRoutes(r, getJWTConfig(), c)
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
	if err := sanitizeInput(&input); err != nil {
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
	limit := 500
	offset := 0

	books, err := c.service.ListBooks(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, books)
}

// queryBooks godoc
// @Summary      Query books with filters and pagination
// @Description  Supports offset and cursor based pagination with catalog filters
// @Tags         Books
// @Produce      json
// @Param        limit           query  int     false  "Result limit (default 12, max 100)"
// @Param        offset          query  int     false  "Result offset (offset mode)"
// @Param        cursor          query  string  false  "Opaque cursor (cursor mode)"
// @Success      200  {object}  domain.BookSearchResult
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /books/search [get]
func (c *bookController) queryBooks(ctx *gin.Context) {
	limit := uint64(12)
	offset := uint64(0)
	var cursor *string

	// Get the limit from the params
	if v := ctx.Query("limit"); v != "" {
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	// Get the offset from the params
	if v := ctx.Query("offset"); v != "" {
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
		offset = parsed
	}

	// Get the cursor from the params
	if v := ctx.Query("cursor"); v != "" {
		cursor = &v
	}

	// Get the filter input 
	filter := domain.BookFilter{}
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize the input
	if err := sanitizeInput(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service method
	result, err := c.service.QueryBooks(
		ctx,
		filter,
		domain.QueryOptions{Limit: limit, Offset: offset, Cursor: cursor},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the result
	ctx.JSON(http.StatusOK, result)
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
	if err := sanitizeInput(&filter); err != nil {
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
	if err := sanitizeInput(&input); err != nil {
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
