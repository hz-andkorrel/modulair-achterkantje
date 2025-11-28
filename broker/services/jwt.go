package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JwtService wraps RSA keys and token generation/validation
type JwtService struct {
	signingMethod jwt.SigningMethod
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	expiryTime    time.Duration
	issuer        string
}

// NewJwtService creates a new JWT service with generated RSA keys
// These are the public and private keys used to sign and verify tokens
// The signing method is RS256, which uses RSA with SHA-256 hashing
// The expiry time and issuer are taken from the provided configuration
func NewJwtService(configuration *Configuration) *JwtService {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	return &JwtService{
		signingMethod: jwt.SigningMethodRS256,
		privateKey:    key,
		publicKey:     &key.PublicKey,
		expiryTime:    configuration.JwtExpiry,
		issuer:        configuration.JwtIssuer,
	}
}

// Create a JWT token for a given subject, which could be the user ID or username
// A claim for issuer, subject, issued at, and expiration is used for the creation of the token
// The token is signed with the RSA private key and returned as a string along with its expiration time
func (service *JwtService) GenerateToken(subject string) (string, time.Time) {
	now := time.Now()
	expireTime := now.Add(service.expiryTime)

	claims := jwt.RegisteredClaims{
		Issuer:    service.issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expireTime),
	}

	rawToken := jwt.NewWithClaims(service.signingMethod, claims)
	signedToken, _ := rawToken.SignedString(service.privateKey)
	return signedToken, expireTime
}

// Parse takes a JWT token string and parses it.
// WARNING: this function does NOT validate the token!
// If the token can be parsed into claims, the claim is returned; otherwise, nil is returned.
// Log messages are printed for parsing errors.
func (service *JwtService) Parse(tokenString string) *jwt.RegisteredClaims {
	parser := jwt.Parser{}
	var claims jwt.RegisteredClaims

	_, err := parser.ParseWithClaims(tokenString, &claims, service.retrieveKey)
	if err != nil {
		log.Println("Failed to parse JWT token:", err)
		return nil
	}

	return &claims
}

// Helper function to provide the public key for token verification
func (service *JwtService) retrieveKey(token *jwt.Token) (any, error) {
	return service.publicKey, nil
}

// For simple dev tooling: expose public key as PEM
func (service *JwtService) PublicKeyPEM() []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(service.publicKey)
	if err != nil {
		log.Printf("Warning: failed to marshal public key: %v", err)
		return nil
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})
	return pemBytes
}
