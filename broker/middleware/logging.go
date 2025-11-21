package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// The RequestLogging function returns a Gin middleware that logs details of each HTTP request.
// The log is written to the console in a structured format:
// [REQUEST] status | duration | client_ip | method path | error.
func RequestLogging() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		query := "?" + context.Request.URL.RawQuery
		if query == "?" {
			query = ""
		}

		start := time.Now()
		context.Next()
		duration := time.Since(start)

		statusCode := context.Writer.Status()
		clientIP := context.ClientIP()
		method := context.Request.Method
		errorMessage := context.Errors.ByType(gin.ErrorTypePrivate).String()

		log.Printf("[REQUEST] %d | %13v | %15s | %-7s %s | %s",
			statusCode,
			duration,
			clientIP,
			method,
			path+query,
			errorMessage,
		)
	}
}
