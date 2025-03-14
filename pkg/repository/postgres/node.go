package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"webblueprint/internal/node"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"
)

type PostgresNodeRepository struct {
	db *sql.DB
}

func NewPostgresNodeRepository(db *sql.DB) repository.NodeRepository {
	return &PostgresNodeRepository{db: db}
}

func (p *PostgresNodeRepository) NodeCreate(ctx context.Context, nodeType *models.NodeType) error {
	if nodeType == nil {
		return errors.New("nodeType is nil")
	}

	query := `
		INSERT INTO node_types (
			id, name, description, category_id, version, 
		    author, author_url, icon, is_core, is_deprecated, 
		    inputs, outputs, properties, metadata, created_at, updated_at
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := p.db.ExecContext(
		ctx,
		query,
		nodeType.ID,
		nodeType.Name,
		nodeType.Description,
		nodeType.CategoryID,
		nodeType.Version,
		nodeType.Author,
		nodeType.AuthorURL,
		nodeType.Icon,
		nodeType.IsCore,
		nodeType.IsDeprecated,
		nodeType.Inputs,
		nodeType.Outputs,
		nodeType.Properties,
		nodeType.Metadata,
		nodeType.CreatedAt,
		nodeType.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create node type: %w", err)
	}

	return nil
}

func (p *PostgresNodeRepository) NodeExists(ctx context.Context, typeId string) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM node_types WHERE id = $1)", typeId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if node exists: %w", err)
	}

	return exists, nil
}

func (p *PostgresNodeRepository) NodeGetAll(ctx context.Context) ([]*models.NodeType, error) {
	var nodeTypes []*models.NodeType

	query := `
		SELECT id, name, description, category_id, version, 
				author, author_url, icon, is_core, is_deprecated, inputs, 
				outputs, properties, metadata, created_at, updated_at
		FROM node_types`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nodeType models.NodeType

		err = rows.Scan(
			&nodeType.ID,
			&nodeType.Name,
			&nodeType.Description,
			&nodeType.CategoryID,
			&nodeType.Version,
			&nodeType.Author,
			&nodeType.AuthorURL,
			&nodeType.Icon,
			&nodeType.IsCore,
			&nodeType.IsDeprecated,
			&nodeType.Inputs,
			&nodeType.Outputs,
			&nodeType.Properties,
			&nodeType.Metadata,
			&nodeType.CreatedAt,
			&nodeType.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node type: %w", err)
		}

		nodeTypes = append(nodeTypes, &nodeType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate node types: %w", err)
	}

	return nodeTypes, nil
}

func (p *PostgresNodeRepository) NodeCategoryCreateIfNot(ctx context.Context, category string) (*models.NodeCategory, error) {
	var nodeCategory = &models.NodeCategory{}

	query := `
		SELECT id, name, description, color, icon, sort_order FROM node_categories WHERE name = $1
	`

	err := p.db.QueryRowContext(ctx, query, category).
		Scan(
			&nodeCategory.ID,
			&nodeCategory.Name,
			&nodeCategory.Description,
			&nodeCategory.Color,
			&nodeCategory.Icon,
			&nodeCategory.SortOrder,
		)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get node category: %w", err)
	}

	createQuery := `
		INSERT INTO node_categories (id, name, description) VALUES (uuid_generate_v1(), $1, $2)
		RETURNING id, name, description
	`

	err = p.db.QueryRowContext(
		ctx,
		createQuery,
		category,
		"core node",
	).Scan(
		&nodeCategory.ID,
		&nodeCategory.Name,
		&nodeCategory.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create node category: %w", err)
	}

	return nodeCategory, nil
}

func (p *PostgresNodeRepository) ToPkgNode(n *models.NodeType) (node.Node, error) {
	return nil, nil
}
