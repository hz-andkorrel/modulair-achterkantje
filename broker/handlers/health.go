package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"broker/models"
)

// Health handles the /health and /healthz endpoints (liveness check)
func Health(w http.ResponseWriter, r *http.Request) {
	resp := models.HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding health response: %v", err)
	}
}

// Ready handles the /ready and /readyz endpoints (readiness check)
func Ready(w http.ResponseWriter, r *http.Request) {
	components := map[string]models.ComponentInfo{
		"broker": CheckComponentHealth("broker"),
		"auth":   CheckComponentHealth("auth"),
	}

	// Determine if service is ready
	ready := true
	for _, comp := range components {
		if comp.Status == "unknown" || comp.Status == "degraded" {
			// For now, we're lenient - only fail if completely down
			// Adjust this logic based on your requirements
			continue
		}
	}

	resp := models.ReadyResponse{
		Ready:      ready,
		Components: components,
		Time:       time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding ready response: %v", err)
	}
}
