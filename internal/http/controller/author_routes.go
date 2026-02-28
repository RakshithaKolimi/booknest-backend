package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterAuthorRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *authorController) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.GET(routes.AuthorsRoute, controller.List)
		protected.GET(routes.AuthorByIDRoute, controller.GetByID)
	}

	admin := r.Group("")
	admin.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		admin.POST(routes.AuthorsRoute, controller.Create)
		admin.PUT(routes.AuthorByIDRoute, controller.Update)
		admin.DELETE(routes.AuthorByIDRoute, controller.Delete)
	}
}
