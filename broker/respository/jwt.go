package respository

import (
	"context"
	"time"
)

// JwtRepository defines methods for storing and validating JWT tokens
type JwtRepository interface {
	SaveToken(ctx context.Context, token, subject string, issuedAt, expiresAt time.Time) error
	IsTokenValid(ctx context.Context, token string) (bool, error)
}
