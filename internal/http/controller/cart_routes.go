package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterCartRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *cartController) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.GET(routes.CartRoute, controller.GetCart)
		protected.POST(routes.CartItemsRoute, controller.AddItem)
		protected.PUT(routes.CartItemsRoute, controller.UpdateItem)
		protected.DELETE(routes.CartItemRoute, controller.RemoveItem)
		protected.POST(routes.CartClearRoute, controller.ClearCart)
	}
}
