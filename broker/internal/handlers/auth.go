package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/database"
	"broker/internal/middleware"
)

// RevokeToken revokes the current user's token
func RevokeToken(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "token revocation not available (database not configured)"})
			return
		}

		// Get the token from the Authorization header
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := header[len(prefix):]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.RevokeToken(ctx, token); err != nil {
			if err.Error() == "token not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token", "details": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
	}
}

// RevokeAllUserTokens revokes all tokens for the current user
func RevokeAllUserTokens(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "token revocation not available (database not configured)"})
			return
		}

		user, ok := middleware.GetCurrentUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.RevokeAllTokensForSubject(ctx, user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke tokens", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "all tokens revoked successfully"})
	}
}

// CleanupExpiredTokens manually triggers token cleanup (admin endpoint)
func CleanupExpiredTokens(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cleanup not available (database not configured)"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		count, err := db.CleanupExpiredTokens(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cleanup failed", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "cleanup completed",
			"deleted": count,
		})
	}
}
