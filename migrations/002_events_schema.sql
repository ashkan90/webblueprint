-- WebBlueprint Event System Migration
-- Add tables for events and event bindings

-- -----------------------------------------------------
-- Events
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    parameters TEXT NOT NULL, -- JSON string of event parameters
    blueprint_id UUID REFERENCES blueprints(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_event_name_per_blueprint UNIQUE (name, blueprint_id)
);

CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);
CREATE INDEX IF NOT EXISTS idx_events_blueprint_id ON events(blueprint_id);

-- -----------------------------------------------------
-- Event Bindings
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS event_bindings (
    id VARCHAR(255) PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    handler_id VARCHAR(255) NOT NULL,
    handler_type VARCHAR(50) NOT NULL,
    blueprint_id UUID REFERENCES blueprints(id) ON DELETE CASCADE,
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_binding UNIQUE (event_id, handler_id, handler_type)
);

CREATE INDEX IF NOT EXISTS idx_event_bindings_event_id ON event_bindings(event_id);
CREATE INDEX IF NOT EXISTS idx_event_bindings_blueprint_id ON event_bindings(blueprint_id);
CREATE INDEX IF NOT EXISTS idx_event_bindings_priority ON event_bindings(priority);

-- -----------------------------------------------------
-- Triggers for timestamp updates
-- -----------------------------------------------------

-- Create trigger for events
DROP TRIGGER IF EXISTS update_events_timestamp ON events;
CREATE TRIGGER update_events_timestamp
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Create trigger for event bindings
DROP TRIGGER IF EXISTS update_event_bindings_timestamp ON event_bindings;
CREATE TRIGGER update_event_bindings_timestamp
    BEFORE UPDATE ON event_bindings
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();
