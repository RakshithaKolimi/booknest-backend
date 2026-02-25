package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetupServer_Success(t *testing.T) {
	t.Setenv("SWAGGER_USER", "swagger")
	t.Setenv("SWAGGER_PASSWORD", "swagger-pass")

	originalConnectGORM := connectGORM
	t.Cleanup(func() {
		connectGORM = originalConnectGORM
	})

	connectGORM = func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}

	router, err := SetupServer(&pgxpool.Pool{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func TestSetupServer_ConnectGORMError(t *testing.T) {
	originalConnectGORM := connectGORM
	t.Cleanup(func() {
		connectGORM = originalConnectGORM
	})

	connectGORM = func() (*gorm.DB, error) {
		return nil, errors.New("db unavailable")
	}

	_, err := SetupServer(&pgxpool.Pool{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestSetupServer_VersionedRoutingAndSwaggerV1(t *testing.T) {
	t.Setenv("SWAGGER_USER", "swagger")
	t.Setenv("SWAGGER_PASSWORD", "swagger-pass")

	originalConnectGORM := connectGORM
	t.Cleanup(func() {
		connectGORM = originalConnectGORM
	})

	connectGORM = func() (*gorm.DB, error) {
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}

	router, err := SetupServer(&pgxpool.Pool{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// v1 route should be mounted.
	wV1 := httptest.NewRecorder()
	reqV1, _ := http.NewRequest(http.MethodPost, "/api/v1/login", nil)
	router.ServeHTTP(wV1, reqV1)
	if wV1.Code == http.StatusNotFound {
		t.Fatalf("expected /api/v1/login to be mounted, got 404")
	}

	// Legacy unversioned route should not exist.
	wLegacy := httptest.NewRecorder()
	reqLegacy, _ := http.NewRequest(http.MethodPost, "/login", nil)
	router.ServeHTTP(wLegacy, reqLegacy)
	if wLegacy.Code != http.StatusNotFound {
		t.Fatalf("expected /login to be unmounted, got %d", wLegacy.Code)
	}

	// Versioned Swagger UI should be protected and mounted.
	wSwagger := httptest.NewRecorder()
	reqSwagger, _ := http.NewRequest(http.MethodGet, "/swagger/v1/index.html", nil)
	router.ServeHTTP(wSwagger, reqSwagger)
	if wSwagger.Code != http.StatusUnauthorized {
		t.Fatalf("expected swagger route auth challenge, got %d", wSwagger.Code)
	}
}

func TestUseCORSMiddleware_PreflightAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(useCORSMiddleware(map[string]bool{"http://localhost:3000": true}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow-origin header, got %q", got)
	}
}

func TestUseCORSMiddleware_GetAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(useCORSMiddleware(map[string]bool{"http://localhost:5173": true}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow-origin header, got %q", got)
	}
}

func TestUseCORSMiddleware_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(useCORSMiddleware(map[string]bool{"http://localhost:5173": true}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://malicious.local")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for GET, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}
