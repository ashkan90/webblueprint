package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// PostgresBlueprintRepository implements BlueprintRepository using PostgreSQL
type PostgresBlueprintRepository struct {
	db *sql.DB
}

// NewBlueprintRepository creates a new PostgreSQL-based blueprint repository
func NewBlueprintRepository(db *sql.DB) repository.BlueprintRepository {
	return &PostgresBlueprintRepository{
		db: db,
	}
}

// Create creates a new blueprint and its initial version
func (r *PostgresBlueprintRepository) Create(ctx context.Context, bp *models.Blueprint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate IDs if not provided
	if bp.ID == "" {
		bp.ID = uuid.New().String()
	}

	// First create the asset record
	assetQuery := `
		INSERT INTO assets (
			id, workspace_id, name, description, type, created_at, updated_at,
			created_by, updated_by, is_public, tags, thumbnail_url, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = tx.ExecContext(
		ctx,
		assetQuery,
		bp.ID,
		bp.WorkspaceID,
		bp.Name,
		bp.Description,
		"blueprint", // Asset type is blueprint
		time.Now(),
		time.Now(),
		bp.CreatedBy,
		bp.UpdatedBy,
		bp.IsPublic,
		bp.Tags,
		bp.ThumbnailURL,
		bp.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create asset record: %w", err)
	}

	// Then create the blueprint record
	blueprintQuery := `
		INSERT INTO blueprints (
			id, node_count, connection_count, entry_points, is_template, category
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(
		ctx,
		blueprintQuery,
		bp.ID,
		bp.NodeCount,
		bp.ConnectionCount,
		bp.EntryPoints,
		bp.IsTemplate,
		bp.Category,
	)
	if err != nil {
		return fmt.Errorf("failed to create blueprint record: %w", err)
	}

	// If there's a current version, create it
	if bp.CurrentVersion != nil {
		versionID := uuid.New().String()
		bp.CurrentVersion.ID = versionID
		bp.CurrentVersion.BlueprintID = bp.ID

		versionQuery := `
			INSERT INTO blueprint_versions (
				id, blueprint_id, version_number, created_at, created_by,
				comment, nodes, connections, variables, functions, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.ExecContext(
			ctx,
			versionQuery,
			versionID,
			bp.ID,
			bp.CurrentVersion.VersionNumber,
			time.Now(),
			bp.CurrentVersion.CreatedBy,
			bp.CurrentVersion.Comment,
			bp.CurrentVersion.Nodes,
			bp.CurrentVersion.Connections,
			bp.CurrentVersion.Variables,
			bp.CurrentVersion.Functions,
			bp.CurrentVersion.Metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to create blueprint version: %w", err)
		}

		// Update the blueprint with the current version ID
		updateQuery := `
			UPDATE blueprints SET current_version_id = $1 WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, updateQuery, versionID, bp.ID)
		if err != nil {
			return fmt.Errorf("failed to update blueprint with current version: %w", err)
		}

		bp.CurrentVersionID = sql.NullString{String: versionID, Valid: true}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a blueprint by its ID
func (r *PostgresBlueprintRepository) GetByID(ctx context.Context, id string) (*models.Blueprint, error) {
	query := `
		SELECT 
			a.id, a.workspace_id, a.name, a.description, a.created_at, a.updated_at,
			a.created_by, a.updated_by, a.is_public, a.tags, a.thumbnail_url, a.metadata,
			b.current_version_id, b.node_count, b.connection_count, b.entry_points, b.is_template, b.category
		FROM blueprints b
		JOIN assets a ON b.id = a.id
		WHERE b.id = $1
	`

	var bp models.Blueprint
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&bp.ID,
		&bp.WorkspaceID,
		&bp.Name,
		&bp.Description,
		&bp.CreatedAt,
		&bp.UpdatedAt,
		&bp.CreatedBy,
		&bp.UpdatedBy,
		&bp.IsPublic,
		&bp.Tags,
		&bp.ThumbnailURL,
		&bp.Metadata,
		&bp.CurrentVersionID,
		&bp.NodeCount,
		&bp.ConnectionCount,
		&bp.EntryPoints,
		&bp.IsTemplate,
		&bp.Category,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("blueprint not found: %s", id)
		}
		return nil, fmt.Errorf("error retrieving blueprint: %w", err)
	}

	// If there's a current version ID, get the current version
	if bp.CurrentVersionID.Valid {
		versionQuery := `
			SELECT 
				id, blueprint_id, version_number, created_at, created_by,
				comment, nodes, connections, variables, functions, metadata
			FROM blueprint_versions
			WHERE id = $1
		`

		var version models.BlueprintVersion
		err = r.db.QueryRowContext(ctx, versionQuery, bp.CurrentVersionID.String).Scan(
			&version.ID,
			&version.BlueprintID,
			&version.VersionNumber,
			&version.CreatedAt,
			&version.CreatedBy,
			&version.Comment,
			&version.Nodes,
			&version.Connections,
			&version.Variables,
			&version.Functions,
			&version.Metadata,
		)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error retrieving current version: %w", err)
		}

		if err == nil {
			bp.CurrentVersion = &version
		}
	}

	return &bp, nil
}

// GetByWorkspaceID retrieves all blueprints in a workspace
func (r *PostgresBlueprintRepository) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.Blueprint, error) {
	query := `
		SELECT 
			a.id, a.workspace_id, a.name, a.description, a.created_at, a.updated_at,
			a.created_by, a.updated_by, a.is_public, a.tags, a.thumbnail_url, a.metadata,
			b.current_version_id, b.node_count, b.connection_count, b.entry_points, b.is_template, b.category
		FROM blueprints b
		JOIN assets a ON b.id = a.id
		WHERE a.workspace_id = $1
		ORDER BY a.name
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("error querying blueprints: %w", err)
	}
	defer rows.Close()

	var blueprints []*models.Blueprint
	for rows.Next() {
		var bp models.Blueprint
		err := rows.Scan(
			&bp.ID,
			&bp.WorkspaceID,
			&bp.Name,
			&bp.Description,
			&bp.CreatedAt,
			&bp.UpdatedAt,
			&bp.CreatedBy,
			&bp.UpdatedBy,
			&bp.IsPublic,
			&bp.Tags,
			&bp.ThumbnailURL,
			&bp.Metadata,
			&bp.CurrentVersionID,
			&bp.NodeCount,
			&bp.ConnectionCount,
			&bp.EntryPoints,
			&bp.IsTemplate,
			&bp.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning blueprint row: %w", err)
		}
		blueprints = append(blueprints, &bp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blueprint rows: %w", err)
	}

	return blueprints, nil
}

func (r *PostgresBlueprintRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Blueprint, error) {
	query := `
		SELECT 
			a.id, a.workspace_id, a.name, a.description, a.created_at, a.updated_at,
			a.created_by, a.updated_by, a.is_public, a.tags, a.thumbnail_url, a.metadata,
			b.current_version_id, b.node_count, b.connection_count, b.entry_points, b.is_template, b.category
		FROM blueprints b
		JOIN assets a ON b.id = a.id
		ORDER BY a.created_at
		LIMIT $1
		OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no blueprint found")
		}
		return nil, fmt.Errorf("error retrieving blueprint: %w", err)
	}

	defer rows.Close()

	var blueprints []*models.Blueprint
	for rows.Next() {
		var bp models.Blueprint
		err := rows.Scan(
			&bp.ID,
			&bp.WorkspaceID,
			&bp.Name,
			&bp.Description,
			&bp.CreatedAt,
			&bp.UpdatedAt,
			&bp.CreatedBy,
			&bp.UpdatedBy,
			&bp.IsPublic,
			&bp.Tags,
			&bp.ThumbnailURL,
			&bp.Metadata,
			&bp.CurrentVersionID,
			&bp.NodeCount,
			&bp.ConnectionCount,
			&bp.EntryPoints,
			&bp.IsTemplate,
			&bp.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning blueprint row: %w", err)
		}
		blueprints = append(blueprints, &bp)
	}

	// If there's a current version ID, get the current version
	for _, bp := range blueprints {
		if bp.CurrentVersionID.Valid {
			versionQuery := `
			SELECT 
				id, blueprint_id, version_number, created_at, created_by,
				comment, nodes, connections, variables, functions, metadata
			FROM blueprint_versions
			WHERE id = $1
		`

			var version models.BlueprintVersion
			err = r.db.QueryRowContext(ctx, versionQuery, bp.CurrentVersionID.String).Scan(
				&version.ID,
				&version.BlueprintID,
				&version.VersionNumber,
				&version.CreatedAt,
				&version.CreatedBy,
				&version.Comment,
				&version.Nodes,
				&version.Connections,
				&version.Variables,
				&version.Functions,
				&version.Metadata,
			)

			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error retrieving current version: %w", err)
			}

			if err == nil {
				bp.CurrentVersion = &version
			}
		}
	}

	return blueprints, nil
}

// Update updates a blueprint's metadata (not its version content)
func (r *PostgresBlueprintRepository) Update(ctx context.Context, bp *models.Blueprint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the asset record
	assetQuery := `
		UPDATE assets
		SET name = $1, description = $2, updated_at = $3, updated_by = $4,
		    is_public = $5, tags = $6, thumbnail_url = $7, metadata = $8
		WHERE id = $9
	`
	_, err = tx.ExecContext(
		ctx,
		assetQuery,
		bp.Name,
		bp.Description,
		time.Now(),
		bp.UpdatedBy,
		bp.IsPublic,
		bp.Tags,
		bp.ThumbnailURL,
		bp.Metadata,
		bp.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset record: %w", err)
	}

	// Update the blueprint record
	blueprintQuery := `
		UPDATE blueprints
		SET is_template = $1, category = $2
		WHERE id = $3
	`
	_, err = tx.ExecContext(
		ctx,
		blueprintQuery,
		bp.IsTemplate,
		bp.Category,
		bp.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update blueprint record: %w", err)
	}

	// If there's an updated current version ID, update it
	if bp.CurrentVersionID.Valid {
		updateQuery := `
			UPDATE blueprints SET current_version_id = $1 WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, updateQuery, bp.CurrentVersionID.String, bp.ID)
		if err != nil {
			return fmt.Errorf("failed to update current version reference: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a blueprint and all its versions
func (r *PostgresBlueprintRepository) Delete(ctx context.Context, id string) error {
	// The asset deletion will cascade to the blueprint and versions due to foreign key constraints
	query := `DELETE FROM assets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete blueprint: %w", err)
	}
	return nil
}

// FindByTags finds blueprints by tags
func (r *PostgresBlueprintRepository) FindByTags(ctx context.Context, tags []string) ([]*models.Blueprint, error) {
	if len(tags) == 0 {
		return nil, errors.New("at least one tag must be provided")
	}

	// Build the query with a variable number of tags
	query := `
		SELECT 
			a.id, a.workspace_id, a.name, a.description, a.created_at, a.updated_at,
			a.created_by, a.updated_by, a.is_public, a.tags, a.thumbnail_url, a.metadata,
			b.current_version_id, b.node_count, b.connection_count, b.entry_points, b.is_template, b.category
		FROM blueprints b
		JOIN assets a ON b.id = a.id
		WHERE a.type = 'blueprint' 
	` // @TODO missing query

	// Build the tag conditions: a.tags @> ARRAY['tag1'] OR a.tags @> ARRAY['tag2'] ...
	conditions := make([]string, len(tags))
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		conditions[i] = fmt.Sprintf("a.tags @> $%d", i+1)
		args[i] = []string{tag} // PostgreSQL array syntax for a single element
	}

	query += strings.Join(conditions, " OR ") + ")"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying blueprints by tags: %w", err)
	}
	defer rows.Close()

	var blueprints []*models.Blueprint
	for rows.Next() {
		var bp models.Blueprint
		err := rows.Scan(
			&bp.ID,
			&bp.WorkspaceID,
			&bp.Name,
			&bp.Description,
			&bp.CreatedAt,
			&bp.UpdatedAt,
			&bp.CreatedBy,
			&bp.UpdatedBy,
			&bp.IsPublic,
			&bp.Tags,
			&bp.ThumbnailURL,
			&bp.Metadata,
			&bp.CurrentVersionID,
			&bp.NodeCount,
			&bp.ConnectionCount,
			&bp.EntryPoints,
			&bp.IsTemplate,
			&bp.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning blueprint row: %w", err)
		}
		blueprints = append(blueprints, &bp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blueprint rows: %w", err)
	}

	return blueprints, nil
}

// CreateVersion creates a new version of a blueprint
func (r *PostgresBlueprintRepository) CreateVersion(ctx context.Context, blueprintID string, version *models.BlueprintVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate ID if not provided
	if version.ID == "" {
		version.ID = uuid.New().String()
	}

	// Make sure blueprint exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM blueprints WHERE id = $1)", blueprintID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking blueprint existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("blueprint not found: %s", blueprintID)
	}

	// Get the next version number if not specified
	if version.VersionNumber <= 0 {
		var maxVersion int
		err = tx.QueryRowContext(
			ctx,
			"SELECT COALESCE(MAX(version_number), 0) FROM blueprint_versions WHERE blueprint_id = $1",
			blueprintID,
		).Scan(&maxVersion)
		if err != nil {
			return fmt.Errorf("error getting max version number: %w", err)
		}
		version.VersionNumber = maxVersion + 1
	}

	// Create the new version
	version.BlueprintID = blueprintID
	versionQuery := `
		INSERT INTO blueprint_versions (
			id, blueprint_id, version_number, created_at, created_by,
			comment, nodes, connections, variables, functions, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(
		ctx,
		versionQuery,
		version.ID,
		version.BlueprintID,
		version.VersionNumber,
		time.Now(),
		version.CreatedBy,
		version.Comment,
		version.Nodes,
		version.Connections,
		version.Variables,
		version.Functions,
		version.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create blueprint version: %w", err)
	}

	// Update the blueprint with the new version and counts
	updateQuery := `
		UPDATE blueprints 
		SET current_version_id = $1,
		    node_count = $2,
		    connection_count = $3
		WHERE id = $4
	`
	nodeCount := 0
	connectionCount := 0
	if version.Nodes != nil {
		nodeCount = len(version.Nodes)
	}
	if version.Connections != nil {
		connectionCount = len(version.Connections)
	}

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		version.ID,
		nodeCount,
		connectionCount,
		blueprintID,
	)
	if err != nil {
		return fmt.Errorf("failed to update blueprint with new version: %w", err)
	}

	// Update the asset's updated_at timestamp
	_, err = tx.ExecContext(
		ctx,
		"UPDATE assets SET updated_at = $1 WHERE id = $2",
		time.Now(),
		blueprintID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset timestamp: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetVersion gets a specific version of a blueprint
func (r *PostgresBlueprintRepository) GetVersion(ctx context.Context, blueprintID string, versionNumber int) (*models.BlueprintVersion, error) {
	query := `
		SELECT 
			id, blueprint_id, version_number, created_at, created_by,
			comment, nodes, connections, variables, functions, metadata
		FROM blueprint_versions
		WHERE blueprint_id = $1 AND version_number = $2
	`

	var version models.BlueprintVersion
	err := r.db.QueryRowContext(ctx, query, blueprintID, versionNumber).Scan(
		&version.ID,
		&version.BlueprintID,
		&version.VersionNumber,
		&version.CreatedAt,
		&version.CreatedBy,
		&version.Comment,
		&version.Nodes,
		&version.Connections,
		&version.Variables,
		&version.Functions,
		&version.Metadata,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("blueprint version not found: %s version %d", blueprintID, versionNumber)
		}
		return nil, fmt.Errorf("error retrieving blueprint version: %w", err)
	}

	return &version, nil
}

// GetVersions gets all versions of a blueprint
func (r *PostgresBlueprintRepository) GetVersions(ctx context.Context, blueprintID string) ([]*models.BlueprintVersion, error) {
	query := `
		SELECT 
			id, blueprint_id, version_number, created_at, created_by,
			comment, metadata
		FROM blueprint_versions
		WHERE blueprint_id = $1
		ORDER BY version_number DESC
	`

	rows, err := r.db.QueryContext(ctx, query, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error querying blueprint versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.BlueprintVersion
	for rows.Next() {
		var version models.BlueprintVersion
		// Note: Not loading the full nodes/connections/variables arrays to save memory
		err := rows.Scan(
			&version.ID,
			&version.BlueprintID,
			&version.VersionNumber,
			&version.CreatedAt,
			&version.CreatedBy,
			&version.Comment,
			&version.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning blueprint version row: %w", err)
		}
		versions = append(versions, &version)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blueprint version rows: %w", err)
	}

	return versions, nil
}

// GetReferences gets all referenced assets (dependencies) of a blueprint
func (r *PostgresBlueprintRepository) GetReferences(ctx context.Context, blueprintID string) ([]*models.AssetReference, error) {
	query := `
		SELECT 
			source_asset_id, target_asset_id, reference_type, reference_count, details
		FROM asset_references
		WHERE source_asset_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error querying asset references: %w", err)
	}
	defer rows.Close()

	var references []*models.AssetReference
	for rows.Next() {
		var ref models.AssetReference
		err := rows.Scan(
			&ref.SourceAssetID,
			&ref.TargetAssetID,
			&ref.ReferenceType,
			&ref.ReferenceCount,
			&ref.Details,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning asset reference row: %w", err)
		}
		references = append(references, &ref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating asset reference rows: %w", err)
	}

	return references, nil
}

// ToPkgBlueprint converts a database blueprint model to a package blueprint
func (r *PostgresBlueprintRepository) ToPkgBlueprint(blueprintModel *models.Blueprint, versionModel *models.BlueprintVersion) (*blueprint.Blueprint, error) {
	if blueprintModel == nil {
		return nil, errors.New("blueprint model cannot be nil")
	}

	// Create a new package blueprint
	bp := blueprint.NewBlueprint(blueprintModel.ID, blueprintModel.Name, "1.0.0") // Version might need to be set differently

	// Set description if provided
	if blueprintModel.Description.Valid {
		bp.Description = blueprintModel.Description.String
	}

	// Convert metadata if present
	if blueprintModel.Metadata != nil {
		bp.Metadata = make(map[string]string)
		for k, v := range blueprintModel.Metadata {
			if strVal, ok := v.(string); ok {
				bp.Metadata[k] = strVal
			} else {
				// Try to convert other types to string
				bp.Metadata[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// If we have a version model, populate nodes and connections
	if versionModel != nil {
		// Convert nodes
		for _, nodeData := range versionModel.Nodes {
			var node blueprint.BlueprintNode
			nodeMap, ok := nodeData.(map[string]interface{})
			if !ok {
				continue // Skip invalid nodes
			}

			// Extract basic node information
			if id, ok := nodeMap["id"].(string); ok {
				node.ID = id
			}
			if nodeType, ok := nodeMap["type"].(string); ok {
				node.Type = nodeType
			}

			// Extract position
			if posMap, ok := nodeMap["position"].(map[string]interface{}); ok {
				var x, y float64
				if xVal, ok := posMap["x"].(float64); ok {
					x = xVal
				}
				if yVal, ok := posMap["y"].(float64); ok {
					y = yVal
				}
				node.Position = blueprint.Position{X: x, Y: y}
			}

			// Extract properties
			if propsArray, ok := nodeMap["properties"].([]interface{}); ok {
				for _, propData := range propsArray {
					propMap, ok := propData.(map[string]interface{})
					if !ok {
						continue
					}

					var prop blueprint.NodeProperty
					if name, ok := propMap["name"].(string); ok {
						prop.Name = name
					}
					if description, ok := propMap["description"].(string); ok {
						prop.Description = description
					}
					if typeMap, ok := propMap["type"].(map[string]interface{}); ok {
						var pinType blueprint.NodePinType
						if typeID, ok := typeMap["id"].(string); ok {
							pinType.ID = typeID
						}
						if typeName, ok := typeMap["name"].(string); ok {
							pinType.Name = typeName
						}
						if typeDesc, ok := typeMap["description"].(string); ok {
							pinType.Description = typeDesc
						}
						prop.Type = &pinType
					}
					if value, ok := propMap["value"]; ok {
						prop.Value = value
					}
					node.Properties = append(node.Properties, prop)
				}
			}

			// Extract data (general purpose map)
			if dataMap, ok := nodeMap["data"].(map[string]interface{}); ok {
				node.Data = dataMap
			}

			bp.AddNode(node)
		}

		// Convert connections
		for _, connData := range versionModel.Connections {
			var conn blueprint.Connection
			connMap, ok := connData.(map[string]interface{})
			if !ok {
				continue // Skip invalid connections
			}

			// Extract connection information
			if id, ok := connMap["id"].(string); ok {
				conn.ID = id
			}
			if sourceNodeID, ok := connMap["source_node"].(string); ok {
				conn.SourceNodeID = sourceNodeID
			}
			if sourcePinID, ok := connMap["source_pin"].(string); ok {
				conn.SourcePinID = sourcePinID
			}
			if targetNodeID, ok := connMap["target_node"].(string); ok {
				conn.TargetNodeID = targetNodeID
			}
			if targetPinID, ok := connMap["target_pin"].(string); ok {
				conn.TargetPinID = targetPinID
			}
			if connType, ok := connMap["connection_type"].(string); ok {
				conn.ConnectionType = connType
			} else {
				// Default to "data" if not specified
				conn.ConnectionType = "data"
			}

			// Extract data (general purpose map)
			if dataMap, ok := connMap["data"].(map[string]interface{}); ok {
				conn.Data = dataMap
			} else {
				conn.Data = make(map[string]any)
			}

			bp.AddConnection(conn)
		}

		// Convert variables
		for _, varData := range versionModel.Variables {
			varMap, ok := varData.(map[string]interface{})
			if !ok {
				continue // Skip invalid variables
			}

			var variable blueprint.Variable
			if name, ok := varMap["name"].(string); ok {
				variable.Name = name
			}
			if varType, ok := varMap["type"].(string); ok {
				variable.Type = varType
			}
			if value, ok := varMap["value"]; ok {
				variable.Value = value
			}

			bp.AddVariable(variable)
		}

		// Convert functions
		for _, funcData := range versionModel.Functions {
			funcMap, ok := funcData.(map[string]interface{})
			if !ok {
				continue // Skip invalid functions
			}

			var function blueprint.Function
			if id, ok := funcMap["id"].(string); ok {
				function.ID = id
			}
			if name, ok := funcMap["name"].(string); ok {
				function.Name = name
			}
			if desc, ok := funcMap["description"].(string); ok {
				function.Description = desc
			}

			// Extract node type data
			if nodeTypeMap, ok := funcMap["node_type"].(map[string]interface{}); ok {
				// Convert node type fields based on the Function struct in the blueprint package
				// This is a simplified version and may need to be expanded

				// Extract inputs
				if inputs, ok := nodeTypeMap["inputs"].([]interface{}); ok {
					for _, inputData := range inputs {
						inputMap, ok := inputData.(map[string]interface{})
						if !ok {
							continue
						}
						var pin blueprint.NodePin
						if id, ok := inputMap["id"].(string); ok {
							pin.ID = id
						}
						if name, ok := inputMap["name"].(string); ok {
							pin.Name = name
						}
						if desc, ok := inputMap["description"].(string); ok {
							pin.Description = desc
						}

						// Extract pin type
						if typeMap, ok := inputMap["type"].(map[string]interface{}); ok {
							var pinType blueprint.NodePinType
							if typeID, ok := typeMap["id"].(string); ok {
								pinType.ID = typeID
							}
							if typeName, ok := typeMap["name"].(string); ok {
								pinType.Name = typeName
							}
							if typeDesc, ok := typeMap["description"].(string); ok {
								pinType.Description = typeDesc
							}
							pin.Type = &pinType
						}

						if optional, ok := inputMap["optional"].(bool); ok {
							pin.Optional = optional
						}
						if defaultVal, ok := inputMap["default"]; ok {
							pin.Default = defaultVal
						}

						function.NodeType.Inputs = append(function.NodeType.Inputs, pin)
					}
				}

				// Extract outputs (similar to inputs)
				if outputs, ok := nodeTypeMap["outputs"].([]interface{}); ok {
					for _, outputData := range outputs {
						outputMap, ok := outputData.(map[string]interface{})
						if !ok {
							continue
						}
						var pin blueprint.NodePin
						if id, ok := outputMap["id"].(string); ok {
							pin.ID = id
						}
						if name, ok := outputMap["name"].(string); ok {
							pin.Name = name
						}
						if desc, ok := outputMap["description"].(string); ok {
							pin.Description = desc
						}

						// Extract pin type
						if typeMap, ok := outputMap["type"].(map[string]interface{}); ok {
							var pinType blueprint.NodePinType
							if typeID, ok := typeMap["id"].(string); ok {
								pinType.ID = typeID
							}
							if typeName, ok := typeMap["name"].(string); ok {
								pinType.Name = typeName
							}
							if typeDesc, ok := typeMap["description"].(string); ok {
								pinType.Description = typeDesc
							}
							pin.Type = &pinType
						}

						if optional, ok := outputMap["optional"].(bool); ok {
							pin.Optional = optional
						}
						if defaultVal, ok := outputMap["default"]; ok {
							pin.Default = defaultVal
						}

						function.NodeType.Outputs = append(function.NodeType.Outputs, pin)
					}
				}
			}

			// Extract function nodes and connections
			if nodesArray, ok := funcMap["nodes"].([]interface{}); ok {
				for _, nodeData := range nodesArray {
					nodeMap, ok := nodeData.(map[string]interface{})
					if !ok {
						continue
					}
					var node blueprint.BlueprintNode
					if id, ok := nodeMap["id"].(string); ok {
						node.ID = id
					}
					if nodeType, ok := nodeMap["type"].(string); ok {
						node.Type = nodeType
					}
					// Extract position and properties similar to blueprint nodes
					function.Nodes = append(function.Nodes, node)
				}
			}

			if connsArray, ok := funcMap["connections"].([]interface{}); ok {
				for _, connData := range connsArray {
					connMap, ok := connData.(map[string]interface{})
					if !ok {
						continue
					}
					var conn blueprint.Connection
					if id, ok := connMap["id"].(string); ok {
						conn.ID = id
					}
					if sourceNodeID, ok := connMap["source_node_id"].(string); ok {
						conn.SourceNodeID = sourceNodeID
					}
					if sourcePinID, ok := connMap["source_pin_id"].(string); ok {
						conn.SourcePinID = sourcePinID
					}
					if targetNodeID, ok := connMap["target_node_id"].(string); ok {
						conn.TargetNodeID = targetNodeID
					}
					if targetPinID, ok := connMap["target_pin_id"].(string); ok {
						conn.TargetPinID = targetPinID
					}
					if connType, ok := connMap["connection_type"].(string); ok {
						conn.ConnectionType = connType
					}
					function.Connections = append(function.Connections, conn)
				}
			}

			// Add metadata and variables
			if metaMap, ok := funcMap["metadata"].(map[string]interface{}); ok {
				function.Metadata = make(map[string]string)
				for k, v := range metaMap {
					if strVal, ok := v.(string); ok {
						function.Metadata[k] = strVal
					} else {
						function.Metadata[k] = fmt.Sprintf("%v", v)
					}
				}
			}

			bp.Functions = append(bp.Functions, function)
		}
	}

	return bp, nil
}

// FromPkgBlueprint converts a package blueprint to database models
func (r *PostgresBlueprintRepository) FromPkgBlueprint(bp *blueprint.Blueprint) (*models.Blueprint, *models.BlueprintVersion, error) {
	if bp == nil {
		return nil, nil, errors.New("blueprint cannot be nil")
	}

	// Create database blueprint model
	blueprintModel := &models.Blueprint{
		Asset: models.Asset{
			ID:   bp.ID,
			Name: bp.Name,
			Type: "blueprint",
		},
		EntryPoints: bp.FindEntryPoints(), // Calculate entry points
	}

	// Set description
	if bp.Description != "" {
		blueprintModel.Description = sql.NullString{
			String: bp.Description,
			Valid:  true,
		}
	}

	// Convert metadata
	if len(bp.Metadata) > 0 {
		blueprintModel.Metadata = make(models.JSONB)
		for k, v := range bp.Metadata {
			blueprintModel.Metadata[k] = v
		}
	}

	// Create blueprint version
	versionModel := &models.BlueprintVersion{
		VersionNumber: 1, // Default to version 1
		CreatedAt:     time.Now(),
	}

	// Convert nodes
	nodes := make([]interface{}, len(bp.Nodes))
	for i, node := range bp.Nodes {
		nodeMap := map[string]interface{}{
			"id":   node.ID,
			"type": node.Type,
			"position": map[string]interface{}{
				"x": node.Position.X,
				"y": node.Position.Y,
			},
		}

		// Convert properties
		if len(node.Properties) > 0 {
			props := make([]interface{}, len(node.Properties))
			for j, prop := range node.Properties {
				props[j] = map[string]interface{}{
					"name":  prop.Name,
					"value": prop.Value,
				}
			}
			nodeMap["properties"] = props
		}

		// Add data map if present
		if node.Data != nil {
			nodeMap["data"] = node.Data
		}

		nodes[i] = nodeMap
	}
	versionModel.Nodes = nodes

	// Convert connections
	connections := make([]interface{}, len(bp.Connections))
	for i, conn := range bp.Connections {
		connMap := map[string]interface{}{
			"id":              conn.ID,
			"source_node":     conn.SourceNodeID,
			"source_pin":      conn.SourcePinID,
			"target_node":     conn.TargetNodeID,
			"target_pin":      conn.TargetPinID,
			"connection_type": conn.ConnectionType,
		}

		// Add data map if present
		if conn.Data != nil {
			connMap["data"] = conn.Data
		}

		connections[i] = connMap
	}
	versionModel.Connections = connections

	// Convert variables
	variables := make([]interface{}, len(bp.Variables))
	for i, variable := range bp.Variables {
		variables[i] = map[string]interface{}{
			"name":  variable.Name,
			"type":  variable.Type,
			"value": variable.Value,
		}
	}
	versionModel.Variables = variables

	// Convert functions
	functions := make([]interface{}, len(bp.Functions))
	for i, function := range bp.Functions {
		funcMap := map[string]interface{}{
			"id":          function.ID,
			"name":        function.Name,
			"description": function.Description,
		}

		// Convert node type
		nodeTypeMap := map[string]interface{}{}

		// Convert inputs
		if len(function.NodeType.Inputs) > 0 {
			inputs := make([]interface{}, len(function.NodeType.Inputs))
			for j, pin := range function.NodeType.Inputs {
				pinMap := map[string]interface{}{
					"id":          pin.ID,
					"name":        pin.Name,
					"description": pin.Description,
					"optional":    pin.Optional,
				}

				if pin.Default != nil {
					pinMap["default"] = pin.Default
				}

				if pin.Type != nil {
					pinMap["type"] = map[string]interface{}{
						"id":          pin.Type.ID,
						"name":        pin.Type.Name,
						"description": pin.Type.Description,
					}
				}

				inputs[j] = pinMap
			}
			nodeTypeMap["inputs"] = inputs
		}

		// Convert outputs (similar to inputs)
		if len(function.NodeType.Outputs) > 0 {
			outputs := make([]interface{}, len(function.NodeType.Outputs))
			for j, pin := range function.NodeType.Outputs {
				pinMap := map[string]interface{}{
					"id":          pin.ID,
					"name":        pin.Name,
					"description": pin.Description,
					"optional":    pin.Optional,
				}

				if pin.Default != nil {
					pinMap["default"] = pin.Default
				}

				if pin.Type != nil {
					pinMap["type"] = map[string]interface{}{
						"id":          pin.Type.ID,
						"name":        pin.Type.Name,
						"description": pin.Type.Description,
					}
				}

				outputs[j] = pinMap
			}
			nodeTypeMap["outputs"] = outputs
		}

		funcMap["node_type"] = nodeTypeMap

		// Convert function nodes
		if len(function.Nodes) > 0 {
			funcNodes := make([]interface{}, len(function.Nodes))
			for j, node := range function.Nodes {
				funcNodes[j] = map[string]interface{}{
					"id":   node.ID,
					"type": node.Type,
					"position": map[string]interface{}{
						"x": node.Position.X,
						"y": node.Position.Y,
					},
				}
			}
			funcMap["nodes"] = funcNodes
		}

		// Convert function connections
		if len(function.Connections) > 0 {
			funcConns := make([]interface{}, len(function.Connections))
			for j, conn := range function.Connections {
				funcConns[j] = map[string]interface{}{
					"id":              conn.ID,
					"source_node_id":  conn.SourceNodeID,
					"source_pin_id":   conn.SourcePinID,
					"target_node_id":  conn.TargetNodeID,
					"target_pin_id":   conn.TargetPinID,
					"connection_type": conn.ConnectionType,
				}
			}
			funcMap["connections"] = funcConns
		}

		// Convert metadata
		if len(function.Metadata) > 0 {
			metaMap := make(map[string]interface{})
			for k, v := range function.Metadata {
				metaMap[k] = v
			}
			funcMap["metadata"] = metaMap
		}

		functions[i] = funcMap
	}
	versionModel.Functions = functions

	// Set counts
	blueprintModel.NodeCount = len(bp.Nodes)
	blueprintModel.ConnectionCount = len(bp.Connections)

	return blueprintModel, versionModel, nil
}
