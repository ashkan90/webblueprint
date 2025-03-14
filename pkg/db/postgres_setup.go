package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/registry"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"
	"webblueprint/pkg/repository/postgres"
)

// Setup initializes the database and sets up the connection
func Setup(ctx context.Context) (*ConnectionManager, repository.RepositoryFactory, error) {
	// Get database configuration from environment variables
	dbConfig := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "root"),
		DBName:   getEnv("DB_NAME", "webblueprint"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Initialize connection
	connectionManager := GetConnectionManager()
	err := connectionManager.Initialize(dbConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database connection: %w", err)
	}

	// Create repository factory
	repoFactory := postgres.NewRepositoryFactory(connectionManager.GetDB())

	// Set up schema if needed
	schemaPath := getEnv("SCHEMA_PATH", "./migrations")
	if err := setupSchema(ctx, connectionManager, schemaPath); err != nil {
		connectionManager.Close()
		return nil, nil, fmt.Errorf("failed to set up database schema: %w", err)
	}

	// Set up default user if needed
	if err := setupDefaultUser(ctx, connectionManager, repoFactory); err != nil {
		connectionManager.Close()
		return nil, nil, fmt.Errorf("failed to set up default user: %w", err)
	}

	if err := migrateNodeTypes(ctx, connectionManager, repoFactory); err != nil {
		log.Printf("failed to migrate node types: %v", err)
	}

	// Migrate in-memory blueprints if DB is empty and in-memory has data
	if err := migrateInMemoryData(ctx, connectionManager, repoFactory); err != nil {
		log.Printf("Warning: Failed to migrate in-memory data: %v", err)
	}

	return connectionManager, repoFactory, nil
}

// setupSchema sets up the database schema
func setupSchema(ctx context.Context, cm *ConnectionManager, schemaPath string) error {
	// Check if schema path exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		log.Printf("Schema path not found: %s", schemaPath)

		// Try to find schema in alternative locations
		alternatives := []string{
			"./migrations",
			"../migrations",
			"../../migrations",
			"./db/migrations",
			"../db/migrations",
		}

		for _, alt := range alternatives {
			if _, err := os.Stat(alt); !os.IsNotExist(err) {
				log.Printf("Found schema at: %s", alt)
				schemaPath = alt
				break
			}
		}
	}

	// If still not found, create a basic schema
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		log.Printf("Creating basic schema directory at: %s", schemaPath)
		if err := os.MkdirAll(schemaPath, 0755); err != nil {
			return fmt.Errorf("failed to create schema directory: %w", err)
		}

		// Write a basic schema file with essential tables
		basicSchemaPath := filepath.Join(schemaPath, "001_initial_schema.sql")
		if err := writeBasicSchema(basicSchemaPath); err != nil {
			return fmt.Errorf("failed to write basic schema: %w", err)
		}
	}

	// Run migrations
	migrationManager := NewMigrationManager(cm.GetDB(), schemaPath)
	if err := migrationManager.SetupSchema(ctx); err != nil {
		return fmt.Errorf("failed to apply schema migrations: %w", err)
	}

	return nil
}

// writeBasicSchema writes a minimal schema file to get started
func writeBasicSchema(path string) error {
	// This is a simplified version of the schema that includes just enough
	// to get the system running. In a real application, you'd have a more
	// complete schema, possibly split into multiple migration files.

	basicSchema := `-- Basic schema for WebBlueprint
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users and Authentication
CREATE TABLE users (
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

-- Workspaces
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user', 'team')),
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    thumbnail_url TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
    -- The validation of owner_id will be handled at the application level
);

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL DEFAULT 'editor',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

-- Assets
CREATE TABLE assets (
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

-- Blueprints
CREATE TABLE blueprints (
    id UUID PRIMARY KEY REFERENCES assets(id) ON DELETE CASCADE,
    current_version_id UUID,
    node_count INT NOT NULL DEFAULT 0,
    connection_count INT NOT NULL DEFAULT 0,
    entry_points TEXT[] DEFAULT ARRAY[]::TEXT[],
    is_template BOOLEAN NOT NULL DEFAULT FALSE,
    category VARCHAR(100)
);

CREATE TABLE blueprint_versions (
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
ALTER TABLE blueprints 
    ADD CONSTRAINT fk_current_version 
    FOREIGN KEY (current_version_id) 
    REFERENCES blueprint_versions(id);

-- Executions
CREATE TABLE executions (
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

CREATE TABLE execution_nodes (
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

CREATE TABLE execution_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    node_id TEXT,
    log_level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    details JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- References
CREATE TABLE asset_references (
    source_asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    target_asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    reference_type VARCHAR(50) NOT NULL,
    reference_count INT NOT NULL DEFAULT 1,
    details JSONB,
    PRIMARY KEY (source_asset_id, target_asset_id, reference_type)
);

-- Indexes
CREATE INDEX idx_assets_workspace ON assets(workspace_id);
CREATE INDEX idx_assets_type ON assets(type);
CREATE INDEX idx_assets_created_by ON assets(created_by);
CREATE INDEX idx_blueprint_versions_blueprint ON blueprint_versions(blueprint_id);
CREATE INDEX idx_executions_blueprint ON executions(blueprint_id);
CREATE INDEX idx_execution_logs_execution ON execution_logs(execution_id);
CREATE INDEX idx_asset_references_target ON asset_references(target_asset_id);
`

	return os.WriteFile(path, []byte(basicSchema), 0644)
}

// setupDefaultUser creates a default user if no users exist
func setupDefaultUser(ctx context.Context, cm *ConnectionManager, repoFactory repository.RepositoryFactory) error {
	userRepo := repoFactory.GetUserRepository()

	// Create migration manager
	migrationManager := NewMigrationManager(cm.GetDB(), "")

	// Try to create or get default user
	userID, err := migrationManager.CreateDefaultUser(ctx, userRepo)
	if err != nil {
		return fmt.Errorf("failed to set up default user: %w", err)
	}

	log.Printf("Using default user ID: %s", userID)
	return nil
}

// migrateInMemoryData migrates in-memory blueprints to the database
func migrateInMemoryData(ctx context.Context, cm *ConnectionManager, repoFactory repository.RepositoryFactory) error {
	// First check if there are any blueprints in the database
	var count int
	err := cm.GetDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM blueprints").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check blueprints: %w", err)
	}

	// If there are already blueprints in the database, don't migrate
	if count > 0 {
		log.Printf("Database already has %d blueprints, skipping in-memory migration", count)
		return nil
	}

	// Check if we have in-memory blueprints
	if len(db.Blueprints) == 0 {
		log.Println("No in-memory blueprints to migrate")
		return nil
	}

	// Get default user ID
	var userID string
	err = cm.GetDB().QueryRowContext(ctx, "SELECT id FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Create a default workspace
	workspaceID := ""

	// Create migration manager
	migrationManager := NewMigrationManager(cm.GetDB(), "")

	// Migrate blueprints
	inMemoryBlueprints := make(map[string]*blueprint.Blueprint)
	for bpID, bp := range db.Blueprints {
		inMemoryBlueprints[bpID] = bp
	}

	if err := migrationManager.MigrateInMemoryBlueprints(
		ctx,
		inMemoryBlueprints,
		repoFactory,
		workspaceID,
		userID,
	); err != nil {
		return fmt.Errorf("failed to migrate in-memory blueprints: %w", err)
	}

	log.Printf("Successfully migrated %d in-memory blueprints to database", len(inMemoryBlueprints))
	return nil
}

func migrateNodeTypes(ctx context.Context, cm *ConnectionManager, repoFactory repository.RepositoryFactory) error {
	nodeTypeRepo := repoFactory.GetNodeRepository()

	for typeID, factory := range registry.GetInstance().GetAllNodeFactories() {
		nodeInstance := factory()
		metadata := nodeInstance.GetMetadata()

		// Mevcut mu kontrol et
		exists, _ := nodeTypeRepo.NodeExists(ctx, typeID)
		category, _ := nodeTypeRepo.NodeCategoryCreateIfNot(ctx, metadata.Category)
		if !exists {
			// Veritabanına ekle
			nodeType := &models.NodeType{
				ID:           typeID,
				Name:         metadata.Name,
				Description:  models.NullString(metadata.Description),
				CategoryID:   models.NullString(category.ID),
				Version:      metadata.Version,
				Author:       models.CoreNodeAuthor,
				AuthorURL:    models.CoreNodeAuthorUrl,
				Icon:         models.CoreNodeAuthorIcon,
				IsCore:       true,
				IsDeprecated: false,
				Inputs:       models.ArrayToArrayJSON(nodeInstance.GetInputPins()),
				Outputs:      models.ArrayToArrayJSON(nodeInstance.GetOutputPins()),
				Properties:   models.ArrayToArrayJSON([]any{}),
				Metadata:     models.StructToJSONB(metadata),
				CreatedAt:    time.Now(),
			}

			if err := nodeTypeRepo.NodeCreate(ctx, nodeType); err != nil {
				log.Printf("Node tipi kaydedilemedi %s: %v", typeID, err)
			}
		}
	}
	return nil
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get environment variable as integer with default
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue := defaultValue
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		log.Printf("Warning: Could not parse %s=%s as int, using default: %d", key, value, defaultValue)
		return defaultValue
	}

	return intValue
}
