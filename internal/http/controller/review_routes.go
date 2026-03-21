package controller

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterReviewRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, controller *reviewController) {
	public := r.Group("")
	{
		public.GET(routes.BookReviewsRoute, controller.ListBookReviews)
	}

	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.POST(routes.BookReviewsRoute, controller.UpsertBookReview)
	}
}
