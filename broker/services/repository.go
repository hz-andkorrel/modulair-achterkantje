package services

import "hotelhub/broker/repository"

// The RepositoryStrategy struct defines a strategy for repository operations.
// A structure will be returned with a selection of repositories based on the method.
// The strategy can be implemented as dependency injection for different repository types.
type RepositoryStrategy struct {
	JtwRepository    repository.JwtRepository
	PluginRepository repository.PluginRepository
}

// NewPostgresRepositoryStrategy creates a new RepositoryStrategy using Postgres repositories.
// Information will be stored and retrieved from a Postgres database.
func NewPostgresRepositoryStrategy() *RepositoryStrategy {
	return &RepositoryStrategy{
		PluginRepository: repository.NewPostgresPluginRepository(),
	}
}
