package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"webblueprint/internal/event"
)

// EventRepository handles database operations for events
type PostgresEventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *sql.DB) *PostgresEventRepository {
	return &PostgresEventRepository{
		db: db,
	}
}

// EventModel represents an event stored in the database
type EventModel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Parameters  string    `json:"parameters"` // JSON string
	BlueprintID string    `json:"blueprintId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Create adds a new event to the database
func (r *PostgresEventRepository) Create(ctx context.Context, event event.EventDefinition) error {
	// Convert parameters to JSON
	parametersJSON, err := json.Marshal(event.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Insert event into database
	query := `
		INSERT INTO events (id, name, description, category, parameters, blueprint_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	_, err = r.db.ExecContext(
		ctx,
		query,
		event.ID,
		event.Name,
		event.Description,
		event.Category,
		string(parametersJSON),
		event.BlueprintID,
		event.CreatedAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetByID retrieves an event by its ID
func (r *PostgresEventRepository) GetByID(ctx context.Context, id string) (event.EventDefinition, error) {
	query := `
		SELECT id, name, description, category, parameters, blueprint_id, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	var model EventModel
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Name,
		&model.Description,
		&model.Category,
		&model.Parameters,
		&model.BlueprintID,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return event.EventDefinition{}, fmt.Errorf("event not found: %s", id)
		}
		return event.EventDefinition{}, fmt.Errorf("failed to query event: %w", err)
	}

	// Parse parameters JSON
	var parameters []event.EventParameter
	if model.Parameters != "" {
		if err := json.Unmarshal([]byte(model.Parameters), &parameters); err != nil {
			return event.EventDefinition{}, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
	}

	// Convert to EventDefinition
	return event.EventDefinition{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		Category:    model.Category,
		Parameters:  parameters,
		BlueprintID: model.BlueprintID,
		CreatedAt:   model.CreatedAt,
	}, nil
}

// GetAll retrieves all events
func (r *PostgresEventRepository) GetAll(ctx context.Context) ([]event.EventDefinition, error) {
	query := `
		SELECT id, name, description, category, parameters, blueprint_id, created_at, updated_at
		FROM events
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []event.EventDefinition
	for rows.Next() {
		var model EventModel
		err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.Description,
			&model.Category,
			&model.Parameters,
			&model.BlueprintID,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		// Parse parameters JSON
		var parameters []event.EventParameter
		if model.Parameters != "" {
			if err := json.Unmarshal([]byte(model.Parameters), &parameters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
			}
		}

		// Convert to EventDefinition
		events = append(events, event.EventDefinition{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			Category:    model.Category,
			Parameters:  parameters,
			BlueprintID: model.BlueprintID,
			CreatedAt:   model.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// GetByBlueprintID retrieves all events for a blueprint
func (r *PostgresEventRepository) GetByBlueprintID(ctx context.Context, blueprintID string) ([]event.EventDefinition, error) {
	query := `
		SELECT id, name, description, category, parameters, blueprint_id, created_at, updated_at
		FROM events
		WHERE blueprint_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []event.EventDefinition
	for rows.Next() {
		var model EventModel
		err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.Description,
			&model.Category,
			&model.Parameters,
			&model.BlueprintID,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		// Parse parameters JSON
		var parameters []event.EventParameter
		if model.Parameters != "" {
			if err := json.Unmarshal([]byte(model.Parameters), &parameters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
			}
		}

		// Convert to EventDefinition
		events = append(events, event.EventDefinition{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			Category:    model.Category,
			Parameters:  parameters,
			BlueprintID: model.BlueprintID,
			CreatedAt:   model.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// Update updates an existing event
func (r *PostgresEventRepository) Update(ctx context.Context, event event.EventDefinition) error {
	// Convert parameters to JSON
	parametersJSON, err := json.Marshal(event.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Update event in database
	query := `
		UPDATE events
		SET name = $1, description = $2, category = $3, parameters = $4, 
		    blueprint_id = $5, updated_at = $6
		WHERE id = $7
	`

	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		query,
		event.Name,
		event.Description,
		event.Category,
		string(parametersJSON),
		event.BlueprintID,
		now,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found: %s", event.ID)
	}

	return nil
}

// Delete removes an event by ID
func (r *PostgresEventRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// AddParameter adds a parameter to an event
func (r *PostgresEventRepository) AddParameter(ctx context.Context, eventID string, parameter event.EventParameter) error {
	// Get current event
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	// Add parameter
	event.Parameters = append(event.Parameters, parameter)

	// Update event
	return r.Update(ctx, event)
}

// RemoveParameter removes a parameter from an event
func (r *PostgresEventRepository) RemoveParameter(ctx context.Context, eventID string, parameterName string) error {
	// Get current event
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	// Remove parameter
	for i, param := range event.Parameters {
		if param.Name == parameterName {
			event.Parameters = append(event.Parameters[:i], event.Parameters[i+1:]...)
			break
		}
	}

	// Update event
	return r.Update(ctx, event)
}

// EventBindingModel represents an event binding stored in the database
type EventBindingModel struct {
	ID          string    `json:"id"`
	EventID     string    `json:"eventId"`
	HandlerID   string    `json:"handlerId"`
	HandlerType string    `json:"handlerType"`
	BlueprintID string    `json:"blueprintId"`
	Priority    int       `json:"priority"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateBinding adds a new event binding to the database
func (r *PostgresEventRepository) CreateBinding(ctx context.Context, binding event.EventBinding) error {
	// Insert binding into database
	query := `
		INSERT INTO event_bindings (id, event_id, handler_id, handler_type, blueprint_id, priority, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		query,
		binding.ID,
		binding.EventID,
		binding.HandlerID,
		binding.HandlerType,
		binding.BlueprintID,
		binding.Priority,
		binding.Enabled,
		binding.CreatedAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to insert binding: %w", err)
	}

	return nil
}

// GetBindingByID retrieves a binding by its ID
func (r *PostgresEventRepository) GetBindingByID(ctx context.Context, id string) (event.EventBinding, error) {
	query := `
		SELECT id, event_id, handler_id, handler_type, blueprint_id, priority, enabled, created_at, updated_at
		FROM event_bindings
		WHERE id = $1
	`

	var model EventBindingModel
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.EventID,
		&model.HandlerID,
		&model.HandlerType,
		&model.BlueprintID,
		&model.Priority,
		&model.Enabled,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return event.EventBinding{}, fmt.Errorf("binding not found: %s", id)
		}
		return event.EventBinding{}, fmt.Errorf("failed to query binding: %w", err)
	}

	// Convert to EventBinding
	return event.EventBinding{
		ID:          model.ID,
		EventID:     model.EventID,
		HandlerID:   model.HandlerID,
		HandlerType: model.HandlerType,
		BlueprintID: model.BlueprintID,
		Priority:    model.Priority,
		Enabled:     model.Enabled,
		CreatedAt:   model.CreatedAt,
	}, nil
}

// GetBindingsByEventID retrieves all bindings for an event
func (r *PostgresEventRepository) GetBindingsByEventID(ctx context.Context, eventID string) ([]event.EventBinding, error) {
	query := `
		SELECT id, event_id, handler_id, handler_type, blueprint_id, priority, enabled, created_at, updated_at
		FROM event_bindings
		WHERE event_id = $1
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bindings: %w", err)
	}
	defer rows.Close()

	var bindings []event.EventBinding
	for rows.Next() {
		var model EventBindingModel
		err := rows.Scan(
			&model.ID,
			&model.EventID,
			&model.HandlerID,
			&model.HandlerType,
			&model.BlueprintID,
			&model.Priority,
			&model.Enabled,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan binding row: %w", err)
		}

		// Convert to EventBinding
		bindings = append(bindings, event.EventBinding{
			ID:          model.ID,
			EventID:     model.EventID,
			HandlerID:   model.HandlerID,
			HandlerType: model.HandlerType,
			BlueprintID: model.BlueprintID,
			Priority:    model.Priority,
			Enabled:     model.Enabled,
			CreatedAt:   model.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating binding rows: %w", err)
	}

	return bindings, nil
}

// DeleteBinding removes a binding by ID
func (r *PostgresEventRepository) DeleteBinding(ctx context.Context, id string) error {
	query := `DELETE FROM event_bindings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete binding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("binding not found: %s", id)
	}

	return nil
}

// DeleteBindingsByEventID removes all bindings for an event
func (r *PostgresEventRepository) DeleteBindingsByEventID(ctx context.Context, eventID string) error {
	query := `DELETE FROM event_bindings WHERE event_id = $1`

	_, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete bindings: %w", err)
	}

	return nil
}

// DeleteBindingsByBlueprintID removes all bindings for a blueprint
func (r *PostgresEventRepository) DeleteBindingsByBlueprintID(ctx context.Context, blueprintID string) error {
	query := `DELETE FROM event_bindings WHERE blueprint_id = $1`

	_, err := r.db.ExecContext(ctx, query, blueprintID)
	if err != nil {
		return fmt.Errorf("failed to delete bindings: %w", err)
	}

	return nil
}

// GetAllBindings retrieves all bindings
func (r *PostgresEventRepository) GetAllBindings(ctx context.Context) ([]event.EventBinding, error) {
	query := `
		SELECT id, event_id, handler_id, handler_type, blueprint_id, priority, enabled, created_at, updated_at
		FROM event_bindings
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bindings: %w", err)
	}
	defer rows.Close()

	var bindings []event.EventBinding
	for rows.Next() {
		var model EventBindingModel
		err := rows.Scan(
			&model.ID,
			&model.EventID,
			&model.HandlerID,
			&model.HandlerType,
			&model.BlueprintID,
			&model.Priority,
			&model.Enabled,
			&model.CreatedAt,
			&model.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan binding row: %w", err)
		}

		// Convert to EventBinding
		bindings = append(bindings, event.EventBinding{
			ID:          model.ID,
			EventID:     model.EventID,
			HandlerID:   model.HandlerID,
			HandlerType: model.HandlerType,
			BlueprintID: model.BlueprintID,
			Priority:    model.Priority,
			Enabled:     model.Enabled,
			CreatedAt:   model.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating binding rows: %w", err)
	}

	return bindings, nil
}
