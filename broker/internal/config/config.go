package config

import (
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
}

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() *Config {
	cfg := &Config{
		ServerPort: getEnv("BROKER_PORT", "8081"),
		JWTExpiry:  getDurationEnv("JWT_EXPIRY", 10*time.Minute),
		JWTIssuer:  getEnv("JWT_ISSUER", "broker-service"),
	}
	
	log.Println("Broker Configuration loaded:")
	log.Printf("  Server port: %s\n", cfg.ServerPort)
	log.Printf("  JWT expiry: %v\n", cfg.JWTExpiry)
	
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
		}
	}
	return defaultValue
}
