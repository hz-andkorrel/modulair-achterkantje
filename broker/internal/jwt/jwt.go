package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Manager wraps RSA keys and token generation/validation
type Manager struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
	ttl     time.Duration
	issuer  string
	db      TokenStore // Optional database for token persistence
}

// TokenStore defines the interface for token storage
type TokenStore interface {
	SaveToken(ctx context.Context, token, subject string, issuedAt, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}

func NewManager(ttl time.Duration, issuer string) (*Manager, error) {
	// generate a dev RSA keypair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	m := &Manager{private: priv, public: &priv.PublicKey, ttl: ttl, issuer: issuer}
	return m, nil
}

// SetDB sets the token store for persistence
func (m *Manager) SetDB(db TokenStore) {
	m.db = db
}

func (m *Manager) GenerateToken(subject string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.private)
	if err != nil {
		return "", time.Time{}, err
	}
	
	// Save token to database if available
	if m.db != nil {
		ctx := context.Background()
		if err := m.db.SaveToken(ctx, signed, subject, now, exp); err != nil {
			log.Printf("Warning: failed to save token to database: %v", err)
			// Don't fail token generation if database save fails
		}
	}
	
	return signed, exp, nil
}

func (m *Manager) ParseAndVerify(tokenStr string) (*jwt.RegisteredClaims, error) {
	parser := jwt.Parser{ValidMethods: []string{jwt.SigningMethodRS256.Name}}
	var claims jwt.RegisteredClaims
	_, err := parser.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		return m.public, nil
	})
	if err != nil {
		return nil, err
	}
	// check exp
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	
	// Check if token has been revoked (if database is available)
	if m.db != nil {
		ctx := context.Background()
		revoked, err := m.db.IsTokenRevoked(ctx, tokenStr)
		if err != nil {
			log.Printf("Warning: failed to check token revocation: %v", err)
			// Continue with validation even if database check fails
		} else if revoked {
			return nil, errors.New("token has been revoked")
		}
	}
	
	return &claims, nil
}

// For simple dev tooling: expose public key as PEM
func (m *Manager) PublicKeyPEM() []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(m.public)
	if err != nil {
		log.Printf("Warning: failed to marshal public key: %v", err)
		return nil
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})
	return pemBytes
}

// AuthMiddleware helper for cmd to reuse without pulling Gin middleware package here
func AuthMiddleware(r *http.Request, m *Manager) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("missing authorization header")
	}
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return "", errors.New("invalid authorization header")
	}
	tok := header[len(prefix):]
	claims, err := m.ParseAndVerify(tok)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
