package handlers

import (
	"hotelhub/broker/middleware"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
	"log"

	"github.com/gin-contrib/cors"
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
func NewRouter(configuration *services.Configuration, jwtService *services.JwtService, repository *repository.RepositoryStrategy) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	corsSettings := cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})

	engine.Use(corsSettings)
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestLogging())

	NewStatusHandler(repository).RegisterRoutes(engine, jwtService)
	NewAuthHandler(repository).RegisterRoutes(engine, jwtService)
	NewUserHandler(repository).RegisterRoutes(engine, jwtService)
	NewEventLogHandler(repository).RegisterRoutes(engine, jwtService)

	return &Router{
		engine: engine,
		port:   configuration.ServerPort,
	}
}

// Run starts the HTTP server on the configured port.
func (router *Router) Run() {
	log.Println("[Router] Starting server on port", router.port)
	router.engine.Run(":" + router.port)
}
