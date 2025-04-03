package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"webblueprint/internal/engine"
	"webblueprint/pkg/blueprint"
)

// TestActorModeConcurrentExecution tests that the actor-based execution mode
// correctly handles concurrent node execution
func TestActorModeConcurrentExecution(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register node types
	RegisterMockNodes(runner)

	// Create and register a parallel blueprint
	bp := CreateParallelBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases (only for actor mode since this tests actor-specific behavior)
	testCases := []BlueprintTestCase{
		{
			Name:        "parallel execution - actor mode",
			BlueprintID: "test_parallel",
			Inputs: map[string]interface{}{
				"data": "test_data",
			},
			ExpectedOutputs: map[string]interface{}{
				"resultA": "test_data_processed_by_A",
				"resultB": "test_data_processed_by_B",
				"resultC": "test_data_processed_by_C",
			},
			ExecutionMode:        engine.ModeActor,
			VerifyNodeExecutions: true,
			ExpectedNodeExecutions: map[string]int{
				"start": 1,
				"split": 1,
				"pathA": 1,
				"pathB": 1,
				"pathC": 1,
				// In actor mode the merge and end node might execute multiple times 
				// due to concurrent processing, so we don't check execution count for these
			},
			Timeout: 10 * time.Second,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// TestActorModeStressTest performs stress testing of the actor-based execution mode
// with many concurrent nodes to ensure it can handle high concurrency
func TestActorModeStressTest(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register node types
	RegisterMockNodes(runner)

	// Create and register a stress test blueprint
	bp := createStressTestBlueprint(50) // Create a blueprint with 50 parallel paths
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test case
	testCase := BlueprintTestCase{
		Name:        "stress test - actor mode with 50 parallel nodes",
		BlueprintID: "test_stress",
		Inputs: map[string]interface{}{
			"data": "test_data",
		},
		ExecutionMode: engine.ModeActor,
		// We're mostly checking that it completes without errors
		// so we don't validate all 50 outputs, just check a few
		ExpectedOutputs: map[string]interface{}{
			"result0":  "test_data_processed_by_A",
			"result25": "test_data_processed_by_A",
			"result49": "test_data_processed_by_A",
		},
		Timeout: 30 * time.Second, // High concurrency might need more time
	}

	// Run test case
	t.Run(testCase.Name, func(t *testing.T) {
		runner.AssertTestCase(t, testCase)
	})
}

// TestActorModeMessagePassingOrder tests that messages are passed in the correct order
// between actors, even with high concurrency
func TestActorModeMessagePassingOrder(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register node types
	RegisterMockNodes(runner)
	runner.RegisterNodeType("sequence-check", NewSequenceCheckNodeFactory())

	// Create and register a message passing test blueprint
	bp := createMessagePassingBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test case
	testCase := BlueprintTestCase{
		Name:        "message passing order - actor mode",
		BlueprintID: "test_message_passing",
		Inputs: map[string]interface{}{
			"data": []interface{}{1, 2, 3, 4, 5},
		},
		ExpectedOutputs: map[string]interface{}{
			"result": true, // Should be true if order is preserved
		},
		ExecutionMode: engine.ModeActor,
		Timeout:       10 * time.Second,
	}

	// Run test case
	t.Run(testCase.Name, func(t *testing.T) {
		runner.AssertTestCase(t, testCase)
	})
}

// TestActorRecoveryAfterError tests that actors can recover after an error
// and continue execution
func TestActorRecoveryAfterError(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register node types
	RegisterMockNodes(runner)
	runner.RegisterNodeType("recoverable-error", NewRecoverableErrorNodeFactory())

	// Create and register a recovery test blueprint
	bp := createRecoveryTestBlueprint()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test case
	testCase := BlueprintTestCase{
		Name:        "actor recovery - after error",
		BlueprintID: "test_recovery",
		Inputs: map[string]interface{}{
			"should_error": true,
		},
		ExpectedOutputs: map[string]interface{}{
			"result": "recovered", // Should indicate recovery happened
		},
		ExecutionMode: engine.ModeActor,
		Timeout:       10 * time.Second,
	}

	// Run test case
	t.Run(testCase.Name, func(t *testing.T) {
		runner.AssertTestCase(t, testCase)
	})
}

// createStressTestBlueprint creates a blueprint with many parallel execution paths
// to stress test the actor-based execution mode
func createStressTestBlueprint(pathCount int) *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_stress", "Stress Test", "1.0.0")

	// Add start and end nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 100},
	})

	// Add split node with many outputs
	splitNode := blueprint.BlueprintNode{
		ID:       "split",
		Type:     "split",
		Position: blueprint.Position{X: 200, Y: 100},
	}
	bp.AddNode(splitNode)

	// Connect start to split
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_split",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "split",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Add merge node with many inputs
	mergeNode := blueprint.BlueprintNode{
		ID:       "merge",
		Type:     "merge",
		Position: blueprint.Position{X: 400, Y: 100},
	}
	bp.AddNode(mergeNode)

	// Connect merge to end
	bp.AddConnection(blueprint.Connection{
		ID:             "merge_to_end",
		SourceNodeID:   "merge",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Add data input node
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_data",
		Type:     "get-variable-data",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	// Create parallel paths
	for i := 0; i < pathCount; i++ {
		// Create process node
		processNode := blueprint.BlueprintNode{
			ID:       fmt.Sprintf("process%d", i),
			Type:     "process-a", // Reuse existing node type
			Position: blueprint.Position{X: 300, Y: 100 + float64(i*30)},
		}
		bp.AddNode(processNode)

		// Create pin IDs based on index (merge and split have limited pins in our mock)
		splitPinID := "out1"
		mergePinID := "in1"
		if i % 3 == 1 {
			splitPinID = "out2"
			mergePinID = "in2"
		} else if i % 3 == 2 {
			splitPinID = "out3"
			mergePinID = "in3"
		}

		// Connect split to process
		bp.AddConnection(blueprint.Connection{
			ID:             fmt.Sprintf("split_to_process%d", i),
			SourceNodeID:   "split",
			SourcePinID:    splitPinID,
			TargetNodeID:   processNode.ID,
			TargetPinID:    "in",
			ConnectionType: "execution",
		})

		// Connect process to merge
		bp.AddConnection(blueprint.Connection{
			ID:             fmt.Sprintf("process%d_to_merge", i),
			SourceNodeID:   processNode.ID,
			SourcePinID:    "out",
			TargetNodeID:   "merge",
			TargetPinID:    mergePinID,
			ConnectionType: "execution",
		})

		// Connect data to process
		bp.AddConnection(blueprint.Connection{
			ID:             fmt.Sprintf("data_to_process%d", i),
			SourceNodeID:   "input_data",
			SourcePinID:    "value",
			TargetNodeID:   processNode.ID,
			TargetPinID:    "data",
			ConnectionType: "data",
		})

		// Add result output node
		resultNode := blueprint.BlueprintNode{
			ID:       fmt.Sprintf("output_result%d", i),
			Type:     fmt.Sprintf("set-variable-result%d", i),
			Position: blueprint.Position{X: 500, Y: 200 + float64(i*30)},
		}
		bp.AddNode(resultNode)

		// Connect process result to output
		bp.AddConnection(blueprint.Connection{
			ID:             fmt.Sprintf("process%d_to_result%d", i, i),
			SourceNodeID:   processNode.ID,
			SourcePinID:    "result",
			TargetNodeID:   resultNode.ID,
			TargetPinID:    "value",
			ConnectionType: "data",
		})
	}

	return bp
}

// createMessagePassingBlueprint creates a blueprint that tests message passing order
// between actors by processing a sequence of data items and verifying the order is preserved
func createMessagePassingBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_message_passing", "Message Passing Test", "1.0.0")

	// Add basic nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "start",
		Type:     "start",
		Position: blueprint.Position{X: 100, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "process_sequence",
		Type:     "sequence-check",
		Position: blueprint.Position{X: 300, Y: 100},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 100},
	})

	// Connect execution flow
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_process",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "process_sequence",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process_to_end",
		SourceNodeID:   "process_sequence",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Add data nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_data",
		Type:     "get-variable-data",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 500, Y: 200},
	})

	// Connect data flow
	bp.AddConnection(blueprint.Connection{
		ID:             "data_to_process",
		SourceNodeID:   "input_data",
		SourcePinID:    "value",
		TargetNodeID:   "process_sequence",
		TargetPinID:    "sequence",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "process_to_result",
		SourceNodeID:   "process_sequence",
		SourcePinID:    "order_preserved",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	return bp
}

// createRecoveryTestBlueprint creates a blueprint that tests actor recovery after errors
func createRecoveryTestBlueprint() *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("test_recovery", "Recovery Test", "1.0.0")

	// Add basic nodes
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
		ID:       "end",
		Type:     "end",
		Position: blueprint.Position{X: 500, Y: 100},
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
		ID:             "try_to_end",
		SourceNodeID:   "try_error",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	// Add data nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:       "input_should_error",
		Type:     "get-variable-should_error",
		Position: blueprint.Position{X: 100, Y: 200},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:       "output_result",
		Type:     "set-variable-result",
		Position: blueprint.Position{X: 500, Y: 200},
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
		ID:             "try_to_result",
		SourceNodeID:   "try_error",
		SourcePinID:    "status",
		TargetNodeID:   "output_result",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	return bp
}
