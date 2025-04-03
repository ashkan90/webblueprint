# WebBlueprint Context System Migration Guide

This guide helps developers migrate from the old context system to the new context management system. The new approach resolves the "context hell" issue by providing a clean, composable approach to context capabilities.

## Migration Overview

1. **Identify Existing Usage**: Determine where and how contexts are currently created and used
2. **Replace Context Creation**: Use the new `ContextManager` to create contexts
3. **Update Node Implementations**: Update nodes to detect capabilities instead of expecting specific context types
4. **Test Thoroughly**: Ensure all functionality works as expected

## Step 1: Add the Context Manager

First, add the Context Manager to your execution engine:

```go
// Create dependencies
errorManager := bperrors.NewErrorManager()
recoveryManager := bperrors.NewRecoveryManager(errorManager)
eventManager := event.NewEventManager()

// Create the context manager
contextManager := engineext.NewContextManager(
    errorManager,
    recoveryManager,
    eventManager,
)
```

## Step 2: Replace Context Creation

Replace old context creation code with the new builder pattern:

### Old Approach:

```go
// Multiple wrappers with error-prone ordering
baseCtx := engine.NewExecutionContext(nodeID, nodeType, blueprintID, executionID, inputs, variables, logger, hooks, activateFlow)
errorCtx := engine.NewErrorAwareExecutionContext(baseCtx, errorManager, recoveryManager)
eventCtx := event.NewDefaultExecutionContextWithEvents(errorCtx, eventManager, false, nil)
```

### New Approach:

```go
// Clean, composable builder pattern
ctx := contextManager.CreateContextBuilder(nodeID, nodeType, blueprintID, executionID, inputs, variables, logger, hooks, activateFlow)
    .WithErrorHandling(errorManager, recoveryManager)
    .WithEventSupport(eventManager, false, nil)
    .WithActorMode()
    .Build()
```

Or use the specialized factory methods:

```go
// For standard contexts
ctx := contextManager.CreateStandardContext(nodeID, nodeType, blueprintID, executionID, inputs, variables, logger, hooks, activateFlow)

// For actor contexts
ctx := contextManager.CreateActorContext(nodeID, nodeType, blueprintID, executionID, inputs, variables, logger, hooks, activateFlow)

// For event-aware contexts
ctx := contextManager.CreateEventAwareContext(nodeID, nodeType, blueprintID, executionID, inputs, variables, logger, hooks, activateFlow, false, nil)

// For function contexts
ctx := contextManager.CreateFunctionContext(nodeID, nodeType, blueprintID, executionID, functionID, inputs, variables, logger)
```

## Step 3: Update Node Implementations

Update node `Execute` methods to detect capabilities instead of assuming a context type:

### Old Approach:

```go
func (n *MyNode) Execute(ctx node.ExecutionContext) error {
    // Assume context has error handling
    errorCtx := ctx.(*engine.ErrorAwareExecutionContext)
    err := errorCtx.ReportError(...)
    
    // Assume context has event capabilities
    eventCtx := ctx.(*event.DefaultExecutionContextWithEvents)
    eventCtx.DispatchEvent(...)
    
    return nil
}
```

### New Approach:

```go
func (n *MyNode) Execute(ctx node.ExecutionContext) error {
    // Detect error handling capabilities
    if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
        // Use error handling
        err := errorCtx.ReportError(...)
    }
    
    // Detect event capabilities
    if eventCtx, ok := ctx.(core.EventAwareContext); ok {
        // Use event handling
        eventCtx.DispatchEvent(...)
    }
    
    // Fallback behavior for other context types
    
    return nil
}
```

## Step 4: Transition Strategy for Existing Code

For gradual migration, use the helper utilities:

```go
// Detect capabilities of an existing context
capabilities := engineext.DetectContextCapabilities(oldContext)

// Upgrade an old context to use the new system
newContext := engineext.UpgradeToNewContextSystem(oldContext, contextManager)

// Extract the base context from a decorator chain
baseContext := engineext.ExtractBaseContext(ctx)
```

## Common Migration Patterns

### Error Handling

```go
// Before
errorCtx := ctx.(*engine.ErrorAwareExecutionContext)
err := errorCtx.ReportError(...)

// After
if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
    err := errorCtx.ReportError(...)
} else {
    // Fallback error handling
    return fmt.Errorf("missing required input")
}
```

### Event Handling

```go
// Before
eventCtx := ctx.(*event.DefaultExecutionContextWithEvents)
eventCtx.DispatchEvent(...)

// After
if eventCtx, ok := ctx.(core.EventAwareContext); ok {
    eventCtx.DispatchEvent(...)
} else {
    // Fallback behavior when events not supported
    log.Printf("Events not supported, skipping dispatch of %s", eventID)
}
```

### Function Context

```go
// Before
functionCtx := ctx.(*engine.FunctionExecutionContext)
functionCtx.StoreInternalOutput(...)

// After
if functionCtx, ok := ctx.(*engineext.FunctionExecutionContext); ok {
    functionCtx.StoreInternalOutput(...)
} else {
    // Fallback behavior when not a function context
    ctx.SetVariable("internal_" + nodeID + "_" + pinID, value)
}
```

## Testing Migration

1. Create a test blueprint that exercises error handling, events, actor mode, and function execution
2. Run the blueprint in both old and new context systems
3. Compare execution results to ensure they match
4. Check for any errors or unexpected behavior
5. Verify performance and resource usage

## Troubleshooting

- **Missing Capabilities**: If a context lacks expected capabilities, check the context creation code
- **Context Casting Errors**: Ensure you're using `ok` checks when type asserting
- **Import Cycles**: Use the core interfaces (in `internal/core`) for cross-package references
- **Performance Issues**: Ensure contexts are created efficiently and not recreated unnecessarily

## Conclusion

The new context system provides a cleaner, more maintainable approach to handling execution contexts in WebBlueprint. By following this migration guide, you can safely transition from the old system to the new one while preserving all functionality.
