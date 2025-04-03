package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"webblueprint/pkg/blueprint"
)

// TestBlueprintData tests the integration with the provided blueprint data
func TestBlueprintData(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Create a test blueprint from the sample data
	bp := createBlueprintFromSampleData()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases
	testCases := []BlueprintTestCase{
		{
			Name:        "Sample blueprint test - standard mode",
			BlueprintID: bp.ID,
			Inputs: map[string]interface{}{
				"test_variable": "test_value",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "test_value",
			},
			ExecutionMode: "standard",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// createBlueprintFromSampleData creates a blueprint based on the sample data
func createBlueprintFromSampleData() *blueprint.Blueprint {
	// Create a new blueprint
	bp := blueprint.NewBlueprint("sample_blueprint", "Sample Blueprint", "1.0.0")

	// Add nodes
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "start",
		Type: "start",
		Position: blueprint.Position{
			X: 100,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "variable_get",
		Type: "get-variable-test_variable",
		Position: blueprint.Position{
			X: 300,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "variable_set",
		Type: "set-variable-result",
		Position: blueprint.Position{
			X: 500,
			Y: 100,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "end",
		Type: "end",
		Position: blueprint.Position{
			X: 700,
			Y: 100,
		},
	})

	// Add connections
	bp.AddConnection(blueprint.Connection{
		ID:             "start_to_variable_set",
		SourceNodeID:   "start",
		SourcePinID:    "out",
		TargetNodeID:   "variable_set",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "variable_set_to_end",
		SourceNodeID:   "variable_set",
		SourcePinID:    "out",
		TargetNodeID:   "end",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "variable_get_to_variable_set",
		SourceNodeID:   "variable_get",
		SourcePinID:    "value",
		TargetNodeID:   "variable_set",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	return bp
}

// TestBlueprintFromJsonData tests the creation of a blueprint from JSON data
func TestBlueprintFromJsonData(t *testing.T) {
	// Create test runner
	runner := NewBlueprintTestRunner()

	// Register mock nodes
	RegisterMockNodes(runner)

	// Parse the given JSON data and create a blueprint
	bp := createBlueprintFromPastedData()
	err := runner.RegisterBlueprint(bp)
	assert.NoError(t, err)

	// Define test cases
	testCases := []BlueprintTestCase{
		{
			Name:        "JSON blueprint test - standard mode",
			BlueprintID: bp.ID,
			Inputs: map[string]interface{}{
				"test_variable": "default_value",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "default value on prop",
			},
			ExecutionMode: "standard",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runner.AssertTestCase(t, tc)
		})
	}
}

// createBlueprintFromPastedData creates a blueprint based on the JSON data
func createBlueprintFromPastedData() *blueprint.Blueprint {
	// Create a new blueprint
	bp := blueprint.NewBlueprint("new_blueprint7", "New Blueprint7", "1.0.0")

	// Add nodes based on the provided JSON data
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "0753e85b-a169-4036-97b5-c5c972c815e7",
		Type: "constant-string",
		Position: blueprint.Position{
			X: 1032.3333333333335,
			Y: 468.2222222222223,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "constantValue",
				DisplayName: "",
				Description: "",
				Value:       "default value on prop",
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
		Type: "print",
		Position: blueprint.Position{
			X: 1265.2222222222224,
			Y: 218.33333333333331,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
		Type: "http-request",
		Position: blueprint.Position{
			X: 534,
			Y: 195,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "input_method",
				DisplayName: "",
				Description: "",
				Value:       "GET",
				Type:        nil,
			},
			{
				Name:        "input_url",
				DisplayName: "",
				Description: "",
				Value:       "https://jsonplaceholder.typicode.com/todos/1",
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "90a3774f-0fe8-4601-915b-c6a721e91153",
		Type: "object-operations",
		Position: blueprint.Position{
			X: 792,
			Y: 155.55555555555566,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "input_operation",
				DisplayName: "",
				Description: "",
				Value:       "get",
				Type:        nil,
			},
			{
				Name:        "input_key",
				DisplayName: "",
				Description: "",
				Value:       "title",
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "50df077e-5303-4cf7-a912-da95f47dc722",
		Type: "print",
		Position: blueprint.Position{
			X: 1052.0000000000007,
			Y: 204.8888888888888,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "33df614d-e269-4113-ba42-c9794c115e23",
		Type: "variable-set",
		Position: blueprint.Position{
			X: 1173,
			Y: 638.5,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "input_name",
				DisplayName: "",
				Description: "",
				Value:       "test_variable",
				Type:        nil,
			},
			{
				Name:        "input_value",
				DisplayName: "",
				Description: "",
				Value:       "default_value",
				Type:        nil,
			},
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
		Type: "print",
		Position: blueprint.Position{
			X: 1542,
			Y: 261.5,
		},
	})

	bp.AddNode(blueprint.BlueprintNode{
		ID:   "c482a3e0-99bc-427d-91db-b994f683d14c",
		Type: "variable-get",
		Position: blueprint.Position{
			X: 1281,
			Y: 469.5,
		},
		Properties: []blueprint.NodeProperty{
			{
				Name:        "input_name",
				DisplayName: "",
				Description: "",
				Value:       "test_variable",
				Type:        nil,
			},
		},
	})

	// Add connections
	bp.AddConnection(blueprint.Connection{
		ID:             "e9c6e13b-61e4-4aab-a6c6-400f52eb8d49",
		SourceNodeID:   "0753e85b-a169-4036-97b5-c5c972c815e7",
		SourcePinID:    "value",
		TargetNodeID:   "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
		TargetPinID:    "message",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "89179cba-e2b2-464f-a947-b60aeacb5c1d",
		SourceNodeID:   "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
		SourcePinID:    "then",
		TargetNodeID:   "90a3774f-0fe8-4601-915b-c6a721e91153",
		TargetPinID:    "exec",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "569da363-f9dd-47d9-9c52-02f10f735001",
		SourceNodeID:   "1dda3f01-6d79-4a31-bd7a-6786ceb80abb",
		SourcePinID:    "response",
		TargetNodeID:   "90a3774f-0fe8-4601-915b-c6a721e91153",
		TargetPinID:    "object",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "e23f9e26-52f2-4e50-8250-cfc12140d015",
		SourceNodeID:   "90a3774f-0fe8-4601-915b-c6a721e91153",
		SourcePinID:    "then",
		TargetNodeID:   "50df077e-5303-4cf7-a912-da95f47dc722",
		TargetPinID:    "exec",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "52d141f7-2b31-47f0-95ca-395e62351de2",
		SourceNodeID:   "90a3774f-0fe8-4601-915b-c6a721e91153",
		SourcePinID:    "result",
		TargetNodeID:   "50df077e-5303-4cf7-a912-da95f47dc722",
		TargetPinID:    "message",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "ba14bdce-114b-42c5-98ce-271c1b48aa43",
		SourceNodeID:   "50df077e-5303-4cf7-a912-da95f47dc722",
		SourcePinID:    "then",
		TargetNodeID:   "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
		TargetPinID:    "exec",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "fd3de6f8-96f6-4827-9a5c-304a757b6808",
		SourceNodeID:   "fbbd33a5-626b-4a2e-997b-b9eb1647a095",
		SourcePinID:    "then",
		TargetNodeID:   "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
		TargetPinID:    "exec",
		ConnectionType: "execution",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "059a0d4a-ae71-460d-818f-3abee09bab8a",
		SourceNodeID:   "c482a3e0-99bc-427d-91db-b994f683d14c",
		SourcePinID:    "value",
		TargetNodeID:   "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
		TargetPinID:    "message",
		ConnectionType: "data",
	})

	// Add a connection to set the result output
	bp.AddNode(blueprint.BlueprintNode{
		ID:   "result_setter",
		Type: "set-variable-result",
		Position: blueprint.Position{
			X: 1800,
			Y: 300,
		},
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "constant_to_result",
		SourceNodeID:   "0753e85b-a169-4036-97b5-c5c972c815e7",
		SourcePinID:    "value",
		TargetNodeID:   "result_setter",
		TargetPinID:    "value",
		ConnectionType: "data",
	})

	bp.AddConnection(blueprint.Connection{
		ID:             "print_to_result_setter",
		SourceNodeID:   "e2a3d94f-e636-4bd0-b86b-73f45620ce1c",
		SourcePinID:    "then",
		TargetNodeID:   "result_setter",
		TargetPinID:    "in",
		ConnectionType: "execution",
	})

	return bp
}
