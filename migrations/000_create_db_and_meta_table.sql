-- The Meta table stores key-value pairs for application metadata.
CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Insert initial metadata
INSERT INTO meta (key, value) VALUES ('name', 'HotelHub');
