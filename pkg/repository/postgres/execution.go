package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// PostgresExecutionRepository implements ExecutionRepository using PostgreSQL
type PostgresExecutionRepository struct {
	db *sql.DB
}

// NewExecutionRepository creates a new PostgreSQL-based execution repository
func NewExecutionRepository(db *sql.DB) repository.ExecutionRepository {
	return &PostgresExecutionRepository{
		db: db,
	}
}

// Create creates a new execution record
func (r *PostgresExecutionRepository) Create(ctx context.Context, execution *models.Execution) error {
	// Generate ID if not provided
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}

	query := `
		INSERT INTO executions (
			id, blueprint_id, version_id, started_at, status, initiated_by,
			execution_mode, initial_variables
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		execution.ID,
		execution.BlueprintID,
		execution.VersionID,
		execution.StartedAt,
		execution.Status,
		execution.InitiatedBy,
		execution.ExecutionMode,
		execution.InitialVariables,
	)

	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	return nil
}

// GetByID retrieves an execution by ID
func (r *PostgresExecutionRepository) GetByID(ctx context.Context, id string) (*models.Execution, error) {
	query := `
		SELECT 
			id, blueprint_id, version_id, started_at, completed_at, status, initiated_by,
			execution_mode, initial_variables, result, error, duration_ms
		FROM executions
		WHERE id = $1
	`

	var execution models.Execution
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&execution.ID,
		&execution.BlueprintID,
		&execution.VersionID,
		&execution.StartedAt,
		&execution.CompletedAt,
		&execution.Status,
		&execution.InitiatedBy,
		&execution.ExecutionMode,
		&execution.InitialVariables,
		&execution.Result,
		&execution.Error,
		&execution.DurationMs,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		return nil, fmt.Errorf("error retrieving execution: %w", err)
	}

	return &execution, nil
}

// GetByBlueprintID retrieves executions by blueprint ID
func (r *PostgresExecutionRepository) GetByBlueprintID(ctx context.Context, blueprintID string) ([]*models.Execution, error) {
	query := `
		SELECT 
			id, blueprint_id, version_id, started_at, completed_at, status, initiated_by,
			execution_mode, initial_variables, result, error, duration_ms
		FROM executions
		WHERE blueprint_id = $1
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error querying executions: %w", err)
	}
	defer rows.Close()

	var executions []*models.Execution
	for rows.Next() {
		var execution models.Execution
		err := rows.Scan(
			&execution.ID,
			&execution.BlueprintID,
			&execution.VersionID,
			&execution.StartedAt,
			&execution.CompletedAt,
			&execution.Status,
			&execution.InitiatedBy,
			&execution.ExecutionMode,
			&execution.InitialVariables,
			&execution.Result,
			&execution.Error,
			&execution.DurationMs,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning execution row: %w", err)
		}
		executions = append(executions, &execution)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating execution rows: %w", err)
	}

	return executions, nil
}

// UpdateStatus updates an execution status
func (r *PostgresExecutionRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE executions SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution not found: %s", id)
	}

	return nil
}

// Complete completes an execution with results
func (r *PostgresExecutionRepository) Complete(
	ctx context.Context,
	id string,
	success bool,
	result map[string]interface{},
	errorMsg string,
) error {
	now := time.Now()

	// Convert result to JSONB
	resultData := models.JSONB(result)

	// Calculate duration if possible
	var durationMs sql.NullInt32

	// Get the started_at time
	var startedAt time.Time
	err := r.db.QueryRowContext(ctx, "SELECT started_at FROM executions WHERE id = $1", id).Scan(&startedAt)
	if err == nil {
		// Calculate duration in milliseconds
		duration := now.Sub(startedAt).Milliseconds()
		durationMs = sql.NullInt32{Int32: int32(duration), Valid: true}
	}

	// Set status based on success
	status := "completed"
	if !success {
		status = "failed"
	}

	// Update execution
	query := `
		UPDATE executions
		SET 
			completed_at = $1,
			status = $2,
			result = $3,
			error = $4,
			duration_ms = $5
		WHERE id = $6
	`

	var nullableError sql.NullString
	if errorMsg != "" {
		nullableError = sql.NullString{String: errorMsg, Valid: true}
	}

	var _result sql.Result

	_result, err = r.db.ExecContext(
		ctx,
		query,
		now,
		status,
		resultData,
		nullableError,
		durationMs,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to complete execution: %w", err)
	}

	rowsAffected, err := _result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution not found: %s", id)
	}

	return nil
}

// RecordNodeExecution records node execution details
func (r *PostgresExecutionRepository) RecordNodeExecution(
	ctx context.Context,
	executionID, nodeID, nodeType string,
	inputs, outputs map[string]interface{},
) error {
	// Check if node execution record already exists
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM execution_nodes WHERE execution_id = $1 AND node_id = $2)",
		executionID,
		nodeID,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking node execution existence: %w", err)
	}

	// Convert inputs and outputs to JSONB
	inputsData := models.JSONB(inputs)
	outputsData := models.JSONB(outputs)

	now := time.Now()

	if exists {
		// Update existing record
		query := `
			UPDATE execution_nodes
			SET 
				node_type = $1,
				inputs = $2,
				outputs = $3,
				updated_at = $4
			WHERE execution_id = $5 AND node_id = $6
		`

		_, err = r.db.ExecContext(
			ctx,
			query,
			nodeType,
			inputsData,
			outputsData,
			now,
			executionID,
			nodeID,
		)
	} else {
		// Insert new record
		query := `
			INSERT INTO execution_nodes (
				execution_id, node_id, node_type, started_at, status, inputs, outputs
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err = r.db.ExecContext(
			ctx,
			query,
			executionID,
			nodeID,
			nodeType,
			now,
			"executing", // Initial status
			inputsData,
			outputsData,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to record node execution: %w", err)
	}

	return nil
}

// UpdateNodeStatus updates a node execution status
func (r *PostgresExecutionRepository) UpdateNodeStatus(ctx context.Context, executionID, nodeID, status string) error {
	now := time.Now()
	var completedAt sql.NullTime
	var durationMs sql.NullInt32

	// If status is "completed" or "error", set completed_at and calculate duration
	if status == "completed" || status == "error" {
		completedAt = sql.NullTime{Time: now, Valid: true}

		// Get the started_at time
		var startedAt time.Time
		err := r.db.QueryRowContext(
			ctx,
			"SELECT started_at FROM execution_nodes WHERE execution_id = $1 AND node_id = $2",
			executionID,
			nodeID,
		).Scan(&startedAt)

		if err == nil {
			// Calculate duration in milliseconds
			duration := now.Sub(startedAt).Milliseconds()
			durationMs = sql.NullInt32{Int32: int32(duration), Valid: true}
		}
	}

	query := `
		UPDATE execution_nodes
		SET 
			status = $1,
			completed_at = $2,
			duration_ms = $3
		WHERE execution_id = $4 AND node_id = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		completedAt,
		durationMs,
		executionID,
		nodeID,
	)

	if err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Node execution record doesn't exist, create a new one
		query := `
			INSERT INTO execution_nodes (
				execution_id, node_id, status, started_at
			) VALUES ($1, $2, $3, $4)
		`

		_, err = r.db.ExecContext(
			ctx,
			query,
			executionID,
			nodeID,
			status,
			now,
		)

		if err != nil {
			return fmt.Errorf("failed to create node execution record: %w", err)
		}
	}

	return nil
}

// AddLogEntry adds an execution log entry
func (r *PostgresExecutionRepository) AddLogEntry(
	ctx context.Context,
	executionID, nodeID, level, message string,
	details map[string]interface{},
) error {
	// Generate a log ID
	logID := uuid.New().String()

	// Convert details to JSONB
	detailsData := models.JSONB(details)

	// Create nullable nodeID
	var nullableNodeID sql.NullString
	if nodeID != "" {
		nullableNodeID = sql.NullString{String: nodeID, Valid: true}
	}

	query := `
		INSERT INTO execution_logs (
			id, execution_id, node_id, log_level, message, details, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		logID,
		executionID,
		nullableNodeID,
		level,
		message,
		detailsData,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to add log entry: %w", err)
	}

	return nil
}

// GetLogs retrieves execution logs
func (r *PostgresExecutionRepository) GetLogs(ctx context.Context, executionID string) ([]*models.ExecutionLog, error) {
	query := `
		SELECT 
			id, execution_id, node_id, log_level, message, details, timestamp
		FROM execution_logs
		WHERE execution_id = $1
		ORDER BY timestamp
	`

	rows, err := r.db.QueryContext(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("error querying execution logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.ExecutionLog
	for rows.Next() {
		var log models.ExecutionLog
		err := rows.Scan(
			&log.ID,
			&log.ExecutionID,
			&log.NodeID,
			&log.LogLevel,
			&log.Message,
			&log.Details,
			&log.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning log row: %w", err)
		}
		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating log rows: %w", err)
	}

	return logs, nil
}
