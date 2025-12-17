package repository

import (
	"time"
)

// JwtRepository defines methods for storing and validating JWT tokens
// It can add, validate and delete tokens
type BaseJwtRepository interface {
	Add(token, subject string, expiresAt time.Time) bool
	AddResetToken(token, subject string, expiresAt time.Time) bool
	IsValid(token string, tokenType string) bool
	Delete(token string) bool
}
