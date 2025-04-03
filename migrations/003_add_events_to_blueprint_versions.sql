-- Add events column to blueprint_versions table to store custom event definitions
ALTER TABLE blueprint_versions
ADD COLUMN events JSONB DEFAULT '[]'::jsonb;

-- Add event_bindings column to blueprint_versions table
ALTER TABLE blueprint_versions
ADD COLUMN event_bindings JSONB DEFAULT '[]'::jsonb;

-- Optional: Add indexes if querying events frequently becomes necessary
CREATE INDEX idx_blueprint_versions_events ON blueprint_versions USING GIN (events);
CREATE INDEX idx_blueprint_versions_event_bindings ON blueprint_versions USING GIN (event_bindings);

-- Add comments on the new columns
COMMENT ON COLUMN blueprint_versions.events IS 'Stores user-defined event definitions specific to this blueprint version as a JSON array.';
COMMENT ON COLUMN blueprint_versions.event_bindings IS 'Stores user-defined event bindings specific to this blueprint version as a JSON array.';