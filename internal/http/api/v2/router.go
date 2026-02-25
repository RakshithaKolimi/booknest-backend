package v2

import "github.com/gin-gonic/gin"

const Version = "v2"

// Router is a scaffold for future v2 endpoints.
type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Version() string {
	return Version
}

func (r *Router) RegisterRoutes(_ *gin.RouterGroup) {
	// Intentionally empty until v2 resources are implemented.
}
