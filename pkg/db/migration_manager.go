package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// MigrationManager handles database migrations and schema setup
type MigrationManager struct {
	db         *sql.DB
	schemaPath string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, schemaPath string) *MigrationManager {
	return &MigrationManager{
		db:         db,
		schemaPath: schemaPath,
	}
}

// SetupSchema sets up the database schema
func (m *MigrationManager) SetupSchema(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	_, err := m.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	// Get list of migration files
	files, err := os.ReadDir(m.schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	// Sort migration files by name (which should follow a versioning convention)
	migrationFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, filepath.Ext(filename))
		if appliedMigrations[version] {
			log.Printf("Skipping already applied migration: %s", version)
			continue
		}

		log.Printf("Applying migration: %s", version)

		// Read the migration file
		filePath := filepath.Join(m.schemaPath, filename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		// Begin transaction
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute the migration
		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record the migration
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)",
			version,
			time.Now(),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		log.Printf("Successfully applied migration: %s", version)
	}

	return nil
}

// MigrateInMemoryBlueprints migrates blueprints from in-memory storage to the database
func (m *MigrationManager) MigrateInMemoryBlueprints(
	ctx context.Context,
	inMemoryBPs map[string]*blueprint.Blueprint,
	repoFactory repository.RepositoryFactory,
	defaultWorkspaceID string,
	defaultUserID string,
) error {
	// Get the blueprint repository
	blueprintRepo := repoFactory.GetBlueprintRepository()

	// Create a workspace if it doesn't exist
	if defaultWorkspaceID == "" {
		defaultWorkspaceID = uuid.New().String()

		workspaceRepo := repoFactory.GetWorkspaceRepository()
		workspace := &models.Workspace{
			ID:        defaultWorkspaceID,
			Name:      "Default Workspace",
			OwnerType: "user",
			OwnerID:   defaultUserID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsPublic:  false,
			Metadata:  models.JSONB{"migrated": true},
		}

		err := workspaceRepo.Create(ctx, workspace)
		if err != nil {
			return fmt.Errorf("failed to create default workspace: %w", err)
		}
	}

	// Migrate each blueprint
	for id, bp := range inMemoryBPs {
		log.Printf("Migrating blueprint: %s (%s)", bp.Name, id)

		// Convert to database models
		blueprintModel, versionModel, err := blueprintRepo.FromPkgBlueprint(bp)
		if err != nil {
			log.Printf("Error converting blueprint %s: %v", id, err)
			continue
		}

		// Set additional metadata
		blueprintModel.WorkspaceID = defaultWorkspaceID
		blueprintModel.CreatedBy = defaultUserID
		blueprintModel.UpdatedBy = defaultUserID
		blueprintModel.CreatedAt = time.Now()
		blueprintModel.UpdatedAt = time.Now()

		// Set version metadata
		versionModel.CreatedBy = defaultUserID
		versionModel.VersionNumber = 1
		versionModel.ID = uuid.New().String()

		// Set the current version
		blueprintModel.CurrentVersion = versionModel
		blueprintModel.CurrentVersionID.String = versionModel.ID
		blueprintModel.CurrentVersionID.Valid = true

		// Create the blueprint
		err = blueprintRepo.Create(ctx, blueprintModel)
		if err != nil {
			log.Printf("Error migrating blueprint %s: %v", id, err)
			continue
		}

		log.Printf("Successfully migrated blueprint: %s", id)
	}

	return nil
}

// CreateDefaultUser creates a default user if no users exist
func (m *MigrationManager) CreateDefaultUser(ctx context.Context, userRepo repository.UserRepository) (string, error) {
	// Check if any users exist
	users, err := m.executeCountQuery(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		return "", fmt.Errorf("failed to check users: %w", err)
	}

	if users > 0 {
		// Get the first user ID
		var userID string
		err := m.db.QueryRowContext(ctx, "SELECT id FROM users LIMIT 1").Scan(&userID)
		if err != nil {
			return "", fmt.Errorf("failed to get user ID: %w", err)
		}
		return userID, nil
	}

	// Create a default admin user
	userID := uuid.New().String()
	defaultUser := &models.User{
		ID:           userID,
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$JVS5i5R8G6g5GN.HB3/yHuZK.5K2FkE6qHKxj1jXQOiZzw9TksVQW", // "admin123"
		FullName:     "System Administrator",
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err = userRepo.Create(ctx, defaultUser)
	if err != nil {
		return "", fmt.Errorf("failed to create default user: %w", err)
	}

	log.Printf("Created default admin user: %s", userID)
	return userID, nil
}

// executeCountQuery executes a query that returns a count
func (m *MigrationManager) executeCountQuery(ctx context.Context, query string) (int, error) {
	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
