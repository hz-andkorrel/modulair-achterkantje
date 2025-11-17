package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/config"
	"broker/internal/database"
	"broker/internal/handlers"
	"broker/internal/jwt"
	"broker/internal/middleware"
	"broker/internal/plugins"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize JWT manager
	jwtManager, err := jwt.NewManager(cfg.JWTExpiry, cfg.JWTIssuer)
	if err != nil {
		log.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Initialize database if configured
	var db *database.DB
	if cfg.DatabaseURL != "" {
		db, err = database.NewDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Initialize database schema
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.InitSchema(ctx); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}

		// Set database on JWT manager for token persistence
		jwtManager.SetDB(db)
		log.Println("Database connected and initialized")
	} else {
		log.Println("Database disabled - tokens will not be persisted")
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
		//   "host": "http://localhost:8080",
		//   "base-api-route": "/kiosk",
		//   "settings-route": "/kiosk/settings",
		//   "api-routes": ["/status", "/reset", "/welcome"],
		//   "enabled": true
		// }

		// Require authentication for plugin registration
		v1.POST("/route", middleware.RequireAuth(jwtManager), handlers.RegisterPlugin)
		v1.GET("/routes", handlers.ListPlugins)
		v1.PUT("/route/:slug", middleware.RequireAuth(jwtManager), handlers.UpdatePlugin)
		v1.DELETE("/route/:slug", middleware.RequireAuth(jwtManager), handlers.DeletePlugin)
	}

	// Catch-all proxy route: forward requests to registered plugins
	// This must be registered AFTER specific routes to avoid conflicts
	router.NoRoute(handlers.ProxyToPlugin)

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
