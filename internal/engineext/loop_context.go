package engineext

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// LoopContext is a special execution context that tracks loop state
type LoopContext struct {
	node.ExecutionContext
	// loopNode       *logic.LoopNode // Removed to break import cycle
	loopVarName    string
	currentIndex   float64
	maxIterations  int
	startIndex     float64
	nodeID         string
	bodyCompleted  chan bool
	executionDone  chan bool
	debugData      map[string]interface{}
	bodyActivated  bool
	iterationsDone int
	startTime      time.Time
	outputs        map[string]types.Value
}

func NewLoopContext(
	ctx *DefaultExecutionContext,
	// loopNode *logic.LoopNode, // Removed parameter
	loopVarName string,
	maxIterations int,
	startIndex float64,
	nodeID string,
	debugData map[string]interface{},
	outputs map[string]types.Value,
) *LoopContext {
	return &LoopContext{
		ExecutionContext: ctx,
		// loopNode:         loopNode, // Removed assignment
		loopVarName:    loopVarName,
		currentIndex:   0,
		maxIterations:  maxIterations,
		startIndex:     startIndex,
		nodeID:         nodeID,
		bodyCompleted:  make(chan bool),
		executionDone:  make(chan bool),
		debugData:      debugData,
		bodyActivated:  false,
		iterationsDone: 0,
		startTime:      time.Now(),
		outputs:        outputs,
	}
}

// ActivateOutputFlow overrides the standard method to track loop body completion
func (ctx *LoopContext) ActivateOutputFlow(pinID string) error {
	if pinID == "loop" {
		// Mark that we've activated the loop body flow
		ctx.bodyActivated = true
		// Let the original flow happen
		ctx.Logger().Info("Activate loop body flow", map[string]interface{}{
			"pinID": pinID,
		})
		return ctx.ExecutionContext.ActivateOutputFlow(pinID)
	} else if pinID == "completed" {
		// This is called when we're done with all iterations
		ctx.executionDone <- true
		return ctx.ExecutionContext.ActivateOutputFlow(pinID)
	}
	// Removed unreliable logic that assumed other pin activations meant body completion.
	// Loop body completion must now be signaled explicitly via SignalIterationComplete().

	// Otherwise, just let the standard activation happen
	return ctx.ExecutionContext.ActivateOutputFlow(pinID)
}

// GetDebugData retrieves debug data from the loop context
func (ctx *LoopContext) GetDebugData() map[string]interface{} {
	// Combine our debug data with the parent context's debug data
	result := make(map[string]interface{})

	// Get base debug data
	baseData := ctx.ExecutionContext.GetDebugData()
	for k, v := range baseData {
		result[k] = v
	}

	// Add loop-specific debug data
	result["loopState"] = map[string]interface{}{
		"currentIndex":   ctx.currentIndex,
		"maxIterations":  ctx.maxIterations,
		"iterationsDone": ctx.iterationsDone,
		"duration":       time.Since(ctx.startTime).String(),
	}

	// Add any other debug data we've collected
	for k, v := range ctx.debugData {
		result[k] = v
	}

	return result
}

// SetOutputValue sets an output value, storing it in our local map
func (ctx *LoopContext) SetOutputValue(pinID string, value types.Value) {
	// Initialize outputs map if needed
	if ctx.outputs == nil {
		ctx.outputs = make(map[string]types.Value)
	}

	// Store the value locally
	ctx.outputs[pinID] = value

	// For the index pin, make a copy to ensure unique references
	if pinID == "index" {
		// Create a new value with the same type and raw value
		// This ensures each connected node gets a distinct copy
		clonedValue := types.NewValue(value.Type, value.RawValue)
		ctx.ExecutionContext.SetOutputValue(pinID, clonedValue)

		// Log the index update for debugging
		ctx.Logger().Debug("Setting loop index output", map[string]interface{}{
			"index":     value.RawValue,
			"iteration": ctx.iterationsDone,
		})

		// Update the debug data to include this index value
		ctx.debugData["currentIndex"] = value.RawValue
	} else {
		// For other pins, pass through normally
		ctx.ExecutionContext.SetOutputValue(pinID, value)
	}
}

// GetOutputValue retrieves an output value from our local map
func (ctx *LoopContext) GetOutputValue(pinID string) (types.Value, bool) {
	// Check our local map first
	if ctx.outputs != nil {
		if value, exists := ctx.outputs[pinID]; exists {
			return value, true
		}
	}

	// Fall back to the parent context
	return ctx.ExecutionContext.(*DefaultExecutionContext).GetOutputValue(pinID)
}

// Removed GetOutputPins override as loopNode field is removed

// --- Getters and Setters for Loop State ---

func (ctx *LoopContext) StartIndex() float64 {
	return ctx.startIndex
}

func (ctx *LoopContext) MaxIterations() int {
	return ctx.maxIterations
}

func (ctx *LoopContext) CurrentIndex() float64 {
	return ctx.currentIndex
}

func (ctx *LoopContext) SetCurrentIndex(index float64) {
	ctx.currentIndex = index
}

func (ctx *LoopContext) IterationsDone() int {
	return ctx.iterationsDone
}

func (ctx *LoopContext) IncrementIterationsDone() {
	ctx.iterationsDone++
}

func (ctx *LoopContext) BodyCompletedSignal() <-chan bool {
	return ctx.bodyCompleted
}

func (ctx *LoopContext) ExecutionDoneSignal() <-chan bool {
	return ctx.executionDone
}

// SignalIterationComplete sends a signal indicating the loop body finished
func (ctx *LoopContext) SignalIterationComplete() {
	// Use non-blocking send in case the loop isn't waiting (e.g., error occurred)
	select {
	case ctx.bodyCompleted <- true:
	default:
		// Log or handle cases where the signal couldn't be sent?
		ctx.Logger().Warn("Could not send loop body completion signal", nil)
	}
}
