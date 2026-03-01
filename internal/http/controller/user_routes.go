package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	"booknest/internal/http/routes"
	"booknest/internal/middleware"
)

func RegisterUserRoutes(r gin.IRouter, jwtConfig middleware.JWTConfig, rdb *redis.Client, controller *userController) {
	auth := r.Group("")
	auth.Use(middleware.LoginRateLimiter(rdb))
	{
		auth.POST(routes.RegisterRoute, controller.Register)
		auth.POST(routes.LoginRoute, controller.Login)
		auth.POST(routes.RefreshRoute, controller.Refresh)
		auth.POST(routes.ForgotPassword, controller.ForgotPassword)
		auth.POST(routes.ResetPasswordByToken, controller.ResetPasswordWithToken)
	}

	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtConfig))
	{
		protected.GET(routes.UserRoute, controller.GetUser)
		protected.DELETE(routes.UserRoute, controller.DeleteUser)
		protected.POST(routes.VerifyEmailRoute, controller.VerifyEmail)
		protected.POST(routes.VerifyMobileRoute, controller.VerifyMobile)
		protected.POST(routes.ResendEmailRoute, controller.ResendEmailVerification)
		protected.POST(routes.ResendMobileOTPRoute, controller.ResendMobileOTP)
		protected.POST(routes.ResetPasswordRoute, controller.ResetPassword)
	}
}
