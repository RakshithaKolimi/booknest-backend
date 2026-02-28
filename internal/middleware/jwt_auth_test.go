package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"booknest/internal/domain"
)

func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := JWTConfig{
		Keys: map[string][]byte{
			domain.CurrentKeyID: []byte("test_jwt_secret"),
		},
	}
	r := gin.New()
	r.Use(JWTAuthMiddleware(jwtConfig))
	r.GET("/private", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/private", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := JWTConfig{
		Keys: map[string][]byte{
			domain.CurrentKeyID: []byte("test_jwt_secret"),
		},
	}
	r := gin.New()
	r.Use(JWTAuthMiddleware(jwtConfig))
	r.GET("/private", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/private", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := JWTConfig{
		Keys: map[string][]byte{
			domain.CurrentKeyID: []byte("test_jwt_secret"),
		},
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "some-id",
		"email":   "a@b.com",
	})

	token.Header["kid"] = domain.CurrentKeyID
	s, err := token.SignedString([]byte("test_jwt_secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := gin.New()
	r.Use(JWTAuthMiddleware(jwtConfig))
	r.GET("/private", func(c *gin.Context) {
		// handler should see user info set by middleware
		if _, ok := c.Get("user_id"); !ok {
			t.Fatalf("expected user_id in context")
		}
		if _, ok := c.Get("email"); !ok {
			t.Fatalf("expected email in context")
		}
		c.Status(200)
	})

	req := httptest.NewRequest("GET", "/private", nil)
	req.Header.Set("Authorization", "Bearer "+s)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
