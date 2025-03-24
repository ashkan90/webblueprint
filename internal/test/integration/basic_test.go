package integration

import (
	"testing"
	"time"

	"webblueprint/internal/engine"
	"webblueprint/pkg/blueprint"
)

// TestSimplestExecution is a minimal test to verify the test framework and node registration
func TestSimplestExecution(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Create a very simple blueprint
	bp := blueprint.NewBlueprint("test_simplest", "Simplest Test", "1.0.0")

	// Add start and end nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "set_var",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 300, Y: 100},
		Data: map[string]interface{}{
			"value": "default_value",
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 100},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_set",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "set_var",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "set_to_end",
		SourceNodeID:   "set_var",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Register blueprint
	err := runner.RegisterBlueprint(bp)
	if err != nil {
		t.Fatalf("Failed to register blueprint: %v", err)
	}

	// Define test case
	testCase := BlueprintTestCase{
		Name:        "simple execution",
		BlueprintID: "test_simplest",
		Inputs:      map[string]interface{}{},
		ExpectedOutputs: map[string]interface{}{
			"result": "default_value",
		},
		ExecutionMode: engine.ModeStandard,
		Timeout:       5 * time.Second,
	}

	// Run test case
	runner.AssertTestCase(t, testCase)
}
