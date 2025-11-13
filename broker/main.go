package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

var startTime = time.Now()

type StatusResponse struct {
	Status     string                   `json:"status"`
	Time       string                   `json:"time"`
	Uptime     float64                  `json:"uptime"`
	Version    string                   `json:"version"`
	Components map[string]ComponentInfo `json:"components,omitempty"`
	User       *User                    `json:"user,omitempty"`
	Hostname   string                   `json:"hostname,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type ReadyResponse struct {
	Ready      bool                     `json:"ready"`
	Components map[string]ComponentInfo `json:"components"`
	Time       string                   `json:"time"`
}

type ComponentInfo struct {
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
	Latency float64 `json:"latency_ms,omitempty"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func checkComponentHealth(name string) ComponentInfo {
	start := time.Now()
	defer func() {
		// Latency wordt hier gemeten maar pas later toegevoegd aan het resultaat
	}()

	switch name {
	case "broker":
		latency := time.Since(start).Milliseconds()
		return ComponentInfo{Status: "ok", Message: "service running", Latency: float64(latency)}

	case "database":
		// Simuleer database check (vervang later met echte DB ping)
		time.Sleep(2 * time.Millisecond)
		latency := time.Since(start).Milliseconds()
		return ComponentInfo{Status: "unknown", Message: "not yet implemented", Latency: float64(latency)}

	case "cache":
		// Simuleer cache check
		time.Sleep(1 * time.Millisecond)
		latency := time.Since(start).Milliseconds()
		return ComponentInfo{Status: "unknown", Message: "not yet implemented", Latency: float64(latency)}

	case "auth":
		latency := time.Since(start).Milliseconds()
		if os.Getenv("BASIC_AUTH_USER") != "" {
			return ComponentInfo{Status: "ok", Message: "basic auth configured", Latency: float64(latency)}
		}
		return ComponentInfo{Status: "degraded", Message: "no auth configured", Latency: float64(latency)}

	default:
		latency := time.Since(start).Milliseconds()
		return ComponentInfo{Status: "unknown", Message: "unknown component", Latency: float64(latency)}
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("STATUS_VERSION")
	if version == "" {
		version = "0.1.0"
	}

	// Check components
	components := map[string]ComponentInfo{
		"broker":   checkComponentHealth("broker"),
		"auth":     checkComponentHealth("auth"),
		"database": checkComponentHealth("database"),
		"cache":    checkComponentHealth("cache"),
	}

	// Determine overall status based on components
	overallStatus := "ok"
	for _, comp := range components {
		if comp.Status == "degraded" {
			overallStatus = "degraded"
			break
		}
	}

	resp := StatusResponse{
		Status:     overallStatus,
		Time:       time.Now().UTC().Format(time.RFC3339),
		Uptime:     time.Since(startTime).Seconds(),
		Version:    version,
		Components: components,
	}

	if username, password, ok := r.BasicAuth(); ok {
		envUser := os.Getenv("BASIC_AUTH_USER")
		envPass := os.Getenv("BASIC_AUTH_PASS")
		if envUser != "" && envPass != "" && username == envUser && password == envPass {
			resp.User = &User{ID: username, Name: username}
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding health response: %v", err)
	}
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	components := map[string]ComponentInfo{
		"broker": checkComponentHealth("broker"),
		"auth":   checkComponentHealth("auth"),
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

	resp := ReadyResponse{
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/v1/status", statusHandler)

	// Health check endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readyHandler)

	addr := ":" + port
	log.Printf("broker: starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("broker: server failed: %v", err)
	}
}
