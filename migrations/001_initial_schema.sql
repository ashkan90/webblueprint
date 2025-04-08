-- WebBlueprint Database Schema
-- Based on PostgreSQL with JSONB architecture

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- -----------------------------------------------------
-- Users and Authentication
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       username VARCHAR(255) NOT NULL UNIQUE,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       full_name VARCHAR(255),
                       avatar_url TEXT,
                       role VARCHAR(50) NOT NULL DEFAULT 'user',
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       last_login_at TIMESTAMPTZ,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS teams (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       name VARCHAR(255) NOT NULL,
                       description TEXT,
                       avatar_url TEXT,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       created_by UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS team_members (
                              team_id UUID NOT NULL REFERENCES teams(id),
                              user_id UUID NOT NULL REFERENCES users(id),
                              role VARCHAR(50) NOT NULL DEFAULT 'member',
                              joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              PRIMARY KEY (team_id, user_id)
);

-- -----------------------------------------------------
-- Workspaces and Core Assets
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS workspaces (
                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                            name VARCHAR(255) NOT NULL,
                            description TEXT,
                            owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user', 'team')),
                            owner_id UUID NOT NULL,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            is_public BOOLEAN NOT NULL DEFAULT FALSE,
                            thumbnail_url TEXT,
                            metadata JSONB DEFAULT '{}'::jsonb,
                            CONSTRAINT unique_workspace_name_per_owner UNIQUE (name, owner_type, owner_id)
--                             CONSTRAINT valid_owner CHECK (
--                                 (owner_type = 'user' AND EXISTS (SELECT 1 FROM users WHERE id = owner_id)) OR
--                                 (owner_type = 'team' AND EXISTS (SELECT 1 FROM teams WHERE id = owner_id))
--                                 )
);

CREATE TABLE IF NOT EXISTS workspace_members (
                                   workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
                                   user_id UUID NOT NULL REFERENCES users(id),
                                   role VARCHAR(50) NOT NULL DEFAULT 'editor',
                                   joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                   PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS assets (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
                        name VARCHAR(255) NOT NULL,
                        description TEXT,
                        type VARCHAR(50) NOT NULL,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        created_by UUID NOT NULL REFERENCES users(id),
                        updated_by UUID NOT NULL REFERENCES users(id),
                        is_public BOOLEAN NOT NULL DEFAULT FALSE,
                        tags TEXT[] DEFAULT ARRAY[]::TEXT[],
                        thumbnail_url TEXT,
                        metadata JSONB DEFAULT '{}'::jsonb,
                        CONSTRAINT unique_asset_name_per_workspace_type UNIQUE (workspace_id, name, type)
);

-- -----------------------------------------------------
-- Blueprints and Their Components
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS blueprints (
                            id UUID PRIMARY KEY REFERENCES assets(id) ON DELETE CASCADE,
                            current_version_id UUID,
                            node_count INT NOT NULL DEFAULT 0,
                            connection_count INT NOT NULL DEFAULT 0,
                            entry_points TEXT[] DEFAULT ARRAY[]::TEXT[],
                            is_template BOOLEAN NOT NULL DEFAULT FALSE,
                            category VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS blueprint_versions (
                                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                    blueprint_id UUID NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
                                    version_number INT NOT NULL,
                                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                    created_by UUID NOT NULL REFERENCES users(id),
                                    comment TEXT,
                                    nodes JSONB NOT NULL DEFAULT '[]'::jsonb,
                                    connections JSONB NOT NULL DEFAULT '[]'::jsonb,
                                    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
                                    functions JSONB NOT NULL DEFAULT '[]'::jsonb,
                                    metadata JSONB DEFAULT '{}'::jsonb,
                                    CONSTRAINT unique_blueprint_version UNIQUE (blueprint_id, version_number)
);

-- Add foreign key after both tables are created
ALTER TABLE blueprints DROP CONSTRAINT IF EXISTS fk_current_version;
ALTER TABLE blueprints
    ADD CONSTRAINT fk_current_version
        FOREIGN KEY (current_version_id)
            REFERENCES blueprint_versions(id);

CREATE TABLE IF NOT EXISTS functions (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           blueprint_id UUID NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
                           blueprint_version_id UUID REFERENCES blueprint_versions(id),
                           name VARCHAR(255) NOT NULL,
                           description TEXT,
    -- Metadata for querying
                           category VARCHAR(100),
                           is_public BOOLEAN NOT NULL DEFAULT FALSE,
                           version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           created_by UUID NOT NULL REFERENCES users(id),
                           updated_by UUID NOT NULL REFERENCES users(id),
    -- Function reference info
                           function_id TEXT NOT NULL, -- ID within the blueprint JSONB
                           input_types JSONB,  -- Simplified schema for querying
                           output_types JSONB, -- Simplified schema for querying
    -- Still includes the function interface for quick access
                           node_interface JSONB NOT NULL,
                           CONSTRAINT unique_function_name_per_blueprint UNIQUE (blueprint_id, name)
);

CREATE TABLE IF NOT EXISTS variables (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           blueprint_id UUID NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
                           blueprint_version_id UUID REFERENCES blueprint_versions(id),
                           name VARCHAR(255) NOT NULL,
                           type VARCHAR(50) NOT NULL,
    -- Quick access fields
                           default_value JSONB,
                           description TEXT,
                           is_exposed BOOLEAN NOT NULL DEFAULT FALSE,
                           category VARCHAR(100),
                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Variable reference info
                           variable_id TEXT NOT NULL, -- ID within the blueprint JSONB
                           CONSTRAINT unique_variable_name_per_blueprint UNIQUE (blueprint_id, name)
);

-- -----------------------------------------------------
-- Node Types and Registry
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS node_categories (
                                 id VARCHAR(100) PRIMARY KEY,
                                 name VARCHAR(255) NOT NULL,
                                 description TEXT,
                                 color VARCHAR(20),
                                 icon TEXT,
                                 sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS node_types (
                            id VARCHAR(255) PRIMARY KEY,
                            name VARCHAR(255) NOT NULL,
                            description TEXT,
                            category_id VARCHAR(100) REFERENCES node_categories(id),
                            version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
                            author VARCHAR(255),
                            author_url TEXT,
                            icon TEXT,
                            is_core BOOLEAN NOT NULL DEFAULT FALSE,
                            is_deprecated BOOLEAN NOT NULL DEFAULT FALSE,
                            inputs JSONB NOT NULL DEFAULT '[]'::jsonb,
                            outputs JSONB NOT NULL DEFAULT '[]'::jsonb,
                            properties JSONB NOT NULL DEFAULT '[]'::jsonb,
                            metadata JSONB DEFAULT '{}'::jsonb,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------
-- Execution and Runtime
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS executions (
                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                            blueprint_id UUID NOT NULL REFERENCES blueprints(id),
                            version_id UUID REFERENCES blueprint_versions(id),
                            started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            completed_at TIMESTAMPTZ,
                            status VARCHAR(50) NOT NULL DEFAULT 'running',
                            initiated_by UUID NOT NULL REFERENCES users(id),
                            execution_mode VARCHAR(50) NOT NULL DEFAULT 'standard',
                            initial_variables JSONB DEFAULT '{}'::jsonb,
                            result JSONB,
                            error TEXT,
                            duration_ms INT
);

CREATE TABLE IF NOT EXISTS execution_nodes (
                                 execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
                                 node_id TEXT NOT NULL,
                                 node_type VARCHAR(255) NOT NULL,
                                 started_at TIMESTAMPTZ,
                                 completed_at TIMESTAMPTZ,
                                 status VARCHAR(50) NOT NULL DEFAULT 'pending',
                                 inputs JSONB DEFAULT '{}'::jsonb,
                                 outputs JSONB DEFAULT '{}'::jsonb,
                                 error TEXT,
                                 duration_ms INT,
                                 debug_data JSONB,
                                 PRIMARY KEY (execution_id, node_id)
);

CREATE TABLE IF NOT EXISTS execution_logs (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
                                node_id TEXT,
                                log_level VARCHAR(20) NOT NULL,
                                message TEXT NOT NULL,
                                details JSONB,
                                timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------
-- References and Relationships
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS asset_references (
                                  source_asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
                                  target_asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
                                  reference_type VARCHAR(50) NOT NULL,
                                  reference_count INT NOT NULL DEFAULT 1,
                                  details JSONB,
                                  PRIMARY KEY (source_asset_id, target_asset_id, reference_type)
);

CREATE TABLE IF NOT EXISTS blueprint_dependencies (
                                        blueprint_id UUID NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
                                        dependency_id UUID NOT NULL REFERENCES blueprints(id) ON DELETE RESTRICT,
                                        dependency_type VARCHAR(50) NOT NULL,
                                        is_optional BOOLEAN NOT NULL DEFAULT FALSE,
                                        version_constraint VARCHAR(100),
                                        PRIMARY KEY (blueprint_id, dependency_id)
);

-- -----------------------------------------------------
-- User Preferences and Settings
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS user_preferences (
                                  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                                  theme VARCHAR(50) DEFAULT 'light',
                                  language VARCHAR(10) DEFAULT 'en',
                                  node_size VARCHAR(20) DEFAULT 'medium',
                                  auto_save BOOLEAN DEFAULT TRUE,
                                  advanced_mode BOOLEAN DEFAULT FALSE,
                                  preferences JSONB DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS workspace_settings (
                                    workspace_id UUID PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
                                    default_blueprint_privacy BOOLEAN DEFAULT FALSE,
                                    enable_comments BOOLEAN DEFAULT TRUE,
                                    enable_versioning BOOLEAN DEFAULT TRUE,
                                    settings JSONB DEFAULT '{}'::jsonb
);

-- -----------------------------------------------------
-- Plugins and Extensions
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS plugins (
                         id VARCHAR(255) PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         description TEXT,
                         version VARCHAR(20) NOT NULL,
                         author VARCHAR(255),
                         author_url TEXT,
                         repository_url TEXT,
                         license VARCHAR(50),
                         is_active BOOLEAN NOT NULL DEFAULT TRUE,
                         is_system BOOLEAN NOT NULL DEFAULT FALSE,
                         installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         config JSONB DEFAULT '{}'::jsonb,
                         manifest JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS plugin_node_types (
                                   plugin_id VARCHAR(255) NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
                                   node_type_id VARCHAR(255) NOT NULL REFERENCES node_types(id) ON DELETE CASCADE,
                                   PRIMARY KEY (plugin_id, node_type_id)
);

-- -----------------------------------------------------
-- Marketplace and Sharing
-- -----------------------------------------------------

CREATE TABLE IF NOT EXISTS marketplace_items (
                                   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                   name VARCHAR(255) NOT NULL,
                                   description TEXT,
                                   type VARCHAR(50) NOT NULL,
                                   owner_id UUID NOT NULL REFERENCES users(id),
                                   is_approved BOOLEAN NOT NULL DEFAULT FALSE,
                                   is_featured BOOLEAN NOT NULL DEFAULT FALSE,
                                   price DECIMAL(10, 2),
                                   is_free BOOLEAN NOT NULL DEFAULT TRUE,
                                   downloads INT NOT NULL DEFAULT 0,
                                   rating DECIMAL(3, 2),
                                   rating_count INT NOT NULL DEFAULT 0,
                                   asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
                                   thumbnail_url TEXT,
                                   screenshots TEXT[],
                                   tags TEXT[],
                                   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                   published_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS marketplace_versions (
                                      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                      item_id UUID NOT NULL REFERENCES marketplace_items(id) ON DELETE CASCADE,
                                      version VARCHAR(20) NOT NULL,
                                      changelog TEXT,
                                      asset_version_id UUID,
                                      downloads INT NOT NULL DEFAULT 0,
                                      is_current BOOLEAN NOT NULL DEFAULT TRUE,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      published_at TIMESTAMPTZ,
                                      CONSTRAINT unique_marketplace_item_version UNIQUE (item_id, version)
);


-- -----------------------------------------------------
-- Defaults
-- -----------------------------------------------------
INSERT INTO users
    (id, username, email, password_hash, full_name, avatar_url, role, created_at, updated_at, last_login_at, is_active)
VALUES (
        '00000000-0000-0000-0000-000000000001', 'default_user', 'default_user@webblueprint.com',
        '$2a$10$pY.BJoIGoL7X71w3LQRDIeZ4juE7Oe653QmZ3ZObgA7g2qL1ptWYC', 'Default User', 'https://i.pravatar.cc/150?u=default_user@webblueprint.com',
        'user', now(), now(), now(), true
    );
-- -----------------------------------------------------
-- Indexes
-- -----------------------------------------------------

-- General indexes
CREATE INDEX IF NOT EXISTS idx_assets_workspace ON assets(workspace_id);
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
CREATE INDEX IF NOT EXISTS idx_assets_created_by ON assets(created_by);
CREATE INDEX IF NOT EXISTS idx_assets_created_at ON assets(created_at);
CREATE INDEX IF NOT EXISTS idx_assets_tags ON assets USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_assets_metadata ON assets USING GIN(metadata);

-- Blueprint specific indexes
CREATE INDEX IF NOT EXISTS idx_blueprint_versions_blueprint ON blueprint_versions(blueprint_id);
CREATE INDEX IF NOT EXISTS idx_blueprint_versions_created_at ON blueprint_versions(created_at);
CREATE INDEX IF NOT EXISTS idx_blueprint_nodes ON blueprint_versions USING GIN((nodes));
CREATE INDEX IF NOT EXISTS idx_blueprint_connections ON blueprint_versions USING GIN((connections));
CREATE INDEX IF NOT EXISTS idx_blueprint_variables ON blueprint_versions USING GIN((variables));
CREATE INDEX IF NOT EXISTS idx_blueprint_functions ON blueprint_versions USING GIN((functions));

-- Execution indexes
CREATE INDEX IF NOT EXISTS idx_executions_blueprint ON executions(blueprint_id);
CREATE INDEX IF NOT EXISTS idx_executions_started_at ON executions(started_at);
CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
CREATE INDEX IF NOT EXISTS idx_execution_nodes_status ON execution_nodes(status);
CREATE INDEX IF NOT EXISTS idx_execution_logs_execution ON execution_logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_node ON execution_logs(node_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_level ON execution_logs(log_level);

-- Reference indexes
CREATE INDEX IF NOT EXISTS idx_asset_references_target ON asset_references(target_asset_id);
CREATE INDEX IF NOT EXISTS idx_blueprint_dependencies_dependency ON blueprint_dependencies(dependency_id);

-- Marketplace indexes
CREATE INDEX IF NOT EXISTS idx_marketplace_items_owner ON marketplace_items(owner_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_items_type ON marketplace_items(type);
CREATE INDEX IF NOT EXISTS idx_marketplace_items_tags ON marketplace_items USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_marketplace_items_price ON marketplace_items(price) WHERE is_free = FALSE;
CREATE INDEX IF NOT EXISTS idx_marketplace_versions_item ON marketplace_versions(item_id);

-- -----------------------------------------------------
-- Functions and Triggers
-- -----------------------------------------------------

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Apply update_timestamp trigger to relevant tables update_users_timestamp;
DROP TRIGGER IF EXISTS update_users_timestamp ON webblueprint.public.users;
CREATE TRIGGER update_users_timestamp BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_teams_timestamp ON webblueprint.public.teams;
CREATE TRIGGER update_teams_timestamp BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_workspaces_timestamp ON workspaces;
CREATE TRIGGER update_workspaces_timestamp BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_assets_timestamp ON assets;
CREATE TRIGGER update_assets_timestamp BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_node_types_timestamp ON node_types;
CREATE TRIGGER update_node_types_timestamp BEFORE UPDATE ON node_types
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_plugins_timestamp ON plugins;
CREATE TRIGGER update_plugins_timestamp BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

DROP TRIGGER IF EXISTS update_marketplace_items_timestamp ON marketplace_items;
CREATE TRIGGER update_marketplace_items_timestamp BEFORE UPDATE ON marketplace_items
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Function to update node and connection counts
CREATE OR REPLACE FUNCTION update_blueprint_counts()
RETURNS TRIGGER AS $$
BEGIN
UPDATE blueprints
SET
    node_count = jsonb_array_length(NEW.nodes),
    connection_count = jsonb_array_length(NEW.connections)
WHERE id = NEW.blueprint_id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Trigger to update counts when a blueprint version is inserted or updated update_blueprint_counts_trigger;
DROP TRIGGER IF EXISTS update_blueprint_counts_trigger ON blueprint_versions;
CREATE TRIGGER update_blueprint_counts_trigger
    AFTER INSERT OR UPDATE OF nodes, connections ON blueprint_versions
    FOR EACH ROW EXECUTE FUNCTION update_blueprint_counts();

-- Function to update current version reference
CREATE OR REPLACE FUNCTION update_current_version()
RETURNS TRIGGER AS $$
BEGIN
UPDATE blueprints
SET current_version_id = NEW.id
WHERE id = NEW.blueprint_id AND (
    current_version_id IS NULL OR
    NEW.version_number > (
        SELECT version_number
        FROM blueprint_versions
        WHERE id = blueprints.current_version_id
    )
    );
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_blueprint_components()
RETURNS TRIGGER AS $$
DECLARE
func_record RECORD;
    var_record RECORD;
    version_id UUID;
BEGIN
    -- Get the version ID
    version_id := NEW.id;

    -- Clear existing entries
DELETE FROM functions WHERE blueprint_version_id = version_id;
DELETE FROM variables WHERE blueprint_version_id = version_id;

-- Extract functions
FOR func_record IN
SELECT * FROM jsonb_array_elements(NEW.functions) AS func
    LOOP
INSERT INTO functions (
    blueprint_id,
    blueprint_version_id,
    name,
    description,
    function_id,
    node_interface,
    input_types,
    output_types
) VALUES (
    NEW.blueprint_id,
    version_id,
    func_record.value->>'name',
    func_record.value->>'description',
    func_record.value->>'id',
    func_record.value->'nodeType',
    func_record.value->'inputTypes',
    func_record.value->'outputTypes'
    );
END LOOP;

    -- Extract variables (similar logic)
FOR var_record IN
SELECT * FROM jsonb_array_elements(NEW.variables) AS var
    LOOP
INSERT INTO variables (
    blueprint_id,
    blueprint_version_id,
    name,
    type,
    default_value,
    description,
    is_exposed,
    variable_id
) VALUES (
    NEW.blueprint_id,
    version_id,
    var_record.value->>'name',
    var_record.value->>'type',
    var_record.value->'defaultValue',
    var_record.value->>'description',
    (var_record.value->>'isExposed')::boolean,
    var_record.value->>'id'
    )
ON CONFLICT(blueprint_id,name)
    DO UPDATE SET
                  name = EXCLUDED.name,
                  type = EXCLUDED.type,
                  default_value = EXCLUDED.default_value,
                  description = EXCLUDED.description,
                  is_exposed = EXCLUDED.is_exposed;
END LOOP;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Trigger to update current version when a new version is created update_current_version_trigger;
DROP TRIGGER IF EXISTS update_current_version_trigger ON webblueprint.public.blueprint_versions;
CREATE TRIGGER update_current_version_trigger
    AFTER INSERT ON blueprint_versions
    FOR EACH ROW EXECUTE FUNCTION update_current_version();

-- Function to update execution status and duration
CREATE OR REPLACE FUNCTION update_execution_completion()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('completed', 'failed', 'cancelled') AND OLD.status = 'running' THEN
        NEW.completed_at = NOW();
        NEW.duration_ms = EXTRACT(EPOCH FROM (NOW() - NEW.started_at)) * 1000;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for execution completion update_execution_completion_trigger;
DROP TRIGGER IF EXISTS update_execution_completion_trigger ON webblueprint.public.executions;
CREATE TRIGGER update_execution_completion_trigger
    BEFORE UPDATE OF status ON executions
    FOR EACH ROW EXECUTE FUNCTION update_execution_completion();

-- Function to insert asset references from blueprint versions
CREATE OR REPLACE FUNCTION extract_blueprint_references()
RETURNS TRIGGER AS $$
DECLARE
node_record RECORD;
    ref_asset_id UUID;
    ref_type TEXT;
BEGIN
    -- Clear existing references for this blueprint
DELETE FROM asset_references
WHERE source_asset_id = NEW.blueprint_id
  AND reference_type LIKE 'node%';

-- Extract references from nodes
FOR node_record IN
SELECT * FROM jsonb_array_elements(NEW.nodes) AS node
    LOOP
        -- Check if this is a reference to another asset
        IF node_record.value->>'type' LIKE 'get-variable-%' THEN
            -- Extract variable reference
            ref_type := 'node_variable_reference';
-- Logic to find the asset ID would go here
-- For now, we'll just insert a placeholder

ELSIF node_record.value->>'type' LIKE '%function%' THEN
            -- Extract function reference
            ref_type := 'node_function_reference';
            -- Logic to find the asset ID would go here

END IF;

        -- If we found a reference, insert it
        IF ref_type IS NOT NULL AND ref_asset_id IS NOT NULL THEN
            INSERT INTO asset_references (
                source_asset_id,
                target_asset_id,
                reference_type,
                reference_count
            ) VALUES (
                NEW.blueprint_id,
                ref_asset_id,
                ref_type,
                1
            )
            ON CONFLICT (source_asset_id, target_asset_id, reference_type)
            DO UPDATE SET reference_count = asset_references.reference_count + 1;
END IF;
END LOOP;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Trigger to extract references when a blueprint version is inserted extract_blueprint_references_trigger;
DROP TRIGGER IF EXISTS extract_blueprint_references_trigger ON webblueprint.public.blueprint_versions;
CREATE TRIGGER extract_blueprint_references_trigger
    AFTER INSERT ON blueprint_versions
    FOR EACH ROW EXECUTE FUNCTION extract_blueprint_references();

DROP TRIGGER IF EXISTS sync_blueprint_components_trigger ON webblueprint.public.blueprint_versions;
CREATE TRIGGER sync_blueprint_components_trigger
    AFTER INSERT OR UPDATE OF functions, variables ON blueprint_versions
    FOR EACH ROW EXECUTE FUNCTION sync_blueprint_components();


-- -----------------------------------------------------
-- Views
-- -----------------------------------------------------

-- Blueprint summary view
CREATE VIEW blueprint_summary AS
SELECT
    b.id,
    a.name,
    a.description,
    a.workspace_id,
    a.created_at,
    a.updated_at,
    a.created_by,
    u.username AS created_by_username,
    a.is_public,
    b.node_count,
    b.connection_count,
    b.entry_points,
    b.is_template,
    b.category,
    bv.version_number AS current_version,
    bv.created_at AS version_date,
    a.tags,
    COUNT(DISTINCT ar.target_asset_id) AS dependency_count,
    COUNT(DISTINCT e.id) AS execution_count
FROM
    blueprints b
        JOIN
    assets a ON b.id = a.id
        JOIN
    users u ON a.created_by = u.id
        LEFT JOIN
    blueprint_versions bv ON b.current_version_id = bv.id
        LEFT JOIN
    asset_references ar ON b.id = ar.source_asset_id
        LEFT JOIN
    executions e ON b.id = e.blueprint_id
GROUP BY
    b.id, a.id, u.id, bv.id;

-- Execution summary view
CREATE VIEW execution_summary AS
SELECT
    e.id,
    e.blueprint_id,
    a.name AS blueprint_name,
    e.started_at,
    e.completed_at,
    e.status,
    e.execution_mode,
    e.duration_ms,
    u.username AS initiated_by,
    COALESCE(e.error, '') AS error,
    COUNT(en.node_id) AS total_nodes,
    SUM(CASE WHEN en.status = 'completed' THEN 1 ELSE 0 END) AS completed_nodes,
    SUM(CASE WHEN en.status = 'error' THEN 1 ELSE 0 END) AS error_nodes,
    COUNT(el.id) AS log_count
FROM
    executions e
        JOIN
    blueprints b ON e.blueprint_id = b.id
        JOIN
    assets a ON b.id = a.id
        JOIN
    users u ON e.initiated_by = u.id
        LEFT JOIN
    execution_nodes en ON e.id = en.execution_id
        LEFT JOIN
    execution_logs el ON e.id = el.execution_id
GROUP BY
    e.id, a.name, u.username;

-- User activity summary
CREATE VIEW user_activity_summary AS
SELECT
    u.id,
    u.username,
    u.email,
    u.last_login_at,
    COUNT(DISTINCT a.id) AS asset_count,
    COUNT(DISTINCT bv.id) AS blueprint_version_count,
    COUNT(DISTINCT e.id) AS execution_count,
    MAX(a.updated_at) AS last_asset_update,
    MAX(e.started_at) AS last_execution
FROM
    users u
        LEFT JOIN
    assets a ON u.id = a.created_by
        LEFT JOIN
    blueprints b ON a.id = b.id
        LEFT JOIN
    blueprint_versions bv ON b.id = bv.blueprint_id AND u.id = bv.created_by
        LEFT JOIN
    executions e ON u.id = e.initiated_by
GROUP BY
    u.id;