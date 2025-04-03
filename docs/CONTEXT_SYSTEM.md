# WebBlueprint Context System

## Overview

The WebBlueprint Context System provides a structured and flexible way to manage execution contexts within the blueprint engine. It resolves the previous "context hell" by implementing a clean decorator pattern and a builder to compose contexts with exactly the capabilities needed.

## Key Components

### Core Components

1. **DefaultExecutionContext**: The base implementation of the `node.ExecutionContext` interface.
2. **ContextBuilder**: Creates execution contexts with specific capabilities.
3. **ContextManager**: Provides a centralized way to create and manage execution contexts.

### Context Decorators

1. **ErrorAwareContext**: Adds error handling capabilities.
2. **EventAwareContext**: Adds event handling capabilities.
3. **ActorExecutionContext**: Supports actor-based parallel execution.

### Specialized Contexts

1. **FunctionExecutionContext**: Specialized for user-defined functions.
2. **LoopContext**: Specialized for loop nodes.

## Architecture

The context system is based on the decorator pattern, which allows functionality to be composed rather than inherited. Each decorator wraps a base context and adds specific capabilities:

```
+-------------------------+
| Base Context            |
| (DefaultExecutionContext)|
+-------------------------+
            ↑
            | wraps
+-------------------------+
| Error Handling          |
| (ErrorAwareContext)     |
+-------------------------+
            ↑
            | wraps
+-------------------------+
| Event Support           |
| (EventAwareContext)     |
+-------------------------+
            ↑
            | wraps
+-------------------------+
| Actor Mode              |
| (ActorExecutionContext) |
+-------------------------+
```

All context types implement the same `node.ExecutionContext` interface, ensuring that nodes can work with any context type.

## Feature Details

### Base Context (DefaultExecutionContext)

The DefaultExecutionContext provides the core functionality for all contexts:

- Getting/setting input and output values
- Getting/setting variables
- Activating output flows
- Recording debug information
- Providing node and blueprint information

### Error Handling (ErrorAwareContext)

The ErrorAwareContext adds error handling capabilities:

- Reporting errors with detailed information
- Attempting recovery from errors
- Getting error summaries
- Providing default values for missing inputs

```go
// Example of error handling
if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
    err := errorCtx.ReportError(
        bperrors.ErrorTypeExecution,
        bperrors.ErrMissingRequiredInput,
        "Required input is missing",
        nil,
    )
    
    // Try to recover
    recovered, details := errorCtx.AttemptRecovery(err)
    if recovered {
        // Use recovery details
    }
}
```

### Event Support (EventAwareContext)

The EventAwareContext adds event handling capabilities:

- Dispatching events
- Checking if handling an event
- Getting event parameters
- Getting event source information

```go
// Example of event handling
if eventCtx, ok := ctx.(core.EventAwareContext); ok {
    // Dispatch an event
    params := make(map[string]types.Value)
    params["message"] = types.NewValue(types.PinTypes.String, "Hello!")
    
    err := eventCtx.DispatchEvent("custom-event", params)
    
    // Check if handling an event
    if eventCtx.IsEventHandlerActive() {
        handlerCtx := eventCtx.GetEventHandlerContext()
        eventID := handlerCtx.EventID
        sourceID := handlerCtx.SourceID
        
        // Access event parameters
        if value, exists := handlerCtx.Parameters["param1"]; exists {
            // Use parameter value
        }
    }
}
```

### Actor Mode (ActorExecutionContext)

The ActorExecutionContext adds actor-based parallel execution:

- Thread-safe access to inputs, outputs, and variables
- Message-based communication
- Asynchronous execution flow

```go
// Example of actor mode
if actorCtx, ok := ctx.(*engineext.ActorExecutionContext); ok {
    // Send a message to this actor
    actorCtx.SendMessage("setValue", "output1", value, sender)
    
    // Stop the actor (when done)
    actorCtx.Stop()
}
```

### Function Context (FunctionExecutionContext)

The FunctionExecutionContext is specialized for user-defined functions:

- Storing internal outputs
- Managing function variables
- Tracking activated flows
- Retrieving function outputs

```go
// Example of function context
if functionCtx, ok := ctx.(*engineext.FunctionExecutionContext); ok {
    // Store internal outputs
    functionCtx.StoreInternalOutput(nodeID, pinID, value)
    
    // Get all outputs
    outputs := functionCtx.GetAllOutputs()
    
    // Get activated flows
    flows := functionCtx.GetActivatedFlows()
}
```

## Usage

### Creating a Standard Context

```go
// Create dependencies
errorManager := bperrors.NewErrorManager()
recoveryManager := bperrors.NewRecoveryManager(errorManager)
eventManager := createEventManager()

// Create context manager
contextManager := engineext.NewContextManager(
    errorManager,
    recoveryManager,
    eventManager,
)

// Create a standard context (with error handling)
ctx := contextManager.CreateStandardContext(
    nodeID,
    nodeType,
    blueprintID,
    executionID,
    inputs,
    variables,
    logger,
    hooks,
    activateFlow,
)
```

### Using the Builder Pattern

```go
// Create a context with specific capabilities
ctx := contextManager.CreateContextBuilder(
    nodeID,
    nodeType,
    blueprintID,
    executionID,
    inputs,
    variables,
    logger,
    hooks,
    activateFlow,
).
    WithErrorHandling(errorManager, recoveryManager).
    WithEventSupport(eventManager, false, nil).
    WithActorMode().
    Build()
```

### Creating a Context with Options

```go
// Specify options for the context
options := engineext.ContextOptions{
    WithErrorHandling: true,
    WithEventSupport:  true,
    WithActorMode:     false,
    IsEventHandler:    false,
}

// Create a context with the specified options
ctx := contextManager.CreateComplexContext(
    nodeID,
    nodeType,
    blueprintID,
    executionID,
    inputs,
    variables,
    logger,
    hooks,
    activateFlow,
    options,
)
```

### Detecting Context Capabilities in Nodes

Node implementations can detect and use context capabilities:

```go
func (n *MyNode) Execute(ctx node.ExecutionContext) error {
    // Get input values
    input1, exists1 := ctx.GetInputValue("input1")
    input2, exists2 := ctx.GetInputValue("input2")
    
    // Check for error handling capabilities
    if !exists1 || !exists2 {
        if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
            // Use error handling
            err := errorCtx.ReportError(
                bperrors.ErrorTypeExecution,
                bperrors.ErrMissingRequiredInput,
                "Missing required input",
                nil,
            )
            
            // Try to recover
            if recovered, _ := errorCtx.AttemptRecovery(err); !recovered {
                return err
            }
        } else {
            // Fallback error handling
            return fmt.Errorf("missing required input")
        }
    }
    
    // Process inputs
    result := processInputs(input1, input2)
    
    // Set output
    ctx.SetOutputValue("result", result)
    
    // Check for event capabilities
    if eventCtx, ok := ctx.(core.EventAwareContext); ok {
        // Dispatch an event
        params := make(map[string]types.Value)
        params["result"] = result
        eventCtx.DispatchEvent("calculation-complete", params)
    }
    
    // Continue execution
    return ctx.ActivateOutputFlow("then")
}
```

## Benefits

1. **Clarity**: Clear separation of concerns between different context types.
2. **Flexibility**: Easy to combine different context features as needed.
3. **Maintainability**: Simpler to understand which contexts are being used and why.
4. **Consistency**: All contexts follow the same patterns and conventions.
5. **Extensibility**: Easy to add new context types without breaking existing code.

## Best Practices

1. Use the `ContextManager` to create contexts rather than instantiating them directly.
2. When detecting capabilities, always provide fallbacks for when they're not available.
3. For specialized contexts (like `FunctionExecutionContext`), use the specific creation methods.
4. When adding a new feature to contexts, implement it as a decorator following the established pattern.
5. Use the `ContextBuilder` for complex context configurations.

## Examples

See the examples in:
- `/cmd/contextdemo/main.go` - Simple demo of the context system
- `/cmd/advancedcontextdemo/main.go` - Advanced demo of all context types

## Related Documentation

- [Context System Implementation](CONTEXT_SYSTEM_IMPLEMENTATION.md)
- [Context System Migration Guide](CONTEXT_SYSTEM_MIGRATION.md)
