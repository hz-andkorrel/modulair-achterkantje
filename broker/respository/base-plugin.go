package respository

import (
	"context"

	"hotelhub/broker/domain"
)

// PluginRepository defines the interface for plugin data persistence.
type PluginRepository interface {
	SavePlugin(ctx context.Context, p *domain.Plugin) error
	GetPlugin(ctx context.Context, slug string) (*domain.Plugin, error)
	GetAllPlugins(ctx context.Context) ([]*domain.Plugin, error)
	UpdatePlugin(ctx context.Context, p *domain.Plugin) error
	DeletePlugin(ctx context.Context, slug string) error
	GetAllBaseRoutes(ctx context.Context) ([]string, error)
}
