package api

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type stubRegistrar struct {
	version string
	called  bool
}

func (s *stubRegistrar) Version() string {
	return s.version
}

func (s *stubRegistrar) RegisterRoutes(r *gin.RouterGroup) {
	s.called = true
	r.GET("/ping", func(c *gin.Context) {})
}

func TestMountVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	v1 := &stubRegistrar{version: "v1"}
	v2 := &stubRegistrar{version: "v2"}

	MountVersions(engine, v1, v2)

	if !v1.called || !v2.called {
		t.Fatal("expected all registrars to register routes")
	}

	routes := engine.Routes()
	foundV1 := false
	foundV2 := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/ping" {
			foundV1 = true
		}
		if route.Method == "GET" && route.Path == "/api/v2/ping" {
			foundV2 = true
		}
	}

	if !foundV1 || !foundV2 {
		t.Fatalf("expected versioned routes to be mounted, got %+v", routes)
	}
}
