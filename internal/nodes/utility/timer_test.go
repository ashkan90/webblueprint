package utility_test

import (
	"testing"
	"time"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/test"
)

func TestTimerNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "get current time",
			Inputs: map[string]interface{}{
				"operation": "current_time",
			},
			ExpectedFlow: "then",
			// We can't predict the exact time value, but the output should exist
		},
		{
			Name: "delay operation",
			Inputs: map[string]interface{}{
				"operation": "delay",
				"duration":  0.001, // Very small delay for testing
			},
			ExpectedFlow: "then",
		},
		{
			Name: "format timestamp",
			Inputs: map[string]interface{}{
				"operation": "format",
				"timestamp": float64(time.Now().Unix()),
				"format":    "2006-01-02",
			},
			ExpectedFlow: "then",
		},
		{
			Name: "invalid operation",
			Inputs: map[string]interface{}{
				"operation": "invalid_op",
			},
			ExpectedFlow: "error",
		},
		{
			Name:         "missing operation",
			Inputs:       map[string]interface{}{},
			ExpectedFlow: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := utility.NewTimerNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
