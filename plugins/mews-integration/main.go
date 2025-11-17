package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// MewsConfig holds the Mews API configuration
type MewsConfig struct {
	ClientToken     string
	AccessToken     string
	PlatformAddress string
	ClientName      string
}

// WebhookPayload is the incoming webhook from Mews
type WebhookPayload struct {
	EnterpriseID  string  `json:"EnterpriseId"`
	IntegrationID string  `json:"IntegrationId"`
	Events        []Event `json:"Events"`
}

// Event represents a single event in the webhook
type Event struct {
	Discriminator string          `json:"Discriminator"`
	Value         json.RawMessage `json:"Value"`
}

// EntityUpdated contains the entity ID
type EntityUpdated struct {
	ID string `json:"Id"`
}

// CheckInEvent represents a guest check-in event to publish to Redis
type CheckInEvent struct {
	Source            string    `json:"source"`
	EventType         string    `json:"event_type"`
	ReservationID     string    `json:"reservation_id"`
	ReservationNumber string    `json:"reservation_number"`
	RoomID            string    `json:"room_id"`
	GuestAccountID    string    `json:"guest_account_id"`
	CheckInTime       time.Time `json:"check_in_time"`
	State             string    `json:"state"`
	Timestamp         time.Time `json:"timestamp"`
}

// GetAllReservationsRequest is the request to fetch reservation details
type GetAllReservationsRequest struct {
	ClientToken    string   `json:"ClientToken"`
	AccessToken    string   `json:"AccessToken"`
	Client         string   `json:"Client"`
	ReservationIds []string `json:"ReservationIds"`
}

// ReservationsResponse is the response from Mews API
type ReservationsResponse struct {
	Reservations []Reservation `json:"Reservations"`
}

// Reservation represents a single reservation from Mews
type Reservation struct {
	ID                 string     `json:"Id"`
	Number             string     `json:"Number"`
	State              string     `json:"State"`
	StartUtc           time.Time  `json:"StartUtc"`
	EndUtc             time.Time  `json:"EndUtc"`
	AssignedResourceID string     `json:"AssignedResourceId"`
	AccountID          string     `json:"AccountId"`
	ActualStartUtc     *time.Time `json:"ActualStartUtc"`
}

var (
	redisClient *redis.Client
	mewsConfig  MewsConfig
	ctx         context.Context
)

func main() {
	ctx = context.Background()

	// Load configuration
	mewsConfig = MewsConfig{
		ClientToken:     os.Getenv("MEWS_CLIENT_TOKEN"),
		AccessToken:     os.Getenv("MEWS_ACCESS_TOKEN"),
		PlatformAddress: getEnvOrDefault("MEWS_PLATFORM_ADDRESS", "https://api.mews-demo.com"),
		ClientName:      getEnvOrDefault("MEWS_CLIENT_NAME", "HotelIntegration 1.0.0"),
	}

	if mewsConfig.ClientToken == "" || mewsConfig.AccessToken == "" {
		log.Fatal("❌ MEWS_CLIENT_TOKEN and MEWS_ACCESS_TOKEN must be set")
	}

	// Connect to Redis
	redisAddr := getEnvOrDefault("REDIS_ADDR", "hub_bus:6379")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis at %s: %v", redisAddr, err)
	}

	log.Println("🏨 Mews Webhook Integration Plugin Started")
	log.Printf("📡 Connected to Redis at %s", redisAddr)
	log.Printf("🔗 Mews API: %s", mewsConfig.PlatformAddress)

	// Setup HTTP server for webhooks
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/health", handleHealth)

	port := getEnvOrDefault("WEBHOOK_PORT", "8080")
	log.Printf("🌐 Webhook server listening on port %s", port)
	log.Println("📩 Waiting for Mews webhooks...")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Failed to start webhook server: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the webhook payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("❌ Error parsing webhook JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("📨 Received webhook with %d event(s) from Enterprise: %s", len(payload.Events), payload.EnterpriseID)

	// Process events asynchronously
	go processEvents(payload.Events)

	// Respond immediately to Mews (required for webhook delivery)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func processEvents(events []Event) {
	reservationIDs := []string{}

	// Collect reservation IDs from ServiceOrderUpdated events
	for _, event := range events {
		if event.Discriminator == "ServiceOrderUpdated" {
			var entity EntityUpdated
			if err := json.Unmarshal(event.Value, &entity); err != nil {
				log.Printf("❌ Error parsing event value: %v", err)
				continue
			}
			reservationIDs = append(reservationIDs, entity.ID)
		}
	}

	if len(reservationIDs) == 0 {
		log.Println("📭 No ServiceOrderUpdated events in webhook")
		return
	}

	log.Printf("🔍 Fetching details for %d reservation(s)", len(reservationIDs))

	// Fetch reservation details from Mews API
	reservations, err := getReservationsByIds(reservationIDs)
	if err != nil {
		log.Printf("❌ Error fetching reservations: %v", err)
		return
	}

	// Publish check-in events for reservations that are checked in
	for _, reservation := range reservations {
		if reservation.State == "Started" {
			if err := publishCheckInEvent(ctx, reservation); err != nil {
				log.Printf("❌ Error publishing event for reservation %s: %v", reservation.Number, err)
			} else {
				log.Printf("✅ Published check-in event for reservation %s (Room: %s)",
					reservation.Number, reservation.AssignedResourceID)
			}
		}
	}
}

func getReservationsByIds(reservationIDs []string) ([]Reservation, error) {
	req := GetAllReservationsRequest{
		ClientToken:    mewsConfig.ClientToken,
		AccessToken:    mewsConfig.AccessToken,
		Client:         mewsConfig.ClientName,
		ReservationIds: reservationIDs,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	url := fmt.Sprintf("%s/api/connector/v1/reservations/getAll", mewsConfig.PlatformAddress)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var mewsResp ReservationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&mewsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return mewsResp.Reservations, nil
}

func publishCheckInEvent(ctx context.Context, reservation Reservation) error {
	event := CheckInEvent{
		Source:            "mews",
		EventType:         "guest.checked_in",
		ReservationID:     reservation.ID,
		ReservationNumber: reservation.Number,
		RoomID:            reservation.AssignedResourceID,
		GuestAccountID:    reservation.AccountID,
		CheckInTime:       reservation.StartUtc,
		State:             reservation.State,
		Timestamp:         time.Now(),
	}

	if reservation.ActualStartUtc != nil {
		event.CheckInTime = *reservation.ActualStartUtc
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to general hotel events channel
	if err := redisClient.Publish(ctx, "hotel.events", string(eventJSON)).Err(); err != nil {
		return fmt.Errorf("failed to publish to hotel.events: %w", err)
	}

	// Also publish to specific check-in channel
	if err := redisClient.Publish(ctx, "hotel.events.guest.checked_in", string(eventJSON)).Err(); err != nil {
		return fmt.Errorf("failed to publish to hotel.events.guest.checked_in: %w", err)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
