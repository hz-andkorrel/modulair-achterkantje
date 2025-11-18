package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"broker/internal/models"
)

// CreateUser creates a new user with hashed password
func (db *DB) CreateUser(ctx context.Context, email, password, name, role string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID (you can use UUID or custom logic)
	userID := fmt.Sprintf("user-%d", time.Now().UnixNano())

	query := `
		INSERT INTO users (id, email, name, password_hash, role, enabled)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, email, name, role, enabled, created_at, updated_at
	`

	var user models.User
	var createdAt, updatedAt time.Time

	err = db.conn.QueryRowContext(ctx, query, userID, email, name, hashedPassword, role).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = createdAt

	return &user, nil
}

// GetUserByEmail retrieves a user by email address
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, enabled, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	var createdAt, updatedAt time.Time
	var lastLoginAt sql.NullTime

	err := db.conn.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.CreatedAt = createdAt

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, enabled, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	var createdAt, updatedAt time.Time
	var lastLoginAt sql.NullTime

	err := db.conn.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.CreatedAt = createdAt

	return &user, nil
}

// ValidateUserPassword checks if the provided password matches the user's hashed password
func (db *DB) ValidateUserPassword(ctx context.Context, email, password string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, enabled, created_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	var passwordHash string
	var enabled bool
	var createdAt time.Time

	err := db.conn.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&passwordHash,
		&user.Role,
		&enabled,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}

	// Check if user is enabled
	if !enabled {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	user.CreatedAt = createdAt

	return &user, nil
}

// UpdateLastLogin updates the last_login_at timestamp for a user
func (db *DB) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_login_at = NOW()
		WHERE id = $1
	`

	_, err := db.conn.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// UpdateUser updates user information (not password)
func (db *DB) UpdateUser(ctx context.Context, userID, name, email string) error {
	query := `
		UPDATE users
		SET name = $2, email = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := db.conn.ExecContext(ctx, query, userID, name, email)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateUserPassword updates a user's password
func (db *DB) UpdateUserPassword(ctx context.Context, userID, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := db.conn.ExecContext(ctx, query, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser removes a user from the database
func (db *DB) DeleteUser(ctx context.Context, userID string) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := db.conn.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetAllUsers retrieves all users from the database
func (db *DB) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, name, role, enabled, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var createdAt, updatedAt time.Time
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&createdAt,
			&updatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.CreatedAt = createdAt
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
