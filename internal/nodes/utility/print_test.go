package utility_test

import (
	"testing"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/test"
)

func TestPrintNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "print string",
			Inputs: map[string]interface{}{
				"message": "Hello, World!",
			},
			ExpectedOutputs: map[string]interface{}{
				"output": "Hello, World!",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "print number",
			Inputs: map[string]interface{}{
				"message": 42.5,
			},
			ExpectedOutputs: map[string]interface{}{
				"output": 42.5,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "print boolean",
			Inputs: map[string]interface{}{
				"message": true,
			},
			ExpectedOutputs: map[string]interface{}{
				"output": true,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "print object",
			Inputs: map[string]interface{}{
				"message": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			ExpectedOutputs: map[string]interface{}{
				"output": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "print with prefix",
			Inputs: map[string]interface{}{
				"message": "World",
				"prefix":  "Hello,",
			},
			ExpectedOutputs: map[string]interface{}{
				"output": "World",
			},
			ExpectedFlow: "then",
		},
		{
			Name:   "print without message",
			Inputs: map[string]interface{}{},
			ExpectedOutputs: map[string]interface{}{
				"output": "[undefined]",
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
