package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
	"webblueprint/pkg/api/dt"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"
)

// WorkspaceService provides high-level operations for managing workspaces
type WorkspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
	assetRepo     repository.AssetRepository
	blueprintRepo repository.BlueprintRepository
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
	assetRepo repository.AssetRepository,
	blueprintRepo repository.BlueprintRepository,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
		assetRepo:     assetRepo,
		blueprintRepo: blueprintRepo,
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(
	ctx context.Context,
	name, description string,
	isPublic bool,
	ownerType string,
	ownerID string,
) (string, error) {
	// Validate owner
	if ownerType == "user" {
		// Check if user exists
		_, err := s.userRepo.GetByID(ctx, ownerID)
		if err != nil {
			return "", fmt.Errorf("user not found: %w", err)
		}
	} else if ownerType == "team" {
		// TODO: Implement team validation when teams are supported
		return "", fmt.Errorf("team workspaces not implemented yet")
	} else {
		return "", fmt.Errorf("invalid owner type: %s", ownerType)
	}

	// Create the workspace
	workspace := &models.Workspace{
		ID:          uuid.New().String(),
		Name:        name,
		Description: models.NullString(description),
		OwnerType:   ownerType,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsPublic:    isPublic,
	}

	// Save to repository
	err := s.workspaceRepo.Create(ctx, workspace)
	if err != nil {
		return "", fmt.Errorf("error creating workspace: %w", err)
	}

	return workspace.ID, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace: %w", err)
	}

	return workspace, nil
}

// GetUserWorkspaces retrieves all workspaces where the user is an owner or member
func (s *WorkspaceService) GetUserWorkspaces(ctx context.Context, userID string) ([]*dt.Workspace, error) {
	// Get workspaces where user is the owner
	ownedWorkspaces, err := s.workspaceRepo.GetByOwnerID(ctx, "user", userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving owned workspaces: %w", err)
	}

	// TODO: Also get workspaces where the user is a member
	// This would require additional repository methods

	var workspaces = make([]*dt.Workspace, 0, len(ownedWorkspaces))
	for _, ownedWorkspace := range ownedWorkspaces {
		workspaces = append(workspaces, s.workspaceRepo.ToDt(ownedWorkspace))
	}

	return workspaces, nil
}

// UpdateWorkspace updates a workspace's information
func (s *WorkspaceService) UpdateWorkspace(
	ctx context.Context,
	id, name, description string,
	isPublic bool,
) error {
	// Get the workspace
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Update fields
	if name != "" {
		workspace.Name = name
	}

	if description != "" {
		workspace.Description = models.NullString(description)
	}

	workspace.IsPublic = isPublic
	workspace.UpdatedAt = time.Now()

	// Save changes
	err = s.workspaceRepo.Update(ctx, workspace)
	if err != nil {
		return fmt.Errorf("error updating workspace: %w", err)
	}

	return nil
}

// DeleteWorkspace deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	// Get the workspace to ensure it exists
	_, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Delete the workspace
	err = s.workspaceRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting workspace: %w", err)
	}

	return nil
}

// GetWorkspaceMembers retrieves all members of a workspace
func (s *WorkspaceService) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*models.WorkspaceMember, error) {
	// Get the workspace to ensure it exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Get members
	members, err := s.workspaceRepo.GetMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace members: %w", err)
	}

	return members, nil
}

// AddWorkspaceMember adds a user to a workspace
func (s *WorkspaceService) AddWorkspaceMember(
	ctx context.Context,
	workspaceID, userID, role string,
) error {
	// Get the workspace to ensure it exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Get the user to ensure they exist
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Set default role if not provided
	if role == "" {
		role = "editor" // Default role
	}

	// Add the member
	err = s.workspaceRepo.AddMember(ctx, workspaceID, userID, role)
	if err != nil {
		return fmt.Errorf("error adding member: %w", err)
	}

	return nil
}

// RemoveWorkspaceMember removes a user from a workspace
func (s *WorkspaceService) RemoveWorkspaceMember(
	ctx context.Context,
	workspaceID, userID string,
) error {
	// Get the workspace to ensure it exists
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Can't remove the owner
	if workspace.OwnerType == "user" && workspace.OwnerID == userID {
		return fmt.Errorf("cannot remove the workspace owner")
	}

	// Remove the member
	err = s.workspaceRepo.RemoveMember(ctx, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("error removing member: %w", err)
	}

	return nil
}

// GetWorkspaceBlueprints retrieves all blueprints in a workspace
func (s *WorkspaceService) GetWorkspaceBlueprints(ctx context.Context, workspaceID string) ([]*models.Blueprint, error) {
	// Get the workspace to ensure it exists
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Get blueprints from repository
	blueprints, err := s.blueprintRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace blueprints: %w", err)
	}

	return blueprints, nil
}
