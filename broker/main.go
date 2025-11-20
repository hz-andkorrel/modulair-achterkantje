package main

import (
	"context"
	"hotelhub/broker/internal/database"
	"hotelhub/broker/internal/handlers"
	"hotelhub/broker/internal/middleware"
	"hotelhub/broker/internal/plugins"
	"hotelhub/broker/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// The entry point of the broker service start by loading the configuration from the .env file.
func main() {
	configuration := services.NewConfiguration()
	jwtService := services.NewJwtService(configuration)

	// Initialize database if configured
	var db *database.DB
	var cleanupStopChan chan struct{}

	if cfg.DatabaseURL != "" {
		db, err = database.NewDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize database schema
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := db.InitSchema(ctx); err != nil {
			cancel()
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
		cancel()

		// Set database on JWT manager for token persistence
		jwtManager.SetDB(db)
		log.Println("Database connected and initialized")

		// Start background cleanup job for expired tokens
		cleanupStopChan = make(chan struct{})
		go startCleanupJob(db, cleanupStopChan)
	} else {
		log.Println("Database disabled - tokens will not be persisted")
	}

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode) // Disable Gin's default logger for custom logging
	router := gin.New()

	// Add recovery middleware to handle panics
	router.Use(gin.Recovery())

	// Add custom request logging middleware
	router.Use(middleware.RequestLogger())

	// Add rate limiting if enabled
	if cfg.RateLimitEnabled {
		router.Use(middleware.IPRateLimit(cfg.RateLimitMaxRequests, cfg.RateLimitWindow))
		log.Printf("Rate limiting enabled: %d requests per %v per IP", cfg.RateLimitMaxRequests, cfg.RateLimitWindow)
	}

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
		v1.GET("/status", middleware.OptionalAuth(jwtManager), handlers.GetStatus)

		// Health check endpoint for plugins
		v1.GET("/health/plugins", handlers.CheckPluginsHealth)

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

	// Configure plugin persistence (loads existing registrations if present)
	if cfg.PluginsPersistPath != "" {
		if err := plugins.Global.SetPersistPath(cfg.PluginsPersistPath); err != nil {
			log.Printf("Warning: failed to set plugin persist path: %v", err)
		}
	}

	// Set configuration for proxy handlers
	handlers.SetProxyConfig(cfg)

	// Create HTTP server with timeouts
	addr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Broker service starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// SIGINT (Ctrl+C), SIGTERM (docker stop, kubernetes)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Stop background cleanup job
	if cleanupStopChan != nil {
		close(cleanupStopChan)
	}

	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}

	log.Println("Server shutdown complete")
}
