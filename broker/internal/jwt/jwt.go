package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
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
	return &claims, nil
}

// For simple dev tooling: expose public key as PEM
func (m *Manager) PublicKeyPEM() []byte {
	pubASN1, _ := x509.MarshalPKIXPublicKey(m.public)
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
