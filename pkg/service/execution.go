package service

import (
	"context"
	"fmt"
	"time"
	"webblueprint/internal/engine"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"

	"github.com/google/uuid"
)

// ExecutionService provides high-level operations for managing blueprint executions
type ExecutionService struct {
	executionRepo   repository.ExecutionRepository
	blueprintRepo   repository.BlueprintRepository
	executionEngine *engine.ExecutionEngine
}

// NewExecutionService creates a new execution service
func NewExecutionService(
	executionRepo repository.ExecutionRepository,
	blueprintRepo repository.BlueprintRepository,
	executionEngine *engine.ExecutionEngine,
) *ExecutionService {
	return &ExecutionService{
		executionRepo:   executionRepo,
		blueprintRepo:   blueprintRepo,
		executionEngine: executionEngine,
	}
}

// StartExecution starts a new blueprint execution
func (s *ExecutionService) StartExecution(
	ctx context.Context,
	blueprintID string,
	initialVariables map[string]interface{},
	userID string,
) (string, error) {
	// Get the blueprint to validate it exists
	blueprintModel, err := s.blueprintRepo.GetByID(ctx, blueprintID)
	if err != nil {
		return "", fmt.Errorf("blueprint not found: %w", err)
	}

	// Create a unique execution ID
	executionID := uuid.New().String()

	// Create execution record
	execution := &models.Execution{
		ID:               executionID,
		BlueprintID:      blueprintID,
		StartedAt:        time.Now(),
		Status:           "running",
		InitiatedBy:      userID,
		ExecutionMode:    "standard", // Could be configurable
		InitialVariables: models.JSONB(initialVariables),
	}

	// Set the version ID if available
	if blueprintModel.CurrentVersionID.Valid {
		execution.VersionID = blueprintModel.CurrentVersionID
	}

	// Save execution record
	err = s.executionRepo.Create(ctx, execution)
	if err != nil {
		return "", fmt.Errorf("failed to create execution record: %w", err)
	}

	// Convert to the format expected by the execution engine
	variables := make(map[string]types.Value)
	for k, v := range initialVariables {
		// Determine type based on value
		var pinType *types.PinType
		switch v.(type) {
		case string:
			pinType = types.PinTypes.String
		case float64, int, int64:
			pinType = types.PinTypes.Number
		case bool:
			pinType = types.PinTypes.Boolean
		case map[string]interface{}:
			pinType = types.PinTypes.Object
		case []interface{}:
			pinType = types.PinTypes.Array
		default:
			pinType = types.PinTypes.Any
		}
		variables[k] = types.NewValue(pinType, v)
	}
	bp, _ := s.blueprintRepo.ToPkgBlueprint(blueprintModel, blueprintModel.CurrentVersion)

	// Execute the blueprint in a goroutine
	go func(bp *blueprint.Blueprint) {
		// Get a background context since the request context will be canceled
		bgCtx := context.Background()

		// Execute the blueprint
		result, err := s.executionEngine.Execute(bp, executionID, variables)

		// Update execution record with result
		if err != nil {
			// Execution failed
			s.executionRepo.Complete(bgCtx, executionID, false, nil, err.Error())
		} else {
			// Execution succeeded
			resultMap := make(map[string]interface{})
			for nodeID, outputs := range result.NodeResults {
				resultMap[nodeID] = outputs
			}
			s.executionRepo.Complete(bgCtx, executionID, true, resultMap, "")
		}

		// Note: The WebSocket message about execution completion is handled by
		// the execution engine's event system. When an execution completes, it emits
		// an EventExecutionEnd event, which is picked up by any registered listeners
		// (such as the ExecutionEventListener) and broadcasted to connected clients.
	}(bp)

	return executionID, nil
}

// GetExecution retrieves execution details by ID
func (s *ExecutionService) GetExecution(ctx context.Context, id string) (*models.Execution, error) {
	execution, err := s.executionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}
	return execution, nil
}

// GetExecutionsByBlueprint retrieves all executions for a blueprint
func (s *ExecutionService) GetExecutionsByBlueprint(ctx context.Context, blueprintID string) ([]*models.Execution, error) {
	// Check if blueprint exists
	_, err := s.blueprintRepo.GetByID(ctx, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found: %w", err)
	}

	// Get executions
	executions, err := s.executionRepo.GetByBlueprintID(ctx, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving executions: %w", err)
	}

	return executions, nil
}

// GetExecutionLogs retrieves logs for an execution
func (s *ExecutionService) GetExecutionLogs(ctx context.Context, executionID string) ([]*models.ExecutionLog, error) {
	// Check if execution exists
	_, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %w", err)
	}

	// Get logs
	logs, err := s.executionRepo.GetLogs(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs: %w", err)
	}

	return logs, nil
}

// RecordNodeExecution records the execution of a node
func (s *ExecutionService) RecordNodeExecution(
	ctx context.Context,
	executionID, nodeID, nodeType string,
	inputs, outputs map[string]interface{},
) error {
	err := s.executionRepo.RecordNodeExecution(ctx, executionID, nodeID, nodeType, inputs, outputs)
	if err != nil {
		return fmt.Errorf("failed to record node execution: %w", err)
	}
	return nil
}

// UpdateNodeStatus updates the status of a node execution
func (s *ExecutionService) UpdateNodeStatus(
	ctx context.Context,
	executionID, nodeID, status string,
) error {
	err := s.executionRepo.UpdateNodeStatus(ctx, executionID, nodeID, status)
	if err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}
	return nil
}

// AddLogEntry adds a log entry to an execution
func (s *ExecutionService) AddLogEntry(
	ctx context.Context,
	executionID, nodeID, level, message string,
	details map[string]interface{},
) error {
	err := s.executionRepo.AddLogEntry(ctx, executionID, nodeID, level, message, details)
	if err != nil {
		return fmt.Errorf("failed to add log entry: %w", err)
	}
	return nil
}

// CancelExecution cancels a running execution
func (s *ExecutionService) CancelExecution(ctx context.Context, executionID string) error {
	// Get the execution
	execution, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	// Check if execution can be canceled
	if execution.Status != "running" {
		return fmt.Errorf("execution cannot be canceled: status is %s", execution.Status)
	}

	// Update status
	err = s.executionRepo.UpdateStatus(ctx, executionID, "cancelled")
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	// TODO: Signal the execution engine to stop this execution
	// This would require adding cancellation capabilities to the engine

	return nil
}
