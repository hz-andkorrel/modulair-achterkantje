package services

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Configuration is a service for interacting with environment variables stored in the .env file.
// The first section of variables are general application settings.
// The second section of variables are related to JWT authentication.
type Configuration struct {
	AppVersion      string
	ServerPort      string
	DatabasePath    string
	PluginDirectory string

	JwtIssuer string
	JwtExpiry time.Duration
}

// To create an instance of the Configuration service, the .env file is loaded.
// The variables are read and assigned to the corresponding field.
// If the .env file cannot be loaded, the application will log a fatal error and terminate.
// Helper functions are used to retrieve the variables in the appropriate format.
// Fatal errors are logged if any variable is missing or cannot be parsed correctly.
// This ensures that the application is configured correctly, preventing unexpected/unintended behavior.
func NewConfiguration() *Configuration {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("[Configuration] Error loading environment variables from .env file:", err)
	}

	return &Configuration{
		AppVersion:      getEnvString("APP_VERSION"),
		ServerPort:      getEnvString("BROKER_PORT"),
		DatabasePath:    getEnvString("DATABASE_PATH"),
		PluginDirectory: getEnvString("PLUGIN_DIRECTORY"),

		JwtIssuer: getEnvString("JWT_ISSUER"),
		JwtExpiry: getEnvDuration("JWT_EXPIRY"),
	}
}

// Helper function to retrieve environment variables as strings.
// If the variable is not found, a fatal error is logged and the application terminates.
func getEnvString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalln("[Configuration] Error loading", key, "variable")
	}

	return value
}

// Helper function to retrieve environment variables as time.Duration.
// The getEnvString function is used to get the variable as a string, which is then parsed into a time.Duration.
// If the variable is not found, a fatal error is logged and the application terminates.
// If the variable cannot be parsed as a duration, a fatal error is logged and the application terminates.
func getEnvDuration(key string) time.Duration {
	value := getEnvString(key)
	duration, err := time.ParseDuration(value)

	if err != nil {
		log.Fatalln("[Configuration] Error parsing", key, "variable to time object:", err)
	}

	return duration
}
