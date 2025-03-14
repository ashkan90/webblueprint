# Enhanced Error Handling and Recovery

This document provides an overview of the Enhanced Error Handling and Recovery system implemented in WebBlueprint.

## Overview

The enhanced error handling system provides:

1. Comprehensive error classification and structured error objects
2. Error recovery mechanisms for common failure scenarios
3. Detailed error analysis and diagnostics
4. UI components for error visualization and recovery
5. API error standardization

## Key Components

### Error Types and Classification

The system categorizes errors into specific types:
- `ErrorTypeExecution`: Errors during blueprint execution
- `ErrorTypeConnection`: Errors related to node connections
- `ErrorTypeValidation`: Errors validating blueprint structure
- `ErrorTypePermission`: Errors related to permissions
- `ErrorTypeDatabase`: Errors interacting with the database
- `ErrorTypeNetwork`: Network-related errors
- `ErrorTypePlugin`: Plugin-related errors
- `ErrorTypeSystem`: System-level errors
- `ErrorTypeUnknown`: Unclassified errors

### Severity Levels

Errors are assigned a severity level:
- `SeverityCritical`: System-breaking errors
- `SeverityHigh`: Errors that prevent operation but not system-breaking
- `SeverityMedium`: Errors that affect functionality but allow continued operation
- `SeverityLow`: Errors that are minor and don't significantly affect operation
- `SeverityInfo`: Informational errors

### Recovery Strategies

The system supports several recovery strategies:
- `RecoveryRetry`: Retry the operation
- `RecoverySkipNode`: Skip the problematic node
- `RecoveryUseDefaultValue`: Use a default value for the operation
- `RecoveryManualIntervention`: Require manual intervention
- `RecoveryNone`: No recovery possible

## Blueprint Error Structure

Each error is represented as a `BlueprintError` object with comprehensive metadata:

```go
type BlueprintError struct {
    Type            ErrorType         
    Code            BlueprintErrorCode 
    Message         string            
    Details         map[string]interface{} 
    Severity        ErrorSeverity     
    Recoverable     bool              
    RecoveryOptions []RecoveryStrategy 
    NodeID          string            
    PinID           string            
    BlueprintID     string            
    ExecutionID     string            
    Timestamp       time.Time         
    OriginalError   error             
    StackTrace      []string          
}
```

## Error Management

The `ErrorManager` handles error recording, analysis, and recovery strategies:

```go
type ErrorManager struct {
    errors            map[string][]*BlueprintError // Maps executionID to errors
    errorHandlers     map[ErrorType][]ErrorHandler
    recoveryStrategies map[BlueprintErrorCode][]RecoveryStrategy
}
```

It provides functionality for:
- Recording errors
- Registering error handlers
- Managing recovery strategies
- Analyzing error patterns
- Attempting recovery

## Recovery Management

The `RecoveryManager` handles error recovery attempts:

```go
type RecoveryManager struct {
    errorManager      *ErrorManager
    defaultProviders  map[*types.PinType]DefaultValueProvider
    recoveryAttempts  map[string]map[string][]RecoveryContext
}
```

It provides functionality for:
- Registering default value providers
- Attempting recovery from errors
- Tracking recovery attempts
- Providing default values

## UI Components

The system includes a dedicated error panel for visualizing and interacting with errors:

- `ErrorPanel.vue`: Displays errors with severity filtering, recovery options, and detailed error information
- `errorHandler.ts`: Manages error state in the frontend

## Integration with Execution Engine

The error handling system is integrated with the execution engine:

- `ErrorAwareExecutionContext`: Extends the execution context with error handling capabilities
- `ErrorAwareExecutionEngine`: Adds error handling to the execution engine
- `ExtendedExecutionResult`: Includes error analysis and recovery information in execution results

## API Error Handling

The system standardizes API error responses:

```go
type ErrorResponse struct {
    Success   bool
    Error     *errors.BlueprintError
    Message   string
    ErrorCode string
    Details   map[string]interface{}
    Timestamp string
}
```

## WebSocket Error Notifications

The system provides real-time error notifications via WebSocket:

- `ErrorNotification`: Sends information about new errors
- `ErrorAnalysisNotification`: Sends error analysis updates
- `RecoveryNotification`: Sends information about recovery attempts

## Usage Examples

### Recording an Error

```go
// Create a BlueprintError
err := errors.New(
    errors.ErrorTypeExecution,
    errors.ErrNodeExecutionFailed,
    "Node execution failed",
    errors.SeverityHigh,
).WithNodeInfo(nodeID, pinID).WithBlueprintInfo(blueprintID, executionID)

// Record the error
errorManager.RecordError(executionID, err)
```

### Attempting Recovery

```go
// Attempt recovery from an error
if success, details := recoveryManager.RecoverFromError(executionID, err); success {
    // Recovery was successful
    log.Println("Recovery successful:", details)
} else {
    // Recovery failed
    log.Println("Recovery failed")
}
```

### Error Analysis

```go
// Get error analysis
analysis := errorManager.AnalyzeErrors(executionID)
fmt.Printf("Total errors: %d, Recoverable: %d\n", 
    analysis["totalErrors"], analysis["recoverableErrors"])
```
