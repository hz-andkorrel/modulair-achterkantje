package middleware

import (
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JwtMiddleware extracts and verifies the JWT token from the Authorization header of the HTTP request
// It returns the subject from the token claims if verification is successful.
// When authorization fails, it logs the reason and returns an empty string.
// Based on the 'required' flag, it either aborts the request with 401 or allows it to proceed.
// Possible reasons are: missing header, invalid format, or token verification failure.
func JwtMiddleware(jwtService *services.JwtService, repository *repository.RepositoryStrategy, required bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		header := context.GetHeader("Authorization")
		if header == "" {
			authenticationError(context, "Authorization header missing", required)
			return
		}

		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			authenticationError(context, "Invalid authorization header format", required)
			return
		}

		token := header[len(prefix):]
		claims := jwtService.Parse(token)
		if claims == nil {
			authenticationError(context, "Token verification failed", required)
			return
		}

		if !repository.JwtRepository.IsValid(token) {
			authenticationError(context, "Token is revoked or expired", required)
			return
		}

		context.Set("user_id", claims.Subject)
		context.Set("authenticated", true)
		context.Next()
	}
}

// Helper function to handle authentication errors
// The function start by logging the error message to the console.
// Based on the 'required' flag, it either aborts the request wiith 401 or allows it to proceed.
// The error message is send along with the 401 response.
func authenticationError(context *gin.Context, message string, required bool) {
	log.Println("[JWT] " + message)

	if required {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "message"})
	} else {
		context.Next()
	}
}
