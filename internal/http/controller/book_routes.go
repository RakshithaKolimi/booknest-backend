package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/middleware"
)

func RegisterBookRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *bookController) {
	public := r.Group("/books")
	{
		public.GET("/search", controller.queryBooks)
		public.GET("/semantic-search", controller.semanticSearch)
		public.POST("/filter", controller.filterBooks)
		public.GET("/:id", controller.getBook)
		public.GET("", controller.listBooks)
	}

	user := r.Group("/books")
	user.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		user.GET("/recommend", controller.recommendBooks)
	}

	admin := r.Group("/books")
	admin.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		admin.POST("", controller.createBook)
		admin.POST("/:id/summary", controller.generateSummary)
		admin.POST("/:id/categories", controller.generateCategories)
		admin.POST("/:id/embeddings", controller.generateEmbeddings)
		admin.PUT("/:id", controller.updateBook)
		admin.DELETE("/:id", controller.deleteBook)
	}
}
