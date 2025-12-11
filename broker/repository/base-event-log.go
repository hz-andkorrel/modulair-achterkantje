package repository

import (
	"context"

	"hotelhub/broker/domain"
)

type EventLogRepository interface {
	Add(ctx context.Context, log *domain.EventLog) error

	GetRecent(ctx context.Context, limit int) ([]*domain.EventLog, error)

	Count(ctx context.Context) (int, error)
}
