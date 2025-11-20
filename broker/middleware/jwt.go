package middleware

import (
	"log"
	"net/http"

	"hotelhub/broker/services"
)

// JwtMiddleware extracts and verifies the JWT token from the Authorization header of the HTTP request
// It returns the subject from the token claims if verification is successful.
// When authorization fails, it logs the reason and returns an empty string.
// Possible reasons are: missing header, invalid format, or token verification failure.
func JwtMiddleware(request *http.Request, jwtService *services.JwtService) string {
	header := request.Header.Get("Authorization")
	if header == "" {
		log.Println("[JWT] Authorization header missing")
		return ""
	}

	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		log.Println("[JWT] Invalid authorization header format")
		return ""
	}

	token := header[len(prefix):]
	claims := jwtService.ParseAndVerify(token)
	if claims == nil {
		log.Println("[JWT] Token verification failed")
		return ""
	}

	return claims.Subject
}
