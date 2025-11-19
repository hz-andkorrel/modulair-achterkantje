package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/plugins"
)

// PluginHealth represents the health status of a plugin
type PluginHealth struct {
	Slug         string        `json:"slug"`
	Name         string        `json:"name"`
	Host         string        `json:"host"`
	Status       string        `json:"status"` // "healthy", "unhealthy", "disabled"
	ResponseTime time.Duration `json:"response_time_ms,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// HealthCheckResponse represents the overall health check response
type HealthCheckResponse struct {
	Status  string          `json:"status"` // "healthy", "degraded", "unhealthy"
	Checked int             `json:"checked"`
	Healthy int             `json:"healthy"`
	Plugins []PluginHealth  `json:"plugins"`
}

// CheckPluginsHealth checks the health of all registered plugins
func CheckPluginsHealth(c *gin.Context) {
	allPlugins := plugins.Global.List()
	
	if len(allPlugins) == 0 {
		c.JSON(http.StatusOK, HealthCheckResponse{
			Status:  "healthy",
			Checked: 0,
			Healthy: 0,
			Plugins: []PluginHealth{},
		})
		return
	}

	// Check all plugins concurrently
	var wg sync.WaitGroup
	results := make([]PluginHealth, len(allPlugins))
	
	for i, plugin := range allPlugins {
		wg.Add(1)
		go func(idx int, p interface{}) {
			defer wg.Done()
			results[idx] = checkPluginHealth(p)
		}(i, plugin)
	}
	
	wg.Wait()

	// Calculate overall status
	healthyCount := 0
	for _, result := range results {
		if result.Status == "healthy" {
			healthyCount++
		}
	}

	overallStatus := "healthy"
	if healthyCount == 0 {
		overallStatus = "unhealthy"
	} else if healthyCount < len(results) {
		overallStatus = "degraded"
	}

	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:  overallStatus,
		Checked: len(results),
		Healthy: healthyCount,
		Plugins: results,
	})
}

// checkPluginHealth checks if a single plugin is healthy
func checkPluginHealth(pluginInterface interface{}) PluginHealth {
	// Type assertion to get the plugin
	plugin, ok := pluginInterface.(*struct {
		Description   string
		Version       string
		Slug          string
		Name          string
		Category      string
		Host          string
		BaseAPIRoute  string
		SettingsRoute string
		APIRoutes     []string
		Enabled       bool
	})
	
	if !ok {
		return PluginHealth{
			Slug:   "unknown",
			Name:   "unknown",
			Host:   "unknown",
			Status: "unhealthy",
			Error:  "invalid plugin type",
		}
	}

	health := PluginHealth{
		Slug: plugin.Slug,
		Name: plugin.Name,
		Host: plugin.Host,
	}

	// Check if plugin is disabled
	if !plugin.Enabled {
		health.Status = "disabled"
		return health
	}

	// Perform HTTP health check
	start := time.Now()
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try to reach the plugin's host
	healthCheckURL := plugin.Host
	if len(plugin.APIRoutes) > 0 {
		// Use first API route if available
		healthCheckURL = plugin.Host + plugin.APIRoutes[0]
	} else if plugin.BaseAPIRoute != "" {
		// Use base API route
		healthCheckURL = plugin.Host + plugin.BaseAPIRoute
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", healthCheckURL, nil)
	if err != nil {
		health.Status = "unhealthy"
		health.Error = "failed to create request: " + err.Error()
		return health
	}

	resp, err := client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health
	}
	defer resp.Body.Close()

	// Consider plugin healthy if it responds (any status code is fine)
	health.Status = "healthy"
	health.ResponseTime = duration

	return health
}
