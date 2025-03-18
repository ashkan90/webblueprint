package logic_test

import (
	"testing"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/test"
)

func TestBranchNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "match case 1",
			Inputs: map[string]interface{}{
				"value": "test",
				"case1": "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 1.0,
			},
			ExpectedFlow: "case1_out",
		},
		{
			Name: "match case 2",
			Inputs: map[string]interface{}{
				"value": "test",
				"case1": "not-match",
				"case2": "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 2.0,
			},
			ExpectedFlow: "case2_out",
		},
		{
			Name: "match case 3",
			Inputs: map[string]interface{}{
				"value": "test",
				"case1": "not-match",
				"case2": "also-not-match",
				"case3": "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 3.0,
			},
			ExpectedFlow: "case3_out",
		},
		{
			Name: "match case 4",
			Inputs: map[string]interface{}{
				"value": "test",
				"case1": "not-match",
				"case2": "also-not-match",
				"case3": "still-not-match",
				"case4": "test",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 4.0,
			},
			ExpectedFlow: "case4_out",
		},
		{
			Name: "no matches - default",
			Inputs: map[string]interface{}{
				"value": "test",
				"case1": "not-match",
				"case2": "also-not-match",
				"case3": "still-not-match",
				"case4": "yet-another-not-match",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "no cases - default",
			Inputs: map[string]interface{}{
				"value": "test",
				// No cases defined
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "missing value - default",
			Inputs: map[string]interface{}{
				// No value provided
				"case1": "case",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "number comparison - match",
			Inputs: map[string]interface{}{
				"value": 42.0,
				"case1": 42.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 1.0,
			},
			ExpectedFlow: "case1_out",
		},
		{
			Name: "boolean comparison - match",
			Inputs: map[string]interface{}{
				"value": true,
				"case2": true,
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 2.0,
			},
			ExpectedFlow: "case2_out",
		},
		{
			Name: "object comparison - match",
			Inputs: map[string]interface{}{
				"value": map[string]interface{}{"key": "value"},
				"case3": map[string]interface{}{"key": "value"},
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 3.0,
			},
			ExpectedFlow: "case3_out",
		},
		{
			Name: "array comparison - match",
			Inputs: map[string]interface{}{
				"value": []interface{}{1, 2, 3},
				"case4": []interface{}{1, 2, 3},
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 4.0,
			},
			ExpectedFlow: "case4_out",
		},
		{
			Name: "object comparison - different sizes",
			Inputs: map[string]interface{}{
				"value": map[string]interface{}{"key1": "value1", "key2": "value2"},
				"case1": map[string]interface{}{"key1": "value1"},
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "array comparison - different sizes",
			Inputs: map[string]interface{}{
				"value": []interface{}{1, 2, 3},
				"case1": []interface{}{1, 2},
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "mixed type comparison - not match",
			Inputs: map[string]interface{}{
				"value": "42",
				"case1": 42,
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
		{
			Name: "number type mixing - int and float match",
			Inputs: map[string]interface{}{
				"value": 42,
				"case1": 42.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 1.0,
			},
			ExpectedFlow: "case1_out",
		},
		{
			Name: "nil values comparison",
			Inputs: map[string]interface{}{
				"value": nil,
				"case1": nil,
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 1.0,
			},
			ExpectedFlow: "case1_out",
		},
		{
			Name: "nil and non-nil comparison",
			Inputs: map[string]interface{}{
				"value": nil,
				"case1": "not nil",
			},
			ExpectedOutputs: map[string]interface{}{
				"matched_case": 0.0,
			},
			ExpectedFlow: "default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := logic.NewBranchNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
