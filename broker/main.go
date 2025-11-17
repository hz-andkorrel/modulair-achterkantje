package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"broker/internal/config"
	"broker/internal/handlers"
	"broker/internal/plugins"
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

		// Plugin registration: plugins call POST /api/v1/route with their metadata
		// Example body:
		// {
		//   "description": "Toont een welkomstscherm voor gebruikers bij binnenkomst",
		//   "version": "1.0.2",
		//   "slug": "kiosk",
		//   "name": "Kiosk Plug-in",
		//   "base-api-route": "/kiosk",
		//   "settings-route": "/kiosk/settings",
		//   "api-routes": ["/status", "/reset", "/welcome"],
		//   "enabled": true
		// }

		// Require authentication for plugin registration
		v1.POST("/route", middleware.RequireAuth(jwtManager), handlers.RegisterPlugin)
		v1.GET("/routes", handlers.ListPlugins)
	}

	// Configure plugin persistence (loads existing registrations if present)
	if cfg.PluginsPersistPath != "" {
		if err := plugins.Global.SetPersistPath(cfg.PluginsPersistPath); err != nil {
			log.Printf("Warning: failed to set plugin persist path: %v", err)
		}
	}

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Broker service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
