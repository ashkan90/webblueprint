# WebBlueprint Enhanced Error Handling

## Overview

The Enhanced Error Handling and Recovery system provides comprehensive error management, diagnostics, and recovery mechanisms for the WebBlueprint platform. This system improves the robustness and user experience of visual programming workflows by making errors easier to understand, diagnose, and recover from.

## Key Features

### 1. Structured Error Classification

Errors are classified by:

- **Type**: Categorizes errors by domain (execution, connection, validation, etc.)
- **Severity**: Indicates criticality (critical, high, medium, low)
- **Code**: Provides specific error codes for precise identification

### 2. Contextual Error Information

Each error contains rich contextual data:

- Node and pin identification
- Blueprint and execution context
- Timestamp and execution path
- Original error details
- Stack trace when available

### 3. Recovery Mechanisms

Multiple recovery strategies:

- **Retry**: Attempts the operation again
- **Skip Node**: Continues execution by skipping problematic nodes
- **Default Values**: Uses fallback values for missing or invalid inputs
- **Manual Intervention**: Allows user interaction to resolve complex errors

### 4. Error Analysis

Comprehensive error analysis capabilities:

- Error frequency patterns
- Most problematic nodes identification
- Error type distribution
- Recovery success rates

### 5. UI Components

Advanced error management interface:

- Error visualization with severity highlighting
- Filtering and sorting capabilities
- Interactive recovery options
- Error testing utilities

## System Architecture

The error handling system is designed as a modular, non-intrusive extension to the existing WebBlueprint architecture:

### Backend Components

1. **Error Types and Classification** (`internal/bperrors/error_types.go`)
   - Defines error types, codes, and severity levels
   - Provides structured error objects with contextual metadata

2. **Error Manager** (`internal/bperrors/error_manager.go`)
   - Records and tracks errors during execution
   - Provides analysis and diagnostic capabilities
   - Registers custom error handlers

3. **Recovery Manager** (`internal/bperrors/recovery.go`)
   - Manages error recovery strategies
   - Tracks recovery attempts
   - Provides default value handling

4. **Error-Aware Context** (`internal/bperrors/error_context.go`)
   - Extends execution context with error handling capabilities
   - Works with both standard and actor-based execution models
   - Compatible with existing node implementations

5. **Error-Aware Engine** (`internal/engine/error_aware_engine.go`)
   - Integrates error handling with the execution engine
   - Supports blueprint validation
   - Enhances execution results with error diagnostics

### Frontend Components

1. **Error Types** (`web/src/types/errors.ts`)
   - TypeScript definitions for error structures
   - Type-safe error handling in UI components

2. **Error Store** (`web/src/stores/errorStore.ts`)
   - Manages error state in the frontend
   - Provides actions for error handling and recovery

3. **Error View Store** (`web/src/stores/errorViewStore.ts`)
   - Manages error display preferences
   - Handles error filtering and interaction

4. **Error Panel** (`web/src/components/debug/ErrorPanel.vue`)
   - Visualizes errors with severity highlighting
   - Provides filtering and recovery options

5. **Error Testing Panel** (`web/src/components/debug/ErrorTestingPanel.vue`)
   - Generates test errors for development and testing
   - Simulates error scenarios

6. **WebSocket Handler** (`web/src/composables/useErrorWebSocketHandler.ts`)
   - Processes real-time error notifications
   - Updates UI when new errors occur

## Integration

The error handling system is designed for minimal intrusion into existing code:

### For Node Developers

To add error handling to a node:

```go
// Check if context supports error handling
errorCtx, isErrorAware := ctx.(bperrors.ErrorAwareContext)

// If supported, use enhanced error handling
if isErrorAware {
    // Report an error with context
    err := errorCtx.ReportError(
        bperrors.ErrorTypeExecution,
        bperrors.ErrNodeExecutionFailed,
        "Error message",
        originalError,
    )
    
    // Try to recover
    success, details := errorCtx.AttemptRecovery(err)
    if success {
        // Handle successful recovery
        // Use details to adjust behavior
    } else {
        // Handle failure case
    }
} else {
    // Fall back to standard error handling
    return fmt.Errorf("error message")
}
```

### Server Integration

```go
// Create base engine
baseEngine := engine.NewExecutionEngine()

// Wrap with error handling
errorAwareEngine := engine.NewErrorAwareEngine(baseEngine)

// Get error components
errorManager := errorAwareEngine.GetErrorManager()
recoveryManager := errorAwareEngine.GetRecoveryManager()

// Register custom error handlers if needed
errorManager.RegisterErrorHandler(bperrors.ErrorTypeExecution, func(err *bperrors.BlueprintError) error {
    log.Printf("[ERROR] %s", err.Error())
    return nil
})
```

### Frontend Integration

```typescript
// Initialize stores
const errorStore = useErrorStore();
const errorViewStore = useErrorViewStore();

// Set up WebSocket handling
const errorWsHandler = useErrorWebSocketHandler(wsConnection);
errorWsHandler.init();

// Handle errors
errorStore.$subscribe((mutation, state) => {
  // React to error changes
});
```

## Example Nodes with Error Handling

The system includes example nodes demonstrating error handling:

1. **HTTP Request Node** (`internal/nodes/web/http_request_with_recovery.go`)
   - Handles network failures with retry
   - Supports fallback URLs
   - Provides detailed error diagnostics

2. **Safe Divide Node** (`internal/nodes/math/safe_divide.go`)
   - Handles division by zero
   - Uses default values for missing inputs
   - Provides error flow path

## Testing and Development

The system includes tools for testing error handling:

1. **Error Testing API** (`internal/api/error_testing_api.go`)
   - Generates test errors with specific properties
   - Simulates error scenarios

2. **Testing Panel** (`web/src/components/debug/ErrorTestingPanel.vue`)
   - User interface for generating test errors
   - Visualizes testing results

## Implementation Notes

The enhanced error handling system follows these design principles:

1. **Non-intrusive**: Works alongside existing code without requiring changes
2. **Backward compatible**: Standard nodes continue to work without modification
3. **Progressive enhancement**: Error-aware nodes get enhanced capabilities
4. **Performance conscious**: Minimal overhead during normal operation
5. **User-focused**: Prioritizes useful error information and recovery options

## Future Enhancements

Planned improvements:

1. **AI-assisted recovery**: Suggest recovery strategies based on error patterns
2. **Error prediction**: Identify potential issues before execution
3. **Blueprint validation rules**: Custom validation rules for blueprints
4. **Error telemetry**: Optional error reporting for improving the platform
5. **Smart defaults**: Context-aware default values for recovery

## Conclusion

The Enhanced Error Handling and Recovery system significantly improves the robustness and user experience of WebBlueprint by providing comprehensive error management capabilities while maintaining compatibility with the existing architecture.
