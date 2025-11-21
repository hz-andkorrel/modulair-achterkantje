-- Create users table for user management and authentication
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    enabled BOOLEAN DEFAULT TRUE,
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);

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
