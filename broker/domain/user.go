package domain

import (
	"golang.org/x/crypto/bcrypt"
)

// A user in the system, with logic for authentication.
type User struct {
	Email        string
	Name         string
	PasswordHash string
	Role         string
	Enabled      bool
}

// Compare a given password with the stored password hash.
func (user *User) ValidatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
