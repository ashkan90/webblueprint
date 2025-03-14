# WebBlueprint Error Handling

## Overview

WebBlueprint's Enhanced Error Handling and Recovery system provides comprehensive error management, reporting, diagnostics, and recovery mechanisms. This feature enables more robust and maintainable visual programming workflows by making errors easier to understand and recover from.

## Key Features

### Error Classification

Errors are classified by:
- **Type**: Categorizes errors by their domain (execution, connection, validation, etc.)
- **Severity**: Indicates how critical the error is (critical, high, medium, low)
- **Code**: Provides a specific error code for precise identification

### Error Metadata

Each error contains rich metadata:
- Context information (node ID, pin ID, blueprint ID, execution ID)
- Detailed error message
- Additional details specific to the error
- Timestamp
- Original error (if wrapping another error)

### Recovery Mechanisms

The system supports multiple recovery strategies:
- **Retry**: Attempts the operation again
- **Skip Node**: Continues execution by skipping the problematic node
- **Use Default Value**: Substitutes a default value when input is missing or invalid
- **Manual Intervention**: Requires user interaction to resolve the error

### Error Analysis

The system provides error analysis capabilities:
- Error frequency by type and severity
- Most problematic nodes identification
- Recovery success rates
- Execution impact assessment

### UI Components

Advanced UI components for error management:
- Error panel with filtering, sorting, and detailed information
- Error recovery interface
- Error testing utility
- Visual highlighting of error sources

## Integration Points

The error handling system integrates with several key components:

- **Execution Engine**: Records, classifies, and attempts recovery from errors during blueprint execution
- **API Layer**: Standardizes error responses with appropriate HTTP status codes
- **WebSocket**: Provides real-time error notifications and updates
- **Node Implementations**: Enables nodes to use enhanced error handling capabilities
- **UI Components**: Visualizes errors and provides recovery interfaces

## Usage Examples

### Reporting and Recovering from Errors in Nodes

```go
// Cast to error-aware context if available
errorAwareCtx, isErrorAware := ctx.(*engine.ErrorAwareExecutionContext)

// Report an error with rich context information
if isErrorAware {
    err := errorAwareCtx.ReportError(
        errors.ErrorTypeExecution,
        errors.ErrNodeExecutionFailed,
        "Division by zero",
        originalError,
    )
    
    // Try to recover
    success, details := errorAwareCtx.AttemptRecovery(err)
    if success {
        // Use recovery details to handle the error
        defaultValue := details["defaultValue"]
        ctx.Logger().Info("Recovered from error", map[string]interface{}{
            "using": defaultValue,
        })
    } else {
        // Failed to recover, activate error flow
        ctx.ActivateOutputFlow("catch")
        return err
    }
}
```

### Testing Error Handling

```go
// Create a test scenario
generator := errors.NewTestErrorGenerator()
analysis, err := generator.SimulateErrorScenario("execution_failure", "test-execution-id")

// Verify error handling behavior
verifier := errors.NewTestVerifier(generator.GetErrorManager(), generator.GetRecoveryManager())
report := verifier.GenerateVerificationReport("test-execution-id")
```

## Components

### Core Components

- **ErrorManager**: Handles error recording, analysis, and classification
- **RecoveryManager**: Manages recovery attempts and strategies
- **BlueprintValidator**: Validates blueprint structure with detailed error reporting
- **TestErrorGenerator**: Creates test error scenarios for development and testing
- **TestVerifier**: Verifies error handling behavior

### UI Components

- **ErrorPanel**: Displays and filters errors
- **ErrorTestingPanel**: Generates test errors for development

### Integration Components

- **ErrorAwareExecutionContext**: Extends execution context with error handling capabilities
- **ErrorAwareExecutionEngine**: Adds error handling to the execution engine
- **ErrorNotificationHandler**: Manages WebSocket notifications for errors

## Documentation

For more detailed information about the error handling system, see:

- [Error Handling Architecture](README_error_handling.md)
- [Error Handling Guide](web/public/error_handling_guide.html)
- [API Reference](internal/errors/)

## Best Practices

1. Use specific error types and codes rather than generic ones
2. Include relevant context in error details
3. Determine recoverability based on the error's nature and severity
4. Prefer automatic recovery for non-critical errors
5. Use manual intervention for critical errors that need human decision making
6. Log all errors, even recoverable ones
7. Test recovery strategies thoroughly
