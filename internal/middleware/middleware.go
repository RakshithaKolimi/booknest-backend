package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs each request with its method, path, status, and latency.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		log.Printf("[%d] %s %s (%v)", status, c.Request.Method, c.Request.URL.Path, latency)
	}
}

// ErrorHandler catches panics and errors and returns a JSON response.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			log.Printf("error: %v", c.Errors.Last().Error())
			c.AbortWithStatusJSON(400, gin.H{
				"error": c.Errors.Last().Error(),
			})
		}
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevents MIME sniffing.. i.e., e.g: the header context sayes text/plain but we find javascript
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		// Cannot be loaded under IFrame
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Next()
	}
}
