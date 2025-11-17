package plugins

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "sync"

    "broker/internal/models"
)

// Registry stores registered plugins in memory and optionally persists them to a file.
type Registry struct {
    mu         sync.RWMutex
    plugins    map[string]*models.Plugin
    persistPath string
}

// Global is the package-level registry instance used by handlers.
var Global = NewRegistry()

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
    return &Registry{
        plugins: make(map[string]*models.Plugin),
    }
}

// SetPersistPath sets a path to persist plugin registrations to. If the file exists,
// it will be loaded. The directory will be created if necessary.
func (r *Registry) SetPersistPath(path string) error {
    if path == "" {
        return errors.New("path is empty")
    }

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }

    // Attempt to load existing file
    if _, err := os.Stat(path); err == nil {
        if err := r.loadFromFile(path); err != nil {
            return err
        }
    }

    r.mu.Lock()
    r.persistPath = path
    r.mu.Unlock()
    return nil
}

// Register adds a plugin to the registry. Returns an error if the slug already exists.
// If a persist path is set, it will attempt to save the registry to disk after registering.
func (r *Registry) Register(p *models.Plugin) error {
    if p == nil {
        return errors.New("plugin is nil")
    }

    if p.Slug == "" {
        return errors.New("plugin slug is required")
    }

    r.mu.Lock()
    if _, ok := r.plugins[p.Slug]; ok {
        r.mu.Unlock()
        return errors.New("plugin with this slug already registered")
    }

    // store a copy to avoid external mutation
    copy := *p
    r.plugins[p.Slug] = &copy

    // create a snapshot to write without holding the lock during IO
    snapshot := make([]*models.Plugin, 0, len(r.plugins))
    for _, v := range r.plugins {
        snapshot = append(snapshot, v)
    }
    persistPath := r.persistPath
    r.mu.Unlock()

    if persistPath != "" {
        // attempt to save; if save fails, we do not roll back registration but return the error
        if err := savePluginsToFile(persistPath, snapshot); err != nil {
            return err
        }
    }

    return nil
}

// List returns all registered plugins.
func (r *Registry) List() []*models.Plugin {
    r.mu.RLock()
    defer r.mu.RUnlock()

    out := make([]*models.Plugin, 0, len(r.plugins))
    for _, p := range r.plugins {
        out = append(out, p)
    }
    return out
}

// Get returns a plugin by slug, or nil if not found.
func (r *Registry) Get(slug string) *models.Plugin {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.plugins[slug]
}

// Update replaces an existing plugin. Returns an error if the plugin doesn't exist.
func (r *Registry) Update(p *models.Plugin) error {
    if p == nil {
        return errors.New("plugin is nil")
    }
    if p.Slug == "" {
        return errors.New("plugin slug is required")
    }

    r.mu.Lock()
    if _, ok := r.plugins[p.Slug]; !ok {
        r.mu.Unlock()
        return errors.New("plugin not found")
    }

    copy := *p
    r.plugins[p.Slug] = &copy

    snapshot := make([]*models.Plugin, 0, len(r.plugins))
    for _, v := range r.plugins {
        snapshot = append(snapshot, v)
    }
    persistPath := r.persistPath
    r.mu.Unlock()

    if persistPath != "" {
        if err := savePluginsToFile(persistPath, snapshot); err != nil {
            return err
        }
    }
    return nil
}

// Delete removes a plugin by slug. Returns an error if the plugin doesn't exist.
func (r *Registry) Delete(slug string) error {
    if slug == "" {
        return errors.New("slug is required")
    }

    r.mu.Lock()
    if _, ok := r.plugins[slug]; !ok {
        r.mu.Unlock()
        return errors.New("plugin not found")
    }

    delete(r.plugins, slug)

    snapshot := make([]*models.Plugin, 0, len(r.plugins))
    for _, v := range r.plugins {
        snapshot = append(snapshot, v)
    }
    persistPath := r.persistPath
    r.mu.Unlock()

    if persistPath != "" {
        if err := savePluginsToFile(persistPath, snapshot); err != nil {
            return err
        }
    }
    return nil
}

// GetAllBaseRoutes returns a list of all base-api-route values for conflict checking.
func (r *Registry) GetAllBaseRoutes() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    routes := make([]string, 0, len(r.plugins))
    for _, p := range r.plugins {
        if p.BaseAPIRoute != "" {
            routes = append(routes, p.BaseAPIRoute)
        }
    }
    return routes
}

// loadFromFile replaces the registry contents with the plugins loaded from the file.
func (r *Registry) loadFromFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    var arr []*models.Plugin
    if err := json.Unmarshal(data, &arr); err != nil {
        return err
    }

    r.mu.Lock()
    defer r.mu.Unlock()
    r.plugins = make(map[string]*models.Plugin, len(arr))
    for _, p := range arr {
        if p != nil && p.Slug != "" {
            copy := *p
            r.plugins[p.Slug] = &copy
        }
    }
    return nil
}

// savePluginsToFile writes the plugin slice to the given path as JSON.
func savePluginsToFile(path string, arr []*models.Plugin) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }

    tmpFile, err := os.CreateTemp(dir, "plugins-*.tmp")
    if err != nil {
        return err
    }
    tmpName := tmpFile.Name()

    // Ensure temp file is removed on any early return
    defer func() {
        tmpFile.Close()
        _ = os.Remove(tmpName)
    }()

    enc := json.NewEncoder(tmpFile)
    enc.SetIndent("", "  ")
    if err := enc.Encode(arr); err != nil {
        return err
    }

    if err := tmpFile.Sync(); err != nil {
        return err
    }
    if err := tmpFile.Close(); err != nil {
        return err
    }

    // Atomic rename
    if err := os.Rename(tmpName, path); err != nil {
        return err
    }

    // Best-effort: sync dir to reduce risk of rename not being persisted
    dirFile, err := os.Open(dir)
    if err == nil {
        _ = dirFile.Sync()
        _ = dirFile.Close()
    }

    return nil
}
