package engineext

import (
	"sync"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ActorExecutionContext is a specialized execution context for actor-based parallel execution
type ActorExecutionContext struct {
	node.ExecutionContext
	baseContext    *DefaultExecutionContext
	activePins     map[string]bool
	outputsMutex   sync.RWMutex
	variablesMutex sync.RWMutex
	debugMutex     sync.RWMutex
}

// NewActorExecutionContext creates a new actor execution context
func NewActorExecutionContext(baseCtx *DefaultExecutionContext) *ActorExecutionContext {
	return &ActorExecutionContext{
		ExecutionContext: baseCtx,
		baseContext:      baseCtx,
		activePins:       make(map[string]bool),
	}
}

// IsInputPinActive checks if the input pin triggered execution
func (ctx *ActorExecutionContext) IsInputPinActive(pinID string) bool {
	// Check if the pin is marked active in our map
	if active, exists := ctx.activePins[pinID]; exists && active {
		return true
	}

	// If we're checking the default "execute" pin and no pins are active,
	// then treat it as active
	if pinID == "execute" && len(ctx.activePins) == 0 {
		return true
	}

	// Otherwise delegate to base context
	return ctx.ExecutionContext.IsInputPinActive(pinID)
}

// SetInputPinActive marks a pin as active
func (ctx *ActorExecutionContext) SetInputPinActive(pinID string) {
	ctx.activePins[pinID] = true
}

// GetInputValue thread-safely gets an input value
func (ctx *ActorExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	return ctx.ExecutionContext.GetInputValue(pinID)
}

// SetOutputValue thread-safely sets an output value
func (ctx *ActorExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.outputsMutex.Lock()
	defer ctx.outputsMutex.Unlock()

	ctx.ExecutionContext.SetOutputValue(pinID, value)
}

// ActivateOutputFlow thread-safely activates an output flow
func (ctx *ActorExecutionContext) ActivateOutputFlow(pinID string) error {
	return ctx.ExecutionContext.ActivateOutputFlow(pinID)
}

// GetVariable thread-safely retrieves a variable
func (ctx *ActorExecutionContext) GetVariable(name string) (types.Value, bool) {
	ctx.variablesMutex.RLock()
	defer ctx.variablesMutex.RUnlock()

	return ctx.ExecutionContext.GetVariable(name)
}

// SetVariable thread-safely sets a variable
func (ctx *ActorExecutionContext) SetVariable(name string, value types.Value) {
	ctx.variablesMutex.Lock()
	defer ctx.variablesMutex.Unlock()

	ctx.ExecutionContext.SetVariable(name, value)
}

// RecordDebugInfo thread-safely records debug information
func (ctx *ActorExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	ctx.debugMutex.Lock()
	defer ctx.debugMutex.Unlock()

	ctx.ExecutionContext.RecordDebugInfo(info)
}

// GetDebugData thread-safely retrieves debug data
func (ctx *ActorExecutionContext) GetDebugData() map[string]interface{} {
	ctx.debugMutex.RLock()
	defer ctx.debugMutex.RUnlock()

	return ctx.ExecutionContext.GetDebugData()
}
