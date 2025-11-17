package models

// Plugin represents a plugin registration payload sent by plugins
type Plugin struct {
    Description   string   `json:"description"`
    Version       string   `json:"version"`
    Slug          string   `json:"slug"`
    Name          string   `json:"name"`
    BaseAPIRoute  string   `json:"base-api-route"`
    SettingsRoute string   `json:"settings-route"`
    APIRoutes     []string `json:"api-routes"`
    Enabled       bool     `json:"enabled"`
}
