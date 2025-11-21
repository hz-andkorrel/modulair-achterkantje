package infrastructure

import (
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
