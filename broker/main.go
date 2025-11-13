package main

import (
	"log"
	"net/http"
	"os"

	"broker/handlers"
	"broker/metrics"
	"broker/middleware"
)

func main() {
	// Initialize metrics tracking
	metrics.Init()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API endpoints with metrics tracking
	mux.HandleFunc("/api/v1/status", middleware.MetricsMiddleware("status", handlers.Status))
	mux.HandleFunc("/api/v1/metrics", middleware.MetricsMiddleware("metrics", handlers.Metrics))

	// Health check endpoints
	mux.HandleFunc("/health", middleware.MetricsMiddleware("health", handlers.Health))
	mux.HandleFunc("/ready", middleware.MetricsMiddleware("ready", handlers.Ready))
	mux.HandleFunc("/healthz", middleware.MetricsMiddleware("health", handlers.Health))
	mux.HandleFunc("/readyz", middleware.MetricsMiddleware("ready", handlers.Ready))

	addr := ":" + port
	log.Printf("broker: starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("broker: server failed: %v", err)
	}
}
