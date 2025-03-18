package data_test

import (
	"testing"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/test"
)

func TestArrayNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "create array",
			Inputs: map[string]interface{}{
				"operation": "create",
				"size":      3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{nil, nil, nil},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "create array with size 0",
			Inputs: map[string]interface{}{
				"operation": "create",
				"size":      0.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "create array missing size",
			Inputs: map[string]interface{}{
				"operation": "create",
				// Missing size
			},
			ExpectedFlow: "error",
		},
		{
			Name: "create array invalid size",
			Inputs: map[string]interface{}{
				"operation": "create",
				"size":      "not a number",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get array element",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     1.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "b",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "get array element at index 0",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     0.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "a",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "get array element at last index",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     2.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": "c",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "get array element missing array",
			Inputs: map[string]interface{}{
				"operation": "get",
				// Missing array
				"index": 1.0,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get array element missing index",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				// Missing index
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get array element invalid array",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     "not an array",
				"index":     1.0,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get array element invalid index",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     "not a number",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set array element",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     1.0,
				"value":     "x",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"a", "x", "c"},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "set array element at index 0",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     0.0,
				"value":     "x",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"x", "b", "c"},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "set array element at last index",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     2.0,
				"value":     "x",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"a", "b", "x"},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "set array element missing array",
			Inputs: map[string]interface{}{
				"operation": "set",
				// Missing array
				"index": 1.0,
				"value": "x",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set array element missing index",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				// Missing index
				"value": "x",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set array element missing value",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     1.0,
				// Missing value
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set array element invalid array",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     "not an array",
				"index":     1.0,
				"value":     "x",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set array element invalid index",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     "not a number",
				"value":     "x",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "push to array",
			Inputs: map[string]interface{}{
				"operation": "push",
				"array":     []interface{}{"a", "b"},
				"value":     "c",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"a", "b", "c"},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "push to empty array",
			Inputs: map[string]interface{}{
				"operation": "push",
				"array":     []interface{}{},
				"value":     "a",
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"a"},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "push complex value to array",
			Inputs: map[string]interface{}{
				"operation": "push",
				"array":     []interface{}{"a", "b"},
				"value":     map[string]interface{}{"key": "value"},
			},
			ExpectedOutputs: map[string]interface{}{
				"result": []interface{}{"a", "b", map[string]interface{}{"key": "value"}},
			},
			ExpectedFlow: "then",
		},
		{
			Name: "push missing array",
			Inputs: map[string]interface{}{
				"operation": "push",
				// Missing array
				"value": "c",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "push missing value",
			Inputs: map[string]interface{}{
				"operation": "push",
				"array":     []interface{}{"a", "b"},
				// Missing value
			},
			ExpectedFlow: "error",
		},
		{
			Name: "push invalid array",
			Inputs: map[string]interface{}{
				"operation": "push",
				"array":     "not an array",
				"value":     "c",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "pop from array",
			Inputs: map[string]interface{}{
				"operation": "pop",
				"array":     []interface{}{"a", "b", "c"},
			},
			ExpectedOutputs: map[string]interface{}{
				"result":      []interface{}{"a", "b"},
				"popped_item": "c",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "pop from array with one element",
			Inputs: map[string]interface{}{
				"operation": "pop",
				"array":     []interface{}{"a"},
			},
			ExpectedOutputs: map[string]interface{}{
				"result":      []interface{}{},
				"popped_item": "a",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "pop from empty array",
			Inputs: map[string]interface{}{
				"operation": "pop",
				"array":     []interface{}{},
			},
			ExpectedFlow: "error",
		},
		{
			Name: "pop missing array",
			Inputs: map[string]interface{}{
				"operation": "pop",
				// Missing array
			},
			ExpectedFlow: "error",
		},
		{
			Name: "pop invalid array",
			Inputs: map[string]interface{}{
				"operation": "pop",
				"array":     "not an array",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get array length",
			Inputs: map[string]interface{}{
				"operation": "length",
				"array":     []interface{}{"a", "b", "c"},
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 3.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "get empty array length",
			Inputs: map[string]interface{}{
				"operation": "length",
				"array":     []interface{}{},
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 0.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "get length missing array",
			Inputs: map[string]interface{}{
				"operation": "length",
				// Missing array
			},
			ExpectedFlow: "error",
		},
		{
			Name: "get length invalid array",
			Inputs: map[string]interface{}{
				"operation": "length",
				"array":     "not an array",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "array index out of bounds",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     10.0,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "negative array index",
			Inputs: map[string]interface{}{
				"operation": "get",
				"array":     []interface{}{"a", "b", "c"},
				"index":     -1.0,
			},
			ExpectedFlow: "error",
		},
		{
			Name: "set index out of bounds",
			Inputs: map[string]interface{}{
				"operation": "set",
				"array":     []interface{}{"a", "b", "c"},
				"index":     10.0,
				"value":     "x",
			},
			ExpectedFlow: "error",
		},
		{
			Name: "missing operation",
			Inputs: map[string]interface{}{
				"array": []interface{}{"a", "b", "c"},
			},
			ExpectedFlow: "error",
		},
		{
			Name: "invalid operation",
			Inputs: map[string]interface{}{
				"operation": "invalid_op",
				"array":     []interface{}{"a", "b", "c"},
			},
			ExpectedFlow: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := data.NewArrayNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
