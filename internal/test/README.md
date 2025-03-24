# WebBlueprint Testing Framework

This document outlines the testing approach and utilities for testing WebBlueprint nodes and integration scenarios.

## Overview

The WebBlueprint testing framework provides a structured way to test both individual nodes and complete blueprint executions through mock objects and test utilities.

## Directory Structure

```
internal/test/
├── integration/                  # Integration tests
│   ├── actor_execution_test.go   # Actor execution tests
│   ├── blueprint_*.go            # Blueprint tests
│   ├── comprehensive_test.go     # Comprehensive execution tests
│   ├── flow_control_test.go      # Flow control tests
│   ├── mock_nodes.go             # Mock node implementations
│   ├── test_framework.go         # Integration test framework
│   ├── variable_scoping_test.go  # Variable scoping tests
│   └── ...
├── mocks/                        # Mock implementations
│   ├── execution_context.go      # Mock ExecutionContext
│   └── logger.go                 # Mock Logger
├── node_test_utils.go            # Test utilities for node testing
└── README.md                     # Documentation (this file)
```

## Node Testing

Individual nodes are tested using the following components:

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

## Integration Testing

The `integration` directory contains comprehensive integration tests for the WebBlueprint execution engine. These tests verify that complete blueprints execute correctly in both standard and actor modes.

### Running Integration Tests

To run all integration tests:
```bash
cd integration && go test -v
```

To run specific test categories:
```bash
cd integration && go test -v -run TestPattern
```

For example:
```bash
# Run simple blueprint tests
go test -v -run TestSimple

# Run variable scoping tests
go test -v -run TestLocalVsGlobalVariables

# Run comprehensive tests (currently skipped)
go test -v -run TestComprehensiveExecution
```

### Integration Test Framework Components

1. **BlueprintTestRunner**: Manages blueprint registration and execution for testing.
2. **Mock Nodes**: Provides mock implementations of various node types.
3. **Assertion Helpers**: Verifies execution results against expected values.
4. **Node Execution Tracking**: Tracks which nodes executed and how many times.

### Known Issues with Integration Tests

- Some complex tests are temporarily skipped (TestVariableLifetime, TestComprehensiveExecution) until they can be properly integrated with updated engine components.
- Actor mode may behave differently from standard mode in some scenarios, particularly with variable scoping.
- Be cautious with timeouts in tests when running in resource-constrained environments.

## Best Practices

1. **Test Coverage**: Aim for at least 80% code coverage.
2. **Edge Cases**: Test both normal operation and edge cases.
3. **Error Handling**: Verify that errors are properly handled and reported.
4. **Variety of Inputs**: Test with different input types and values.
5. **Documentation**: Document unusual test cases or complex scenarios.
6. **Consistency**: Follow the same testing pattern across all node types.

## Running All Tests

Run node tests using the standard Go testing command:

```bash
go test ./internal/nodes/... ./internal/test/integration/...
```

To view coverage information:

```bash
go test ./internal/... -cover
```

To generate a detailed coverage report:

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```