package repository

import (
	"hotelhub/broker/domain"
	"hotelhub/broker/services"
	"log"
)

type PostgresUserRepository struct {
	database *services.Postgres
}

func NewPostgresUserRepository(database *services.Postgres) BaseUserRepository {
	return &PostgresUserRepository{
		database: database,
	}
}

func (repo *PostgresUserRepository) Get(username string) *domain.User {
	var user domain.User
	query := "SELECT id, email, name, password_hash, role, enabled FROM users WHERE id = $1"

	row := repo.database.QueryRow(query, username)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve user:", username)
		return nil
	}

	err := row.Scan(&user.Id, &user.Email, &user.Name, &user.PasswordHash, &user.Role, &user.Enabled)
	if err != nil {
		log.Println("[Postgres] Error scanning row:", err)
		return nil
	}

	return &user
}
