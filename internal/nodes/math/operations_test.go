package math_test

import (
	"testing"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/test"
)

func TestAddNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "add positive numbers",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 8.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "add negative numbers",
			Inputs: map[string]interface{}{
				"a": -5.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": -8.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "add positive and negative",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "add zeros",
			Inputs: map[string]interface{}{
				"a": 0.0,
				"b": 0.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 0.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "missing input a",
			Inputs: map[string]interface{}{
				"b": 3.0,
			},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "missing input b",
			Inputs: map[string]interface{}{
				"a": 5.0,
			},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "invalid input a",
			Inputs: map[string]interface{}{
				"a": "not a number",
				"b": 3.0,
			},
			ExpectedError: true,
			ErrorContains: "cannot convert string",
		},
		{
			Name: "invalid input b",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": "not a number",
			},
			ExpectedError: true,
			ErrorContains: "cannot convert string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := math.NewAddNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

func TestSubtractNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "subtract positive numbers",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "subtract to negative",
			Inputs: map[string]interface{}{
				"a": 3.0,
				"b": 5.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": -2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "subtract negative numbers",
			Inputs: map[string]interface{}{
				"a": -3.0,
				"b": -5.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "subtract zeros",
			Inputs: map[string]interface{}{
				"a": 0.0,
				"b": 0.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 0.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "missing inputs",
			Inputs: map[string]interface{}{
				"a": 5.0,
			},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "invalid inputs",
			Inputs: map[string]interface{}{
				"a": "not a number",
				"b": 3.0,
			},
			ExpectedError: true,
			ErrorContains: "cannot convert string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := math.NewSubtractNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

func TestMultiplyNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "multiply positive numbers",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 15.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "multiply by zero",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 0.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 0.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "multiply negative numbers",
			Inputs: map[string]interface{}{
				"a": -5.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 15.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "multiply positive and negative",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": -15.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "missing inputs",
			Inputs: map[string]interface{}{
				"a": 5.0,
			},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "invalid inputs",
			Inputs: map[string]interface{}{
				"a": "not a number",
				"b": 3.0,
			},
			ExpectedError: true,
			ErrorContains: "cannot convert string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := math.NewMultiplyNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}

func TestDivideNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "divide positive numbers",
			Inputs: map[string]interface{}{
				"a": 6.0,
				"b": 3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "divide to fraction",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 2.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.5,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "divide negative numbers",
			Inputs: map[string]interface{}{
				"a": -6.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "divide positive by negative",
			Inputs: map[string]interface{}{
				"a": 6.0,
				"b": -3.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": -2.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "divide by zero",
			Inputs: map[string]interface{}{
				"a": 5.0,
				"b": 0.0,
			},
			ExpectedFlow: "error",
			ExpectedOutputs: map[string]interface{}{
				"result": "Division by zero",
			},
		},
		{
			Name: "missing inputs",
			Inputs: map[string]interface{}{
				"a": 5.0,
			},
			ExpectedError: true,
			ErrorContains: "missing required input",
		},
		{
			Name: "invalid inputs",
			Inputs: map[string]interface{}{
				"a": "not a number",
				"b": 3.0,
			},
			ExpectedError: true,
			ErrorContains: "cannot convert string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := math.NewDivideNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
