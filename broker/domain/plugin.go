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

	// ContainerID stores the identifier of the container instance running this plugin.
	// It is set when the plugin container is created/started and is not exposed via JSON.
	ContainerID string `json:"-"`
	// ImageName stores the name (and optionally tag) of the container image used for this plugin.
	// It should be populated when the plugin image is selected or pulled, and is internal-only.
	ImageName   string `json:"-"`
}
