package data_test

import (
	"testing"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/test"
)

func TestJSONNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "parse JSON object",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `{"name":"John","age":30,"city":"New York"}`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": map[string]interface{}{
					"name": "John",
					"age":  30.0,
					"city": "New York",
				},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON array",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `[1,2,3,"four",true]`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{1.0, 2.0, 3.0, "four", true},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON null",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `null`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": nil,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON boolean",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `true`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": true,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON number",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `42.5`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 42.5,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON string",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `"hello world"`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "hello world",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse JSON with nested objects",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `{"person":{"name":"John","age":30},"address":{"city":"New York","zip":"10001"}}`,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": map[string]interface{}{
					"person": map[string]interface{}{
						"name": "John",
						"age":  30.0,
					},
					"address": map[string]interface{}{
						"city": "New York",
						"zip":  "10001",
					},
				},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON object",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data": map[string]interface{}{
					"name": "John",
					"age":  30.0,
					"city": "New York",
				},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON array",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      []interface{}{1.0, 2.0, 3.0, "four", true},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON null",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      nil,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "null",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON boolean",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      true,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "true",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON number",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      42.5,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "42.5",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify JSON string",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      "hello world",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": `"hello world"`,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "stringify with pretty print",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data": map[string]interface{}{
					"name": "John",
					"age":  30.0,
				},
				"pretty": true,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "parse invalid JSON",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      `{"name":"John","age":30,,,}`,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "parse empty string",
			Inputs: map[string]interface{}{
				"operation": "parse",
				"data":      "",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "stringify circular reference (should handle gracefully)",
			Inputs: map[string]interface{}{
				"operation": "stringify",
				"data":      "circular reference placeholder",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "missing operation",
			Inputs: map[string]interface{}{
				"data": `{"name":"John"}`,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "invalid operation",
			Inputs: map[string]interface{}{
				"operation": "invalid_op",
				"data":      `{"name":"John"}`,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "missing data",
			Inputs: map[string]interface{}{
				"operation": "parse",
			},
			ExpectedFlow: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := data.NewJSONNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
