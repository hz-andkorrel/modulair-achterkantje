package domain

import "time"

type EventLog struct {
	ID         int       `json:"id"`
	Channel    string    `json:"channel"`
	Action     string    `json:"action"`
	UserEmail  string    `json:"user_email"`
	Payload    string    `json:"payload"`
	PluginSlug string    `json:"plugin_slug"`
	CreatedAt  time.Time `json:"created_at"`
}
