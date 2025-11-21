package respository

import (
	"time"
)

// JwtRepository defines methods for storing and validating JWT tokens
type JwtRepository interface {
	SaveToken(token, subject string, issuedAt, expiresAt time.Time) error
	IsTokenValid(token string) (bool, error)
}
