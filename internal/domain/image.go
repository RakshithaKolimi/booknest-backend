package domain

import "github.com/gin-gonic/gin"

type ImageController interface {
	RegisterRoutes(r gin.IRouter)
}
