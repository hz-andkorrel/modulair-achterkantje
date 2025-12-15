package repository

import "hotelhub/broker/domain"

// A user can be created and retrieved from a data source.
// It is also possible to update the password hash of a user.
type BaseUserRepository interface {
	Create(user *domain.User) *domain.User
	Get(username string) *domain.User
	Update(username string, passwordHash string) *domain.User
}
