package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterCategoryRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *categoryController) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.GET(routes.CategoriesRoute, controller.List)
		protected.GET(routes.CategoryByIDRoute, controller.GetByID)
	}

	admin := r.Group("")
	admin.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		admin.POST(routes.CategoriesRoute, controller.Create)
		admin.PUT(routes.CategoryByIDRoute, controller.Update)
		admin.DELETE(routes.CategoryByIDRoute, controller.Delete)
	}
}
