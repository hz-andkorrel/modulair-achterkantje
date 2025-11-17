package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// PluginRegistration represents the registration payload sent to the broker
type PluginRegistration struct {
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Category     string   `json:"category,omitempty"`
	Host         string   `json:"host"`
	BaseAPIRoute string   `json:"base-api-route"`
	SettingsRoute string  `json:"settings-route,omitempty"`
	APIRoutes    []string `json:"api-routes,omitempty"`
	Enabled      bool     `json:"enabled"`
}

// RegisterWithBroker registers this service with the broker on startup
func RegisterWithBroker() error {
	brokerURL := os.Getenv("BROKER_URL")
	if brokerURL == "" {
		brokerURL = "http://localhost:8081" // Default broker URL
	}

	brokerAuthToken := os.Getenv("BROKER_AUTH_TOKEN")
	if brokerAuthToken == "" {
		log.Println("Warning: BROKER_AUTH_TOKEN not set, registration may fail")
	}

	serviceHost := os.Getenv("HOST")
	if serviceHost == "" {
		serviceHost = "localhost"
	}
	servicePort := os.Getenv("PORT")
	if servicePort == "" {
		servicePort = "8080"
	}

	registration := PluginRegistration{
		Description:   "Hotel Internal API - Gateway for user portal and admin services",
		Version:       "2.0.0",
		Slug:          "internal-api",
		Name:          "Hotel Internal API",
		Category:      "gateway",
		Host:          fmt.Sprintf("http://%s:%s", serviceHost, servicePort),
		BaseAPIRoute:  "/api/v1",
		SettingsRoute: "/admin/system/stats",
		APIRoutes: []string{
			"/api/v1/albums",
			"/api/auth/login",
			"/api/auth/logout",
			"/admin/users",
			"/admin/roles",
			"/health",
		},
		Enabled: true,
	}

	payload, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	req, err := http.NewRequest("POST", brokerURL+"/api/v1/route", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if brokerAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+brokerAuthToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("registration failed with status %d: %v", resp.StatusCode, errResp)
	}

	log.Printf("Successfully registered with broker at %s", brokerURL)
	return nil
}

// Example usage: Add this to your InternalAPI main.go
func main() {
	// Register with broker on startup
	if err := RegisterWithBroker(); err != nil {
		log.Printf("Failed to register with broker: %v", err)
		// Don't fail startup if broker registration fails - service should still work standalone
	}

	// Continue with normal InternalAPI startup...
	fmt.Println("InternalAPI service starting...")
}
