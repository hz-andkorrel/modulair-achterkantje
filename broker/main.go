package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"broker/internal/config"
	"broker/internal/handlers"
	"broker/internal/jwt"
	"broker/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize JWT manager
	jwtManager, err := jwt.NewManager(cfg.JWTExpiry, cfg.JWTIssuer)
	if err != nil {
		log.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Status endpoint with optional authentication
		v1.GET("/status", middleware.OptionalAuth(jwtManager), handlers.GetStatus)
	}

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Broker service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
