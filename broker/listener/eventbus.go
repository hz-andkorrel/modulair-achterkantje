package listener

import (
	"context"
	"encoding/json"
	"log"

	"hotelhub/broker/domain"
	"hotelhub/broker/repository"

	"github.com/redis/go-redis/v9"
)

type EventBusListener struct {
	redis      *redis.Client
	repository *repository.RepositoryStrategy
}

func NewEventBusListener(repository *repository.RepositoryStrategy) *EventBusListener {
	client := redis.NewClient(&redis.Options{
		Addr: "hotelhub-bus:6379",
		DB:   0,
	})

	return &EventBusListener{
		redis:      client,
		repository: repository,
	}
}

func (listener *EventBusListener) Start() {
	ctx := context.Background()

	// Subscribe to both eventbus channels
	subscription := listener.redis.Subscribe(ctx, "events", "hotel.events")
	channel := subscription.Channel()

	log.Println("[EventBusListener] 🎧 Listening on 'events' and 'hotel.events' channels...")

	// Loop through all incoming messages
	for msg := range channel {
		listener.handleMessage(ctx, msg)
	}
}

// handleMessage processes a single Redis message and logs it to the database.
func (listener *EventBusListener) handleMessage(ctx context.Context, msg *redis.Message) {
	// Parse JSON payload
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(msg.Payload), &parsed)
	if err != nil {
		log.Printf("[EventBusListener] ⚠️  Failed to parse JSON from [%s]: %v\n", msg.Channel, err)
		log.Printf("[EventBusListener]     Raw payload: %s\n", msg.Payload)
		// Continue logging even with parse error - store raw payload
	}

	// Extract fields from JSON (use empty string if missing)
	action, _ := parsed["action"].(string)
	userEmail, _ := parsed["user_email"].(string)
	pluginSlug, _ := parsed["plugin_slug"].(string)

	// Create EventLog entry
	eventLog := &domain.EventLog{
		Channel:    msg.Channel,
		Action:     action,
		UserEmail:  userEmail,
		Payload:    msg.Payload,
		PluginSlug: pluginSlug,
	}

	// Save to database
	err = listener.repository.EventLogRepository.Add(ctx, eventLog)
	if err != nil {
		log.Printf("[EventBusListener] ❌ Failed to save event to database: %v\n", err)
		log.Printf("[EventBusListener]     Channel: %s, Action: %s\n", msg.Channel, action)
	} else {
		// Log success with action (or 'unknown' if empty)
		displayAction := action
		if displayAction == "" {
			displayAction = "(no action)"
		}
		log.Printf("[EventBusListener] ✅ Logged: [%s] %s\n", msg.Channel, displayAction)
	}
}
