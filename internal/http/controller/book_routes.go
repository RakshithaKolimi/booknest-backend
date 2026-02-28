package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/middleware"
)

func RegisterBookRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *bookController) {
	public := r.Group("/books")
	{
		public.POST("/filter", controller.filterBooks)
		public.GET("/:id", controller.getBook)
		public.GET("", controller.listBooks)
	}

	admin := r.Group("/books")
	admin.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		admin.POST("", controller.createBook)
		admin.PUT("/:id", controller.updateBook)
		admin.DELETE("/:id", controller.deleteBook)
	}
}
