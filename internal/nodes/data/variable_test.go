package data_test

import (
	"testing"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/test"
)

func TestVariableNodes(t *testing.T) {
	// VariableSetNode tests
	t.Run("VariableSetNode", func(t *testing.T) {
		testCases := []test.NodeTestCase{
			{
				Name: "set string variable",
				Inputs: map[string]interface{}{
					"name":  "myVar",
					"value": "test string",
				},
				ExpectedFlow: "then",
			},
			{
				Name: "set number variable",
				Inputs: map[string]interface{}{
					"name":  "myNum",
					"value": 42.5,
				},
				ExpectedFlow: "then",
			},
			{
				Name: "set object variable",
				Inputs: map[string]interface{}{
					"name": "myObj",
					"value": map[string]interface{}{
						"key": "value",
					},
				},
				ExpectedFlow: "then",
			},
			{
				Name: "missing variable name",
				Inputs: map[string]interface{}{
					"value": "test",
				},
				ExpectedFlow: "error",
			},
			{
				Name: "invalid variable name",
				Inputs: map[string]interface{}{
					"name":  123,
					"value": "test",
				},
				ExpectedFlow: "error",
			},
			{
				Name: "missing value",
				Inputs: map[string]interface{}{
					"name": "myVar",
				},
				ExpectedFlow: "error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				node := data.NewVariableSetNode()
				test.ExecuteNodeTestCase(t, node, tc)
			})
		}
	})

	// VariableGetNode tests
	t.Run("VariableGetNode", func(t *testing.T) {
		testCases := []test.NodeTestCase{
			{
				Name: "get existing variable",
				Inputs: map[string]interface{}{
					"name": "existingVar",
				},
				ExpectedFlow: "then",
				// Can't predict the value, but it should activate "then" flow
			},
			{
				Name: "get non-existent variable",
				Inputs: map[string]interface{}{
					"name": "nonExistentVar",
				},
				ExpectedFlow: "error",
			},
			{
				Name:         "missing variable name",
				Inputs:       map[string]interface{}{},
				ExpectedFlow: "error",
			},
			{
				Name: "invalid variable name",
				Inputs: map[string]interface{}{
					"name": 123,
				},
				ExpectedFlow: "error",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				node := data.NewVariableGetNode()
				test.ExecuteNodeTestCase(t, node, tc)
			})
		}
	})
}

func TestConstantNodes(t *testing.T) {
	// StringConstantNode
	t.Run("StringConstantNode", func(t *testing.T) {
		testCase := test.NodeTestCase{
			Name:         "string constant",
			Inputs:       map[string]interface{}{},
			ExpectedFlow: "then",
		}
		node := data.NewStringConstantNode()
		test.ExecuteNodeTestCase(t, node, testCase)
	})

	// NumberConstantNode
	t.Run("NumberConstantNode", func(t *testing.T) {
		testCase := test.NodeTestCase{
			Name:         "number constant",
			Inputs:       map[string]interface{}{},
			ExpectedFlow: "then",
		}
		node := data.NewNumberConstantNode()
		test.ExecuteNodeTestCase(t, node, testCase)
	})

	// BooleanConstantNode
	t.Run("BooleanConstantNode", func(t *testing.T) {
		testCase := test.NodeTestCase{
			Name:         "boolean constant",
			Inputs:       map[string]interface{}{},
			ExpectedFlow: "then",
		}
		node := data.NewBooleanConstantNode()
		test.ExecuteNodeTestCase(t, node, testCase)
	})
}
