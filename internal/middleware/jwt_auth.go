package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"booknest/internal/domain"
)

type JWTConfig struct {
	Keys map[string][]byte
}

func LoadJWTConfigFromEnv() (JWTConfig, error) {
	keys := map[string][]byte{
		domain.PrevKeyID:    []byte(os.Getenv("JWT_SECRET_V0")),
		domain.CurrentKeyID: []byte(os.Getenv("JWT_SECRET_V1")),
	}

	cfg := JWTConfig{Keys: make(map[string][]byte, len(keys))}
	for kid, key := range keys {
		if len(key) == 0 {
			continue
		}
		cfg.Keys[kid] = key
	}
	if len(cfg.Keys) == 0 {
		return JWTConfig{}, fmt.Errorf("at least one JWT key must be configured (%s, %s)", domain.CurrentKeyID, domain.PrevKeyID)
	}
	return cfg, nil
}

// JWTAuthMiddleware verifies JWT token and injects user info into context.
func JWTAuthMiddleware(jwtConfig JWTConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			ctx.Abort()
			return
		}

		if len(jwtConfig.Keys) == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "server auth is not configured"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("missing kid")
			}

			key, exists := jwtConfig.Keys[kid]
			if !exists {
				return nil, errors.New("invalid kid")
			}

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return key, nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims["user_id"])
		ctx.Set("email", claims["email"])
		ctx.Set("user_role", claims["user_role"])
		ctx.Next()
	}
}
