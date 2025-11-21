package infrastructure

import (
	"database/sql"
	"hotelhub/broker/services"
)

// Postgres represents a PostgreSQL database connection
type Postgres struct {
	connectionString string
}

// When creating, the connection string is retrieved from the application configuration.
func NewPostgres(configuration *services.Configuration) *Postgres {
	return &Postgres{
		connectionString: configuration.DatabasePath,
	}
}

// ExecuteQuery executes a given SQL query on the PostgreSQL database.
// The result of the query is ignored.
// Errors occuring during the execution are directly returned.
func (postgres *Postgres) ExecuteQuery(query string) error {
	connection, err := sql.Open("postgres", postgres.connectionString)
	if err != nil {
		return err
	}

	defer connection.Close()
	_, err = connection.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
