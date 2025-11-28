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
// If an error occurs, it logs the error and returns an empty row.
func (postgres *Postgres) QueryRow(query string, args ...any) *sql.Row {
	db, err := sql.Open("postgres", postgres.connectionString)
	if err != nil {
		log.Println("[Postgres] cannot connect to database:", err)
		return &sql.Row{}
	}

	defer db.Close()
	row := db.QueryRow(query, args...)
	return row
}

// Execute executes a query and returns the number of affected rows.
// If an error occurs during the connection or query execution, it logs the error and returns zero.
func (postgres *Postgres) Execute(query string, args ...any) int {
	db, err := sql.Open("postgres", postgres.connectionString)
	if err != nil {
		log.Println("[Postgres] cannot connect to database:", err)
		return 0
	}

	defer db.Close()
	result, err := db.Exec(query, args...)
	if err != nil {
		log.Println("[Postgres] cannot execute query:", err)
		return 0
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("[Postgres] cannot retrieve affected rows:", err)
		return 0
	}

	return int(rowsAffected)
}
