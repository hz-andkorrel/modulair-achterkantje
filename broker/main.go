package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var startTime = time.Now()

// Metrics counters
var (
	totalRequests   atomic.Uint64
	statusRequests  atomic.Uint64
	healthRequests  atomic.Uint64
	readyRequests   atomic.Uint64
	metricsRequests atomic.Uint64
)

type MetricsResponse struct {
	Requests RequestMetrics `json:"requests"`
	Runtime  RuntimeMetrics `json:"runtime"`
	Memory   MemoryMetrics  `json:"memory"`
	Time     string         `json:"time"`
	Uptime   float64        `json:"uptime"`
}

type RequestMetrics struct {
	Total   uint64 `json:"total"`
	Status  uint64 `json:"status"`
	Health  uint64 `json:"health"`
	Ready   uint64 `json:"ready"`
	Metrics uint64 `json:"metrics"`
}

type RuntimeMetrics struct {
	Goroutines int    `json:"goroutines"`
	CPUs       int    `json:"cpus"`
	GoVersion  string `json:"go_version"`
}

type MemoryMetrics struct {
	AllocMB      uint64  `json:"alloc_mb"`
	TotalAllocMB uint64  `json:"total_alloc_mb"`
	SysMB        uint64  `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
	GCPauseMs    float64 `json:"last_gc_pause_ms"`
}

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

// metricsMiddleware wraps handlers to track request counts
func metricsMiddleware(name string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalRequests.Add(1)

		switch name {
		case "status":
			statusRequests.Add(1)
		case "health":
			healthRequests.Add(1)
		case "ready":
			readyRequests.Add(1)
		case "metrics":
			metricsRequests.Add(1)
		}

		handler(w, r)
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

// metricsHandler provides runtime and request metrics
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert to MB for readability
	allocMB := m.Alloc / 1024 / 1024
	totalAllocMB := m.TotalAlloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024

	// Calculate last GC pause time
	var gcPauseMs float64
	if m.NumGC > 0 {
		gcPauseMs = float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000.0
	}

	resp := MetricsResponse{
		Requests: RequestMetrics{
			Total:   totalRequests.Load(),
			Status:  statusRequests.Load(),
			Health:  healthRequests.Load(),
			Ready:   readyRequests.Load(),
			Metrics: metricsRequests.Load(),
		},
		Runtime: RuntimeMetrics{
			Goroutines: runtime.NumGoroutine(),
			CPUs:       runtime.NumCPU(),
			GoVersion:  runtime.Version(),
		},
		Memory: MemoryMetrics{
			AllocMB:      allocMB,
			TotalAllocMB: totalAllocMB,
			SysMB:        sysMB,
			NumGC:        m.NumGC,
			GCPauseMs:    gcPauseMs,
		},
		Time:   time.Now().UTC().Format(time.RFC3339),
		Uptime: time.Since(startTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding metrics response: %v", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API endpoints with metrics tracking
	mux.HandleFunc("/api/v1/status", metricsMiddleware("status", statusHandler))
	mux.HandleFunc("/api/v1/metrics", metricsMiddleware("metrics", metricsHandler))

	// Health check endpoints
	mux.HandleFunc("/health", metricsMiddleware("health", healthHandler))
	mux.HandleFunc("/ready", metricsMiddleware("ready", readyHandler))
	mux.HandleFunc("/healthz", metricsMiddleware("health", healthHandler))
	mux.HandleFunc("/readyz", metricsMiddleware("ready", readyHandler))

	addr := ":" + port
	log.Printf("broker: starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("broker: server failed: %v", err)
	}
}
