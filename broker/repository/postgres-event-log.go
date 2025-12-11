package repository

import (
	"context"
	"fmt"
	"hotelhub/broker/domain"
	"hotelhub/broker/services"
	"log"
)

type PostgresEventLogRepository struct {
	database *services.Postgres
}

func NewPostgresEventLogRepository(database *services.Postgres) EventLogRepository {
	return &PostgresEventLogRepository{
		database: database,
	}
}

// Add onderdeel
func (repo *PostgresEventLogRepository) Add(ctx context.Context, log *domain.EventLog) error {
	query := "INSERT INTO event_logs (channel, action, user_email, payload, plugin_slug) VALUES ($1, $2, $3, $4, $5)"

	affectedRows := repo.database.Execute(query, log.Channel, log.Action, log.UserEmail, log.Payload, log.PluginSlug)
	if affectedRows == 0 {
		log := fmt.Sprintf("failed to insert event log for channel: %s", log.Channel)
		return fmt.Errorf(log)
	}

	return nil
}

// Pak recent fucntie
func (repo *PostgresEventLogRepository) GetRecent(ctx context.Context, limit int) ([]*domain.EventLog, error) {
	query := "SELECT id, channel, action, user_email, payload, plugin_slug, created_at FROM event_logs ORDER BY created_at DESC LIMIT 1"

	row := repo.database.QueryRow(query)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve recent event log")
		return []*domain.EventLog{}, nil
	}

	eventLog := &domain.EventLog{}
	err := row.Scan(&eventLog.ID, &eventLog.Channel, &eventLog.Action, &eventLog.UserEmail, &eventLog.Payload, &eventLog.PluginSlug, &eventLog.CreatedAt)
	if err != nil {
		log.Println("[Postgres] Error scanning event log row:", err)
		return []*domain.EventLog{}, nil
	}

	return []*domain.EventLog{eventLog}, nil
}

// Count functie
func (repo *PostgresEventLogRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM event_logs"

	row := repo.database.QueryRow(query)
	if row == nil {
		log.Println("[Postgres] Cannot retrieve event log count")
		return 0, fmt.Errorf("failed to retrieve count")
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Println("[Postgres] Error scanning count:", err)
		return 0, err
	}

	return count, nil
}
