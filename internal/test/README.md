# WebBlueprint Node Testing Framework

This document outlines the testing approach and utilities for testing WebBlueprint nodes.

## Overview

The WebBlueprint testing framework provides a structured way to test node behavior through mock objects and test utilities. Tests focus on:

1. Node initialization
2. Input validation
3. Business logic correctness
4. Output values
5. Flow activation
6. Error handling

## Directory Structure

```
internal/test/
├── mocks/                     # Mock implementations
│   ├── execution_context.go   # Mock ExecutionContext
│   └── logger.go              # Mock Logger
├── node_test_utils.go         # Test utilities for node testing
└── README.md                  # Documentation (this file)
```

## Test Components

### Mock Objects

#### MockExecutionContext

Simulates a node execution context with the following capabilities:
- Setting and retrieving input values
- Capturing output values
- Tracking activated flows
- Simulating variable access
- Capturing debug information

#### MockLogger

Simulates a logger with the following capabilities:
- Recording log messages at different levels (Debug, Info, Warn, Error)
- Retrieving logged messages for assertions
- Filtering logs by level
- Converting logs to string for comparison

### Testing Utilities

The `ExecuteNodeTestCase` function facilitates node testing by:

1. Setting up the test environment with mock objects
2. Configuring input values based on test case specification
3. Executing the node
4. Validating outputs against expected values
5. Checking flow activation
6. Handling error cases

## Writing Node Tests

To test a node, create a test file with a structure similar to this:

```go
package mypackage_test

import (
	"testing"
	"webblueprint/internal/nodes/mypackage"
	"webblueprint/internal/test"
)

func TestMyNode(t *testing.T) {
	testCases := []test.NodeTestCase{
		{
			Name: "normal operation",
			Inputs: map[string]interface{}{
				"input1": value1,
				"input2": value2,
			},
			ExpectedOutputs: map[string]interface{}{
				"output1": expectedValue1,
			},
			ExpectedFlow: "then",
		},
		{
			Name: "error case",
			Inputs: map[string]interface{}{
				"input1": invalidValue,
			},
			ExpectedError: true,
			ErrorContains: "expected error substring",
		},
		// Additional test cases...
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			node := mypackage.NewMyNode()
			test.ExecuteNodeTestCase(t, node, tc)
		})
	}
}
```

### NodeTestCase Structure

- `Name`: Descriptive name for the test case
- `Inputs`: Map of input pin names to values
- `ExpectedOutputs`: Map of output pin names to expected values
- `ExpectedFlow`: The expected flow pin to be activated
- `ExpectedError`: Whether an error is expected
- `ErrorContains`: Expected substring in the error message

## Best Practices

1. **Test Coverage**: Aim for at least 80% code coverage.
2. **Edge Cases**: Test both normal operation and edge cases.
3. **Error Handling**: Verify that errors are properly handled and reported.
4. **Variety of Inputs**: Test with different input types and values.
5. **Documentation**: Document unusual test cases or complex scenarios.
6. **Consistency**: Follow the same testing pattern across all node types.

## Running Tests

Run node tests using the standard Go testing command:

```bash
go test ./internal/nodes/...
```

To view coverage information:

```bash
go test ./internal/nodes/... -cover
```

To generate a detailed coverage report:

```bash
go test ./internal/nodes/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Common Testing Issues

### Mocking External Dependencies

When testing nodes that interact with external systems (like HTTP requests or file operations):

1. Create specialized mock objects that simulate the external system
2. Use dependency injection to replace real implementations with mocks
3. For HTTP, consider using `httptest` package to create a mock server
4. Focus on testing the node's error handling behavior

### Testing Asynchronous Behavior

For nodes with asynchronous behavior (like timers or events):

1. Use dependency injection to control time-related functions
2. Add specific test modes that synchronize execution
3. Implement timeout mechanisms in tests to avoid hanging

### Testing Internal State

To verify internal state that isn't exposed through outputs:

1. Use debug info capture to record internal state
2. Add test-specific accessor methods
3. Verify state through logs when appropriate

### Handling Non-Deterministic Behavior

For nodes with non-deterministic outcomes:

1. Mock random number generators or similar functions
2. Focus tests on behavior patterns rather than exact values
3. Use ranges or conditionals for assertions when appropriate
