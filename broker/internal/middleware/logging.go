package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs all incoming requests with method, path, status, and duration
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Build query string if present
		if query != "" {
			path = path + "?" + query
		}

		// Log format: [timestamp] status | duration | client_ip | method path | error
		if errorMessage != "" {
			log.Printf("[REQUEST] %d | %13v | %15s | %-7s %s | ERROR: %s",
				statusCode,
				duration,
				clientIP,
				method,
				path,
				errorMessage,
			)
		} else {
			log.Printf("[REQUEST] %d | %13v | %15s | %-7s %s",
				statusCode,
				duration,
				clientIP,
				method,
				path,
			)
		}
	}
}

// ProxyLogger logs requests being proxied to plugins
func ProxyLogger(pluginSlug, pluginHost string) {
	log.Printf("[PROXY] Forwarding to plugin '%s' at %s", pluginSlug, pluginHost)
}
