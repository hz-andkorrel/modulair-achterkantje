package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"broker/metrics"
	"broker/models"
)

// Metrics handles the /api/v1/metrics endpoint
func Metrics(w http.ResponseWriter, r *http.Request) {
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

	resp := models.MetricsResponse{
		Requests: models.RequestMetrics{
			Total:   metrics.GetTotal(),
			Status:  metrics.GetStatus(),
			Health:  metrics.GetHealth(),
			Ready:   metrics.GetReady(),
			Metrics: metrics.GetMetrics(),
		},
		Runtime: models.RuntimeMetrics{
			Goroutines: runtime.NumGoroutine(),
			CPUs:       runtime.NumCPU(),
			GoVersion:  runtime.Version(),
		},
		Memory: models.MemoryMetrics{
			AllocMB:      allocMB,
			TotalAllocMB: totalAllocMB,
			SysMB:        sysMB,
			NumGC:        m.NumGC,
			GCPauseMs:    gcPauseMs,
		},
		Time:   time.Now().UTC().Format(time.RFC3339),
		Uptime: metrics.GetUptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed encoding metrics response: %v", err)
	}
}
