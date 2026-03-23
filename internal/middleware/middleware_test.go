package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"booknest/internal/domain"
)

// Test that LoggingMiddleware does not break request flow and returns handler response
func TestLoggingMiddleware_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggingMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pong") {
		t.Fatalf("expected body to contain pong, got %s", w.Body.String())
	}
}

// Test ErrorHandler returns JSON for errors accumulated in context
func TestErrorHandler_ReportsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/err", func(c *gin.Context) {
		_ = c.Error(errors.New("bad request"))
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "bad request") {
		t.Fatalf("expected error message in body, got %s", w.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected deny frame header, got %q", got)
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		role       any
		setRole    bool
		wantStatus int
	}{
		{name: "missing role", wantStatus: http.StatusUnauthorized},
		{name: "invalid role type", setRole: true, role: 123, wantStatus: http.StatusUnauthorized},
		{name: "non admin role", setRole: true, role: domain.UserRoleUser, wantStatus: http.StatusForbidden},
		{name: "admin role string", setRole: true, role: "ADMIN", wantStatus: http.StatusOK},
		{name: "admin role typed", setRole: true, role: domain.UserRoleAdmin, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.setRole {
					c.Set("user_role", tc.role)
				}
				c.Next()
			})
			r.Use(RequireAdmin())
			r.GET("/admin", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}
