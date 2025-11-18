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

// startCleanupJob runs a background job to clean up expired tokens
func startCleanupJob(db *database.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Started background token cleanup job (runs every hour)")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		count, err := db.CleanupExpiredTokens(ctx)
		cancel()

		if err != nil {
			log.Printf("Token cleanup failed: %v", err)
		} else if count > 0 {
			log.Printf("Cleaned up %d expired tokens", count)
		}
	}
}

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

		// Start background cleanup job for expired tokens
		go startCleanupJob(db)
	} else {
		log.Println("Database disabled - tokens will not be persisted")
	}

	// Initialize Gin router
	router := gin.Default()

	// Initialize plugin registry with database if available
	if db != nil {
		if err := plugins.Global.SetDB(db); err != nil {
			log.Printf("Warning: failed to set plugin database: %v", err)
		} else {
			log.Println("Plugin registry connected to database")
		}
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Status endpoint with optional authentication
		v1.GET("/status", middleware.OptionalAuth(jwtManager, db), handlers.GetStatus)

		// Authentication endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register(db, jwtManager))
			auth.POST("/login", handlers.Login(db, jwtManager))
			auth.GET("/me", middleware.RequireAuth(jwtManager, db), handlers.GetMe(db))
			auth.POST("/revoke", middleware.RequireAuth(jwtManager, db), handlers.RevokeToken(db))
			auth.POST("/revoke-all", middleware.RequireAuth(jwtManager, db), handlers.RevokeAllUserTokens(db))
		}

		// Admin endpoints
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireAuth(jwtManager, db))
		{
			admin.POST("/cleanup-tokens", handlers.CleanupExpiredTokens(db))
		}

		// Plugin registration: plugins call POST /api/v1/route with their metadata
		// Example body:
		// {
		//   "description": "Toont een welkomstscherm voor gebruikers bij binnenkomst",
		//   "version": "1.0.2",
		//   "slug": "kiosk",
		//   "name": "Kiosk Plug-in",
		//   "category": "user-interface",
		//   "host": "http://localhost:8080",
		//   "base-api-route": "/kiosk",
		//   "settings-route": "/kiosk/settings",
		//   "api-routes": ["/status", "/reset", "/welcome"],
		//   "enabled": true
		// }

		// Require authentication for plugin registration
		v1.POST("/route", middleware.RequireAuth(jwtManager, db), handlers.RegisterPlugin)
		v1.GET("/routes", handlers.ListPlugins)
		v1.GET("/routes/categories", handlers.GetCategories)
		v1.GET("/routes/category/:category", handlers.ListPluginsByCategory)
		v1.PUT("/route/:slug", middleware.RequireAuth(jwtManager, db), handlers.UpdatePlugin)
		v1.DELETE("/route/:slug", middleware.RequireAuth(jwtManager, db), handlers.DeletePlugin)
	}

	// Catch-all proxy route: forward requests to registered plugins
	// This must be registered AFTER specific routes to avoid conflicts
	router.NoRoute(handlers.ProxyToPlugin)

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Broker service starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
