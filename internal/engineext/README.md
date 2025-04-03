# WebBlueprint Engine Extensions

This package provides a modernized context management system for WebBlueprint, addressing the "context hell" issue with a clean, composable approach to execution contexts.

## Overview

The `engineext` package implements a flexible and maintainable approach to context management with:

1. **Clean Decorator Pattern**: Each capability (error handling, events, actor model) is implemented as a separate decorator
2. **Context Builder**: A fluent API for constructing contexts with desired capabilities
3. **Context Manager**: Factory methods for common context types
4. **Specialized Contexts**: Purpose-built contexts for functions, actors, and more

## Components

### Core Components

- **DefaultExecutionContext**: Base implementation of the `node.ExecutionContext` interface
- **ErrorAwareContext**: Adds error handling capabilities to contexts
- **EventAwareContext**: Adds event handling capabilities to contexts
- **ActorExecutionContext**: Adds actor model capabilities for parallel execution
- **FunctionExecutionContext**: Specialized context for function nodes

### Context Management

- **ContextManager**: Centralized factory for creating contexts
- **ContextBuilder**: Fluent API for building contexts with specific capabilities
- **ContextOptions**: Configuration options for context creation

## Usage

### Basic Usage

```go
// Create a context manager
contextManager := engineext.NewContextManager(
    errorManager,
    recoveryManager,
    eventManager,
)

// Create a standard context (with error handling)
ctx := contextManager.CreateStandardContext(
    nodeID, nodeType, blueprintID, executionID,
    inputs, variables, logger, hooks, activateFlow,
)
```

### Advanced Usage

```go
// Create a context with specific capabilities
ctx := contextManager.CreateComplexContext(
    nodeID, nodeType, blueprintID, executionID,
    inputs, variables, logger, hooks, activateFlow,
    engineext.ContextOptions{
        WithErrorHandling: true,
        WithEventSupport:  true,
        WithActorMode:     true,
    },
)
```

### Using Contexts in Nodes

```go
func (n *MyNode) Execute(ctx node.ExecutionContext) error {
    // Check for error handling capabilities
    if errorCtx, ok := ctx.(*engineext.ErrorAwareContext); ok {
        // Use error handling capabilities
    }
    
    // Check for event capabilities
    if eventCtx, ok := ctx.(core.EventAwareContext); ok {
        // Use event capabilities
    }
    
    // Base context functionality is always available
    value, exists := ctx.GetInputValue("input")
    ctx.SetOutputValue("output", result)
    return ctx.ActivateOutputFlow("then")
}
```

## Benefits

- **Avoids Import Cycles**: Designed to prevent circular dependencies between packages
- **Consistent Interface**: All contexts implement the same core interface
- **Capability Detection**: Nodes can detect available capabilities at runtime
- **Thread Safety**: Actor contexts provide synchronized access in parallel execution
- **Easily Extensible**: New capabilities can be added as decorators without modifying existing code

## Documentation

See the comprehensive documentation in `docs/`:

- [Context System](../../docs/CONTEXT_SYSTEM.md)
- [Context System Implementation](../../docs/CONTEXT_SYSTEM_IMPLEMENTATION.md)
- [Context System Migration Guide](../../docs/CONTEXT_SYSTEM_MIGRATION.md)

## Examples

Working examples are provided in the `cmd/` directory:

- [Basic Demo](../../cmd/contextdemo/main.go)
- [Advanced Demo](../../cmd/advancedcontextdemo/main.go)
