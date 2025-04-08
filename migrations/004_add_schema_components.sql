-- migrations/004_add_schema_components.sql
CREATE TABLE schema_components (
    id TEXT PRIMARY KEY,             -- Using TEXT for UUIDs or similar identifiers
    name TEXT NOT NULL,              -- User-defined name for the schema
    schema_definition TEXT NOT NULL, -- The actual schema definition (e.g., JSON string)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Optional: Index on name for faster lookups if needed
-- CREATE INDEX idx_schema_components_name ON schema_components(name);

-- Optional: Trigger to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_schema_components_updated_at
BEFORE UPDATE ON schema_components
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();