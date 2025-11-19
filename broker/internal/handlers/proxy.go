package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"broker/internal/middleware"
	"broker/internal/models"
	"broker/internal/plugins"
)

// ProxyToPlugin forwards requests to registered plugins based on the base-api-route
func ProxyToPlugin(c *gin.Context) {
	requestPath := c.Request.URL.Path

	// Find matching plugin by base-api-route
	plugin := findPluginByRoute(requestPath)
	if plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no plugin registered for this route"})
		return
	}

	if !plugin.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "plugin is disabled"})
		return
	}

	// Log proxy request
	middleware.ProxyLogger(plugin.Slug, plugin.Host)

	// Parse plugin host URL
	targetURL, err := url.Parse(plugin.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid plugin host URL"})
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize the director to rewrite the path
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host

		// Rewrite path: remove base-api-route prefix and forward the rest
		// e.g., /api/v1/albums -> /api/v1/albums (keep full path for InternalAPI)
		req.URL.Path = requestPath
	}

	// Error handler for proxy failures
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[PROXY ERROR] Plugin '%s' failed: %v", plugin.Slug, err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "plugin request failed",
			"plugin":  plugin.Slug,
			"details": err.Error(),
		})
	}

	// Forward the request
	proxy.ServeHTTP(c.Writer, c.Request)
}

// findPluginByRoute finds a plugin that matches the request path
func findPluginByRoute(requestPath string) *models.Plugin {
	allPlugins := plugins.Global.List()

	// Sort by base-api-route length (longest first) to match most specific routes
	var bestMatch *models.Plugin
	longestMatch := 0

	for _, p := range allPlugins {
		if strings.HasPrefix(requestPath, p.BaseAPIRoute) {
			matchLen := len(p.BaseAPIRoute)
			if matchLen > longestMatch {
				longestMatch = matchLen
				bestMatch = p
			}
		}
	}

	return bestMatch
}
