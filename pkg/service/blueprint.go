package service

import (
	"context"
	"fmt"
	"time"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// BlueprintService provides high-level operations for managing blueprints
type BlueprintService struct {
	blueprintRepo repository.BlueprintRepository
	workspaceRepo repository.WorkspaceRepository
	assetRepo     repository.AssetRepository
	executionRepo repository.ExecutionRepository
}

// NewBlueprintService creates a new blueprint service
func NewBlueprintService(
	blueprintRepo repository.BlueprintRepository,
	workspaceRepo repository.WorkspaceRepository,
	assetRepo repository.AssetRepository,
	executionRepo repository.ExecutionRepository,
) *BlueprintService {
	return &BlueprintService{
		blueprintRepo: blueprintRepo,
		workspaceRepo: workspaceRepo,
		assetRepo:     assetRepo,
		executionRepo: executionRepo,
	}
}

// CreateBlueprint creates a new blueprint from a package blueprint
func (s *BlueprintService) CreateBlueprint(
	ctx context.Context,
	bp *blueprint.Blueprint,
	workspaceID string,
	userID string,
) (string, error) {
	// First, check if the workspace exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("workspace not found: %w", err)
	}

	// Convert package blueprint to database model
	blueprintModel, versionModel, err := s.blueprintRepo.FromPkgBlueprint(bp)
	if err != nil {
		return "", fmt.Errorf("error converting blueprint: %w", err)
	}

	// Set additional metadata
	blueprintModel.WorkspaceID = workspaceID
	blueprintModel.CreatedBy = userID
	blueprintModel.UpdatedBy = userID
	blueprintModel.CreatedAt = time.Now()
	blueprintModel.UpdatedAt = time.Now()

	// Set version metadata
	versionModel.CreatedBy = userID
	versionModel.VersionNumber = 1
	versionModel.ID = uuid.New().String()

	// Set the current version
	blueprintModel.CurrentVersion = versionModel
	blueprintModel.CurrentVersionID.String = versionModel.ID
	blueprintModel.CurrentVersionID.Valid = true

	// Create the blueprint
	err = s.blueprintRepo.Create(ctx, blueprintModel)
	if err != nil {
		return "", fmt.Errorf("error creating blueprint: %w", err)
	}

	return blueprintModel.ID, nil
}

// GetBlueprint ...
func (s *BlueprintService) GetBlueprint(ctx context.Context, id string) (*blueprint.Blueprint, error) {
	// Get the blueprint model
	blueprintModel, err := s.blueprintRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving blueprint: %w", err)
	}

	// Convert to package blueprint
	bp, err := s.blueprintRepo.ToPkgBlueprint(blueprintModel, blueprintModel.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("error converting blueprint: %w", err)
	}

	return bp, nil
}

// GetBlueprintOrCreate gets a blueprint by ID
func (s *BlueprintService) GetBlueprintOrCreate(ctx context.Context, bp *blueprint.Blueprint, workspaceID, userID string) (*blueprint.Blueprint, error) {
	// Get the blueprint model
	blueprintModel, err := s.blueprintRepo.GetByID(ctx, bp.ID)
	if err != nil {
		bpModel, _, err := s.blueprintRepo.FromPkgBlueprint(bp)
		if err != nil {
			return nil, fmt.Errorf("error converting blueprint: %w", err)
		}

		bpModel.WorkspaceID = workspaceID
		bpModel.CreatedBy = userID
		bpModel.UpdatedBy = userID

		err = s.blueprintRepo.Create(ctx, bpModel)
		if err != nil {
			return nil, fmt.Errorf("error creating blueprint: %w", err)
		}
		blueprintModel = bpModel
	}

	// Convert to package blueprint
	bp, err = s.blueprintRepo.ToPkgBlueprint(blueprintModel, blueprintModel.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("error converting blueprint: %w", err)
	}

	return bp, nil
}

// GetByWorkspace gets a blueprints by workspace id
func (s *BlueprintService) GetByWorkspace(ctx context.Context, id string) ([]*blueprint.Blueprint, error) {
	blueprintModels, err := s.blueprintRepo.GetByWorkspaceID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving blueprints: %w", err)
	}

	var pkgBlueprints = make([]*blueprint.Blueprint, len(blueprintModels))
	for _, blueprintModel := range blueprintModels {
		pkgBlueprint, err := s.blueprintRepo.ToPkgBlueprint(blueprintModel, blueprintModel.CurrentVersion)
		if err != nil {
			return nil, fmt.Errorf("error converting blueprint: %w", err)
		}

		pkgBlueprints = append(pkgBlueprints, pkgBlueprint)
	}

	return pkgBlueprints, nil
}

func (s *BlueprintService) GetAll(ctx context.Context, limit, offset int) ([]*blueprint.Blueprint, error) {
	blueprintModels, err := s.blueprintRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving blueprints: %w", err)
	}

	var blueprints = make([]*blueprint.Blueprint, len(blueprintModels))
	for _, blueprintModel := range blueprintModels {
		pkgBlueprint, err := s.blueprintRepo.ToPkgBlueprint(blueprintModel, blueprintModel.CurrentVersion)
		if err != nil {
			return nil, fmt.Errorf("error converting blueprint: %w", err)
		}

		blueprints = append(blueprints, pkgBlueprint)
	}

	return blueprints, nil
}

// SaveVersion saves a new version of a blueprint
func (s *BlueprintService) SaveVersion(
	ctx context.Context,
	blueprintID string,
	bp *blueprint.Blueprint,
	comment string,
	userID string,
) (int, error) {
	// Get the current blueprint model
	_, err := s.blueprintRepo.GetByID(ctx, blueprintID) // ?
	if err != nil {
		return 0, fmt.Errorf("error retrieving blueprint: %w", err)
	}

	// Convert package blueprint to version model
	_, versionModel, err := s.blueprintRepo.FromPkgBlueprint(bp)
	if err != nil {
		return 0, fmt.Errorf("error converting blueprint: %w", err)
	}

	// Set version metadata
	versionModel.BlueprintID = blueprintID
	versionModel.CreatedBy = userID
	versionModel.ID = uuid.New().String()
	if comment != "" {
		versionModel.Comment = models.NullString(comment)
	}

	// Get the next version number
	versions, err := s.blueprintRepo.GetVersions(ctx, blueprintID)
	if err != nil {
		return 0, fmt.Errorf("error retrieving versions: %w", err)
	}

	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[0].VersionNumber + 1
	}
	versionModel.VersionNumber = nextVersion

	// Create the new version
	err = s.blueprintRepo.CreateVersion(ctx, blueprintID, versionModel)
	if err != nil {
		return 0, fmt.Errorf("error creating version: %w", err)
	}

	return nextVersion, nil
}

// GetVersion gets a specific version of a blueprint
func (s *BlueprintService) GetVersion(
	ctx context.Context,
	blueprintID string,
	versionNumber int,
) (*blueprint.Blueprint, error) {
	// Get the blueprint model
	blueprintModel, err := s.blueprintRepo.GetByID(ctx, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving blueprint: %w", err)
	}

	// Get the specific version
	versionModel, err := s.blueprintRepo.GetVersion(ctx, blueprintID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("error retrieving version: %w", err)
	}

	// Convert to package blueprint
	bp, err := s.blueprintRepo.ToPkgBlueprint(blueprintModel, versionModel)
	if err != nil {
		return nil, fmt.Errorf("error converting blueprint: %w", err)
	}

	return bp, nil
}

// GetVersions gets all versions of a blueprint
func (s *BlueprintService) GetVersions(ctx context.Context, blueprintID string) ([]VersionInfo, error) {
	// Get all versions
	versions, err := s.blueprintRepo.GetVersions(ctx, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving versions: %w", err)
	}

	// Convert to version info
	versionInfos := make([]VersionInfo, len(versions))
	for i, v := range versions {
		versionInfos[i] = VersionInfo{
			VersionNumber: v.VersionNumber,
			CreatedAt:     v.CreatedAt,
			CreatedBy:     v.CreatedBy,
			Comment:       v.Comment.String,
		}
	}

	return versionInfos, nil
}

// DeleteBlueprint deletes a blueprint
func (s *BlueprintService) DeleteBlueprint(ctx context.Context, id string) error {
	return s.blueprintRepo.Delete(ctx, id)
}

// ExecuteBlueprint executes a blueprint
func (s *BlueprintService) ExecuteBlueprint(
	ctx context.Context,
	blueprintID string,
	initialVariables map[string]interface{},
	userID string,
) (string, error) {
	// Create an execution record
	execution := &models.Execution{
		ID:               uuid.New().String(),
		BlueprintID:      blueprintID,
		StartedAt:        time.Now(),
		Status:           "running",
		InitiatedBy:      userID,
		ExecutionMode:    "standard",
		InitialVariables: models.JSONB(initialVariables),
	}

	// Set current version ID if available
	blueprintModel, err := s.blueprintRepo.GetByID(ctx, blueprintID)
	if err != nil {
		return "", fmt.Errorf("error retrieving blueprint: %w", err)
	}

	if blueprintModel.CurrentVersionID.Valid {
		execution.VersionID = models.NullString(blueprintModel.CurrentVersionID.String)
	}

	// Save the execution record
	err = s.executionRepo.Create(ctx, execution)
	if err != nil {
		return "", fmt.Errorf("error creating execution record: %w", err)
	}

	// In a real implementation, we would execute the blueprint here using the engine
	// For this example, we'll just return the execution ID
	return execution.ID, nil
}

// VersionInfo represents metadata about a blueprint version
type VersionInfo struct {
	VersionNumber int       `json:"versionNumber"`
	CreatedAt     time.Time `json:"createdAt"`
	CreatedBy     string    `json:"createdBy"`
	Comment       string    `json:"comment"`
}
