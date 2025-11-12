package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"broker/internal/jwt"
	"broker/internal/models"
)

// OptionalAuth is middleware that checks for JWT but doesn't require it
// If a valid token is present, it adds user info to the context
func OptionalAuth(jm *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Authorization header
		header := c.GetHeader("Authorization")
		if header == "" {
			// No auth header, continue without user info
			c.Next()
			return
		}

		// Validate Bearer token format
		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			// Invalid format, continue without user info
			c.Next()
			return
		}

		// Parse and validate the token
		tokenString := header[len(prefix):]
		claims, err := jm.ParseAndVerify(tokenString)
		if err != nil {
			// Invalid token, continue without user info
			c.Next()
			return
		}

		// For this MVP, we create a minimal user object from the JWT subject
		// In a real system, you might fetch full user details from a database
		user := &models.User{
			ID:    claims.Subject,
			Email: claims.Subject, // Using subject as email for now
		}

		// Add user to context
		c.Set("currentUser", user)
		c.Set("userID", user.ID)
		c.Set("authenticated", true)

		c.Next()
	}
}

// GetCurrentUser retrieves the current user from the context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	userInterface, exists := c.Get("currentUser")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*models.User)
	return user, ok
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	
	auth, ok := authenticated.(bool)
	return ok && auth
}
