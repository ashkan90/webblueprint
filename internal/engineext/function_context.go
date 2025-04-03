package engineext

import (
	"sync"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// FunctionExecutionContext is a specialized context for function nodes
type FunctionExecutionContext struct {
	node.ExecutionContext
	functionID     string
	internalValues map[string]map[string]types.Value // NodeID -> PinID -> Value
	activatedFlows map[string]bool                   // Record which flows were activated
	mutex          sync.RWMutex
}

// NewFunctionExecutionContext creates a new function execution context
func NewFunctionExecutionContext(baseCtx node.ExecutionContext, functionID string) *FunctionExecutionContext {
	return &FunctionExecutionContext{
		ExecutionContext: baseCtx,
		functionID:       functionID,
		internalValues:   make(map[string]map[string]types.Value),
		activatedFlows:   make(map[string]bool),
	}
}

// GetFunctionID returns the ID of the function being executed
func (ctx *FunctionExecutionContext) GetFunctionID() string {
	return ctx.functionID
}

// StoreInternalOutput stores an output value for a node within the function
func (ctx *FunctionExecutionContext) StoreInternalOutput(nodeID, pinID string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// Create the node map if it doesn't exist
	if _, exists := ctx.internalValues[nodeID]; !exists {
		ctx.internalValues[nodeID] = make(map[string]types.Value)
	}

	// Store the value
	ctx.internalValues[nodeID][pinID] = value
}

// GetInternalOutput gets an output value for a node within the function
func (ctx *FunctionExecutionContext) GetInternalOutput(nodeID, pinID string) (types.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Check if the node exists
	if nodeMap, exists := ctx.internalValues[nodeID]; exists {
		// Check if the pin exists
		if value, pinExists := nodeMap[pinID]; pinExists {
			return value, true
		}
	}

	return types.Value{}, false
}

// RecordActivatedFlow records that a flow was activated
func (ctx *FunctionExecutionContext) RecordActivatedFlow(pinID string) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.activatedFlows[pinID] = true
}

// WasFlowActivated checks if a flow was activated
func (ctx *FunctionExecutionContext) WasFlowActivated(pinID string) bool {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	return ctx.activatedFlows[pinID]
}

// GetAllOutputs returns all stored outputs
func (ctx *FunctionExecutionContext) GetAllOutputs() map[string]map[string]types.Value {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Create a copy to avoid external modification
	result := make(map[string]map[string]types.Value)
	for nodeID, nodeMap := range ctx.internalValues {
		result[nodeID] = make(map[string]types.Value)
		for pinID, value := range nodeMap {
			result[nodeID][pinID] = value
		}
	}

	return result
}

// GetActivatedFlows returns all activated flows
func (ctx *FunctionExecutionContext) GetActivatedFlows() []string {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Create a slice of activated flow IDs
	result := make([]string, 0, len(ctx.activatedFlows))
	for pinID := range ctx.activatedFlows {
		result = append(result, pinID)
	}

	return result
}
