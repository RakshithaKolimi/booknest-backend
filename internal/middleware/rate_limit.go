package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[string]*visitor)
var mu sync.Mutex

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(5, 10) // 5 req/sec, burst 10
		visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Calm down.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Global state
func LoginRateLimit() gin.HandlerFunc {
	limiter := rate.NewLimiter(1, 5) // 1 req/sec, burst 5

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimiter uses Redis when provided; otherwise it falls back to
// in-memory per-IP limiting so route setup remains safe without Redis wiring.
func LoginRateLimiter(rdb ...*redis.Client) gin.HandlerFunc {
	var redisClient *redis.Client
	if len(rdb) > 0 {
		redisClient = rdb[0]
	}

	return func(c *gin.Context) {
		if redisClient == nil {
			limiter := getVisitor("login:" + c.ClientIP())
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many login attempts. Try later.",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		ip := c.ClientIP()

		key := "login_limit:" + ip
		limit := 5
		window := 10 * time.Minute

		count, _ := redisClient.Incr(key).Result()

		if count == 1 {
			redisClient.Expire(key, window)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Try later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
