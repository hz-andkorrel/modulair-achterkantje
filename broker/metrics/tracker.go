package metrics

import (
	"sync/atomic"
	"time"
)

var (
	startTime time.Time

	// Request counters
	totalRequests   atomic.Uint64
	statusRequests  atomic.Uint64
	healthRequests  atomic.Uint64
	readyRequests   atomic.Uint64
	metricsRequests atomic.Uint64
)

// Init initializes the metrics tracker
func Init() {
	startTime = time.Now()
}

// IncrementTotal increments the total request counter
func IncrementTotal() {
	totalRequests.Add(1)
}

// IncrementStatus increments the status endpoint counter
func IncrementStatus() {
	statusRequests.Add(1)
}

// IncrementHealth increments the health endpoint counter
func IncrementHealth() {
	healthRequests.Add(1)
}

// IncrementReady increments the ready endpoint counter
func IncrementReady() {
	readyRequests.Add(1)
}

// IncrementMetrics increments the metrics endpoint counter
func IncrementMetrics() {
	metricsRequests.Add(1)
}

// GetTotal returns the total request count
func GetTotal() uint64 {
	return totalRequests.Load()
}

// GetStatus returns the status endpoint request count
func GetStatus() uint64 {
	return statusRequests.Load()
}

// GetHealth returns the health endpoint request count
func GetHealth() uint64 {
	return healthRequests.Load()
}

// GetReady returns the ready endpoint request count
func GetReady() uint64 {
	return readyRequests.Load()
}

// GetMetrics returns the metrics endpoint request count
func GetMetrics() uint64 {
	return metricsRequests.Load()
}

// GetUptime returns the uptime in seconds
func GetUptime() float64 {
	return time.Since(startTime).Seconds()
}

// GetStartTime returns when the service started
func GetStartTime() time.Time {
	return startTime
}
