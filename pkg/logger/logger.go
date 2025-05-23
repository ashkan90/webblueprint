package nodes

import (
	"fmt"
	"time"
	"webblueprint/internal/types"
)

// NodeMetadata contains information about a node type
type NodeMetadata struct {
	TypeID      string // Unique identifier for the node type
	Name        string // Human-readable name
	Description string // Description of what the node does
	Category    string // Category for grouping nodes in the UI
	Version     string // Version of the node implementation
}

// Node is the interface that all node types must implement
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

// Logger interface for node execution logging
type Logger interface {
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

	// Execution control
	ActivateOutputFlow(pinID string) error

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

// BaseNode provides common implementation for all nodes
type BaseNode struct {
	Metadata NodeMetadata
	Inputs   []types.Pin
	Outputs  []types.Pin
}

// GetMetadata returns the node's metadata
func (n *BaseNode) GetMetadata() NodeMetadata {
	return n.Metadata
}

// GetInputPins returns the node's input pins
func (n *BaseNode) GetInputPins() []types.Pin {
	return n.Inputs
}

// GetOutputPins returns the node's output pins
func (n *BaseNode) GetOutputPins() []types.Pin {
	return n.Outputs
}

// NodeFactory is a function that creates a new instance of a node
type NodeFactory func() Node

// ExecutionHooks provides callbacks for execution events
type ExecutionHooks struct {
	OnNodeStart    func(nodeID, nodeType string)
	OnNodeComplete func(nodeID, nodeType string)
	OnNodeError    func(nodeID string, err error)
	OnPinValue     func(nodeID, pinName string, value interface{})
	OnLog          func(nodeID, message string)
}

// DefaultExecutionContext is the standard implementation of ExecutionContext
type DefaultExecutionContext struct {
	nodeID       string
	nodeType     string
	blueprintID  string
	executionID  string
	inputs       map[string]types.Value
	outputs      map[string]types.Value
	variables    map[string]types.Value
	debugData    map[string]interface{}
	logger       Logger
	hooks        *ExecutionHooks
	activateFlow func(nodeID, pinID string) error
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger Logger,
	hooks *ExecutionHooks,
	activateFlow func(nodeID, pinID string) error,
) *DefaultExecutionContext {
	return &DefaultExecutionContext{
		nodeID:       nodeID,
		nodeType:     nodeType,
		blueprintID:  blueprintID,
		executionID:  executionID,
		inputs:       inputs,
		outputs:      make(map[string]types.Value),
		variables:    variables,
		debugData:    make(map[string]interface{}),
		logger:       logger,
		hooks:        hooks,
		activateFlow: activateFlow,
	}
}

// GetInputValue retrieves an input value by pin ID
func (ctx *DefaultExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.inputs[pinID]

	// Log the input access
	if exists && ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
		ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
	}

	return value, exists
}

// SetOutputValue sets an output value by pin ID
func (ctx *DefaultExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.outputs[pinID] = value

	// Log the output value
	if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
		ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
	}
}

// ActivateOutputFlow activates an output execution flow
func (ctx *DefaultExecutionContext) ActivateOutputFlow(pinID string) error {
	if ctx.activateFlow == nil {
		return fmt.Errorf("no activation function provided")
	}
	return ctx.activateFlow(ctx.nodeID, pinID)
}

// GetVariable retrieves a variable by name
func (ctx *DefaultExecutionContext) GetVariable(name string) (types.Value, bool) {
	value, exists := ctx.variables[name]
	return value, exists
}

// SetVariable sets a variable by name
func (ctx *DefaultExecutionContext) SetVariable(name string, value types.Value) {
	ctx.variables[name] = value
}

// Logger returns the execution logger
func (ctx *DefaultExecutionContext) Logger() Logger {
	return ctx.logger
}

// RecordDebugInfo stores debug information
func (ctx *DefaultExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// Add the debug info to our collection
	key := fmt.Sprintf("debug_%d", time.Now().UnixNano())
	ctx.debugData[key] = info
}

// GetDebugData returns all debug data
func (ctx *DefaultExecutionContext) GetDebugData() map[string]interface{} {
	return ctx.debugData
}

// GetNodeID returns the ID of the executing node
func (ctx *DefaultExecutionContext) GetNodeID() string {
	return ctx.nodeID
}

// GetNodeType returns the type of the executing node
func (ctx *DefaultExecutionContext) GetNodeType() string {
	return ctx.nodeType
}

// GetBlueprintID returns the ID of the executing blueprint
func (ctx *DefaultExecutionContext) GetBlueprintID() string {
	return ctx.blueprintID
}

// GetExecutionID returns the current execution ID
func (ctx *DefaultExecutionContext) GetExecutionID() string {
	return ctx.executionID
}

// DefaultLogger is a simple logger implementation
type DefaultLogger struct {
	nodeID string
}

// NewDefaultLogger creates a new logger for a node
func NewDefaultLogger(nodeID string) *DefaultLogger {
	return &DefaultLogger{
		nodeID: nodeID,
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields map[string]interface{}) {
	// Add node ID to fields
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["nodeID"] = l.nodeID
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, fields map[string]interface{}) {
	// Add node ID to fields
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["nodeID"] = l.nodeID
	fmt.Printf("[INFO] %s: %s %v\n", l.nodeID, msg, fields)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, fields map[string]interface{}) {
	// Add node ID to fields
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["nodeID"] = l.nodeID
	fmt.Printf("[WARN] %s: %s %v\n", l.nodeID, msg, fields)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, fields map[string]interface{}) {
	// Add node ID to fields
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["nodeID"] = l.nodeID
	fmt.Printf("[ERROR] %s: %s %v\n", l.nodeID, msg, fields)
}
