package repository

import "hotelhub/broker/domain"

// A user can be created and retrieved from a data source.
type BaseUserRepository interface {
	Create(user *domain.User) *domain.User
	Get(username string) *domain.User
}
