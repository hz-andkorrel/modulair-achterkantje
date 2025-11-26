package repository

import "hotelhub/broker/domain"

type BaseUserRepository interface {
	Get(username string) *domain.User
}
