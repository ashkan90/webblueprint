package node

import (
	"webblueprint/internal/types"
)

// NodeMetadata contains information about a node type
type NodeMetadata struct {
	TypeID      string           // Unique identifier for the node type
	Name        string           // Human-readable name
	Description string           // Description of what the node does
	Category    string           // Category for grouping nodes in the UI
	Version     string           // Version of the node implementation
	Properties  []types.Property // Node properties
	InputPins   []types.Pin      // Input pins for the node
	OutputPins  []types.Pin      // Output pins for the node
}

// Node is the interface that all node types must implement
type Node interface {
	// GetMetadata returns metadata about the node type
	GetMetadata() NodeMetadata

	// GetInputPins returns the input pins for this node
	GetInputPins() []types.Pin

	// SetInputPins sets the input pins for this node
	SetInputPins(pins []types.Pin)

	// GetOutputPins returns the output pins for this node
	GetOutputPins() []types.Pin

	// SetOutputPins sets the output pins for this node
	SetOutputPins(pins []types.Pin)

	// GetProperties returns the node properties
	GetProperties() []types.Property

	// Execute runs the node's logic with the given execution context
	Execute(ctx ExecutionContext) error
}

// Logger interface for node execution logging
type Logger interface {
	Opts(map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
}

// ExecutionContext provides services to nodes during execution
type ExecutionContext interface {
	// Input/output access
	GetInputValue(pinID string) (types.Value, bool)
	SetOutputValue(pinID string, value types.Value)
	IsInputPinActive(pinID string) bool

	// Execution control
	ActivateOutputFlow(pinID string) error

	// Direct execution (for nodes that need synchronous execution)
	ExecuteConnectedNodes(pinID string) error

	// State management
	GetVariable(name string) (types.Value, bool)
	SetVariable(name string, value types.Value)

	// Logging and debugging
	Logger() Logger

	// Debugging
	RecordDebugInfo(info types.DebugInfo)
	GetDebugData() map[string]interface{}

	// Node information
	GetNodeID() string
	GetNodeType() string

	// Blueprint information
	GetBlueprintID() string
	GetExecutionID() string
}

// ExtendedExecutionContext adds additional methods for engine implementation
// This is needed for internal engine access to context data
type ExtendedExecutionContext interface {
	ExecutionContext

	// Additional methods for engine implementation
	GetOutputValue(pinID string) (types.Value, bool)
	GetAllOutputs() map[string]types.Value
	GetActivatedOutputFlows() []string
}

// ExecutionHooks provides callbacks for execution events
type ExecutionHooks struct {
	OnNodeStart    func(nodeID, nodeType string)
	OnNodeComplete func(nodeID, nodeType string)
	OnNodeError    func(nodeID string, err error)
	OnPinValue     func(nodeID, pinName string, value interface{})
	OnLog          func(nodeID, message string)
}

// ActivationAwareContext is an interface for checking input pin activation
type ActivationAwareContext interface {
	IsInputPinActive(pinID string) bool
}

// NodeFactory is a function that creates a new instance of a node
type NodeFactory func() Node
