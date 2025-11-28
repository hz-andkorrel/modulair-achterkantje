package repository

import (
	"hotelhub/broker/services"
	"log"
	"time"
)

// PostgresJwtRepository is a postgres implementation of the BaseJwtRepository interface.
// It requires an instance of the Postgres service to interact with the database.
type PostgresJwtRepository struct {
	database *services.Postgres
}

// NewPostgresJwtRepository creates a new instance of PostgresJwtRepository.
// The provided database instance is used for database operations.
func NewPostgresJwtRepository(database *services.Postgres) BaseJwtRepository {
	return &PostgresJwtRepository{
		database: database,
	}
}

// Add stores a JWT token with its subject and expiration time in the Postgres database.
// During storing, the issued_at field is set to the current time.
// It returns true if the token was successfully added; otherwise, it returns false.
func (repo *PostgresJwtRepository) Add(token, subject string, expiresAt time.Time) bool {
	query := "INSERT INTO tokens (token, subject, expires_at) VALUES ($1, $2, $3)"

	affectedRows := repo.database.Execute(query, token, subject, expiresAt)
	if affectedRows == 0 {
		log.Println("[Postgres] Cannot add token:", token)
		return false
	}

	return true
}

// IsValid checks if a token is valid by checking its existence, expiration and revoke state.
// It returns true if the token is valid; otherwise, it returns false.
func (repo *PostgresJwtRepository) IsValid(token string) bool {
	var expiresAt time.Time
	var revoked bool

	query := "SELECT expires_at, revoked FROM tokens WHERE token = $1"
	row := repo.database.QueryRow(query, token)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve token:", token)
		return false
	}

	err := row.Scan(&expiresAt, &revoked)
	if err != nil {
		log.Println("[Postgres] Error scanning token row:", err)
		return false
	}

	if revoked || time.Now().After(expiresAt) {
		return false
	}

	return true
}

// Delete sets the revoked field of a token to true in the Postgres database.
// It returns true if the token was successfully revoked; otherwise, it returns false.
func (repo *PostgresJwtRepository) Delete(token string) bool {
	query := "UPDATE tokens SET revoked = TRUE WHERE token = $1"

	affectedRows := repo.database.Execute(query, token)
	if affectedRows == 0 {
		log.Println("[Postgres] Cannot revoke token:", token)
		return false
	}

	return true
}
