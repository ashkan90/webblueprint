package engine

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// FunctionExecutionContext is a specialized execution context for user-defined functions
type FunctionExecutionContext struct {
	ParentContext   node.ExecutionContext
	nodeID          string
	nodeType        string
	functionID      string
	executionID     string
	inputs          map[string]types.Value
	outputs         map[string]types.Value
	variables       map[string]types.Value
	debugData       map[string]interface{}
	logger          node.Logger
	activatedFlows  []string
	internalOutputs map[string]map[string]types.Value // nodeID -> pinID -> value
	mutex           sync.RWMutex
}

// NewFunctionExecutionContext creates a new function execution context
func NewFunctionExecutionContext(
	parentCtx node.ExecutionContext,
	nodeID string,
	nodeType string,
	functionID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
) *FunctionExecutionContext {
	return &FunctionExecutionContext{
		ParentContext:   parentCtx,
		nodeID:          nodeID,
		nodeType:        nodeType,
		functionID:      functionID,
		executionID:     executionID,
		inputs:          inputs,
		outputs:         make(map[string]types.Value),
		variables:       variables,
		debugData:       make(map[string]interface{}),
		logger:          logger,
		activatedFlows:  make([]string, 0),
		internalOutputs: make(map[string]map[string]types.Value),
	}
}

// Implementation of the node.ExecutionContext interface

// GetInputValue retrieves an input value by pin ID
func (ctx *FunctionExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	// First check inputs directly passed to the function
	if value, exists := ctx.inputs[pinID]; exists {
		return value, true
	}

	// Then check internal outputs (from other nodes in the function)
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	for _, nodeOutputs := range ctx.internalOutputs {
		if value, exists := nodeOutputs[pinID]; exists {
			return value, true
		}
	}

	// If not found, check parent context's variables
	// (useful for accessing variables defined in the parent blueprint)
	if value, exists := ctx.ParentContext.GetVariable(pinID); exists {
		return value, true
	}

	return types.Value{}, false
}

// SetOutputValue sets an output value by pin ID
func (ctx *FunctionExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.outputs[pinID] = value

	// If this is an output of the function, also store it in the parent context
	ctx.ParentContext.SetOutputValue(pinID, value)
}

// ActivateOutputFlow activates an output execution flow
func (ctx *FunctionExecutionContext) ActivateOutputFlow(pinID string) error {
	ctx.mutex.Lock()
	ctx.activatedFlows = append(ctx.activatedFlows, pinID)
	ctx.mutex.Unlock()

	// Forward to parent context if this is a function output pin
	return ctx.ParentContext.ActivateOutputFlow(pinID)
}

// ExecuteConnectedNodes executes nodes connected to the given output pin
func (ctx *FunctionExecutionContext) ExecuteConnectedNodes(pinID string) error {
	// This would be implemented to handle synchronous execution
	// For now, just record the flow activation
	return ctx.ActivateOutputFlow(pinID)
}

// GetVariable retrieves a variable by name
func (ctx *FunctionExecutionContext) GetVariable(name string) (types.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	if value, exists := ctx.variables[name]; exists {
		return value, true
	}

	// Try parent context if not found locally
	return ctx.ParentContext.GetVariable(name)
}

// SetVariable sets a variable by name
func (ctx *FunctionExecutionContext) SetVariable(name string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.variables[name] = value

	// Also set in parent context to propagate outside the function
	ctx.ParentContext.SetVariable(name, value)
}

// Logger returns the execution logger
func (ctx *FunctionExecutionContext) Logger() node.Logger {
	return ctx.logger
}

// RecordDebugInfo stores debug information
func (ctx *FunctionExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// Add the debug info to our collection
	key := fmt.Sprintf("debug_%d", time.Now().UnixNano())

	ctx.mutex.Lock()
	ctx.debugData[key] = info
	ctx.mutex.Unlock()

	// Also record in parent context
	ctx.ParentContext.RecordDebugInfo(info)
}

// GetDebugData returns all debug data
func (ctx *FunctionExecutionContext) GetDebugData() map[string]interface{} {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Create a copy to avoid concurrency issues
	dataCopy := make(map[string]interface{})
	for k, v := range ctx.debugData {
		dataCopy[k] = v
	}

	return dataCopy
}

// GetNodeID returns the ID of the executing node
func (ctx *FunctionExecutionContext) GetNodeID() string {
	return ctx.nodeID
}

// GetNodeType returns the type of the executing node
func (ctx *FunctionExecutionContext) GetNodeType() string {
	return ctx.nodeType
}

// GetBlueprintID returns the ID of the executing blueprint
func (ctx *FunctionExecutionContext) GetBlueprintID() string {
	return ctx.functionID
}

// GetExecutionID returns the current execution ID
func (ctx *FunctionExecutionContext) GetExecutionID() string {
	return ctx.executionID
}

// Additional methods for function context

// GetActivatedFlows returns the activated output flows
func (ctx *FunctionExecutionContext) GetActivatedFlows() []string {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	flows := make([]string, len(ctx.activatedFlows))
	copy(flows, ctx.activatedFlows)
	return flows
}

// StoreInternalOutput stores an output from a node inside the function
func (ctx *FunctionExecutionContext) StoreInternalOutput(nodeID, pinID string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if _, exists := ctx.internalOutputs[nodeID]; !exists {
		ctx.internalOutputs[nodeID] = make(map[string]types.Value)
	}

	ctx.internalOutputs[nodeID][pinID] = value
}

// GetInternalOutput retrieves an output from a node inside the function
func (ctx *FunctionExecutionContext) GetInternalOutput(nodeID, pinID string) (types.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	if nodeOutputs, exists := ctx.internalOutputs[nodeID]; exists {
		if value, exists := nodeOutputs[pinID]; exists {
			return value, true
		}
	}

	return types.Value{}, false
}

// GetAllOutputs returns all outputs from the function
func (ctx *FunctionExecutionContext) GetAllOutputs() map[string]types.Value {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	outputs := make(map[string]types.Value)
	for k, v := range ctx.outputs {
		outputs[k] = v
	}

	return outputs
}

// GetAllVariables returns all variables from the function
func (ctx *FunctionExecutionContext) GetAllVariables() map[string]types.Value {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	variables := make(map[string]types.Value)
	for k, v := range ctx.variables {
		variables[k] = v
	}

	return variables
}
