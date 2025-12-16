package repository

import (
	"database/sql"
	"hotelhub/broker/services"
	"log"
)

// PostgresMetaRepository is a Postgres implementation of the BaseMetaRepository interface.
// It requires an instance of the Postgres service to interact with the database.
type PostgresMetaRepository struct {
	database *services.Postgres
}

// NewPostgresMetaRepository creates a new instance of PostgresMetaRepository.
func NewPostgresMetaRepository(database *services.Postgres) BaseMetaRepository {
	return &PostgresMetaRepository{
		database: database,
	}
}

// Get retrieves the value for the provided key from the meta table.
// Returns an empty string on error or if the key does not exist.
func (repo *PostgresMetaRepository) Get(key string) string {
	var value sql.NullString
	query := "SELECT value FROM meta WHERE key = $1"

	row := repo.database.QueryRow(query, key)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve meta key:", key)
		return ""
	}

	err := row.Scan(&value)
	if err != nil {
		log.Println("[Postgres] Error scanning meta row:", err)
		return ""
	}

	return value.String
}

// Set updates the value for the provided key if it exists, otherwise inserts a new row.
// Returns the key on success and an empty string on failure.
func (repo *PostgresMetaRepository) Set(key string, val string) string {
	updateQuery := "UPDATE meta SET value = $1 WHERE key = $2"
	rows := repo.database.Execute(updateQuery, val, key)

	if rows == 0 {
		insertQuery := "INSERT INTO meta (key, value) VALUES ($1, $2)"
		rows = repo.database.Execute(insertQuery, key, val)
		if rows == 0 {
			log.Println("[Postgres] Cannot set meta key:", key)
			return ""
		}
	}

	return key
}
