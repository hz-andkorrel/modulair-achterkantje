package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Config holds all configuration for the broker service
type Config struct {
	// Server settings
	ServerPort string
	
	// JWT settings
	JWTExpiry time.Duration
	JWTIssuer string
    
	// Application version (can be overridden via env or build flags)
	Version string
    
	// Plugins persistence path (file). If empty, persistence is disabled.
	PluginsPersistPath string
	
	// Database settings
	DatabaseURL string
	
	// Proxy settings
	ProxyTimeout time.Duration
	
	// Rate limiting settings
	RateLimitEnabled     bool
	RateLimitMaxRequests int
	RateLimitWindow      time.Duration
}

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() *Config {
	cfg := &Config{
		ServerPort: getEnv("BROKER_PORT", "8081"),
		JWTExpiry:  getDurationEnv("JWT_EXPIRY", 10*time.Minute),
		JWTIssuer:  getEnv("JWT_ISSUER", "broker-service"),
		PluginsPersistPath: getEnv("PLUGINS_PERSIST_PATH", "data/plugins.json"),
		Version:            getEnv("APP_VERSION", "1.0.0"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		ProxyTimeout:       getDurationEnv("PROXY_TIMEOUT", 30*time.Second),
		RateLimitEnabled:   getBoolEnv("RATE_LIMIT_ENABLED", true),
		RateLimitMaxRequests: getIntEnv("RATE_LIMIT_MAX_REQUESTS", 100),
		RateLimitWindow:    getDurationEnv("RATE_LIMIT_WINDOW", 1*time.Minute),
	}
	
	log.Println("Broker Configuration loaded:")
	log.Printf("  Server port: %s\n", cfg.ServerPort)
	log.Printf("  JWT expiry: %v\n", cfg.JWTExpiry)
	log.Printf("  Plugins persist path: %s\n", cfg.PluginsPersistPath)
	log.Printf("  Version: %s\n", cfg.Version)
	log.Printf("  Proxy timeout: %v\n", cfg.ProxyTimeout)
	if cfg.DatabaseURL != "" {
		log.Printf("  Database: enabled\n")
	} else {
		log.Printf("  Database: disabled (tokens will not be persisted)\n")
	}
	if cfg.RateLimitEnabled {
		log.Printf("  Rate limiting: enabled (%d requests per %v)\n", cfg.RateLimitMaxRequests, cfg.RateLimitWindow)
	} else {
		log.Printf("  Rate limiting: disabled\n")
	}
	
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		} else {
			log.Printf("Warning: Failed to parse %s value '%s', using default: %v", key, value, defaultValue)
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
		log.Printf("Warning: Invalid boolean value for %s: '%s', using default: %v", key, value, defaultValue)
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
		log.Printf("Warning: Failed to parse %s value '%s', using default: %d", key, value, defaultValue)
	}
	return defaultValue
}
