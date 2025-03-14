package bperrors

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"webblueprint/pkg/blueprint"
)

// TestErrorGenerator creates predictable errors for testing
type TestErrorGenerator struct {
	errorManager    *ErrorManager
	recoveryManager *RecoveryManager
	validator       *BlueprintValidator
	mutex           sync.Mutex
}

// NewTestErrorGenerator creates a test error generator
func NewTestErrorGenerator() *TestErrorGenerator {
	errorManager := NewErrorManager()
	recoveryManager := NewRecoveryManager(errorManager)
	validator := NewBlueprintValidator(errorManager)

	return &TestErrorGenerator{
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
		validator:       validator,
	}
}

// GenerateTestError creates a test error with the specified type and code
func (g *TestErrorGenerator) GenerateTestError(
	errType ErrorType,
	code BlueprintErrorCode,
	message string,
	severity ErrorSeverity,
) *BlueprintError {
	err := New(errType, code, message, severity)
	err.Timestamp = time.Now()

	// Add some sample details
	err.WithDetails(map[string]interface{}{
		"generated": true,
		"timestamp": time.Now().Unix(),
		"test":      true,
	})

	// Determine if recoverable based on error type and code
	switch errType {
	case ErrorTypeExecution:
		if code == ErrNodeExecutionFailed || code == ErrExecutionTimeout {
			err.WithRecoveryOptions(RecoveryRetry, RecoverySkipNode)
		}
	case ErrorTypeConnection:
		if code == ErrMissingRequiredInput || code == ErrTypeMismatch {
			err.WithRecoveryOptions(RecoveryUseDefaultValue)
		}
	case ErrorTypeDatabase:
		if code == ErrDatabaseConnection {
			err.WithRecoveryOptions(RecoveryRetry)
		}
	}

	return err
}

// SimulateErrorScenario simulates a full error scenario
func (g *TestErrorGenerator) SimulateErrorScenario(scenarioType string, executionID string) (map[string]interface{}, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Clear previous errors
	g.errorManager.ClearErrors(executionID)

	// Simulate different error scenarios
	switch scenarioType {
	case "execution_failure":
		return g.simulateExecutionFailure(executionID)
	case "connection_problem":
		return g.simulateConnectionProblem(executionID)
	case "validation_error":
		return g.simulateValidationError(executionID)
	case "database_error":
		return g.simulateDatabaseError(executionID)
	case "recoverable_errors":
		return g.simulateRecoverableErrors(executionID)
	case "multi_node_errors":
		return g.simulateMultiNodeErrors(executionID)
	default:
		return nil, fmt.Errorf("unknown scenario type: %s", scenarioType)
	}
}

// simulateExecutionFailure simulates a node execution failure
func (g *TestErrorGenerator) simulateExecutionFailure(executionID string) (map[string]interface{}, error) {
	// Create sample errors
	err1 := g.GenerateTestError(
		ErrorTypeExecution,
		ErrNodeExecutionFailed,
		"Node execution failed: division by zero",
		SeverityHigh,
	).WithNodeInfo("node-123", "output-1").WithBlueprintInfo("bp-456", executionID)

	err2 := g.GenerateTestError(
		ErrorTypeExecution,
		ErrNodeExecutionFailed,
		"Node execution failed: null reference",
		SeverityMedium,
	).WithNodeInfo("node-789", "input-1").WithBlueprintInfo("bp-456", executionID)

	// Record errors
	g.errorManager.RecordError(executionID, err1)
	g.errorManager.RecordError(executionID, err2)

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	return analysis, err1
}

// simulateConnectionProblem simulates connection problems
func (g *TestErrorGenerator) simulateConnectionProblem(executionID string) (map[string]interface{}, error) {
	// Create sample errors
	err1 := g.GenerateTestError(
		ErrorTypeConnection,
		ErrNodeDisconnected,
		"Node is disconnected from the graph",
		SeverityMedium,
	).WithNodeInfo("node-abc", "").WithBlueprintInfo("bp-456", executionID)

	err2 := g.GenerateTestError(
		ErrorTypeConnection,
		ErrTypeMismatch,
		"Type mismatch between connected pins",
		SeverityMedium,
	).WithNodeInfo("node-def", "input-2").WithBlueprintInfo("bp-456", executionID)

	// Record errors
	g.errorManager.RecordError(executionID, err1)
	g.errorManager.RecordError(executionID, err2)

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	return analysis, err1
}

// simulateValidationError simulates blueprint validation errors
func (g *TestErrorGenerator) simulateValidationError(executionID string) (map[string]interface{}, error) {
	// Create a test blueprint with validation issues
	bp := &blueprint.Blueprint{
		ID:   "test-blueprint",
		Name: "",
		Nodes: []blueprint.BlueprintNode{
			{
				ID:   "node-1",
				Type: "constant-string",
			},
			{
				ID:   "node-2",
				Type: "invalid-node-type",
			},
		},
		Connections: []blueprint.Connection{
			{
				ID:             "conn-1",
				SourceNodeID:   "node-1",
				SourcePinID:    "output",
				TargetNodeID:   "node-3", // Non-existent node
				TargetPinID:    "input",
				ConnectionType: "data",
			},
		},
	}

	// Validate the blueprint
	result := g.validator.ValidateBlueprint(bp)

	// Record errors
	for _, err := range result.Errors {
		var bpErr *BlueprintError
		if errors.As(err, &bpErr) {
			err.(*BlueprintError).ExecutionID = executionID
			g.errorManager.RecordError(executionID, err.(*BlueprintError))
		}
	}

	for _, warn := range result.Warnings {
		var bpErr *BlueprintError
		if errors.As(warn, &bpErr) {
			warn.(*BlueprintError).ExecutionID = executionID
			g.errorManager.RecordError(executionID, warn.(*BlueprintError))
		}
	}

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	var mainError error
	if len(result.Errors) > 0 {
		mainError = result.Errors[0]
	}

	return analysis, mainError
}

// simulateDatabaseError simulates database errors
func (g *TestErrorGenerator) simulateDatabaseError(executionID string) (map[string]interface{}, error) {
	// Create sample errors
	err1 := g.GenerateTestError(
		ErrorTypeDatabase,
		ErrDatabaseConnection,
		"Failed to connect to database: connection timeout",
		SeverityHigh,
	).WithBlueprintInfo("bp-456", executionID)

	err2 := g.GenerateTestError(
		ErrorTypeDatabase,
		ErrDatabaseQuery,
		"Query execution failed: syntax error",
		SeverityMedium,
	).WithNodeInfo("db-query-node", "query").WithBlueprintInfo("bp-456", executionID)

	// Record errors
	g.errorManager.RecordError(executionID, err1)
	g.errorManager.RecordError(executionID, err2)

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	return analysis, err1
}

// simulateRecoverableErrors simulates errors with recovery
func (g *TestErrorGenerator) simulateRecoverableErrors(executionID string) (map[string]interface{}, error) {
	// Create recoverable errors
	err1 := g.GenerateTestError(
		ErrorTypeExecution,
		ErrNodeExecutionFailed,
		"Node execution timed out",
		SeverityMedium,
	).WithNodeInfo("node-123", "output-1").WithBlueprintInfo("bp-456", executionID)
	err1.WithRecoveryOptions(RecoveryRetry, RecoverySkipNode)

	err2 := g.GenerateTestError(
		ErrorTypeConnection,
		ErrMissingRequiredInput,
		"Missing required input value",
		SeverityMedium,
	).WithNodeInfo("node-456", "input-1").WithBlueprintInfo("bp-456", executionID)
	err2.WithRecoveryOptions(RecoveryUseDefaultValue)

	// Record errors
	g.errorManager.RecordError(executionID, err1)
	g.errorManager.RecordError(executionID, err2)

	// Attempt recovery
	success1, details1 := g.recoveryManager.RecoverFromError(executionID, err1)
	success2, details2 := g.recoveryManager.RecoverFromError(executionID, err2)

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	// Add recovery information
	analysis["recoveryAttempts"] = []map[string]interface{}{
		{
			"error":     err1.Code,
			"nodeId":    err1.NodeID,
			"success":   success1,
			"details":   details1,
			"timestamp": time.Now(),
		},
		{
			"error":     err2.Code,
			"nodeId":    err2.NodeID,
			"success":   success2,
			"details":   details2,
			"timestamp": time.Now(),
		},
	}

	return analysis, nil
}

// simulateMultiNodeErrors simulates errors across multiple nodes
func (g *TestErrorGenerator) simulateMultiNodeErrors(executionID string) (map[string]interface{}, error) {
	// Create errors across multiple nodes
	nodeIDs := []string{"node-1", "node-2", "node-3", "node-1", "node-2", "node-1"}
	errorTypes := []ErrorType{
		ErrorTypeExecution, ErrorTypeConnection, ErrorTypeValidation,
		ErrorTypeExecution, ErrorTypeExecution, ErrorTypeExecution,
	}
	errorCodes := []BlueprintErrorCode{
		ErrNodeExecutionFailed, ErrTypeMismatch, ErrInvalidNodeConfiguration,
		ErrExecutionTimeout, ErrNodeExecutionFailed, ErrNoEntryPoints,
	}
	severities := []ErrorSeverity{
		SeverityHigh, SeverityMedium, SeverityMedium,
		SeverityHigh, SeverityLow, SeverityCritical,
	}
	messages := []string{
		"Division by zero", "Type mismatch on connection", "Invalid node configuration",
		"Execution timed out", "Warning: Performance issue detected", "No entry points found",
	}

	var mainError *BlueprintError

	// Create and record errors
	for i := 0; i < len(nodeIDs); i++ {
		err := g.GenerateTestError(
			errorTypes[i],
			errorCodes[i],
			messages[i],
			severities[i],
		).WithNodeInfo(nodeIDs[i], "").WithBlueprintInfo("bp-456", executionID)

		g.errorManager.RecordError(executionID, err)

		if i == 0 {
			mainError = err
		}
	}

	// Generate analysis
	analysis := g.errorManager.AnalyzeErrors(executionID)

	return analysis, mainError
}

// GetErrorManager returns the error manager for testing
func (g *TestErrorGenerator) GetErrorManager() *ErrorManager {
	return g.errorManager
}

// GetRecoveryManager returns the recovery manager for testing
func (g *TestErrorGenerator) GetRecoveryManager() *RecoveryManager {
	return g.recoveryManager
}

// GetValidator returns the validator for testing
func (g *TestErrorGenerator) GetValidator() *BlueprintValidator {
	return g.validator
}
