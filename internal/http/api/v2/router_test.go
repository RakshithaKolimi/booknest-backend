package v2

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRouterAndRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter()
	if router.Version() != Version {
		t.Fatalf("expected version %q, got %q", Version, router.Version())
	}

	engine := gin.New()
	group := engine.Group("/api/v2")
	router.RegisterRoutes(group)

	if len(engine.Routes()) != 0 {
		t.Fatalf("expected no v2 routes yet, got %+v", engine.Routes())
	}
}
