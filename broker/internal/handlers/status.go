package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/config"
	"broker/internal/middleware"
)

// StatusResponse represents the API status response
type StatusResponse struct {
	Status        string      `json:"status"`
	Timestamp     time.Time   `json:"timestamp"`
	Version       string      `json:"version"`
	Authenticated bool        `json:"authenticated"`
	User          interface{} `json:"user,omitempty"`
}

// GetStatus returns the API status with optional user info if authenticated
func GetStatus(c *gin.Context) {
	cfg := config.LoadConfig()

	response := StatusResponse{
		Status:        "healthy",
		Timestamp:     time.Now(),
		Version:       cfg.Version,
		Authenticated: middleware.IsAuthenticated(c),
	}

	// If authenticated, include user info
	if user, ok := middleware.GetCurrentUser(c); ok {
		response.User = gin.H{
			"id":    user.ID,
			"email": user.Email,
		}
	}

	c.JSON(http.StatusOK, response)
}
