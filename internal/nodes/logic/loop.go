package logic

import (
	"fmt"
	"time"

	// "webblueprint/internal/engineext" // Removed import to break cycle
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// LoopNode implements a loop that repeats execution a specified number of times
type LoopNode struct {
	node.BaseNode
}

// NewLoopNode creates a new Loop node
func NewLoopNode() node.Node {
	return &LoopNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "loop",
				Name:        "Loop",
				Description: "Executes a sequence of nodes multiple times",
				Category:    "Logic",
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
					ID:          "iterations",
					Name:        "Iterations",
					Description: "Number of times to loop",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "startValue",
					Name:        "Start Value",
					Description: "Initial index value (default: 0)",
					Type:        types.PinTypes.Number,
					Optional:    true,
					Default:     0,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "loop",
					Name:        "Loop Body",
					Description: "Executed for each iteration",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "completed",
					Name:        "Completed",
					Description: "Executed when all iterations are done",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "index",
					Name:        "Index",
					Description: "Current loop index",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *LoopNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Loop node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get the number of iterations
	iterationsValue, exists := ctx.GetInputValue("iterations")
	if !exists {
		err := fmt.Errorf("missing required input: iterations")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing iterations",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	iterations, err := iterationsValue.AsNumber()
	if err != nil {
		logger.Error("Invalid iterations value", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid iterations",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Get the start value (default to 0)
	startValue := float64(0)
	if startInput, exists := ctx.GetInputValue("startValue"); exists {
		if val, err := startInput.AsNumber(); err == nil {
			startValue = val
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"iterations": iterations,
		"startValue": startValue,
	}

	// Create a context variable to store loop state
	nodeID := ctx.GetNodeID()
	// loopVarName := fmt.Sprintf("_loop_%s", nodeID) // Removed unused variable

	// Record debug info for loop start
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      nodeID,
		Description: "Loop Start",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Convert iterations to int
	maxIterations := int(iterations)
	if maxIterations <= 0 {
		// If iterations <= 0, skip the loop body
		logger.Info("Loop skipped (iterations <= 0)", map[string]interface{}{
			"iterations": maxIterations,
		})

		debugData["execution"] = "skipped"
		debugData["reason"] = "iterations <= 0"
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      nodeID,
			Description: "Loop Skipped",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return ctx.ActivateOutputFlow("completed")
	}
	// --- Actor-Managed Loop Initialization ---

	// Get the underlying actor instance via the context
	// We need the concrete ActorExecutionContext to get the actor reference
	actorCtx, ok := ctx.(*engine.ActorExecutionContext) // Use concrete type from engine
	if !ok {
		// This node requires the actor model context to function correctly
		err := fmt.Errorf("LoopNode requires ActorExecutionContext in actor mode, received %T", ctx)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		// Cannot proceed, maybe activate error flow?
		// For now, return error. Loop won't start.
		return err
	}
	loopActor := actorCtx.GetActor() // Get the actor reference

	// Initialize loop state within the actor
	loopActor.InitializeLoop(startValue, maxIterations) // Assuming InitializeLoop method exists

	logger.Info("Initialized loop state in actor", map[string]interface{}{
		"iterations": maxIterations,
		"startIndex": startValue,
	})

	// Record loop start debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      nodeID,
		Description: "Loop Start Initialized",
		Value:       debugData, // Contains inputs
		Timestamp:   time.Now(),
	})

	// Send the first "loop_next" message to self asynchronously to start the iteration process
	loopActor.SendAsync(engine.NodeMessage{Type: "loop_next"})
	// Execute returns immediately; the loop runs via actor messages
	return nil
}
