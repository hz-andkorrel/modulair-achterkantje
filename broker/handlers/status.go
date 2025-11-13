package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"broker/metrics"
	"broker/models"
)

// CheckComponentHealth checks the health of a specific component
func CheckComponentHealth(name string) models.ComponentInfo {
	start := time.Now()

	switch name {
	case "broker":
		latency := time.Since(start).Milliseconds()
		return models.ComponentInfo{Status: "ok", Message: "service running", Latency: float64(latency)}

	case "database":
		// Simuleer database check (vervang later met echte DB ping)
		time.Sleep(2 * time.Millisecond)
		latency := time.Since(start).Milliseconds()
		return models.ComponentInfo{Status: "unknown", Message: "not yet implemented", Latency: float64(latency)}

	case "cache":
		// Simuleer cache check
		time.Sleep(1 * time.Millisecond)
		latency := time.Since(start).Milliseconds()
		return models.ComponentInfo{Status: "unknown", Message: "not yet implemented", Latency: float64(latency)}

	case "auth":
		latency := time.Since(start).Milliseconds()
		if os.Getenv("BASIC_AUTH_USER") != "" {
			return models.ComponentInfo{Status: "ok", Message: "basic auth configured", Latency: float64(latency)}
		}
		return models.ComponentInfo{Status: "degraded", Message: "no auth configured", Latency: float64(latency)}

	default:
		latency := time.Since(start).Milliseconds()
		return models.ComponentInfo{Status: "unknown", Message: "unknown component", Latency: float64(latency)}
	}
}

// Status handles the /api/v1/status endpoint
func Status(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("STATUS_VERSION")
	if version == "" {
		version = "0.1.0"
	}

	// Check components
	components := map[string]models.ComponentInfo{
		"broker":   CheckComponentHealth("broker"),
		"auth":     CheckComponentHealth("auth"),
		"database": CheckComponentHealth("database"),
		"cache":    CheckComponentHealth("cache"),
	}

	// Determine overall status based on components
	overallStatus := "ok"
	for _, comp := range components {
		if comp.Status == "degraded" {
			overallStatus = "degraded"
			break
		}
	}

	resp := models.StatusResponse{
		Status:     overallStatus,
		Time:       time.Now().UTC().Format(time.RFC3339),
		Uptime:     metrics.GetUptime(),
		Version:    version,
		Components: components,
	}

	if username, password, ok := r.BasicAuth(); ok {
		envUser := os.Getenv("BASIC_AUTH_USER")
		envPass := os.Getenv("BASIC_AUTH_PASS")
		if envUser != "" && envPass != "" && username == envUser && password == envPass {
			resp.User = &models.User{ID: username, Name: username}
		}
	}

	// Add hostname info
	if hostname, err := os.Hostname(); err == nil {
		resp.Hostname = hostname
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding status response: %v", err)
	}
}
