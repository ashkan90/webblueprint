package math_test

import (
	"testing"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/test"
)

func TestSafeDivideNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "normal division",
			Inputs: map[string]interface{}{
				"dividend": 10.0,
				"divisor":  2.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 5.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "division with fraction result",
			Inputs: map[string]interface{}{
				"dividend": 5.0,
				"divisor":  2.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": 2.5,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "division by zero",
			Inputs: map[string]interface{}{
				"dividend": 10.0,
				"divisor":  0.0,
			},
			ExpectedFlow: "catch",
			ExpectedOutputs: map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Division by zero",
				},
			},
		},
		{
			Name: "division with negative numbers",
			Inputs: map[string]interface{}{
				"dividend": -10.0,
				"divisor":  2.0,
			},
			ExpectedOutputs: map[string]interface{}{
				"result": -5.0,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "missing dividend input",
			Inputs: map[string]interface{}{
				"divisor": 2.0,
			},
			ExpectedError: true,
			// Don't check specific error message as it may depend on implementation
		},
		{
			Name: "missing divisor input",
			Inputs: map[string]interface{}{
				"dividend": 10.0,
			},
			ExpectedError: true,
			// Don't check specific error message as it may depend on implementation
		},
		{
			Name: "invalid dividend type",
			Inputs: map[string]interface{}{
				"dividend": "not a number",
				"divisor":  2.0,
			},
			ExpectedFlow: "catch",
		},
		{
			Name: "invalid divisor type",
			Inputs: map[string]interface{}{
				"dividend": 10.0,
				"divisor":  "not a number",
			},
			ExpectedFlow: "catch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := math.NewSafeDivideNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
