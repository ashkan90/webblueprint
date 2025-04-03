package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"webblueprint/internal/engine"
	"webblueprint/pkg/blueprint"
)

// TestIfConditionExecution tests if condition execution
func TestIfConditionExecution(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Register additional variable types specific to this test
	RegisterVariableSetNodes(runner, []string{"branch", "result"})
	RegisterVariableGetNodes(runner, []string{"branch"})

	// Create and register a conditional blueprint
	bp := createConditionalBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases for various conditions and execution modes
	testCases := []BlueprintTestCase{
		{
			Name:        "if condition - true path - standard mode",
			BlueprintID: "test_if_condition",
			Inputs: map[string]interface{}{
				"condition": true,
				"value":     "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "test_processed_by_A",
			},
			ExecutionMode:        engine.ModeStandard,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start":      1,
				"if_node":    1,
				"true_path":  1,
				"false_path": 0,
				"end":        1,
			},
		},
		{
			Name:        "if condition - false path - standard mode",
			BlueprintID: "test_if_condition",
			Inputs: map[string]interface{}{
				"condition": false,
				"value":     "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "test_processed_by_B",
			},
			ExecutionMode:        engine.ModeStandard,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start":      1,
				"if_node":    1,
				"true_path":  0,
				"false_path": 1,
				"end":        1,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// Helper function to create a conditional blueprint
func createConditionalBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_if_condition", "Conditional Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "if_node",
		Type:     "if",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "true_path",
		Type:     "process-a",
		Position: blueprint.Position{X: 500, Y: 50},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "false_path",
		Type:     "process-b",
		Position: blueprint.Position{X: 500, Y: 150},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 700, Y: 100},
	})

	// Add variable nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_condition",
		Type:     "get-variable-condition",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_value",
		Type:     "get-variable-value",
		Position: blueprint.Position{X: 100, Y: 250},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 700, Y: 200},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_if",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "if_node",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_true_to_true_path",
		SourceNodeID:   "if_node",
		SourcePinID:    "true",
		TargetNodeID:   "true_path",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "if_false_to_false_path",
		SourceNodeID:   "if_node",
		SourcePinID:    "false",
		TargetNodeID:   "false_path",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "true_path_to_end",
		SourceNodeID:   "true_path",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "false_path_to_end",
		SourceNodeID:   "false_path",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Connect data flow
	bp.AddConnection(blueprint.Connection{
		ID:             "condition_to_if",
		SourceNodeID:   "input_condition",
		SourcePinID:    "value",
		TargetNodeID:   "if_node",
		TargetPinID:    "condition",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "value_to_true_path",
		SourceNodeID:   "input_value",
		SourcePinID:    "value",
		TargetNodeID:   "true_path",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "value_to_false_path",
		SourceNodeID:   "input_value",
		SourcePinID:    "value",
		TargetNodeID:   "false_path",
		TargetPinID:    "data",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "true_path_to_result",
		SourceNodeID:   "true_path",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "false_path_to_result",
		SourceNodeID:   "false_path",
		SourcePinID:    "result",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	return bp
}
