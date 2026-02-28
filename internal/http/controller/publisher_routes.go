package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterPublisherRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *publisherController) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.GET(routes.PublisherRoute, controller.List)
		protected.POST(routes.PublisherRoute, controller.Create)
		protected.GET(routes.PublisherByIDRoute, controller.GetByID)
		protected.PUT(routes.PublisherByIDRoute, controller.Update)
		protected.PATCH(routes.PublisherStatusRoute, controller.SetActive)
		protected.DELETE(routes.PublisherByIDRoute, controller.Delete)
	}
}
