package api

import "github.com/gin-gonic/gin"

const (
	RootPath = "/api"
)

// VersionRegistrar mounts one API version (v1, v2, etc.) under /api.
type VersionRegistrar interface {
	Version() string
	RegisterRoutes(r *gin.RouterGroup)
}

// MountVersions mounts each version registrar under /api/{version}.
func MountVersions(r *gin.Engine, registrars ...VersionRegistrar) {
	api := r.Group(RootPath)
	for _, registrar := range registrars {
		versionGroup := api.Group("/" + registrar.Version())
		registrar.RegisterRoutes(versionGroup)
	}
}
