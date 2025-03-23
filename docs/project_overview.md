# Blueprint Engine Project Overview

## Introduction

This document provides an overview of the Blueprint Engine project, focusing on data types, execution flow, and the context in which blueprints are executed.

## Core Data Types

### Blueprint Structure

The project uses a node-based blueprint system where:

- **Blueprint**: The top-level container that holds nodes, connections, functions, and variables
- **Node**: Individual processing units with specific functionality
- **Connection**: Links between nodes that define data and execution flow
- **Pin**: Input/output points on nodes where connections attach
- **Variable**: Named values that can be accessed throughout the blueprint

Key data structures:

```go
type Blueprint struct {
    ID          string            // Unique identifier
    Name        string            // Human-readable name
    Description string            // Optional description
    Version     string            // Version information
    Nodes       []BlueprintNode   // Collection of nodes
    Functions   []Function        // Custom functions
    Connections []Connection      // Links between nodes
    Variables   []Variable        // Blueprint-level variables
    Metadata    map[string]string // Additional information
}

type BlueprintNode struct {
    ID         string                 // Unique identifier
    Type       string                 // Node type identifier
    Position   Position               // UI position
    Properties []NodeProperty         // Node configuration
    Data       map[string]interface{} // Additional data
}

type Connection struct {
    ID             string         // Unique identifier
    SourceNodeID   string         // Origin node
    SourcePinID    string         // Origin pin
    TargetNodeID   string         // Destination node
    TargetPinID    string         // Destination pin
    ConnectionType string         // "execution" or "data"
    Data           map[string]any // Additional metadata
}
```

### Pin Types and Values

The system uses strongly-typed pins for data flow:

```go
type PinType struct {
    ID          string                                       // Type identifier
    Name        string                                       // Human-readable name
    Description string                                       // Description
    Validator   func(value interface{}) error                // Validation function
    Converter   func(value interface{}) (interface{}, error) // Conversion function
}

type Value struct {
    Type     *PinType    // Type information
    RawValue interface{} // The actual data
}

type Pin struct {
    ID          string      // Unique identifier
    Name        string      // Human-readable name
    Description string      // Description
    Type        *PinType    // Data type
    Optional    bool        // Whether required
    Default     interface{} // Default value
}
```

## Execution System

### Execution Flow

Blueprint execution follows these steps:

1. An execution is initiated with a specific blueprint
2. The engine identifies entry points in the blueprint
3. For each node in the execution path:
   - Input values are gathered from connected nodes or defaults
   - The node's logic is executed
   - Output values are computed and passed to connected nodes
   - Execution flow continues along execution connections

### Execution Context

The execution context provides the environment in which nodes operate:

```go
type ExecutionContext interface {
    // Basic information
    GetNodeID() string
    GetNodeType() string
    GetBlueprintID() string
    GetExecutionID() string
    
    // Data access
    GetInputValue(pinID string) (types.Value, error)
    SetOutputValue(pinID string, value types.Value) error
    GetVariable(name string) (types.Value, error)
    SetVariable(name string, value types.Value) error
    
    // Flow control
    ActivateOutputFlow(pinID string) error
    
    // Logging
    Logger() node.Logger
    
    // Debug information
    AddDebugInfo(pinID string, description string, value interface{}) error
}
```

The context is implemented by various types:

- `DefaultExecutionContext`: Standard implementation
- `ActorExecutionContext`: For actor-based execution
- `FunctionExecutionContext`: For function execution
- `MockExecutionContext`: For testing

### Accessing Node Inputs and Outputs

During execution, nodes access inputs and produce outputs through the context:

```go
// Reading an input
inputValue, err := ctx.GetInputValue("inputPinID")
if err != nil {
    ctx.Logger().Error("Failed to get input", map[string]interface{}{"error": err})
    return err
}

// Processing logic
result := processData(inputValue.RawValue)

// Setting an output
err = ctx.SetOutputValue("outputPinID", types.Value{
    Type:     outputPinType,
    RawValue: result,
})
if err != nil {
    ctx.Logger().Error("Failed to set output", map[string]interface{}{"error": err})
    return err
}

// Activating execution flow
err = ctx.ActivateOutputFlow("nextPinID")
if err != nil {
    ctx.Logger().Error("Failed to activate flow", map[string]interface{}{"error": err})
    return err
}
```

## Error Handling

The system uses structured error types:

```go
type BlueprintError struct {
    Type            ErrorType              // Error category
    Code            BlueprintErrorCode     // Specific error code
    Message         string                 // Human-readable message
    Details         map[string]interface{} // Additional information
    Severity        ErrorSeverity          // Impact level
    Recoverable     bool                   // Can be recovered from
    RecoveryOptions []RecoveryStrategy     // Recovery methods
    NodeID          string                 // Related node
    PinID           string                 // Related pin
}
```

Errors are logged through the context's logger and can be handled based on severity and recoverability.

## Persistence Layer

The project uses a PostgreSQL database for persistence with models that map to database tables:

```go
type Execution struct {
    ID               string
    BlueprintID      string
    VersionID        sql.NullString
    StartedAt        time.Time
    CompletedAt      sql.NullTime
    Status           string
    InitiatedBy      string
    ExecutionMode    string
    InitialVariables JSONB
}

type ExecutionNode struct {
    ExecutionID string
    NodeID      string
    NodeType    string
    StartedAt   sql.NullTime
    CompletedAt sql.NullTime
    Status      string
    Inputs      JSONB
    Outputs     JSONB
    Error       sql.NullString
}

type ExecutionLog struct {
    ID          string
    ExecutionID string
    NodeID      sql.NullString
    LogLevel    string
    Message     string
    Details     JSONB
    Timestamp   time.Time
}
```

## Web Interface

The web interface provides visualization and interaction with blueprints through TypeScript interfaces that mirror the Go structures:

```typescript
export interface Blueprint {
    id: string
    name: string
    description: string
    version: string
    nodes: Node[]
    functions: Function[]
    connections: Connection[]
    variables: Variable[]
    metadata: Record<string, string>
}
```

## Execution Lifecycle

1. **Initialization**: A blueprint execution is created with initial variables
2. **Node Execution**: Each node receives its context with access to inputs
3. **Data Flow**: Data flows through connections between nodes
4. **Execution Flow**: Control flow follows execution connections
5. **Completion**: Execution ends when all paths are complete or an error occurs
6. **Persistence**: Execution results and logs are stored in the database

## Debugging and Monitoring

The system provides:

- Structured logging at different levels (Debug, Info, Warn, Error)
- Debug snapshots of node inputs/outputs
- Execution history and status tracking
- Error analysis and reporting

## Conclusion

The Blueprint Engine provides a flexible, type-safe system for defining and executing node-based workflows. The execution context serves as the bridge between nodes, providing access to inputs, outputs, variables, and logging capabilities.
