-- Create users table for user management and authentication
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

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);

-- Add comments
COMMENT ON TABLE users IS 'Stores registered users for authentication and authorization';
COMMENT ON COLUMN users.id IS 'Unique user identifier (UUID or custom ID)';
COMMENT ON COLUMN users.email IS 'User email address (used for login)';
COMMENT ON COLUMN users.name IS 'Display name of the user';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.role IS 'User role (user, admin, etc.) for access control';
COMMENT ON COLUMN users.enabled IS 'Whether the user account is active';
COMMENT ON COLUMN users.created_at IS 'When the user account was created';
COMMENT ON COLUMN users.updated_at IS 'When the user account was last updated';
COMMENT ON COLUMN users.last_login_at IS 'When the user last logged in';

-- Create trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_users_updated_at();

-- Create default admin user (password: 'admin123' - CHANGE THIS IN PRODUCTION!)
-- Password hash generated with: bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
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
