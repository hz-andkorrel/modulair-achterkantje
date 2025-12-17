package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"hotelhub/broker/domain"
	"hotelhub/broker/services"
	"log"
)

type PostgresPluginRepository struct {
	database *services.Postgres
}

func NewPostgresPluginRepository(database *services.Postgres) PluginRepository {
	return &PostgresPluginRepository{
		database: database,
	}
}

func (repo *PostgresPluginRepository) SavePlugin(ctx context.Context, p *domain.Plugin) error {
	apiRoutesJSON, _ := json.Marshal(p.APIRoutes)
	query := `INSERT INTO plugins (slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	affectedRows := repo.database.Execute(query, p.Slug, p.Name, p.Description, p.Version, p.Category, p.Host, p.BaseApiRoute, p.SettingsRoute, apiRoutesJSON, p.Enabled)
	if affectedRows == 0 {
		log.Println("[Postgres] Failed to save plugin:", p.Slug)
		return sql.ErrNoRows
	}

	return nil
}

func (repo *PostgresPluginRepository) GetPlugin(ctx context.Context, slug string) (*domain.Plugin, error) {
	query := `SELECT slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled 
              FROM plugins WHERE slug = $1`

	row := repo.database.QueryRow(query, slug)
	if row == nil {
		return nil, sql.ErrNoRows
	}

	var plugin domain.Plugin
	var apiRoutesJSON []byte
	err := row.Scan(&plugin.Slug, &plugin.Name, &plugin.Description, &plugin.Version, &plugin.Category, &plugin.Host, &plugin.BaseApiRoute, &plugin.SettingsRoute, &apiRoutesJSON, &plugin.Enabled)
	if err != nil {
		log.Println("[Postgres] Error scanning plugin row:", err)
		return nil, err
	}

	json.Unmarshal(apiRoutesJSON, &plugin.APIRoutes)
	return &plugin, nil
}

func (repo *PostgresPluginRepository) GetAllPlugins(ctx context.Context) ([]*domain.Plugin, error) {
	query := `SELECT slug, name, description, version, category, host, base_api_route, settings_route, api_routes, enabled 
              FROM plugins ORDER BY name`

	rows, err := repo.database.Query(query)
	if err != nil {
		log.Println("[Postgres] Cannot retrieve plugins:", err)
		return []*domain.Plugin{}, err
	}
	defer rows.Close()

	var plugins []*domain.Plugin
	for rows.Next() {
		var plugin domain.Plugin
		var apiRoutesJSON []byte
		err := rows.Scan(&plugin.Slug, &plugin.Name, &plugin.Description, &plugin.Version, &plugin.Category, &plugin.Host, &plugin.BaseApiRoute, &plugin.SettingsRoute, &apiRoutesJSON, &plugin.Enabled)
		if err != nil {
			log.Println("[Postgres] Error scanning plugin row:", err)
			continue
		}
		json.Unmarshal(apiRoutesJSON, &plugin.APIRoutes)
		plugins = append(plugins, &plugin)
	}

	return plugins, nil
}

func (repo *PostgresPluginRepository) UpdatePlugin(ctx context.Context, p *domain.Plugin) error {
	apiRoutesJSON, _ := json.Marshal(p.APIRoutes)
	query := `UPDATE plugins SET name = $2, description = $3, version = $4, category = $5, host = $6, 
              base_api_route = $7, settings_route = $8, api_routes = $9, enabled = $10 WHERE slug = $1`

	affectedRows := repo.database.Execute(query, p.Slug, p.Name, p.Description, p.Version, p.Category, p.Host, p.BaseApiRoute, p.SettingsRoute, apiRoutesJSON, p.Enabled)
	if affectedRows == 0 {
		log.Println("[Postgres] Failed to update plugin:", p.Slug)
		return sql.ErrNoRows
	}

	return nil
}

func (repo *PostgresPluginRepository) DeletePlugin(ctx context.Context, slug string) error {
	query := `DELETE FROM plugins WHERE slug = $1`

	affectedRows := repo.database.Execute(query, slug)
	if affectedRows == 0 {
		log.Println("[Postgres] Failed to delete plugin:", slug)
		return sql.ErrNoRows
	}

	return nil
}

func (repo *PostgresPluginRepository) GetAllBaseRoutes(ctx context.Context) ([]string, error) {
	query := `SELECT base_api_route FROM plugins WHERE enabled = true ORDER BY base_api_route`

	rows, err := repo.database.Query(query)
	if err != nil {
		log.Println("[Postgres] Cannot retrieve base routes:", err)
		return []string{}, err
	}
	defer rows.Close()

	var routes []string
	for rows.Next() {
		var route string
		err := rows.Scan(&route)
		if err != nil {
			log.Println("[Postgres] Error scanning route:", err)
			continue
		}
		routes = append(routes, route)
	}

	return routes, nil
}
