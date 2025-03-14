# WebBlueprint: Developer Documentation

## Overview

WebBlueprint is a web-based visual programming platform inspired by Unreal Engine's Blueprint system. It provides a node-based interface for creating logic flows without traditional coding. The system allows users to create, connect, and execute nodes representing various operations, from basic math to complex web requests.

## System Architecture

WebBlueprint follows a modular architecture composed of several key components:

### Core Components

1. **Node System**: The foundation of WebBlueprint, defining the structure and behavior of nodes.
2. **Execution Engine**: Manages the execution of blueprints, supporting both sequential and actor-based parallel execution models.
3. **Blueprint Model**: Represents the structure of a blueprint, including nodes, connections, and metadata.
4. **Types System**: Handles data typing, conversions, and validation between nodes.
5. **WebSocket Communication**: Provides real-time updates during blueprint execution.

### Code Structure

```
webblueprint/
├── internal/
│   ├── node/             # Node interface definitions and base implementations
│   ├── nodes/            # Various node implementations by category
│   │   ├── data/         # Data manipulation nodes
│   │   ├── logic/        # Control flow nodes
│   │   ├── math/         # Mathematical operation nodes
│   │   ├── utility/      # Utility nodes
│   │   └── web/          # Web-related nodes
│   ├── types/            # Type system definitions
│   ├── engine/           # Execution engine implementation
│   ├── api/              # HTTP and WebSocket API
│   ├── registry/         # Global node registry
│   └── db/               # Database interfaces
├── pkg/
│   └── blueprint/        # Blueprint model definitions
└── cmd/
    └── server/           # Main application entry point
```

## Node System

### Node Interface

The core of WebBlueprint is the `Node` interface defined in `internal/node/interface.go`:

```go
type Node interface {
    // GetMetadata returns metadata about the node type
    GetMetadata() NodeMetadata

    // GetInputPins returns the input pins for this node
    GetInputPins() []types.Pin

    // GetOutputPins returns the output pins for this node
    GetOutputPins() []types.Pin

    // Execute runs the node's logic with the given execution context
    Execute(ctx ExecutionContext) error
}
```

All nodes must implement this interface. The `BaseNode` struct provides a common implementation that node developers can embed in their custom nodes.

### Execution Context

Nodes communicate with the execution engine through the `ExecutionContext` interface:

```go
type ExecutionContext interface {
    // Input/output access
    GetInputValue(pinID string) (types.Value, bool)
    SetOutputValue(pinID string, value types.Value)

    // Execution control
    ActivateOutputFlow(pinID string) error

    // State management
    GetVariable(name string) (types.Value, bool)
    SetVariable(name string, value types.Value)

    // Logging and debugging
    Logger() Logger
    RecordDebugInfo(info types.DebugInfo)
    GetDebugData() map[string]interface{}

    // Information retrieval
    GetNodeID() string
    GetNodeType() string
    GetBlueprintID() string
    GetExecutionID() string
}
```

This allows nodes to:
- Access input values
- Set output values
- Control execution flow
- Access and modify variables
- Log information and debug data

### Creating a New Node Type

To create a new node type:

1. Define a struct that embeds `node.BaseNode`
2. Implement the `Execute` method
3. Create a factory function that returns a new instance

Example:

```go
// MyCustomNode implements a custom operation
type MyCustomNode struct {
    node.BaseNode
}

// NewMyCustomNode creates a new custom node
func NewMyCustomNode() node.Node {
    return &MyCustomNode{
        BaseNode: node.BaseNode{
            Metadata: node.NodeMetadata{
                TypeID:      "my-custom-node",
                Name:        "My Custom Node",
                Description: "Performs a custom operation",
                Category:    "Custom",
                Version:     "1.0.0",
            },
            Inputs: []types.Pin{
                {
                    ID:          "exec",
                    Name:        "Execute",
                    Description: "Execution input",
                    Type:        types.PinTypes.Execution,
                },
                {
                    ID:          "input",
                    Name:        "Input",
                    Description: "Custom input value",
                    Type:        types.PinTypes.String,
                },
            },
            Outputs: []types.Pin{
                {
                    ID:          "then",
                    Name:        "Then",
                    Description: "Execution continues",
                    Type:        types.PinTypes.Execution,
                },
                {
                    ID:          "result",
                    Name:        "Result",
                    Description: "Operation result",
                    Type:        types.PinTypes.String,
                },
            },
        },
    }
}

// Execute runs the node logic
func (n *MyCustomNode) Execute(ctx node.ExecutionContext) error {
    logger := ctx.Logger()
    logger.Debug("Executing Custom node", nil)

    // Get input value
    inputValue, exists := ctx.GetInputValue("input")
    if !exists {
        return fmt.Errorf("missing required input: input")
    }

    // Convert to string
    input, err := inputValue.AsString()
    if err != nil {
        return err
    }

    // Perform custom operation
    result := customOperation(input)

    // Set output value
    ctx.SetOutputValue("result", types.NewValue(types.PinTypes.String, result))

    // Continue execution
    return ctx.ActivateOutputFlow("then")
}

func customOperation(input string) string {
    // Implement your custom logic here
    return strings.ToUpper(input)
}
```

## Execution Engine

The execution engine is responsible for executing blueprints. WebBlueprint supports two execution modes:

1. **Standard Mode**: Sequential execution of nodes
2. **Actor Mode**: Parallel execution using an actor model

### Standard Execution

In standard mode, nodes are executed sequentially following the execution flow defined by connections. This is suitable for simple blueprints with deterministic execution paths.

### Actor-Based Execution

The actor model provides parallel execution capabilities by treating each node as an actor that can process messages asynchronously. This mode is more suitable for complex blueprints with multiple execution paths or when performance is critical.

### Node Execution Flow

1. Node receives execution request
2. Input values are collected from connected nodes
3. Node's `Execute` method is called
4. Node processes inputs and produces outputs
5. Node activates output execution flow
6. Connected nodes are executed based on the activated flow

## Type System

The type system in WebBlueprint is responsible for handling data types, conversions, and validation.

### Built-in Types

WebBlueprint provides several built-in types:

- **Execution**: Controls execution flow (not a data type)
- **String**: Text values
- **Number**: Numeric values (float64)
- **Boolean**: True/false values
- **Object**: Key-value structures (maps)
- **Array**: Lists of values
- **Any**: Can hold any type of value

### Type Conversion

Types can be converted using the methods provided by the `Value` struct:

- `AsString()`: Convert to string
- `AsNumber()`: Convert to number
- `AsBoolean()`: Convert to boolean
- `AsObject()`: Convert to object (map)
- `AsArray()`: Convert to array (slice)

Example:

```go
// Convert a value to string
strValue, err := inputValue.AsString()
if err != nil {
    return err
}

// Convert a value to number
numValue, err := inputValue.AsNumber()
if err != nil {
    return err
}
```

## Blueprint Model

Blueprints are represented by the `Blueprint` struct defined in `pkg/blueprint/blueprint.go`. A blueprint consists of:

- **Nodes**: Individual components with inputs and outputs
- **Connections**: Links between node pins that define data and execution flow
- **Variables**: Named values that can be accessed by nodes
- **Functions**: Reusable components that can be called from other blueprints
- **Metadata**: Additional information about the blueprint

### Blueprint Node

Each node in a blueprint is represented by the `BlueprintNode` struct:

```go
type BlueprintNode struct {
    ID         string                 // Unique identifier
    Type       string                 // Node type ID
    Position   Position               // Position on the canvas
    Properties []NodeProperty         // Node properties
    Data       map[string]interface{} // Additional data
}
```

### Connection

Connections between nodes are represented by the `Connection` struct:

```go
type Connection struct {
    ID             string         // Unique identifier
    SourceNodeID   string         // ID of the source node
    SourcePinID    string         // ID of the source pin
    TargetNodeID   string         // ID of the target node
    TargetPinID    string         // ID of the target pin
    ConnectionType string         // "execution" or "data"
    Data           map[string]any // Additional data
}
```

## Registering New Node Types

To make a new node type available in the system, it must be registered with the execution engine and global registry:

```go
// In main.go or similar initialization code
import (
    "webblueprint/internal/api"
    "webblueprint/internal/nodes/custom"
    "webblueprint/internal/registry"
)

func main() {
    // ...
    
    // Get global node registry
    globalRegistry := registry.GetInstance()
    
    // Register custom node type
    registerNodeType(apiServer, globalRegistry, "my-custom-node", custom.NewMyCustomNode)
    
    // ...
}

// Helper function to register a node type
func registerNodeType(apiServer *api.APIServer, globalRegistry *registry.GlobalNodeRegistry, typeID string, factory func() node.Node) {
    apiServer.RegisterNodeType(typeID, factory)
}
```

## Debugging Tools

WebBlueprint provides built-in debugging capabilities through the `DebugManager` and execution context's `RecordDebugInfo` method.

### Recording Debug Information

Nodes can record debug information during execution:

```go
// Record debug info
ctx.RecordDebugInfo(types.DebugInfo{
    NodeID:      ctx.GetNodeID(),
    Description: "Custom Operation",
    Value: map[string]interface{}{
        "input":  input,
        "result": result,
        "timing": operationTime,
    },
    Timestamp: time.Now(),
})
```

### Retrieving Debug Data

Debug data can be accessed through the API:

```
GET /api/executions/{id}/nodes/{nodeId}
```

## Event System

The execution engine emits events during blueprint execution, which can be captured by execution listeners. This is particularly useful for updating the UI in real-time.

### Execution Events

- `EventNodeStarted`: Node execution started
- `EventNodeCompleted`: Node execution completed
- `EventNodeError`: Node execution error
- `EventValueProduced`: Value produced on a node output
- `EventValueConsumed`: Value consumed by a node input
- `EventExecutionStart`: Blueprint execution started
- `EventExecutionEnd`: Blueprint execution ended
- `EventDebugData`: Debug data available

### Implementing an Execution Listener

To listen for execution events, implement the `ExecutionListener` interface:

```go
type CustomExecutionListener struct {
    // Custom fields
}

func (l *CustomExecutionListener) OnExecutionEvent(event engine.ExecutionEvent) {
    switch event.Type {
    case engine.EventNodeStarted:
        // Handle node start
    case engine.EventNodeCompleted:
        // Handle node completion
    case engine.EventNodeError:
        // Handle node error
    // Handle other event types
    }
}

// Add the listener to the execution engine
executionEngine.AddExecutionListener(customListener)
```

## WebSocket Communication

WebBlueprint uses WebSockets to provide real-time updates during blueprint execution. The WebSocket server is implemented in `internal/api/websocket.go`.

### Message Types

- `node.intro`: Node type introduction
- `node.start`: Node execution started
- `node.complete`: Node execution completed
- `node.error`: Node execution error
- `data.flow`: Data flowing between nodes
- `debug.data`: Debug data available
- `execution.start`: Blueprint execution started
- `execution.end`: Blueprint execution ended
- `execution.status`: Execution status update
- `result`: Pin output value
- `log`: Log message

## Database Strategy

Based on the architecture decision record, WebBlueprint uses PostgreSQL with JSONB as the primary database technology.

### Schema Design

The database schema includes:

- **Workspaces**: Top-level containers
- **Assets**: Named resources within workspaces
- **Blueprints**: Visual programming graphs
- **Blueprint Versions**: Versioned blueprint states
- **Asset References**: Relationships between assets

### Key Features

- JSONB storage for complete blueprint structures
- GIN indexing for efficient searches within JSONB documents
- Transaction support to ensure data integrity
- Relational model for critical relationships

## Function Nodes

WebBlueprint supports user-defined functions through the function node system. Functions allow reusing logic across different blueprints.

### Function Structure

A function is represented by the `Function` struct:

```go
type Function struct {
    ID          string            // Unique identifier
    Name        string            // Human-readable name
    Description string            // Optional description
    NodeType    BlueprintNodeType // Function interface
    Nodes       []BlueprintNode   // Internal nodes
    Connections []Connection      // Internal connections
    Variables   []Variable        // Function variables
    Metadata    map[string]string // Additional metadata
}
```

### Creating a Function

Functions are created by defining a set of nodes and connections, along with input and output interfaces. The system automatically creates a node type for each function.

## User Variables

WebBlueprint supports user-defined variables that can be accessed across a blueprint.

### Variable Types

Variables can be of any supported type:

- String
- Number
- Boolean
- Object
- Array

### Accessing Variables

Variables can be accessed using special variable getter and setter nodes:

- `get-variable-{name}`: Gets the value of a variable
- `set-variable-{name}`: Sets the value of a variable

These nodes are automatically generated when a variable is defined.

## Future Extensions

The WebBlueprint architecture is designed to be extensible. Some areas for future development include:

1. **Enhanced Event System**: Implementing a comprehensive event system inspired by Unreal Engine
2. **Custom Type System**: Supporting user-defined types
3. **Advanced Debugging Tools**: Step-by-step debugging, breakpoints, etc.
4. **Distributed Execution**: Scaling blueprint execution across multiple servers
5. **Plugin System**: Allowing third-party developers to extend WebBlueprint

## Conclusion

This documentation provides an overview of the WebBlueprint system architecture and development guidelines. For more detailed information, refer to the source code and specific component documentation.