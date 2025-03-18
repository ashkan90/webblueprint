package logic_test

import (
	"testing"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/test"
)

func TestIfConditionNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "true condition",
			Inputs: map[string]interface{}{
				"condition": true,
			},
			ExpectedFlow: "true",
		},
		{
			Name: "false condition",
			Inputs: map[string]interface{}{
				"condition": false,
			},
			ExpectedFlow: "false",
		},
		{
			Name:          "missing condition",
			Inputs:        map[string]interface{}{},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "non-boolean condition - string",
			Inputs: map[string]interface{}{
				"condition": "not a boolean",
			},
			ExpectedFlow: "true", // Non-empty strings evaluate to true in Go
		},
		// Additional test cases to increase coverage
		{
			Name: "non-boolean condition - number non-zero",
			Inputs: map[string]interface{}{
				"condition": 42.0,
			},
			ExpectedFlow: "true", // Non-zero numbers evaluate to true
		},
		{
			Name: "non-boolean condition - number zero",
			Inputs: map[string]interface{}{
				"condition": 0.0,
			},
			ExpectedFlow: "false", // Zero evaluates to false
		},
		{
			Name: "non-boolean condition - empty string",
			Inputs: map[string]interface{}{
				"condition": "",
			},
			ExpectedFlow: "false", // Empty string evaluates to false
		},
		{
			Name: "non-boolean condition - nil",
			Inputs: map[string]interface{}{
				"condition": nil,
			},
			ExpectedFlow: "false", // Nil evaluates to false
		},
		{
			Name: "non-boolean condition - object",
			Inputs: map[string]interface{}{
				"condition": map[string]interface{}{"key": "value"},
			},
			ExpectedFlow: "true", // Non-nil objects evaluate to true
		},
		{
			Name: "non-boolean condition - array",
			Inputs: map[string]interface{}{
				"condition": []interface{}{1, 2, 3},
			},
			ExpectedFlow: "true", // Non-empty arrays evaluate to true
		},
		{
			Name: "non-boolean condition - empty array",
			Inputs: map[string]interface{}{
				"condition": []interface{}{},
			},
			ExpectedFlow: "true", // Even empty arrays are not nil, so evaluate to true
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := logic.NewIfConditionNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
