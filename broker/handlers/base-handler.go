package handlers

import (
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// BaseHandler defines a common interface for all handlers to implement.
// The constructor should return BaseHandler and setup anything required for the handler.
// Each handler must implement the RegisterRoutes method to define its routes.
type BaseHandler interface {
	RegisterRoutes(engine *gin.Engine, jwtService *services.JwtService)
}
