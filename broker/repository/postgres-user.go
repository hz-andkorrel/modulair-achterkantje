package repository

import (
	"database/sql"
	"hotelhub/broker/domain"
	"hotelhub/broker/services"
	"log"
)

// PostgresUserRepository is a postgres implementation of the BaseUserRepository interface.
// It requires an instance of the Postgres service to interact with the database.
type PostgresUserRepository struct {
	database *services.Postgres
}

// NewPostgresUserRepository creates a new instance of PostgresUserRepository.
// The provided database instance is used for database operations.
func NewPostgresUserRepository(database *services.Postgres) BaseUserRepository {
	return &PostgresUserRepository{
		database: database,
	}
}

// Creating a new user in the Postgres database using an instance of domain.User.
// The user's details are inserted into the 'users' table.
// If the operation is successful, it returns the created user; otherwise, it returns nil.
// Errors are logged for debugging purposes.
func (repo *PostgresUserRepository) Create(user *domain.User) *domain.User {
	query := "INSERT INTO users (email, name, role) VALUES ($1, $2, $3)"

	err := repo.database.Execute(query, user.Email, user.Name, user.Role)
	if err == 0 {
		log.Println("[Postgres] Error creating user:", user.Email)
		return nil
	}

	return user
}

// Update modifies the password hash of an existing user in the Postgres database.
// It identifies the user by their username (email) and updates the 'password_hash' field.
// If the operation is successful, it returns the updated user; otherwise, it returns nil.
// Errors are logged for debugging purposes.
func (repo *PostgresUserRepository) Update(username string, passwordHash string) *domain.User {
	query := "UPDATE users SET password_hash = $1 WHERE email = $2"

	err := repo.database.Execute(query, passwordHash, username)
	if err == 0 {
		log.Println("[Postgres] Error updating user password:", username)
		return nil
	}

	return repo.Get(username)
}

// Get retrieves a user by their username or email from the Postgres database.
// This fields corresponds to the 'id' and 'email' column in the 'users' table respectively.
// If the user is found, it returns a pointer to a domain.User struct; otherwise, it returns nil.
// Errors are logged for debugging purposes.
func (repo *PostgresUserRepository) Get(username string) *domain.User {
	var user domain.User
	var passwordHash sql.NullString

	query := "SELECT email, name, password_hash, role, enabled FROM users WHERE email = $1"

	row := repo.database.QueryRow(query, username)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve user:", username)
		return nil
	}

	err := row.Scan(&user.Email, &user.Name, &passwordHash, &user.Role, &user.Enabled)
	if err != nil {
		log.Println("[Postgres] Error scanning row:", err)
		return nil
	}

	user.PasswordHash = passwordHash.String
	return &user
}
