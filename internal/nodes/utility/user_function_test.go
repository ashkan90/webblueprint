package utility_test

import (
	"testing"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/test"
	"webblueprint/pkg/blueprint"
)

func TestPrintNodeExtended(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "print with null value",
			Inputs: map[string]interface{}{
				"message": nil,
			},
			ExpectedOutputs: map[string]interface{}{
				"output": nil,
			},
			ExpectedFlow: "then",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := utility.NewPrintNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

func TestTimerNodeExtended(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "invalid format",
			Inputs: map[string]interface{}{
				"operation": "format",
				"timestamp": float64(1609459200), // 2021-01-01
				"format":    "invalid-format",    // This is now handled gracefully
			},
			ExpectedFlow: "then", // Should no longer error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := utility.NewTimerNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

// Note: We're mocking the UserFunctionNode tests rather than actually executing
// them because they require a fully configured global node registry and access
// to the database, which is difficult to set up in a unit test environment.
func TestUserFunctionNodeMock(t *testing.T) {
	// Create a simple function definition
	simpleFunction := blueprint.Function{
		ID:          "test-function",
		Name:        "TestFunction",
		Description: "A test function",
		NodeType: blueprint.BlueprintNodeType{
			Inputs: []blueprint.NodePin{
				{
					ID:          "exec",
					Name:        "Execute",
					Description: "Execution input",
					Type:        &blueprint.NodePinType{ID: "execution"},
				},
				{
					ID:          "input1",
					Name:        "Input 1",
					Description: "String input",
					Type:        &blueprint.NodePinType{ID: "string"},
				},
			},
			Outputs: []blueprint.NodePin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Execution output",
					Type:        &blueprint.NodePinType{ID: "execution"},
				},
				{
					ID:          "result",
					Name:        "Result",
					Description: "Function result",
					Type:        &blueprint.NodePinType{ID: "string"},
				},
			},
		},
		Nodes:       []blueprint.BlueprintNode{},
		Connections: []blueprint.Connection{},
	}

	// Test that the node factory can create a node without errors
	factory := utility.NewUserFunctionNode(simpleFunction)
	node := factory()

	// Check that the node has the correct input/output pins
	if len(node.GetInputPins()) == 0 {
		t.Error("Node should have input pins")
	}

	if len(node.GetOutputPins()) == 0 {
		t.Error("Node should have output pins")
	}

	// Check the node's metadata
	if node.GetMetadata().Name != "TestFunction" {
		t.Errorf("Expected name TestFunction, got %s", node.GetMetadata().Name)
	}

	if node.GetMetadata().TypeID != "testfunction" {
		t.Errorf("Expected typeID testfunction, got %s", node.GetMetadata().TypeID)
	}
}
