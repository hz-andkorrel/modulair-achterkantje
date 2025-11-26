package handlers

import (
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// StatusHandler is used to handle a status check, implementing the BaseHandler interface.
type StatusHandler struct{}

// NewStatusHandler creates a new instance of StatusHandler.
// There is no additional setup required for this handler.
func NewStatusHandler() BaseHandler {
	return &StatusHandler{}
}

// RegisterRoutes registers the status check route with the provided Gin engine.
func (h *StatusHandler) RegisterRoutes(engine *gin.Engine, _ *services.JwtService) {
	engine.GET("/status", h.getStatus())
}

// GetHandler returns the Gin handler function for the status check.
// It responds with a JSON object indicating the service status.
func (h *StatusHandler) getStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	}
}
