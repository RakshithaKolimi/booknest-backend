package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"booknest/internal/domain"
	"booknest/internal/middleware"
)

func signedImageTestToken(t *testing.T, secret string, role domain.UserRole) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   "user-1",
		"email":     "admin@booknest.test",
		"user_role": string(role),
	})
	token.Header["kid"] = domain.CurrentKeyID

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return signed
}

func TestImageUploadRequiresImageFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "image-test-secret"
	jwtConfig := middleware.JWTConfig{
		Keys: map[string][]byte{
			domain.CurrentKeyID: []byte(secret),
		},
	}

	engine := gin.New()
	RegisterImageRoutes(engine, jwtConfig, &imageController{})

	req := httptest.NewRequest(http.MethodPost, "/images/upload", strings.NewReader("not multipart"))
	req.Header.Set("Authorization", "Bearer "+signedImageTestToken(t, secret, domain.UserRoleAdmin))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "image required") {
		t.Fatalf("expected image required error, got %s", w.Body.String())
	}
}

func TestImageUploadRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "image-test-secret"
	jwtConfig := middleware.JWTConfig{
		Keys: map[string][]byte{
			domain.CurrentKeyID: []byte(secret),
		},
	}

	engine := gin.New()
	RegisterImageRoutes(engine, jwtConfig, &imageController{})

	req := httptest.NewRequest(http.MethodPost, "/images/upload", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+signedImageTestToken(t, secret, domain.UserRoleUser))

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "admin access required") {
		t.Fatalf("expected admin access error, got %s", w.Body.String())
	}
}
