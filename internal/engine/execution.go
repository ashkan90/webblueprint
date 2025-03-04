package engine

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// DefaultExecutionContext is the implementation of node.ExecutionContext
type DefaultExecutionContext struct {
	nodeID             string
	nodeType           string
	blueprintID        string
	executionID        string
	inputs             map[string]types.Value
	outputs            map[string]types.Value
	variables          map[string]types.Value
	debugData          map[string]interface{}
	logger             node.Logger
	hooks              *node.ExecutionHooks
	activateFlow       func(ctx *DefaultExecutionContext, nodeID, pinID string) error
	activatedFlows     []string // Track which output pins were activated
	activatedFlowMutex sync.Mutex
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) *DefaultExecutionContext {
	logger.Opts(map[string]interface{}{"nodeId": nodeID})
	return &DefaultExecutionContext{
		nodeID:         nodeID,
		nodeType:       nodeType,
		blueprintID:    blueprintID,
		executionID:    executionID,
		inputs:         inputs,
		outputs:        make(map[string]types.Value),
		variables:      variables,
		debugData:      make(map[string]interface{}),
		logger:         logger,
		hooks:          hooks,
		activateFlow:   activateFlow,
		activatedFlows: make([]string, 0),
	}
}

// GetInputValue retrieves an input value by pin ID
func (ctx *DefaultExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.inputs[pinID]

	// If the value exists, return it
	if exists {
		// Log the input access
		if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
			ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
		}
		return value, true
	}

	// If the value doesn't exist, try to find a default value
	// First check the node properties for input_[pinID]
	bp, err := db.Blueprints.GetBlueprint(ctx.GetBlueprintID())
	if err != nil {
		return types.Value{}, false
	}

	_node := bp.FindNode(ctx.GetNodeID())
	if _node == nil {
		return types.Value{}, false
	}

	for _, prop := range _node.Properties {
		if prop.Name == fmt.Sprintf("input_%s", pinID) || prop.Name == "constantValue" {
			// Create a value from the default
			defaultValue := types.NewValue(types.PinTypes.Any, prop.Value)

			// Log the default value usage
			if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
				ctx.hooks.OnPinValue(ctx.nodeID, pinID, defaultValue.RawValue)
			}

			// Add to debug data
			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.nodeID,
				PinID:       pinID,
				Description: "Default value used",
				Value: map[string]interface{}{
					"default": defaultValue.RawValue,
					"source":  "node property",
				},
				Timestamp: time.Now(),
			})

			return defaultValue, true
		}

		if prop.Name == fmt.Sprintf("_loop_%s", pinID) {
			// Create a value from the default
			defaultValue := types.NewValue(types.PinTypes.Any, prop.Value)

			// Log the default value usage
			if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
				ctx.hooks.OnPinValue(ctx.nodeID, pinID, defaultValue.RawValue)
			}

			// Add to debug data
			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.nodeID,
				PinID:       pinID,
				Description: "Default value used",
				Value: map[string]interface{}{
					"default": defaultValue.RawValue,
					"source":  "node property",
				},
				Timestamp: time.Now(),
			})

			return defaultValue, true
		}
	}

	// No value or default found
	return types.Value{}, false
}

// SetOutputValue sets an output value by pin ID
func (ctx *DefaultExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.outputs[pinID] = value

	// Log the output value
	if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
		ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
	}
}

// GetOutputValue retrieves an output value by pin ID
func (ctx *DefaultExecutionContext) GetOutputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.outputs[pinID]

	// Log for debugging
	if exists {
		fmt.Printf("[DEBUG] Getting output value for %s.%s: %v (type: %T)\n",
			ctx.nodeID, pinID, value.RawValue, value.RawValue)
	} else {
		fmt.Printf("[DEBUG] Output value not found for %s.%s\n", ctx.nodeID, pinID)
	}

	return value, exists
}

// ActivateOutputFlow activates an output execution flow
// This now just stores which pins to activate, actual activation happens later
func (ctx *DefaultExecutionContext) ActivateOutputFlow(pinID string) error {
	ctx.activatedFlowMutex.Lock()
	defer ctx.activatedFlowMutex.Unlock()

	// Store the activated pin for later execution
	ctx.activatedFlows = append(ctx.activatedFlows, pinID)
	fmt.Printf("[DEBUG] Queued output flow activation for %s.%s\n", ctx.nodeID, pinID)

	return nil
}

// ExecuteConnectedNodes executes all nodes connected to the given output pin
// This is different from ActivateOutputFlow because it executes the nodes immediately
// rather than just marking them for execution
func (ctx *DefaultExecutionContext) ExecuteConnectedNodes(pinID string) error {
	logger := ctx.Logger()
	logger.Debug("Executing connected nodes", map[string]interface{}{
		"pin": pinID,
	})

	// We must directly execute connected nodes one by one
	// This requires calling the activateFlow function and waiting for it to complete
	return ctx.activateFlow(ctx, ctx.nodeID, pinID)
}

// GetActivatedOutputFlows returns the list of output pins that were activated
func (ctx *DefaultExecutionContext) GetActivatedOutputFlows() []string {
	ctx.activatedFlowMutex.Lock()
	defer ctx.activatedFlowMutex.Unlock()

	return ctx.activatedFlows
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
func (ctx *DefaultExecutionContext) Logger() node.Logger {
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
	fmt.Printf("[DEBUG] %s: %s %v\n", l.nodeID, msg, fields)
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
