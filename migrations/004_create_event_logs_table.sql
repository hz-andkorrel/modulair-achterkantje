
CREATE TABLE IF NOT EXISTS event_logs (
    id SERIAL PRIMARY KEY,
    channel VARCHAR(255) NOT NULL,
    action VARCHAR(255),
    user_email VARCHAR(255),
    payload TEXT NOT NULL,
    plugin_slug VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_event_logs_channel ON event_logs(channel);
CREATE INDEX idx_event_logs_created_at ON event_logs(created_at DESC);
CREATE INDEX idx_event_logs_action ON event_logs(action);
CREATE INDEX idx_event_logs_user_email ON event_logs(user_email);
CREATE INDEX idx_event_logs_plugin_slug ON event_logs(plugin_slug);


