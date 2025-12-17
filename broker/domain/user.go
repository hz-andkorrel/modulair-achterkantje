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
// If the password of the user has not been set yet, this will always return false.
func (user *User) ValidatePassword(password string) bool {
	if user.PasswordHash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

// Check whether the user has a specific role.
// Valid options are "", "user" and "admin".
// "admin" role includes "user" privileges.
func (user *User) HasRole(role string) bool {
	if role == "" {
		return true
	}

	if role == "user" {
		return user.Role == "user" || user.Role == "admin"
	}

	if role == "admin" {
		return user.Role == "admin"
	}

	return false
}
