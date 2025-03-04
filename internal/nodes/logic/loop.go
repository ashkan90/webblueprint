package logic

import (
	"fmt"
	"time"
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
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Loop Start",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Execute the loop
	maxIterations := int(iterations)
	if maxIterations <= 0 {
		// If iterations <= 0, skip the loop body
		logger.Info("Loop skipped (iterations <= 0)", map[string]interface{}{
			"iterations": maxIterations,
		})

		debugData["execution"] = "skipped"
		debugData["reason"] = "iterations <= 0"
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Loop Skipped",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return ctx.ActivateOutputFlow("completed")
	}

	// Store the node and pin IDs for loop iteration
	nodeID := ctx.GetNodeID()

	// Create a variable to store the loop iteration
	loopVarName := fmt.Sprintf("_loop_%s", nodeID)

	// Initialize the loop variable
	ctx.SetVariable(loopVarName, types.NewValue(types.PinTypes.Number, startValue))

	// Set output for index
	ctx.SetOutputValue("index", types.NewValue(types.PinTypes.Number, startValue))

	// Register a hook function to continue the loop after each iteration
	// This has to be implemented by creating a special execution context for the loop
	// that intercepts the completion of the loop body and either starts the next iteration
	// or exits the loop.

	// For the current implementation, we'll execute the loop body and then immediately
	// activate the "completed" flow.

	logger.Info("Starting loop execution", map[string]interface{}{
		"iterations": maxIterations,
		"startValue": startValue,
	})

	debugData["execution"] = "started"
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Loop Started",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Activate the loop body flow
	err = ctx.ActivateOutputFlow("loop")
	if err != nil {
		return err
	}

	// After the loop body execution is complete, we would normally update the index and
	// check if more iterations are needed. For now, we'll just activate the "completed" flow.

	return ctx.ActivateOutputFlow("completed")
}
