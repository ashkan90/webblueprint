package repository

import (
	"context"
	"webblueprint/internal/node"
	"webblueprint/pkg/api/dt"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
)

// Repository interface for managing assets
type AssetRepository interface {
	// Create a new asset
	Create(ctx context.Context, asset *models.Asset) error

	// Get asset by ID
	GetByID(ctx context.Context, id string) (*models.Asset, error)

	// Get assets by workspace ID
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.Asset, error)

	// Get assets by type
	GetByType(ctx context.Context, assetType string) ([]*models.Asset, error)

	// Update an asset
	Update(ctx context.Context, asset *models.Asset) error

	// Delete an asset by ID
	Delete(ctx context.Context, id string) error

	// Search assets by name, tags, or description
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Asset, int, error)
}

// Repository interface for managing blueprints
type BlueprintRepository interface {
	// Create a new blueprint
	Create(ctx context.Context, bp *models.Blueprint) error

	// GetByID Get blueprint by ID
	GetByID(ctx context.Context, id string) (*models.Blueprint, error)

	// GetByWorkspaceID Get blueprints by workspace ID
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.Blueprint, error)

	// GetAll returns all blueprints (development only)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Blueprint, error)

	// Update a blueprint
	Update(ctx context.Context, bp *models.Blueprint) error

	// Delete a blueprint
	Delete(ctx context.Context, id string) error

	// FindByTags Find blueprints by tag
	FindByTags(ctx context.Context, tags []string) ([]*models.Blueprint, error)

	// CreateVersion  a new version of a blueprint
	CreateVersion(ctx context.Context, blueprintID string, version *models.BlueprintVersion) error

	// GetVersion Get a specific version of a blueprint
	GetVersion(ctx context.Context, blueprintID string, versionNumber int) (*models.BlueprintVersion, error)

	// GetVersions Get all versions of a blueprint
	GetVersions(ctx context.Context, blueprintID string) ([]*models.BlueprintVersion, error)

	// GetReferences Get referenced assets (dependencies)
	GetReferences(ctx context.Context, blueprintID string) ([]*models.AssetReference, error)

	// ToPkgBlueprint Convert database blueprint to package blueprint format
	ToPkgBlueprint(blueprint *models.Blueprint, version *models.BlueprintVersion) (*blueprint.Blueprint, error)

	// FromPkgBlueprint Convert package blueprint to database format
	FromPkgBlueprint(bp *blueprint.Blueprint) (*models.Blueprint, *models.BlueprintVersion, error)
}

// Repository interface for managing workspaces
type WorkspaceRepository interface {
	// Create a new workspace
	Create(ctx context.Context, workspace *models.Workspace) error

	// Get workspace by ID
	GetByID(ctx context.Context, id string) (*models.Workspace, error)

	// Get workspaces by owner ID
	GetByOwnerID(ctx context.Context, ownerType string, ownerID string) ([]*models.Workspace, error)

	// Update a workspace
	Update(ctx context.Context, workspace *models.Workspace) error

	// Delete a workspace
	Delete(ctx context.Context, id string) error

	// Add a member to a workspace
	AddMember(ctx context.Context, workspaceID, userID, role string) error

	// Remove a member from a workspace
	RemoveMember(ctx context.Context, workspaceID, userID string) error

	// Get members of a workspace
	GetMembers(ctx context.Context, workspaceID string) ([]*models.WorkspaceMember, error)

	ToDt(workspace *models.Workspace) *dt.Workspace
}

// Repository interface for managing users
type UserRepository interface {
	// Create a new user
	Create(ctx context.Context, user *models.User) error

	// Get user by ID
	GetByID(ctx context.Context, id string) (*models.User, error)

	// Get user by username
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// Get user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update a user
	Update(ctx context.Context, user *models.User) error

	// Delete a user
	Delete(ctx context.Context, id string) error

	// Update last login time
	UpdateLastLogin(ctx context.Context, id string) error

	ToDt(user *models.User) *dt.User
}

// Repository interface for managing executions
type ExecutionRepository interface {
	// Create a new execution record
	Create(ctx context.Context, execution *models.Execution) error

	// Get execution by ID
	GetByID(ctx context.Context, id string) (*models.Execution, error)

	// Get executions by blueprint ID
	GetByBlueprintID(ctx context.Context, blueprintID string) ([]*models.Execution, error)

	// Update execution status
	UpdateStatus(ctx context.Context, id, status string) error

	// Complete an execution (set completed_at, duration, etc.)
	Complete(ctx context.Context, id string, success bool, result map[string]interface{}, errorMsg string) error

	// Record node execution
	RecordNodeExecution(ctx context.Context, executionID, nodeID, nodeType, execState string, inputs, outputs map[string]interface{}) error

	// Update node execution status
	UpdateNodeStatus(ctx context.Context, executionID, nodeID, status string) error

	// Add execution log entry
	AddLogEntry(ctx context.Context, executionID, nodeID, level, message string, details map[string]interface{}) error

	// Get execution logs
	GetLogs(ctx context.Context, executionID string) ([]*models.ExecutionLog, error)
}

type NodeRepository interface {
	// NodeCreate creates node type reference into database
	NodeCreate(ctx context.Context, nodeType *models.NodeType) error

	// NodeExists check whether node type exists or not
	NodeExists(ctx context.Context, typeId string) (bool, error)

	// NodeGetAll returns all registered core node types
	NodeGetAll(ctx context.Context) ([]*models.NodeType, error)

	// NodeCategoryCreateIfNot gets node category with category name, creates if not exist
	NodeCategoryCreateIfNot(ctx context.Context, category string) (*models.NodeCategory, error)

	ToPkgNode(node *models.NodeType) (node.Node, error)
}

// Repository factory interface for creating repository instances
type RepositoryFactory interface {
	// Get asset repository
	GetAssetRepository() AssetRepository

	// Get blueprint repository
	GetBlueprintRepository() BlueprintRepository

	// Get workspace repository
	GetWorkspaceRepository() WorkspaceRepository

	// Get user repository
	GetUserRepository() UserRepository

	// Get execution repository
	GetExecutionRepository() ExecutionRepository

	// Get node repository
	GetNodeRepository() NodeRepository
}
