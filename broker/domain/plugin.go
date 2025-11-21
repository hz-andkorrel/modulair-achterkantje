package domain

// The Plugin struct represents a plugin with its metadata and configuration.
// The struct fields are annotated with JSON tags for serialization and deserialization.
type Plugin struct {
	Description   string   `json:"description"`
	Version       string   `json:"version"`
	Slug          string   `json:"slug"`
	Name          string   `json:"name"`
	Category      string   `json:"category,omitempty"`
	Host          string   `json:"host"`
	BaseApiRoute  string   `json:"base-api-route"`
	SettingsRoute string   `json:"settings-route"`
	APIRoutes     []string `json:"api-routes"`
	Enabled       bool     `json:"enabled"`
}
