package infrastructure

import (
	"hotelhub/broker/internal/handlers"
	"hotelhub/broker/middleware"
	"hotelhub/broker/services"

	"github.com/gin-gonic/gin"
)

// The router is an abstraction over Gin that sets up the HTTP server with routes and middleware.
// Server information is retrieved from the application configuration.
// It provides a method to start the server.
type Router struct {
	engine *gin.Engine
	port   string
}

// NewRouter initializes a new Router instance with the provided configuration and JWT service.
// It sets up the Gin engine, applies middleware, and configures the API routes.
func NewRouter(configuration *services.Configuration, jwtService *services.JwtService) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestLogging())

	setupV1(engine, jwtService)

	return &Router{
		engine: engine,
	}
}

// Run starts the HTTP server on the configured port.
func (router *Router) Run() {
	router.engine.Run(":" + router.port)
}

// Helper functions for setting up the routes for version 1 of the API.
func setupV1(engine *gin.Engine, jwtManager *services.JwtService) {
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/health", handlers.CheckPluginsHealth)

		v1.GET("/plugin", handlers.ListPlugins)
		v1.POST("/plugin", handlers.RegisterPlugin)
		v1.PUT("/plugin/:slug", handlers.UpdatePlugin)
		v1.DELETE("/plugin/:slug", handlers.DeletePlugin)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register(nil, jwtManager))
			auth.POST("/login", handlers.Login(nil, jwtManager))
		}
	}
}
