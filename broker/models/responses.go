package models

// MetricsResponse contains runtime and request metrics
type MetricsResponse struct {
	Requests RequestMetrics `json:"requests"`
	Runtime  RuntimeMetrics `json:"runtime"`
	Memory   MemoryMetrics  `json:"memory"`
	Time     string         `json:"time"`
	Uptime   float64        `json:"uptime"`
}

// RequestMetrics tracks HTTP request counts
type RequestMetrics struct {
	Total   uint64 `json:"total"`
	Status  uint64 `json:"status"`
	Health  uint64 `json:"health"`
	Ready   uint64 `json:"ready"`
	Metrics uint64 `json:"metrics"`
}

// RuntimeMetrics contains Go runtime information
type RuntimeMetrics struct {
	Goroutines int    `json:"goroutines"`
	CPUs       int    `json:"cpus"`
	GoVersion  string `json:"go_version"`
}

// MemoryMetrics contains memory usage statistics
type MemoryMetrics struct {
	AllocMB      uint64  `json:"alloc_mb"`
	TotalAllocMB uint64  `json:"total_alloc_mb"`
	SysMB        uint64  `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
	GCPauseMs    float64 `json:"last_gc_pause_ms"`
}

// StatusResponse contains overall system status
type StatusResponse struct {
	Status     string                   `json:"status"`
	Time       string                   `json:"time"`
	Uptime     float64                  `json:"uptime"`
	Version    string                   `json:"version"`
	Components map[string]ComponentInfo `json:"components,omitempty"`
	User       *User                    `json:"user,omitempty"`
	Hostname   string                   `json:"hostname,omitempty"`
}

// HealthResponse is a simple liveness check response
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ReadyResponse indicates if the service is ready to accept traffic
type ReadyResponse struct {
	Ready      bool                     `json:"ready"`
	Components map[string]ComponentInfo `json:"components"`
	Time       string                   `json:"time"`
}

// ComponentInfo contains health information for a component
type ComponentInfo struct {
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
	Latency float64 `json:"latency_ms,omitempty"`
}

// User represents an authenticated user
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
