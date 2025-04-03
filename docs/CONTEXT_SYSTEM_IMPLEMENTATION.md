# WebBlueprint Context System Implementation

## Overview

The WebBlueprint Context System has been refactored to address the "context hell" issue where multiple overlapping context types were creating maintenance challenges and import cycles. The new implementation provides a clean and modular approach to context management with clear separation of responsibilities.

## Implementation Details

### Key Components

1. **Core Interfaces**: Located in `internal/core` package, these interfaces define the contracts for various context capabilities without creating import cycles.

2. **Base Context**: A minimal `DefaultExecutionContext` that implements the `node.ExecutionContext` interface and provides the basic functionality.

3. **Context Decorators**: Clean decorators like `ErrorAwareContext` and `EventAwareContext` that add specific capabilities to contexts.

4. **Context Manager**: A centralized factory for creating contexts with desired capabilities.

### Package Structure

- **internal/core**: Contains core interfaces for cross-package communication
- **internal/engineext**: Contains the simplified context system implementations
- **internal/bperrors**: Contains error handling utilities 
- **internal/node**: Contains base node interfaces
- **internal/types**: Contains type system definitions

### Resolving Import Cycles

The previous implementation suffered from import cycles between packages:

```
engine → event → engine  # Cycle!
```

We resolved this by:

1. Moving shared interfaces to the `core` package that both depend on
2. Creating a simplified implementation in `engineext` that doesn't import problematic packages
3. Using interface-based design to allow extension without direct dependencies

### Context Builder Pattern

The `ContextManager` now allows creating contexts with specific capabilities:

```go
// Create a standard context with error handling
ctx := contextManager.CreateStandardContext(
    nodeID, nodeType, blueprintID, executionID,
    inputs, variables, logger, hooks, activateFlow,
)

// Create a context with event capabilities
ctx := contextManager.CreateEventAwareContext(
    nodeID, nodeType, blueprintID, executionID,
    inputs, variables, logger, hooks, activateFlow,
    isEventHandler, eventHandlerContext,
)
```

### Context Detection in Nodes

Nodes can detect context capabilities and use them when available:

```go
// Detect error handling capabilities
if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
    // Use error handling features
    err := errorCtx.ReportError(...)
    recovered, details := errorCtx.AttemptRecovery(err)
}

// Detect event capabilities
if eventCtx, ok := ctx.(core.EventAwareContext); ok {
    // Use event features
    eventCtx.DispatchEvent(...)
}
```

### Integration with Existing Code

For backward compatibility, we provide a migration utility in `context_migration.go` that can:

1. Detect capabilities of existing contexts
2. Upgrade old contexts to the new system
3. Extract base contexts from decorator chains

## Next Steps

1. Gradually migrate existing code to use the new context system
2. Update documentation to reflect the new design
3. Add additional context capabilities as needed
4. Remove unused or redundant context implementations
5. Add comprehensive tests for the new context system

## Benefits

The new context system provides:

1. **Clarity**: Clear separation of concerns between different context types
2. **Flexibility**: Easy to combine different context features as needed
3. **Maintainability**: Simpler to understand which contexts are being used and why
4. **Consistency**: All contexts follow the same patterns and conventions
5. **Extensibility**: Easy to add new context types without breaking existing code
6. **No Import Cycles**: Clean package structure without circular dependencies
