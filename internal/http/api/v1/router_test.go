package v1

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type stubController struct {
	called bool
	path   string
}

func (s *stubController) RegisterRoutes(r gin.IRouter) {
	s.called = true
	r.GET(s.path, func(c *gin.Context) {})
}

func TestNewRouterAndRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user := &stubController{path: "/user"}
	book := &stubController{path: "/book"}
	author := &stubController{path: "/author"}
	category := &stubController{path: "/category"}
	publisher := &stubController{path: "/publisher"}
	cart := &stubController{path: "/cart"}
	order := &stubController{path: "/order"}
	review := &stubController{path: "/review"}
	image := &stubController{path: "/image"}

	router := NewRouter(user, book, author, category, publisher, cart, order, review, image)
	if router.Version() != Version {
		t.Fatalf("expected version %q, got %q", Version, router.Version())
	}

	engine := gin.New()
	group := engine.Group("/api/v1")
	router.RegisterRoutes(group)

	stubs := []*stubController{user, book, author, category, publisher, cart, order, review, image}
	for _, stub := range stubs {
		if !stub.called {
			t.Fatal("expected all controller registrars to be called")
		}
	}
}
