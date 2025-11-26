package services

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Postgres represents a PostgreSQL database connection
type Postgres struct {
	connectionString string
}

// When creating, the connection string is retrieved from the application configuration.
func NewPostgres(configuration *Configuration) *Postgres {
	return &Postgres{
		connectionString: configuration.DatabasePath,
	}
}

// QueryRow executes a query that is expected to return at most one row.
// If an error occurs during the connection or query execution, it logs the error and returns nil.
func (postgres *Postgres) QueryRow(query string, args ...any) *sql.Row {
	db, err := sql.Open("postgres", postgres.connectionString)
	if err != nil {
		log.Println("[Postgres] cannot connect to database:", err)
		return nil
	}

	defer db.Close()
	row := db.QueryRow(query, args...)
	return row
}
