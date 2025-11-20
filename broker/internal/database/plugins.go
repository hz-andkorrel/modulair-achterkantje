package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"broker/internal/models"
)

// SavePlugin inserts a new plugin into the database
func (db *DB) SavePlugin(ctx context.Context, p *models.Plugin) error {
	apiRoutesJSON, err := json.Marshal(p.APIRoutes)
	if err != nil {
		return fmt.Errorf("failed to marshal api_routes: %w", err)
	}

	query := `
		INSERT INTO plugins (slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = db.conn.ExecContext(ctx, query,
		p.Slug,
		p.Name,
		p.Description,
		p.Version,
		p.Category,
		p.Host,
		p.BaseAPIRoute,
		p.SettingsRoute,
		apiRoutesJSON,
		p.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to save plugin: %w", err)
	}

	return nil
}

// GetPlugin retrieves a plugin by slug
func (db *DB) GetPlugin(ctx context.Context, slug string) (*models.Plugin, error) {
	query := `
		SELECT slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled, created_at, updated_at
		FROM plugins
		WHERE slug = $1
	`

	var p models.Plugin
	var apiRoutesJSON []byte
	var createdAt, updatedAt time.Time

	err := db.conn.QueryRowContext(ctx, query, slug).Scan(
		&p.Slug,
		&p.Name,
		&p.Description,
		&p.Version,
		&p.Category,
		&p.Host,
		&p.BaseAPIRoute,
		&p.SettingsRoute,
		&apiRoutesJSON,
		&p.Enabled,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Plugin not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	// Unmarshal api_routes
	if err := json.Unmarshal(apiRoutesJSON, &p.APIRoutes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal api_routes: %w", err)
	}

	return &p, nil
}

// GetAllPlugins retrieves all plugins from the database
func (db *DB) GetAllPlugins(ctx context.Context) ([]*models.Plugin, error) {
	query := `
		SELECT slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled, created_at, updated_at
		FROM plugins
		ORDER BY created_at DESC
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.Plugin
	for rows.Next() {
		var p models.Plugin
		var apiRoutesJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&p.Slug,
			&p.Name,
			&p.Description,
			&p.Version,
			&p.Category,
			&p.Host,
			&p.BaseAPIRoute,
			&p.SettingsRoute,
			&apiRoutesJSON,
			&p.Enabled,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}

		// Unmarshal api_routes
		if err := json.Unmarshal(apiRoutesJSON, &p.APIRoutes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal api_routes: %w", err)
		}

		plugins = append(plugins, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plugins: %w", err)
	}

	return plugins, nil
}

// UpdatePlugin updates an existing plugin in the database
func (db *DB) UpdatePlugin(ctx context.Context, p *models.Plugin) error {
	apiRoutesJSON, err := json.Marshal(p.APIRoutes)
	if err != nil {
		return fmt.Errorf("failed to marshal api_routes: %w", err)
	}

	query := `
		UPDATE plugins
		SET name = $2, description = $3, version = $4, category = $5, host = $6, 
		    base_api_route = $7, settings_route = $8, api_routes = $9, enabled = $10, updated_at = NOW()
		WHERE slug = $1
	`

	result, err := db.conn.ExecContext(ctx, query,
		p.Slug,
		p.Name,
		p.Description,
		p.Version,
		p.Category,
		p.Host,
		p.BaseAPIRoute,
		p.SettingsRoute,
		apiRoutesJSON,
		p.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// DeletePlugin removes a plugin from the database
func (db *DB) DeletePlugin(ctx context.Context, slug string) error {
	query := `
		DELETE FROM plugins
		WHERE slug = $1
	`

	result, err := db.conn.ExecContext(ctx, query, slug)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// GetPluginsByCategory retrieves all plugins in a specific category
func (db *DB) GetPluginsByCategory(ctx context.Context, category string) ([]*models.Plugin, error) {
	query := `
		SELECT slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled, created_at, updated_at
		FROM plugins
		WHERE category = $1
		ORDER BY name ASC
	`

	rows, err := db.conn.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugins by category: %w", err)
	}
	defer rows.Close()

	var plugins []*models.Plugin
	for rows.Next() {
		var p models.Plugin
		var apiRoutesJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&p.Slug,
			&p.Name,
			&p.Description,
			&p.Version,
			&p.Category,
			&p.Host,
			&p.BaseAPIRoute,
			&p.SettingsRoute,
			&apiRoutesJSON,
			&p.Enabled,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}

		// Unmarshal api_routes
		if err := json.Unmarshal(apiRoutesJSON, &p.APIRoutes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal api_routes: %w", err)
		}

		plugins = append(plugins, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plugins: %w", err)
	}

	return plugins, nil
}

// GetAllCategories retrieves all unique plugin categories
func (db *DB) GetAllCategories(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM plugins
		WHERE category IS NOT NULL AND category != ''
		ORDER BY category ASC
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// GetAllBaseRoutes retrieves all base_api_route values for conflict checking
func (db *DB) GetAllBaseRoutes(ctx context.Context) ([]string, error) {
	query := `
		SELECT base_api_route
		FROM plugins
		WHERE base_api_route IS NOT NULL AND base_api_route != ''
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get base routes: %w", err)
	}
	defer rows.Close()

	var routes []string
	for rows.Next() {
		var route string
		if err := rows.Scan(&route); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routes: %w", err)
	}

	return routes, nil
}
