package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/database"
	"broker/internal/jwt"
	"broker/internal/models"
)

// OptionalAuth is middleware that checks for JWT but doesn't require it
// If a valid token is present, it adds user info to the context
func OptionalAuth(jm *jwt.Manager, db *database.DB) gin.HandlerFunc {
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

		// Fetch real user from database
		var user *models.User
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			user, err = db.GetUserByID(ctx, claims.Subject)
			if err != nil || user == nil {
				// User not found in database, continue without user info
				c.Next()
				return
			}
		} else {
			// Fallback to minimal user object if database not available
			user = &models.User{
				ID:    claims.Subject,
				Email: claims.Subject,
			}
		}

		// Add user to context
		c.Set("currentUser", user)
		c.Set("userID", user.ID)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequireAuth is middleware that enforces a valid JWT. If the token is missing or invalid,
// the request is aborted with HTTP 401.
func RequireAuth(jm *jwt.Manager, db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Authorization header
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		// Validate Bearer token format
		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		// Parse and validate the token
		tokenString := header[len(prefix):]
		claims, err := jm.ParseAndVerify(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			return
		}

		// Fetch real user from database
		var user *models.User
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			user, err = db.GetUserByID(ctx, claims.Subject)
			if err != nil || user == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found in database"})
				return
			}
		} else {
			// Fallback to minimal user object if database not available
			user = &models.User{
				ID:    claims.Subject,
				Email: claims.Subject,
			}
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
