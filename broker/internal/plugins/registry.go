package plugins

import (
	"context"
	"errors"
	"sync"
	"time"

	"broker/internal/models"
)

// PluginStore defines the interface for plugin storage (database)
type PluginStore interface {
	SavePlugin(ctx context.Context, p *models.Plugin) error
	GetPlugin(ctx context.Context, slug string) (*models.Plugin, error)
	GetAllPlugins(ctx context.Context) ([]*models.Plugin, error)
	UpdatePlugin(ctx context.Context, p *models.Plugin) error
	DeletePlugin(ctx context.Context, slug string) error
	GetAllBaseRoutes(ctx context.Context) ([]string, error)
}

// Registry stores registered plugins in memory and syncs with database
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*models.Plugin
	db      PluginStore
}

// Global is the package-level registry instance used by handlers.
var Global = NewRegistry()

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*models.Plugin),
	}
}

// SetDB sets the database for plugin persistence and loads existing plugins
func (r *Registry) SetDB(db PluginStore) error {
	if db == nil {
		return errors.New("database is nil")
	}

	r.mu.Lock()
	r.db = db
	r.mu.Unlock()

	// Load existing plugins from database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plugins, err := db.GetAllPlugins(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.plugins = make(map[string]*models.Plugin, len(plugins))
	for _, p := range plugins {
		if p != nil && p.Slug != "" {
			copy := *p
			r.plugins[p.Slug] = &copy
		}
	}
	r.mu.Unlock()

	return nil
}

// Register adds a plugin to the registry and saves it to the database
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
	db := r.db
	r.mu.Unlock()

	// Save to database if available
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.SavePlugin(ctx, p); err != nil {
			// Rollback in-memory change on database error
			r.mu.Lock()
			delete(r.plugins, p.Slug)
			r.mu.Unlock()
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

// Update replaces an existing plugin and updates it in the database
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

	oldPlugin := r.plugins[p.Slug]
	copy := *p
	r.plugins[p.Slug] = &copy
	db := r.db
	r.mu.Unlock()

	// Update in database if available
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.UpdatePlugin(ctx, p); err != nil {
			// Rollback in-memory change on database error
			r.mu.Lock()
			r.plugins[p.Slug] = oldPlugin
			r.mu.Unlock()
			return err
		}
	}

	return nil
}

// Delete removes a plugin by slug and deletes it from the database
func (r *Registry) Delete(slug string) error {
	if slug == "" {
		return errors.New("slug is required")
	}

	r.mu.Lock()
	if _, ok := r.plugins[slug]; !ok {
		r.mu.Unlock()
		return errors.New("plugin not found")
	}

	oldPlugin := r.plugins[slug]
	delete(r.plugins, slug)
	db := r.db
	r.mu.Unlock()

	// Delete from database if available
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.DeletePlugin(ctx, slug); err != nil {
			// Rollback in-memory change on database error
			r.mu.Lock()
			r.plugins[slug] = oldPlugin
			r.mu.Unlock()
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
