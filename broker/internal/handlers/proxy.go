package handlers

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"broker/internal/config"
	"broker/internal/middleware"
	"broker/internal/models"
	"broker/internal/plugins"
)

var proxyConfig *config.Config

// SetProxyConfig sets the configuration for the proxy handlers
func SetProxyConfig(cfg *config.Config) {
	proxyConfig = cfg
}

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

	// Create custom transport with timeout
	timeout := 30 * time.Second
	if proxyConfig != nil {
		timeout = proxyConfig.ProxyTimeout
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = transport

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
		log.Printf("[PROXY ERROR] Plugin '%s' at %s failed: %v", plugin.Slug, plugin.Host, err)
		
		// Determine error type and provide appropriate response
		var statusCode int
		var errorType string
		
		if err == context.DeadlineExceeded {
			statusCode = http.StatusGatewayTimeout
			errorType = "timeout"
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			statusCode = http.StatusGatewayTimeout
			errorType = "timeout"
		} else if strings.Contains(err.Error(), "connection refused") {
			statusCode = http.StatusServiceUnavailable
			errorType = "connection_refused"
		} else if strings.Contains(err.Error(), "no such host") {
			statusCode = http.StatusBadGateway
			errorType = "dns_error"
		} else {
			statusCode = http.StatusBadGateway
			errorType = "proxy_error"
		}
		
		c.JSON(statusCode, gin.H{
			"error":      errorType,
			"message":    "plugin request failed",
			"plugin":     plugin.Slug,
			"plugin_host": plugin.Host,
			"details":    err.Error(),
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
