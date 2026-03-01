package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"golang.org/x/time/rate"
)

func resetVisitors() {
	mu.Lock()
	defer mu.Unlock()
	visitors = make(map[string]*visitor)
}

func TestRateLimitMiddleware_AllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	r := gin.New()
	r.Use(RateLimitMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	mu.Lock()
	visitors["1.2.3.4"] = &visitor{limiter: rate.NewLimiter(0, 0)}
	mu.Unlock()

	r := gin.New()
	r.Use(RateLimitMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestLoginRateLimit_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(LoginRateLimit())
	r.POST("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	var lastCode int
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		lastCode = w.Code
	}

	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("expected 6th request to return 429, got %d", lastCode)
	}
}

func TestLoginRateLimiter_FallbackWithoutRedis_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	r := gin.New()
	r.Use(LoginRateLimiter())
	r.POST("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLoginRateLimiter_FallbackNilRedisArg_TooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	mu.Lock()
	visitors["login:1.2.3.4"] = &visitor{limiter: rate.NewLimiter(0, 0)}
	mu.Unlock()

	r := gin.New()
	r.Use(LoginRateLimiter((*redis.Client)(nil)))
	r.POST("/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
