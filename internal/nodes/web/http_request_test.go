package web_test

import (
	"testing"
	"webblueprint/internal/nodes/web"
	"webblueprint/internal/test"
)

func TestHTTPRequestNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "missing url",
			Inputs: map[string]interface{}{
				"method": "GET",
			},
			ExpectedFlow: "catch",
		},
		{
			Name: "invalid url type",
			Inputs: map[string]interface{}{
				"url":    123,
				"method": "GET",
			},
			ExpectedFlow: "catch",
		},
		// The below tests will attempt real HTTP connections
		// In a real environment, these could be mocked with the mockTransport
		// For now, we'll expect they may succeed or fail based on network conditions
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := web.NewHTTPRequestNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
