package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps database operations for token storage
type DB struct {
	conn *sql.DB
}

// Token represents a stored JWT token
type Token struct {
	ID        int64
	Token     string
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// NewDB creates a new database connection
func NewDB(connString string) (*DB, error) {
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// InitSchema creates the necessary database tables
func (db *DB) InitSchema(ctx context.Context) error {
	schema := `
	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		enabled BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		last_login_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);

	-- Create trigger for users updated_at
	CREATE OR REPLACE FUNCTION update_users_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;
	CREATE TRIGGER trigger_update_users_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_users_updated_at();

	-- Create default admin user if not exists
	INSERT INTO users (id, email, name, password_hash, role, enabled)
	VALUES (
		'admin-001',
		'admin@example.com',
		'System Administrator',
		'$2a$10$rKvE7VE.h5LGW5Y5YnXzIOP7JqF.N8j8hP5F2KVxFJXpqN5vqYQOy',
		'admin',
		true
	)
	ON CONFLICT (id) DO NOTHING;

	-- Create plugins table
	CREATE TABLE IF NOT EXISTS plugins (
		id SERIAL PRIMARY KEY,
		slug TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		version TEXT NOT NULL,
		category TEXT,
		host TEXT NOT NULL,
		base_api_route TEXT NOT NULL,
		settings_route TEXT,
		api_routes JSONB DEFAULT '[]'::jsonb,
		enabled BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_plugins_slug ON plugins(slug);
	CREATE INDEX IF NOT EXISTS idx_plugins_category ON plugins(category);
	CREATE INDEX IF NOT EXISTS idx_plugins_enabled ON plugins(enabled);
	CREATE INDEX IF NOT EXISTS idx_plugins_base_api_route ON plugins(base_api_route);

	-- Create trigger for plugins updated_at
	CREATE OR REPLACE FUNCTION update_plugins_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS trigger_update_plugins_updated_at ON plugins;
	CREATE TRIGGER trigger_update_plugins_updated_at
		BEFORE UPDATE ON plugins
		FOR EACH ROW
		EXECUTE FUNCTION update_plugins_updated_at();

	-- Create tokens table
	CREATE TABLE IF NOT EXISTS tokens (
		id SERIAL PRIMARY KEY,
		token TEXT NOT NULL UNIQUE,
		subject TEXT NOT NULL,
		issued_at TIMESTAMP NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		revoked BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT NOW(),
		CONSTRAINT fk_tokens_user FOREIGN KEY (subject) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_tokens_subject ON tokens(subject);
	CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);
	CREATE INDEX IF NOT EXISTS idx_tokens_revoked ON tokens(revoked);
	`

	_, err := db.conn.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// SaveToken stores a new token in the database
func (db *DB) SaveToken(ctx context.Context, token, subject string, issuedAt, expiresAt time.Time) error {
	query := `
		INSERT INTO tokens (token, subject, issued_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.conn.ExecContext(ctx, query, token, subject, issuedAt, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// GetToken retrieves a token by its value
func (db *DB) GetToken(ctx context.Context, tokenValue string) (*Token, error) {
	query := `
		SELECT id, token, subject, issued_at, expires_at, revoked, created_at
		FROM tokens
		WHERE token = $1
	`

	var t Token
	err := db.conn.QueryRowContext(ctx, query, tokenValue).Scan(
		&t.ID, &t.Token, &t.Subject, &t.IssuedAt, &t.ExpiresAt, &t.Revoked, &t.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &t, nil
}

// RevokeToken marks a token as revoked
func (db *DB) RevokeToken(ctx context.Context, tokenValue string) error {
	query := `
		UPDATE tokens
		SET revoked = TRUE
		WHERE token = $1
	`

	result, err := db.conn.ExecContext(ctx, query, tokenValue)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (db *DB) IsTokenRevoked(ctx context.Context, tokenValue string) (bool, error) {
	query := `
		SELECT revoked
		FROM tokens
		WHERE token = $1
	`

	var revoked bool
	err := db.conn.QueryRowContext(ctx, query, tokenValue).Scan(&revoked)

	if err == sql.ErrNoRows {
		return false, nil // Token not found, consider not revoked
	}
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}

	return revoked, nil
}

// GetTokensBySubject retrieves all tokens for a given subject
func (db *DB) GetTokensBySubject(ctx context.Context, subject string) ([]*Token, error) {
	query := `
		SELECT id, token, subject, issued_at, expires_at, revoked, created_at
		FROM tokens
		WHERE subject = $1
		ORDER BY created_at DESC
	`

	rows, err := db.conn.QueryContext(ctx, query, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens by subject: %w", err)
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		var t Token
		err := rows.Scan(&t.ID, &t.Token, &t.Subject, &t.IssuedAt, &t.ExpiresAt, &t.Revoked, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	return tokens, nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (db *DB) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM tokens
		WHERE expires_at < NOW()
	`

	result, err := db.conn.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// RevokeAllTokensForSubject revokes all tokens for a given subject
func (db *DB) RevokeAllTokensForSubject(ctx context.Context, subject string) error {
	query := `
		UPDATE tokens
		SET revoked = TRUE
		WHERE subject = $1 AND revoked = FALSE
	`

	_, err := db.conn.ExecContext(ctx, query, subject)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens for subject: %w", err)
	}

	return nil
}
