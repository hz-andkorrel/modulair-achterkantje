package domain

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// A user in the system, with logic for authentication.
type User struct {
	Id           string
	Email        string
	Name         string
	PasswordHash string
	Role         string
	Enabled      bool
}

// Compare a given password with the stored password hash.
func (user *User) ValidatePassword(password string) bool {
	log.Println("Comparison result:", bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)))
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
