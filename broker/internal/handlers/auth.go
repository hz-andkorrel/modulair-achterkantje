package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/database"
	"broker/internal/jwt"
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

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func Register(db *database.DB, jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "registration not available (database not configured)"})
			return
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check if user already exists
		existingUser, err := db.GetUserByEmail(ctx, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
			return
		}

		// Create user
		user, err := db.CreateUser(ctx, req.Email, req.Password, req.Name, "user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
			return
		}

		// Generate JWT token
		token, expiresAt, err := jwtManager.GenerateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "user registered successfully",
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
			"token":      token,
			"expires_at": expiresAt,
		})
	}
}

// Login handles user authentication
func Login(db *database.DB, jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "login not available (database not configured)"})
			return
		}

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Validate credentials
		user, err := db.ValidateUserPassword(ctx, req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Update last login
		if err := db.UpdateLastLogin(ctx, user.ID); err != nil {
			// Log error but don't fail login
		}

		// Generate JWT token
		token, expiresAt, err := jwtManager.GenerateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "login successful",
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
			"token":      token,
			"expires_at": expiresAt,
		})
	}
}

// GetMe returns the current authenticated user's information
func GetMe(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user info not available (database not configured)"})
			return
		}

		user, ok := middleware.GetCurrentUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Fetch full user details from database
		fullUser, err := db.GetUserByID(ctx, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
			return
		}
		if fullUser == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":         fullUser.ID,
				"email":      fullUser.Email,
				"name":       fullUser.Name,
				"role":       fullUser.Role,
				"created_at": fullUser.CreatedAt,
			},
		})
	}
}
