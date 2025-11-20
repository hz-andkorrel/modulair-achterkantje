-- Create plugins table for plugin registration storage
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

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_plugins_slug ON plugins(slug);
CREATE INDEX IF NOT EXISTS idx_plugins_category ON plugins(category);
CREATE INDEX IF NOT EXISTS idx_plugins_enabled ON plugins(enabled);
CREATE INDEX IF NOT EXISTS idx_plugins_base_api_route ON plugins(base_api_route);

-- Add comments
COMMENT ON TABLE plugins IS 'Stores all registered plugins';
COMMENT ON COLUMN plugins.slug IS 'Unique identifier for the plugin';
COMMENT ON COLUMN plugins.name IS 'Display name of the plugin';
COMMENT ON COLUMN plugins.description IS 'Description of what the plugin does';
COMMENT ON COLUMN plugins.version IS 'Version number of the plugin';
COMMENT ON COLUMN plugins.category IS 'Category for grouping plugins (e.g., user-interface, authentication)';
COMMENT ON COLUMN plugins.host IS 'Host URL where the plugin is running';
COMMENT ON COLUMN plugins.base_api_route IS 'Base route for the plugin API';
COMMENT ON COLUMN plugins.settings_route IS 'Route for plugin settings page';
COMMENT ON COLUMN plugins.api_routes IS 'JSON array of API routes provided by the plugin';
COMMENT ON COLUMN plugins.enabled IS 'Whether the plugin is currently enabled';
COMMENT ON COLUMN plugins.created_at IS 'When the plugin was first registered';
COMMENT ON COLUMN plugins.updated_at IS 'When the plugin was last updated';

-- Create trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_plugins_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_plugins_updated_at
    BEFORE UPDATE ON plugins
    FOR EACH ROW
    EXECUTE FUNCTION update_plugins_updated_at();
