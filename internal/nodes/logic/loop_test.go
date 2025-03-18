package logic_test

import (
	"testing"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/test"
	"webblueprint/internal/test/mocks"
	"webblueprint/internal/types"
)

func TestLoopNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "normal loop execution",
			Inputs: map[string]interface{}{
				"iterations": 3.0,
			},
			ExpectedFlow: "completed", // The final flow will be "completed" after all iterations
		},
		{
			Name: "loop with start value",
			Inputs: map[string]interface{}{
				"iterations": 3.0,
				"startValue": 10.0,
			},
			ExpectedFlow: "completed",
		},
		{
			Name: "zero iterations - should skip to completed",
			Inputs: map[string]interface{}{
				"iterations": 0.0,
			},
			ExpectedFlow: "completed", // Should skip loop and go straight to completed
		},
		{
			Name: "negative iterations - should skip to completed",
			Inputs: map[string]interface{}{
				"iterations": -5.0,
			},
			ExpectedFlow: "completed", // Should skip loop and go straight to completed
		},
		{
			Name:   "missing iterations - should return error",
			Inputs: map[string]interface{}{
				// No iterations provided
			},
			ExpectedError: true,
		},
		{
			Name: "invalid iterations - should return error",
			Inputs: map[string]interface{}{
				"iterations": "not a number",
			},
			ExpectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := logic.NewLoopNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

// This test tests multiple iterations using a custom testing approach
// because our standard test utils can't easily handle multiple loop iterations
func TestLoopNodeMultipleIterations(t *testing.T) {
	// Create a LoopNode
	node := logic.NewLoopNode()
	logger := mocks.NewMockLogger()

	// Create a mock execution context that can track iterations
	ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)

	// Set up for a 3-iteration loop
	ctx.SetInputValue("iterations", types.NewValue(types.PinTypes.Number, 3.0))

	// Keep track of iterations
	iterationCount := 0
	indexes := []float64{}

	// Create a specialized execution context that simulates
	// the execution of a loop over multiple iterations
	iteratingCtx := &mockLoopExecutionContext{
		MockExecutionContext: ctx,
		onExecuteConnectedNodes: func(pinID string) error {
			// Called each time the loop body should be executed
			if pinID == "loop" {
				iterationCount++

				// Get the current loop index
				indexValue, exists := ctx.GetOutputValue("index")
				if exists {
					index, _ := indexValue.AsNumber()
					indexes = append(indexes, index)
				}

				// Continue the loop unless we've done all iterations
				if iterationCount < 3 {
					// Execute again to trigger the next iteration
					return node.Execute(ctx)
				}
			}
			return nil
		},
	}

	// Execute the node with our special context
	err := node.Execute(iteratingCtx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that we got the right number of iterations
	if iterationCount != 3 {
		t.Errorf("Expected 3 iterations, got %d", iterationCount)
	}

	// Check that indexes were incremented correctly (should be 0, 1, 2)
	if len(indexes) != 3 || indexes[0] != 0 || indexes[1] != 1 || indexes[2] != 2 {
		t.Errorf("Incorrect indexes: %v", indexes)
	}

	// The final flow should be "completed"
	if ctx.GetActivatedFlow() != "completed" {
		t.Errorf("Expected final flow to be 'completed', got '%s'", ctx.GetActivatedFlow())
	}
}

// Custom test to check the index output of the loop node for each iteration
func TestLoopNodeIndexOutput(t *testing.T) {
	// Create a LoopNode
	node := logic.NewLoopNode()
	logger := mocks.NewMockLogger()

	// Test with standard iterations
	t.Run("standard index incrementing", func(t *testing.T) {
		ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)
		ctx.SetInputValue("iterations", types.NewValue(types.PinTypes.Number, 3.0))

		actualIndexes := []float64{}

		mockCtx := &mockLoopExecutionContext{
			MockExecutionContext: ctx,
			onExecuteConnectedNodes: func(pinID string) error {
				if pinID == "loop" {
					// Record the index output
					indexValue, exists := ctx.GetOutputValue("index")
					if exists {
						index, _ := indexValue.AsNumber()
						actualIndexes = append(actualIndexes, index)
					}
				}
				return nil
			},
		}

		// Execute the node
		err := node.Execute(mockCtx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check the indexes
		expectedIndexes := []float64{0, 1, 2}
		if len(actualIndexes) != len(expectedIndexes) {
			t.Errorf("Expected %d indexes, got %d", len(expectedIndexes), len(actualIndexes))
		} else {
			for i, expected := range expectedIndexes {
				if actualIndexes[i] != expected {
					t.Errorf("Index at position %d: expected %f, got %f", i, expected, actualIndexes[i])
				}
			}
		}
	})

	// Test with custom start value
	t.Run("custom start value", func(t *testing.T) {
		ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)
		ctx.SetInputValue("iterations", types.NewValue(types.PinTypes.Number, 3.0))
		ctx.SetInputValue("startValue", types.NewValue(types.PinTypes.Number, 10.0))

		actualIndexes := []float64{}

		mockCtx := &mockLoopExecutionContext{
			MockExecutionContext: ctx,
			onExecuteConnectedNodes: func(pinID string) error {
				if pinID == "loop" {
					// Record the index output
					indexValue, exists := ctx.GetOutputValue("index")
					if exists {
						index, _ := indexValue.AsNumber()
						actualIndexes = append(actualIndexes, index)
					}
				}
				return nil
			},
		}

		// Execute the node
		err := node.Execute(mockCtx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check the indexes
		expectedIndexes := []float64{10, 11, 12}
		if len(actualIndexes) != len(expectedIndexes) {
			t.Errorf("Expected %d indexes, got %d", len(expectedIndexes), len(actualIndexes))
		} else {
			for i, expected := range expectedIndexes {
				if actualIndexes[i] != expected {
					t.Errorf("Index at position %d: expected %f, got %f", i, expected, actualIndexes[i])
				}
			}
		}
	})
}

// Mock execution context for loop testing
type mockLoopExecutionContext struct {
	*mocks.MockExecutionContext
	onExecuteConnectedNodes func(pinID string) error
}

// Implement the ExecuteConnectedNodes interface method
func (m *mockLoopExecutionContext) ExecuteConnectedNodes(pinID string) error {
	if m.onExecuteConnectedNodes != nil {
		return m.onExecuteConnectedNodes(pinID)
	}
	return nil
}

func TestBranchCompareValues(t *testing.T) {
	testCases := []struct {
		name     string
		valueA   interface{}
		valueB   interface{}
		expected bool
	}{
		{
			name:     "nil equals nil",
			valueA:   nil,
			valueB:   nil,
			expected: true,
		},
		{
			name:     "nil not equals string",
			valueA:   nil,
			valueB:   "hello",
			expected: false,
		},
		{
			name:     "string equals string",
			valueA:   "test",
			valueB:   "test",
			expected: true,
		},
		{
			name:     "string not equals different string",
			valueA:   "test1",
			valueB:   "test2",
			expected: false,
		},
		{
			name:     "float equals float",
			valueA:   42.5,
			valueB:   42.5,
			expected: true,
		},
		{
			name:     "float equals int",
			valueA:   42.0,
			valueB:   42,
			expected: true,
		},
		{
			name:     "bool equals bool",
			valueA:   true,
			valueB:   true,
			expected: true,
		},
		{
			name:     "bool not equals different bool",
			valueA:   true,
			valueB:   false,
			expected: false,
		},
		{
			name:     "map string comparison",
			valueA:   map[string]interface{}{"key": "value"},
			valueB:   map[string]interface{}{"key": "value"},
			expected: true, // This test assumes string representation comparison
		},
		{
			name:     "array string comparison",
			valueA:   []interface{}{1, 2, 3},
			valueB:   []interface{}{1, 2, 3},
			expected: true, // This test assumes string representation comparison
		},
	}

	// Create a branch node instance to access compareValues method
	branchNode := logic.NewBranchNode()

	// Using reflection to access the private compareValues function
	// This is a bit of a hack for testing purposes
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Since we can't directly access the private function, we'll test through the Execute method
			inputs := map[string]interface{}{
				"value": tc.valueA,
				"case1": tc.valueB,
			}

			var expectedFlow string
			if tc.expected {
				expectedFlow = "case1_out"
			} else {
				expectedFlow = "default"
			}

			testCase := test.NodeTestCase{
				Name:         tc.name,
				Inputs:       inputs,
				ExpectedFlow: expectedFlow,
			}

			test.ExecuteNodeTestCase(t, branchNode, testCase)
		})
	}
}
