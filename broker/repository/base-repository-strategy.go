package repository

import "hotelhub/broker/services"

// The RepositoryStrategy struct defines a strategy for repository operations.
// A structure will be returned with a selection of repositories based on the method.
// The strategy can be implemented as dependency injection for different repository types.
type RepositoryStrategy struct {
	JwtRepository    BaseJwtRepository
	PluginRepository PluginRepository
	UserRepository   BaseUserRepository
}

// NewPostgresRepositoryStrategy creates a new RepositoryStrategy using Postgres repositories.
// Information will be stored and retrieved from a Postgres database.
func NewPostgresRepositoryStrategy(postgres *services.Postgres) *RepositoryStrategy {
	return &RepositoryStrategy{
		UserRepository: NewPostgresUserRepository(postgres),
		JwtRepository:  NewPostgresJwtRepository(postgres),
	}
}
