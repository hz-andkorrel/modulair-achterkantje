package middleware

import (
	"net/http"

	"broker/metrics"
)

// MetricsMiddleware wraps handlers to track request counts
func MetricsMiddleware(name string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.IncrementTotal()

		switch name {
		case "status":
			metrics.IncrementStatus()
		case "health":
			metrics.IncrementHealth()
		case "ready":
			metrics.IncrementReady()
		case "metrics":
			metrics.IncrementMetrics()
		}

		handler(w, r)
	}
}
