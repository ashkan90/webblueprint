package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"webblueprint/internal/engine"
	"webblueprint/pkg/blueprint"
)

// TestErrorPropagation tests error handling and propagation in the execution engine
func TestErrorPropagation(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Create and register an error propagation blueprint
	bp := createErrorPropagationBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases for both execution modes
	testCases := []BlueprintTestCase{
		{
			Name:        "error propagation - standard mode",
			BlueprintID: "test_error_propagation",
			Inputs: map[string]interface{}{
				"shouldError": true,
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure",
			ExecutionMode:        engine.ModeStandard,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "error propagation - actor mode",
			BlueprintID: "test_error_propagation",
			Inputs: map[string]interface{}{
				"shouldError": true,
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure",
			ExecutionMode:        engine.ModeActor,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "no error - standard mode",
			BlueprintID: "test_error_propagation",
			Inputs: map[string]interface{}{
				"shouldError": false,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "completed successfully",
			},
			ExecutionMode: engine.ModeStandard,
			Timeout:       5 * time.Second,
		},
		{
			Name:        "no error - actor mode",
			BlueprintID: "test_error_propagation",
			Inputs: map[string]interface{}{
				"shouldError": false,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "completed successfully",
			},
			ExecutionMode: engine.ModeActor,
			Timeout:       5 * time.Second,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// TestCascadingErrors tests how errors cascade through the execution path
func TestCascadingErrors(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Create and register a cascading error blueprint
	bp := createCascadingErrorBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases for both execution modes
	testCases := []BlueprintTestCase{
		{
			Name:        "cascading errors - first node fails - standard mode",
			BlueprintID: "test_cascading_errors",
			Inputs: map[string]interface{}{
				"errorAt": "first",
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure in first node",
			ExecutionMode:        engine.ModeStandard,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "cascading errors - second node fails - standard mode",
			BlueprintID: "test_cascading_errors",
			Inputs: map[string]interface{}{
				"errorAt": "second",
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure in second node",
			ExecutionMode:        engine.ModeStandard,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "cascading errors - first node fails - actor mode",
			BlueprintID: "test_cascading_errors",
			Inputs: map[string]interface{}{
				"errorAt": "first",
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure in first node",
			ExecutionMode:        engine.ModeActor,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "cascading errors - second node fails - actor mode",
			BlueprintID: "test_cascading_errors",
			Inputs: map[string]interface{}{
				"errorAt": "second",
			},
			ExpectError:          true,
			ExpectedErrorMessage: "intentional failure in second node",
			ExecutionMode:        engine.ModeActor,
			Timeout:              5 * time.Second,
		},
		{
			Name:        "cascading errors - no errors - standard mode",
			BlueprintID: "test_cascading_errors",
			Inputs: map[string]interface{}{
				"errorAt": "none",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "all nodes completed",
			},
			ExecutionMode: engine.ModeStandard,
			Timeout:       5 * time.Second,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// TestErrorRecovery tests error recovery mechanisms
func TestErrorRecovery(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Create and register an error recovery blueprint
	bp := createErrorRecoveryBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases for both execution modes
	testCases := []BlueprintTestCase{
		{
			Name:        "error recovery - standard mode",
			BlueprintID: "test_error_recovery",
			Inputs: map[string]interface{}{
				"shouldError": true,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "recovered",
				"path":   "recovery",
			},
			ExecutionMode:        engine.ModeStandard,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start":         1,
				"try_error":     1,
				"recovery_path": 1,
				"normal_path":   0,
				"end":           1,
			},
			Timeout: 5 * time.Second,
		},
		{
			Name:        "error recovery - normal path - standard mode",
			BlueprintID: "test_error_recovery",
			Inputs: map[string]interface{}{
				"shouldError": false,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "normal",
				"path":   "normal",
			},
			ExecutionMode:        engine.ModeStandard,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start":         1,
				"try_error":     1,
				"recovery_path": 0,
				"normal_path":   1,
				"end":           1,
			},
			Timeout: 5 * time.Second,
		},
		{
			Name:        "error recovery - actor mode",
			BlueprintID: "test_error_recovery",
			Inputs: map[string]interface{}{
				"shouldError": true,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "recovered",
				"path":   "recovery",
			},
			ExecutionMode:        engine.ModeActor,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start":         1,
				"try_error":     1,
				"recovery_path": 1,
				"normal_path":   0,
				"end":           1,
			},
			Timeout: 5 * time.Second,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// Helper function to create an error propagation blueprint
func createErrorPropagationBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_error_propagation", "Error Propagation Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "if_error",
		Type:     "if",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "error_node",
		Type:     "error-node",
		Position: blueprint.Position{X: 500, Y: 50},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "success_node",
		Type:     "process-a",
		Position: blueprint.Position{X: 500, Y: 150},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 700, Y: 100},
	})

	// Add variable nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_should_error",
		Type:     "get-variable-shouldError",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 600, Y: 200},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_if",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "if_error",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_true_to_error",
		SourceNodeID:   "if_error",
		SourcePinID:    "true",
		TargetNodeID:   "error_node",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_false_to_success",
		SourceNodeID:   "if_error",
		SourcePinID:    "false",
		TargetNodeID:   "success_node",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_to_end",
		SourceNodeID:   "success_node",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Connect data flow
	bp.AddConnection(blueprint.Connection{
		ID:             "should_error_to_if",
		SourceNodeID:   "input_should_error",
		SourcePinID:    "value",
		TargetNodeID:   "if_error",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_to_output",
		SourceNodeID:   "success_node",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	// Add constant for success output
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "success_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 300, Y: 250},
		Data: map[string]interface{}{
			"value": "completed successfully",
		},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_constant_to_success",
		SourceNodeID:   "success_constant",
		SourcePinID:    "value",
		TargetNodeID:   "success_node",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	return bp
}

// Helper function to create a cascading error blueprint
func createCascadingErrorBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_cascading_errors", "Cascading Errors Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "if_first",
		Type:     "if",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "error_first",
		Type:     "error-node",
		Position: blueprint.Position{X: 500, Y: 50},
		Data: map[string]interface{}{
			"errorMessage": "intentional failure in first node",
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "if_second",
		Type:     "if",
		Position: blueprint.Position{X: 500, Y: 150},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "error_second",
		Type:     "error-node",
		Position: blueprint.Position{X: 700, Y: 100},
		Data: map[string]interface{}{
			"errorMessage": "intentional failure in second node",
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "success",
		Type:     "process-a",
		Position: blueprint.Position{X: 700, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 900, Y: 150},
	})

	// Add variable nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_error_at",
		Type:     "get-variable-errorAt",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 800, Y: 250},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_if_first",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "if_first",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_first_true_to_error",
		SourceNodeID:   "if_first",
		SourcePinID:    "true",
		TargetNodeID:   "error_first",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_first_false_to_if_second",
		SourceNodeID:   "if_first",
		SourcePinID:    "false",
		TargetNodeID:   "if_second",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_second_true_to_error",
		SourceNodeID:   "if_second",
		SourcePinID:    "true",
		TargetNodeID:   "error_second",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_second_false_to_success",
		SourceNodeID:   "if_second",
		SourcePinID:    "false",
		TargetNodeID:   "success",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_to_end",
		SourceNodeID:   "success",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Connect data flow
	// Create comparison nodes for each error location
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "is_first",
		Type:     "if",
		Position: blueprint.Position{X: 200, Y: 250},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "is_second",
		Type:     "if",
		Position: blueprint.Position{X: 400, Y: 250},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "error_at_to_is_first",
		SourceNodeID:   "input_error_at",
		SourcePinID:    "value",
		TargetNodeID:   "is_first",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "error_at_to_is_second",
		SourceNodeID:   "input_error_at",
		SourcePinID:    "value",
		TargetNodeID:   "is_second",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	// Add constant for success output
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "success_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 600, Y: 300},
		Data: map[string]interface{}{
			"value": "all nodes completed",
		},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_constant_to_success",
		SourceNodeID:   "success_constant",
		SourcePinID:    "value",
		TargetNodeID:   "success",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "success_to_output",
		SourceNodeID:   "success",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	// Condition constants
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "first_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 150, Y: 300},
		Data: map[string]interface{}{
			"value": "first",
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "second_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 350, Y: 300},
		Data: map[string]interface{}{
			"value": "second",
		},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "first_constant_to_is_first",
		SourceNodeID:   "first_constant",
		SourcePinID:    "value",
		TargetNodeID:   "is_first",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "second_constant_to_is_second",
		SourceNodeID:   "second_constant",
		SourcePinID:    "value",
		TargetNodeID:   "is_second",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	return bp
}

// Helper function to create an error recovery blueprint
func createErrorRecoveryBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_error_recovery", "Error Recovery Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "try_error",
		Type:     "recoverable-error",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "if_path",
		Type:     "if",
		Position: blueprint.Position{X: 500, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "recovery_path",
		Type:     "process-a",
		Position: blueprint.Position{X: 700, Y: 50},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "normal_path",
		Type:     "process-b",
		Position: blueprint.Position{X: 700, Y: 150},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 900, Y: 100},
	})

	// Add variable nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_should_error",
		Type:     "get-variable-shouldError",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 800, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_path",
		Type:     "set-variable-path",
		Position: blueprint.Position{X: 800, Y: 250},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_try",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "try_error",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "try_to_if",
		SourceNodeID:   "try_error",
		SourcePinID:    "out",
		TargetNodeID:   "if_path",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_true_to_recovery",
		SourceNodeID:   "if_path",
		SourcePinID:    "true",
		TargetNodeID:   "recovery_path",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_false_to_normal",
		SourceNodeID:   "if_path",
		SourcePinID:    "false",
		TargetNodeID:   "normal_path",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "recovery_to_end",
		SourceNodeID:   "recovery_path",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "normal_to_end",
		SourceNodeID:   "normal_path",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Connect data flow
	bp.AddConnection(blueprint.Connection{
		ID:             "should_error_to_try",
		SourceNodeID:   "input_should_error",
		SourcePinID:    "value",
		TargetNodeID:   "try_error",
		TargetPinID:    "should_error",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "try_status_to_if",
		SourceNodeID:   "try_error",
		SourcePinID:    "status",
		TargetNodeID:   "if_path",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	// Add constant for path outputs
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "recovery_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 600, Y: 50},
		Data: map[string]interface{}{
			"value": "recovery",
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "normal_constant",
		Type:     "constant-string",
		Position: blueprint.Position{X: 600, Y: 150},
		Data: map[string]interface{}{
			"value": "normal",
		},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "recovery_constant_to_recovery",
		SourceNodeID:   "recovery_constant",
		SourcePinID:    "value",
		TargetNodeID:   "recovery_path",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "normal_constant_to_normal",
		SourceNodeID:   "normal_constant",
		SourcePinID:    "value",
		TargetNodeID:   "normal_path",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "recovery_path_to_output",
		SourceNodeID:   "recovery_path",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "normal_path_to_output",
		SourceNodeID:   "normal_path",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "recovery_constant_to_path",
		SourceNodeID:   "recovery_constant",
		SourcePinID:    "value",
		TargetNodeID:   "output_path",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "normal_constant_to_path",
		SourceNodeID:   "normal_constant",
		SourcePinID:    "value",
		TargetNodeID:   "output_path",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	return bp
}
