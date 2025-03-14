package bperrors_test

import (
	"fmt"
	"testing"
	errors "webblueprint/internal/bperrors"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

func TestErrorClassification(t *testing.T) {
	// Create an error
	err := errors.New(
		errors.ErrorTypeExecution,
		errors.ErrNodeExecutionFailed,
		"Test error message",
		errors.SeverityHigh,
	)

	// Check type
	if err.Type != errors.ErrorTypeExecution {
		t.Errorf("Expected error type %s, got %s", errors.ErrorTypeExecution, err.Type)
	}

	// Check code
	if err.Code != errors.ErrNodeExecutionFailed {
		t.Errorf("Expected error code %s, got %s", errors.ErrNodeExecutionFailed, err.Code)
	}

	// Check severity
	if err.Severity != errors.SeverityHigh {
		t.Errorf("Expected severity %s, got %s", errors.SeverityHigh, err.Severity)
	}

	// Check message
	if err.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", err.Message)
	}

	// Check error interface implementation
	if err.Error() != "[execution-E001] high: Test error message" {
		t.Errorf("Unexpected error string: %s", err.Error())
	}
}

func TestErrorWrapping(t *testing.T) {
	// Create original error
	originalErr := fmt.Errorf("original error")

	// Wrap the error
	wrappedErr := errors.Wrap(
		originalErr,
		errors.ErrorTypeDatabase,
		errors.ErrDatabaseConnection,
		"Database connection failed",
		errors.SeverityHigh,
	)

	// Check that the original error is preserved
	if wrappedErr.OriginalError == nil {
		t.Error("Original error not preserved in wrapped error")
	}

	// Check unwrapping
	unwrapped := wrappedErr.Unwrap()
	if unwrapped == nil {
		t.Error("Unwrap() returned nil")
	}
	if unwrapped.Error() != "original error" {
		t.Errorf("Unwrapped error has wrong message: %s", unwrapped.Error())
	}
}

func TestErrorRecovery(t *testing.T) {
	// Create error manager and recovery manager
	errorManager := errors.NewErrorManager()
	recoveryManager := errors.NewRecoveryManager(errorManager)

	// Create a recoverable error
	err := errors.New(
		errors.ErrorTypeExecution,
		errors.ErrNodeExecutionFailed,
		"Recoverable error",
		errors.SeverityMedium,
	).WithRecoveryOptions(errors.RecoveryRetry, errors.RecoverySkipNode)

	// Record the error
	executionID := "test-execution-1"
	nodeID := "test-node-1"
	err.WithNodeInfo(nodeID, "")
	errorManager.RecordError(executionID, err)

	// Get errors for the execution
	executionErrors := errorManager.GetErrors(executionID)
	if len(executionErrors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(executionErrors))
	}

	// Get errors for the node
	nodeErrors := errorManager.GetNodeErrors(executionID, nodeID)
	if len(nodeErrors) != 1 {
		t.Errorf("Expected 1 node error, got %d", len(nodeErrors))
	}

	// Try recovery
	success, details := recoveryManager.RecoverFromError(executionID, err)
	if !success {
		t.Error("Recovery attempt failed")
	}
	if details["recoveryType"] != "retry" {
		t.Errorf("Expected recovery type 'retry', got '%v'", details["recoveryType"])
	}

	// Check recovery attempts
	attempts := recoveryManager.GetRecoveryAttempts(executionID, nodeID)
	if len(attempts) != 1 {
		t.Errorf("Expected 1 recovery attempt, got %d", len(attempts))
	}

	// Count recovery attempts
	retryCount := recoveryManager.CountRecoveryAttempts(executionID, nodeID, errors.RecoveryRetry)
	if retryCount != 1 {
		t.Errorf("Expected 1 retry attempt, got %d", retryCount)
	}
}

func TestErrorAnalysis(t *testing.T) {
	// Create error manager
	errorManager := errors.NewErrorManager()

	// Create different types of errors
	executionID := "test-execution-analysis"

	// Error 1: Execution error in node 1
	err1 := errors.New(
		errors.ErrorTypeExecution,
		errors.ErrNodeExecutionFailed,
		"Execution failed in node 1",
		errors.SeverityHigh,
	).WithNodeInfo("node-1", "")
	errorManager.RecordError(executionID, err1)

	// Error 2: Another execution error in node 1
	err2 := errors.New(
		errors.ErrorTypeExecution,
		errors.ErrExecutionTimeout,
		"Execution timed out in node 1",
		errors.SeverityMedium,
	).WithNodeInfo("node-1", "")
	errorManager.RecordError(executionID, err2)

	// Error 3: Validation error in node 2
	err3 := errors.New(
		errors.ErrorTypeValidation,
		errors.ErrTypeMismatch,
		"Type mismatch in node 2",
		errors.SeverityLow,
	).WithNodeInfo("node-2", "")
	errorManager.RecordError(executionID, err3)

	// Run analysis
	analysis := errorManager.AnalyzeErrors(executionID)

	// Check total errors
	if totalErrors, ok := analysis["totalErrors"].(int); !ok || totalErrors != 3 {
		t.Errorf("Expected 3 total errors in analysis, got %v", analysis["totalErrors"])
	}

	// Check breakdown by type
	typeBreakdown, ok := analysis["typeBreakdown"].(map[errors.ErrorType]int)
	if !ok {
		t.Error("Type breakdown missing or wrong type")
	} else {
		if typeBreakdown[errors.ErrorTypeExecution] != 2 {
			t.Errorf("Expected 2 execution errors, got %d", typeBreakdown[errors.ErrorTypeExecution])
		}
		if typeBreakdown[errors.ErrorTypeValidation] != 1 {
			t.Errorf("Expected 1 validation error, got %d", typeBreakdown[errors.ErrorTypeValidation])
		}
	}

	// Check top problem nodes
	topNodes, ok := analysis["topProblemNodes"].([]map[string]interface{})
	if !ok {
		t.Error("Top problem nodes missing or wrong type")
	} else if len(topNodes) == 0 {
		t.Error("Top problem nodes is empty")
	} else {
		if topNodes[0]["nodeId"] != "node-1" {
			t.Errorf("Expected node-1 to be top problem node, got %v", topNodes[0]["nodeId"])
		}
		if count, ok := topNodes[0]["count"].(int); !ok || count != 2 {
			t.Errorf("Expected node-1 to have 2 errors, got %v", topNodes[0]["count"])
		}
	}
}

func TestValidation(t *testing.T) {
	// Create error manager and validator
	errorManager := errors.NewErrorManager()
	validator := errors.NewBlueprintValidator(errorManager)

	// Create an invalid blueprint
	bp := &blueprint.Blueprint{
		ID:   "", // Invalid: missing ID
		Name: "Test Blueprint",
		Nodes: []blueprint.BlueprintNode{
			{
				ID:   "node-1",
				Type: "constant-string",
			},
		},
	}

	// Validate
	result := validator.ValidateBlueprint(bp)

	// Check result
	if result.Valid {
		t.Error("Validation should have failed")
	}

	// Check errors
	if len(result.Errors) == 0 {
		t.Error("No validation errors reported")
	} else {
		found := false
		for _, err := range result.Errors {
			if err.Code == errors.ErrInvalidBlueprintStructure {
				found = true
				break
			}
		}
		if !found {
			t.Error("Missing expected error for invalid blueprint structure")
		}
	}
}

func TestDefaultValueProvider(t *testing.T) {
	// Create error manager and recovery manager
	errorManager := errors.NewErrorManager()
	recoveryManager := errors.NewRecoveryManager(errorManager)

	// Test default value for string
	stringValue, err := recoveryManager.GetDefaultValue(types.PinTypes.String)
	if err != nil {
		t.Errorf("GetDefaultValue failed for string: %v", err)
	}

	if str := stringValue.AsString(); str != "" {
		t.Errorf("Expected empty string as default, got '%s'", str)
	}

	// Test default value for number
	numberValue, err := recoveryManager.GetDefaultValue(types.PinTypes.Number)
	if err != nil {
		t.Errorf("GetDefaultValue failed for number: %v", err)
	}
	if num := numberValue.AsNumber(); num != 0 {
		t.Errorf("Expected 0 as default number, got %f", num)
	}

	// Test default value for boolean
	boolValue, err := recoveryManager.GetDefaultValue(types.PinTypes.Boolean)
	if err != nil {
		t.Errorf("GetDefaultValue failed for boolean: %v", err)
	}
	if booly := boolValue.AsBoolean(); booly != false {
		t.Errorf("Expected false as default boolean, got %v", booly)
	}
}
