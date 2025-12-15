package handlers

import (
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// StatusHandler is used to handle a status check, implementing the BaseHandler interface.
// It requires a RepositoryStrategy to access the data layer if needed.
type StatusHandler struct {
	repository *repository.RepositoryStrategy
}

// NewStatusHandler creates a new instance of StatusHandler.
// It takes a RepositoryStrategy as a parameter to access the data layer if needed.
func NewStatusHandler(repository *repository.RepositoryStrategy) BaseHandler {
	return &StatusHandler{
		repository: repository,
	}
}

// RegisterRoutes registers the status check route with the provided Gin engine.
func (handler *StatusHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
	engine.GET("/status", middleware.JwtMiddleware(jwt, handler.repository, "", "access"), handler.getStatus())
}

// GetHandler returns the Gin handler function for the status check.
// It responds with a JSON object indicating the service status.
func (handler *StatusHandler) getStatus() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status":        "ok",
			"authenticated": context.GetBool("authenticated"),
		})
	}
}
