package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterOrderRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *orderController) {
	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.POST(routes.OrderCheckoutRoute, controller.Checkout)
		protected.POST(routes.OrderConfirmRoute, controller.ConfirmPayment)
		protected.GET(routes.OrdersRoute, controller.ListMyOrders)
	}

	admin := r.Group("")
	admin.Use(middleware.JWTAuthMiddleware(jwtConfig), middleware.RequireAdmin())
	{
		admin.GET(routes.AdminOrdersRoute, controller.ListAllOrders)
	}
}
