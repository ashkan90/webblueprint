package db

import (
	"database/sql"
	"fmt"
	"time"
	"webblueprint/pkg/dto"
	"webblueprint/pkg/models"

	"github.com/google/uuid" // For generating IDs
	// Add your SQL driver import here, e.g., _ "github.com/lib/pq" for PostgreSQL
)

// SchemaComponentStore defines the interface for schema component database operations.
type SchemaComponentStore interface {
	CreateSchemaComponent(name, schemaDefinition string) (*models.SchemaComponent, error)
	GetSchemaComponent(id string) (*models.SchemaComponent, error)
	ListSchemaComponents() ([]models.SchemaComponent, error)
	UpdateSchemaComponent(id, name, schemaDefinition string) (*models.SchemaComponent, error)
	DeleteSchemaComponent(id string) error

	// ToPkgSchema Convert database schema to package schema format
	ToPkgSchema(schema *models.SchemaComponent) (*dto.SchemaDefinition, error)
}

// SQLSchemaComponentStore implements SchemaComponentStore using database/sql.
type SQLSchemaComponentStore struct {
	db *sql.DB
}

// NewSQLSchemaComponentStore creates a new SQLSchemaComponentStore.
func NewSQLSchemaComponentStore(db *sql.DB) *SQLSchemaComponentStore {
	return &SQLSchemaComponentStore{db: db}
}

// CreateSchemaComponent adds a new schema component to the database.
func (s *SQLSchemaComponentStore) CreateSchemaComponent(name, schemaDefinition string) (*models.SchemaComponent, error) {
	newID := uuid.New().String() // Generate a new UUID
	now := time.Now()

	query := `
		INSERT INTO schema_components (id, name, schema_definition, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, schema_definition, created_at, updated_at
	`
	row := s.db.QueryRow(query, newID, name, schemaDefinition, now, now)

	sc := &models.SchemaComponent{}
	err := row.Scan(&sc.ID, &sc.Name, &sc.SchemaDefinition, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating schema component: %w", err)
	}
	return sc, nil
}

// GetSchemaComponent retrieves a schema component by its ID.
func (s *SQLSchemaComponentStore) GetSchemaComponent(id string) (*models.SchemaComponent, error) {
	query := `SELECT id, name, schema_definition, created_at, updated_at FROM schema_components WHERE id = $1`
	row := s.db.QueryRow(query, id)

	sc := &models.SchemaComponent{}
	err := row.Scan(&sc.ID, &sc.Name, &sc.SchemaDefinition, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema component not found with id: %s", id)
		}
		return nil, fmt.Errorf("error getting schema component: %w", err)
	}
	return sc, nil
}

// ListSchemaComponents retrieves all schema components.
func (s *SQLSchemaComponentStore) ListSchemaComponents() ([]models.SchemaComponent, error) {
	query := `SELECT id, name, schema_definition, created_at, updated_at FROM schema_components ORDER BY name ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error listing schema components: %w", err)
	}
	defer rows.Close()

	components := []models.SchemaComponent{}
	for rows.Next() {
		sc := models.SchemaComponent{}
		err := rows.Scan(&sc.ID, &sc.Name, &sc.SchemaDefinition, &sc.CreatedAt, &sc.UpdatedAt)
		if err != nil {
			// Log error but continue processing other rows
			fmt.Printf("Error scanning schema component row: %v\n", err) // Replace with proper logging
			continue
		}
		components = append(components, sc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema component rows: %w", err)
	}

	return components, nil
}

// UpdateSchemaComponent updates an existing schema component.
func (s *SQLSchemaComponentStore) UpdateSchemaComponent(id, name, schemaDefinition string) (*models.SchemaComponent, error) {
	query := `
		UPDATE schema_components
		SET name = $2, schema_definition = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, name, schema_definition, created_at, updated_at
	`
	row := s.db.QueryRow(query, id, name, schemaDefinition, time.Now())

	sc := &models.SchemaComponent{}
	err := row.Scan(&sc.ID, &sc.Name, &sc.SchemaDefinition, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema component not found with id for update: %s", id)
		}
		return nil, fmt.Errorf("error updating schema component: %w", err)
	}
	return sc, nil
}

// DeleteSchemaComponent removes a schema component from the database.
func (s *SQLSchemaComponentStore) DeleteSchemaComponent(id string) error {
	query := `DELETE FROM schema_components WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting schema component: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log this error but don't necessarily fail the operation if deletion occurred
		fmt.Printf("Error getting rows affected after delete: %v\n", err) // Replace with proper logging
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schema component not found with id for deletion: %s", id)
	}

	return nil
}

func (s *SQLSchemaComponentStore) ToPkgSchema(schema *models.SchemaComponent) (*dto.SchemaDefinition, error) {
	return &dto.SchemaDefinition{
		ID:               schema.ID,
		Name:             schema.Name,
		SchemaDefinition: schema.SchemaDefinition,
		CreatedAt:        schema.CreatedAt,
		UpdatedAt:        schema.UpdatedAt,
	}, nil
}
