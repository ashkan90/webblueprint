package postgres

import (
	"database/sql"
	"webblueprint/internal/db" // Added import for internal db package
	"webblueprint/pkg/repository"
)

// PostgresRepositoryFactory is an implementation of RepositoryFactory for PostgreSQL
type PostgresRepositoryFactory struct {
	db *sql.DB

	// Cached repository instances
	assetRepo             repository.AssetRepository
	blueprintRepo         repository.BlueprintRepository
	blueprintVariableRepo repository.BlueprintVariableRepository
	workspaceRepo         repository.WorkspaceRepository
	userRepo              repository.UserRepository
	executionRepo         repository.ExecutionRepository
	nodeRepo              repository.NodeRepository
	eventRepo             repository.EventRepository
	schemaComponentStore  db.SchemaComponentStore // Added field
}

// NewRepositoryFactory creates a new PostgreSQL repository factory
func NewRepositoryFactory(db *sql.DB) repository.RepositoryFactory {
	return &PostgresRepositoryFactory{
		db: db,
	}
}

// GetAssetRepository returns an AssetRepository implementation
func (f *PostgresRepositoryFactory) GetAssetRepository() repository.AssetRepository {
	if f.assetRepo == nil {
		f.assetRepo = NewAssetRepository(f.db)
	}
	return f.assetRepo
}

// GetBlueprintRepository returns a BlueprintRepository implementation
func (f *PostgresRepositoryFactory) GetBlueprintRepository() repository.BlueprintRepository {
	if f.blueprintRepo == nil {
		f.blueprintRepo = NewBlueprintRepository(f.db)
	}
	return f.blueprintRepo
}

// GetBlueprintVariableRepository returns a BlueprintVariableRepository implementation
func (f *PostgresRepositoryFactory) GetBlueprintVariableRepository() repository.BlueprintVariableRepository {
	if f.blueprintVariableRepo == nil {
		f.blueprintVariableRepo = NewBlueprintVariableRepository(f.db)
	}
	return f.blueprintVariableRepo
}

// GetWorkspaceRepository returns a WorkspaceRepository implementation
func (f *PostgresRepositoryFactory) GetWorkspaceRepository() repository.WorkspaceRepository {
	if f.workspaceRepo == nil {
		f.workspaceRepo = NewWorkspaceRepository(f.db)
	}
	return f.workspaceRepo
}

// GetUserRepository returns a UserRepository implementation
func (f *PostgresRepositoryFactory) GetUserRepository() repository.UserRepository {
	if f.userRepo == nil {
		f.userRepo = NewUserRepository(f.db)
	}
	return f.userRepo
}

// GetExecutionRepository returns an ExecutionRepository implementation
func (f *PostgresRepositoryFactory) GetExecutionRepository() repository.ExecutionRepository {
	if f.executionRepo == nil {
		f.executionRepo = NewExecutionRepository(f.db)
	}
	return f.executionRepo
}

func (f *PostgresRepositoryFactory) GetNodeRepository() repository.NodeRepository {
	if f.nodeRepo == nil {
		f.nodeRepo = NewPostgresNodeRepository(f.db)
	}

	return f.nodeRepo
}

func (f *PostgresRepositoryFactory) GetEventRepository() repository.EventRepository {
	if f.eventRepo == nil {
		f.eventRepo = NewEventRepository(f.db)
	}

	return f.eventRepo
}

// GetSchemaComponentStore returns a SchemaComponentStore implementation
func (f *PostgresRepositoryFactory) GetSchemaComponentStore() db.SchemaComponentStore {
	if f.schemaComponentStore == nil {
		// Assuming NewSQLSchemaComponentStore exists in internal/db
		f.schemaComponentStore = db.NewSQLSchemaComponentStore(f.db)
	}
	return f.schemaComponentStore
}
