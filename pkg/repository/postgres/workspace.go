package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"webblueprint/pkg/api/dt"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// PostgresWorkspaceRepository implements WorkspaceRepository using PostgreSQL
type PostgresWorkspaceRepository struct {
	db *sql.DB
}

// NewWorkspaceRepository creates a new PostgreSQL-based workspace repository
func NewWorkspaceRepository(db *sql.DB) repository.WorkspaceRepository {
	return &PostgresWorkspaceRepository{
		db: db,
	}
}

// Create creates a new workspace
func (r *PostgresWorkspaceRepository) Create(ctx context.Context, workspace *models.Workspace) error {
	// Generate ID if not provided
	if workspace.ID == "" {
		workspace.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	if workspace.CreatedAt.IsZero() {
		workspace.CreatedAt = time.Now()
	}
	if workspace.UpdatedAt.IsZero() {
		workspace.UpdatedAt = workspace.CreatedAt
	}

	query := `
		INSERT INTO workspaces (
			id, name, description, owner_type, owner_id, created_at, updated_at,
			is_public, thumbnail_url, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		workspace.ID,
		workspace.Name,
		workspace.Description,
		workspace.OwnerType,
		workspace.OwnerID,
		workspace.CreatedAt,
		workspace.UpdatedAt,
		workspace.IsPublic,
		workspace.ThumbnailURL,
		workspace.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

// GetByID retrieves a workspace by ID
func (r *PostgresWorkspaceRepository) GetByID(ctx context.Context, id string) (*models.Workspace, error) {
	query := `
		SELECT 
			id, name, description, owner_type, owner_id, created_at, updated_at,
			is_public, thumbnail_url, metadata
		FROM workspaces
		WHERE id = $1
	`

	var workspace models.Workspace
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.Description,
		&workspace.OwnerType,
		&workspace.OwnerID,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
		&workspace.IsPublic,
		&workspace.ThumbnailURL,
		&workspace.Metadata,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("workspace not found: %s", id)
		}
		return nil, fmt.Errorf("error retrieving workspace: %w", err)
	}

	return &workspace, nil
}

// GetByOwnerID retrieves workspaces by owner ID
func (r *PostgresWorkspaceRepository) GetByOwnerID(ctx context.Context, ownerType string, ownerID string) ([]*models.Workspace, error) {
	query := `
		SELECT 
			id, name, description, owner_type, owner_id, created_at, updated_at,
			is_public, thumbnail_url, metadata
		FROM workspaces
		WHERE owner_type = $1 AND owner_id = $2
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, ownerType, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error querying workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces = make([]*models.Workspace, 0)
	for rows.Next() {
		var workspace models.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.Description,
			&workspace.OwnerType,
			&workspace.OwnerID,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&workspace.IsPublic,
			&workspace.ThumbnailURL,
			&workspace.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning workspace row: %w", err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspace rows: %w", err)
	}

	return workspaces, nil
}

// Update updates a workspace
func (r *PostgresWorkspaceRepository) Update(ctx context.Context, workspace *models.Workspace) error {
	// Always update the updated_at timestamp
	workspace.UpdatedAt = time.Now()

	query := `
		UPDATE workspaces
		SET name = $1, description = $2, updated_at = $3, is_public = $4, 
		    thumbnail_url = $5, metadata = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		workspace.Name,
		workspace.Description,
		workspace.UpdatedAt,
		workspace.IsPublic,
		workspace.ThumbnailURL,
		workspace.Metadata,
		workspace.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found: %s", workspace.ID)
	}

	return nil
}

// Delete deletes a workspace by ID
func (r *PostgresWorkspaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workspaces WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace not found: %s", id)
	}

	return nil
}

// AddMember adds a user to a workspace
func (r *PostgresWorkspaceRepository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	query := `
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, user_id) 
		DO UPDATE SET role = $3
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		workspaceID,
		userID,
		role,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add member to workspace: %w", err)
	}

	return nil
}

// RemoveMember removes a user from a workspace
func (r *PostgresWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	query := `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member from workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found in workspace: %s, %s", workspaceID, userID)
	}

	return nil
}

// GetMembers gets all members of a workspace
func (r *PostgresWorkspaceRepository) GetMembers(ctx context.Context, workspaceID string) ([]*models.WorkspaceMember, error) {
	query := `
		SELECT 
			wm.workspace_id, wm.user_id, wm.role, wm.joined_at,
			u.username, u.email, u.full_name, u.avatar_url
		FROM workspace_members wm
		JOIN users u ON wm.user_id = u.id
		WHERE wm.workspace_id = $1
		ORDER BY wm.role, u.username
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("error querying workspace members: %w", err)
	}
	defer rows.Close()

	var members []*models.WorkspaceMember
	for rows.Next() {
		var member models.WorkspaceMember
		var user models.User

		err := rows.Scan(
			&member.WorkspaceID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&user.Username,
			&user.Email,
			&user.FullName,
			&user.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning workspace member row: %w", err)
		}

		// Set the related user
		user.ID = member.UserID
		member.User = &user

		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspace member rows: %w", err)
	}

	return members, nil
}

func (r *PostgresWorkspaceRepository) ToDt(workspace *models.Workspace) *dt.Workspace {
	return &dt.Workspace{
		ID:           workspace.ID,
		Name:         workspace.Name,
		Description:  workspace.Description.String,
		OwnerType:    workspace.OwnerType,
		OwnerID:      workspace.OwnerID,
		CreatedAt:    workspace.CreatedAt,
		UpdatedAt:    workspace.UpdatedAt,
		IsPublic:     workspace.IsPublic,
		ThumbnailURL: workspace.ThumbnailURL.String,
		Metadata:     workspace.Metadata,
	}
}
