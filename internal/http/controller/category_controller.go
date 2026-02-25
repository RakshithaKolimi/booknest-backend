package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

type categoryController struct {
	service domain.CategoryService
}

func NewCategoryController(service domain.CategoryService) domain.CategoryController {
	return &categoryController{service: service}
}

func (c *categoryController) RegisterRoutes(r gin.IRouter) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET(routes.CategoriesRoute, c.List)
		protected.GET(routes.CategoryByIDRoute, c.GetByID)
	}

	admin := r.Group("")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.RequireAdmin())
	{
		admin.POST(routes.CategoriesRoute, c.Create)
		admin.PUT(routes.CategoryByIDRoute, c.Update)
		admin.DELETE(routes.CategoryByIDRoute, c.Delete)
	}
}

func (c *categoryController) Create(ctx *gin.Context) {
	var input domain.CategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := c.service.Create(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, category)
}

func (c *categoryController) List(ctx *gin.Context) {
	limit := 20
	offset := 0

	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}
	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	categories, err := c.service.List(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, categories)
}

func (c *categoryController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := c.service.FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (c *categoryController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var input domain.CategoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := c.service.Update(ctx, id, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (c *categoryController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := c.service.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
