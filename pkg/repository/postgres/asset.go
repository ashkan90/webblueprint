package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// PostgresAssetRepository implements AssetRepository using PostgreSQL
type PostgresAssetRepository struct {
	db *sql.DB
}

// NewAssetRepository creates a new PostgreSQL-based asset repository
func NewAssetRepository(db *sql.DB) repository.AssetRepository {
	return &PostgresAssetRepository{
		db: db,
	}
}

// Create creates a new asset
func (r *PostgresAssetRepository) Create(ctx context.Context, asset *models.Asset) error {
	// Generate ID if not provided
	if asset.ID == "" {
		asset.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	if asset.CreatedAt.IsZero() {
		asset.CreatedAt = time.Now()
	}
	if asset.UpdatedAt.IsZero() {
		asset.UpdatedAt = asset.CreatedAt
	}

	query := `
		INSERT INTO assets (
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.ID,
		asset.WorkspaceID,
		asset.Name,
		asset.Description,
		asset.Type,
		asset.CreatedAt,
		asset.UpdatedAt,
		asset.CreatedBy,
		asset.UpdatedBy,
		asset.IsPublic,
		asset.Tags,
		asset.ThumbnailURL,
		asset.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	return nil
}

// GetByID retrieves an asset by ID
func (r *PostgresAssetRepository) GetByID(ctx context.Context, id string) (*models.Asset, error) {
	query := `
		SELECT 
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		FROM assets
		WHERE id = $1
	`

	var asset models.Asset
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&asset.ID,
		&asset.WorkspaceID,
		&asset.Name,
		&asset.Description,
		&asset.Type,
		&asset.CreatedAt,
		&asset.UpdatedAt,
		&asset.CreatedBy,
		&asset.UpdatedBy,
		&asset.IsPublic,
		&asset.Tags,
		&asset.ThumbnailURL,
		&asset.Metadata,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("asset not found: %s", id)
		}
		return nil, fmt.Errorf("error retrieving asset: %w", err)
	}

	return &asset, nil
}

// GetByWorkspaceID retrieves all assets in a workspace
func (r *PostgresAssetRepository) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.Asset, error) {
	query := `
		SELECT 
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		FROM assets
		WHERE workspace_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("error querying assets: %w", err)
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.WorkspaceID,
			&asset.Name,
			&asset.Description,
			&asset.Type,
			&asset.CreatedAt,
			&asset.UpdatedAt,
			&asset.CreatedBy,
			&asset.UpdatedBy,
			&asset.IsPublic,
			&asset.Tags,
			&asset.ThumbnailURL,
			&asset.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning asset row: %w", err)
		}
		assets = append(assets, &asset)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating asset rows: %w", err)
	}

	return assets, nil
}

// GetByType retrieves assets by type
func (r *PostgresAssetRepository) GetByType(ctx context.Context, assetType string) ([]*models.Asset, error) {
	query := `
		SELECT 
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		FROM assets
		WHERE type = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, assetType)
	if err != nil {
		return nil, fmt.Errorf("error querying assets by type: %w", err)
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.WorkspaceID,
			&asset.Name,
			&asset.Description,
			&asset.Type,
			&asset.CreatedAt,
			&asset.UpdatedAt,
			&asset.CreatedBy,
			&asset.UpdatedBy,
			&asset.IsPublic,
			&asset.Tags,
			&asset.ThumbnailURL,
			&asset.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning asset row: %w", err)
		}
		assets = append(assets, &asset)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating asset rows: %w", err)
	}

	return assets, nil
}

// Update updates an asset
func (r *PostgresAssetRepository) Update(ctx context.Context, asset *models.Asset) error {
	// Always update the updated_at timestamp
	asset.UpdatedAt = time.Now()

	query := `
		UPDATE assets
		SET name = $1, description = $2, updated_at = $3, updated_by = $4,
			is_public = $5, tags = $6, thumbnail_url = $7, metadata = $8
		WHERE id = $9
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		asset.Name,
		asset.Description,
		asset.UpdatedAt,
		asset.UpdatedBy,
		asset.IsPublic,
		asset.Tags,
		asset.ThumbnailURL,
		asset.Metadata,
		asset.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not found: %s", asset.ID)
	}

	return nil
}

// Delete deletes an asset by ID
func (r *PostgresAssetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM assets WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not found: %s", id)
	}

	return nil
}

// Search searches for assets by name, tags, or description
func (r *PostgresAssetRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Asset, int, error) {
	// Prepare the search terms
	searchTerms := "%" + strings.ToLower(query) + "%"

	// Count total matches query
	countQuery := `
		SELECT COUNT(*)
		FROM assets
		WHERE LOWER(name) LIKE $1
		   OR LOWER(description) LIKE $1
		   OR EXISTS (
			  SELECT 1 FROM unnest(tags) tag
			  WHERE LOWER(tag) LIKE $1
		   )
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, searchTerms).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting search results: %w", err)
	}

	// Search query with pagination
	searchQuery := `
		SELECT 
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		FROM assets
		WHERE LOWER(name) LIKE $1
		   OR LOWER(description) LIKE $1
		   OR EXISTS (
			  SELECT 1 FROM unnest(tags) tag
			  WHERE LOWER(tag) LIKE $1
		   )
		ORDER BY 
			CASE WHEN LOWER(name) LIKE $1 THEN 0
				 WHEN EXISTS (SELECT 1 FROM unnest(tags) tag WHERE LOWER(tag) LIKE $1) THEN 1
				 ELSE 2
			END,
			name
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerms, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error searching assets: %w", err)
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.WorkspaceID,
			&asset.Name,
			&asset.Description,
			&asset.Type,
			&asset.CreatedAt,
			&asset.UpdatedAt,
			&asset.CreatedBy,
			&asset.UpdatedBy,
			&asset.IsPublic,
			&asset.Tags,
			&asset.ThumbnailURL,
			&asset.Metadata,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning asset row: %w", err)
		}
		assets = append(assets, &asset)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating asset rows: %w", err)
	}

	return assets, total, nil
}
