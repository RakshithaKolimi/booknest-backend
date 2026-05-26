package controller

import (
	"testing"

	"github.com/gin-gonic/gin"

	"booknest/internal/middleware"
)

func hasRoute(engine *gin.Engine, method, path string) bool {
	for _, route := range engine.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func TestConfigSetters(t *testing.T) {
	cfg := middleware.JWTConfig{Keys: map[string][]byte{"kid": []byte("secret")}}
	SetJWTConfig(cfg)
	if got := getJWTConfig(); string(got.Keys["kid"]) != "secret" {
		t.Fatalf("expected jwt config to be set, got %+v", got)
	}

	SetRedisClient(nil)
	if got := getRedisClient(); got != nil {
		t.Fatalf("expected nil redis client, got %+v", got)
	}
}

func TestRegisterRouteHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	jwtConfig := middleware.JWTConfig{Keys: map[string][]byte{"kid": []byte("secret")}}

	RegisterAuthorRoutes(engine, jwtConfig, &authorController{})
	RegisterBookRoutes(engine, jwtConfig, &bookController{})
	RegisterCategoryRoutes(engine, jwtConfig, &categoryController{})
	RegisterCartRoutes(engine, jwtConfig, &cartController{})
	RegisterOrderRoutes(engine, jwtConfig, &orderController{})
	RegisterReviewRoutes(engine, jwtConfig, &reviewController{})
	RegisterImageRoutes(engine, jwtConfig, &imageController{})

	expected := []struct {
		method string
		path   string
	}{
		{method: "GET", path: "/authors"},
		{method: "PUT", path: "/authors/:id"},
		{method: "GET", path: "/books/search"},
		{method: "DELETE", path: "/books/:id"},
		{method: "GET", path: "/categories"},
		{method: "POST", path: "/cart/items"},
		{method: "POST", path: "/orders/cancel"},
		{method: "PUT", path: "/admin/orders/status"},
		{method: "GET", path: "/books/:id/reviews"},
		{method: "POST", path: "/images/upload"},
	}

	for _, tc := range expected {
		if !hasRoute(engine, tc.method, tc.path) {
			t.Fatalf("expected route %s %s to be registered", tc.method, tc.path)
		}
	}
}

func TestControllerRegisterRoutesWrappers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetJWTConfig(middleware.JWTConfig{Keys: map[string][]byte{"kid": []byte("secret")}})

	engine := gin.New()
	NewAuthorController(&mockAuthorService{}).RegisterRoutes(engine)
	NewBookController(&mockBookServiceController{}).RegisterRoutes(engine)
	NewCartController(&mockCartServiceController{}).RegisterRoutes(engine)
	NewCategoryController(&mockCategoryService{}).RegisterRoutes(engine)
	NewOrderController(&mockOrderServiceController{}).RegisterRoutes(engine)
	NewReviewController(&mockReviewServiceController{}).RegisterRoutes(engine)
	NewImageController().RegisterRoutes(engine)

	expected := []struct {
		method string
		path   string
	}{
		{method: "GET", path: "/authors"},
		{method: "GET", path: "/books"},
		{method: "GET", path: "/cart"},
		{method: "GET", path: "/categories"},
		{method: "GET", path: "/admin/orders"},
		{method: "GET", path: "/books/:id/reviews"},
		{method: "POST", path: "/images/upload"},
	}

	for _, tc := range expected {
		if !hasRoute(engine, tc.method, tc.path) {
			t.Fatalf("expected route %s %s to be registered via controller wrapper", tc.method, tc.path)
		}
	}
}
