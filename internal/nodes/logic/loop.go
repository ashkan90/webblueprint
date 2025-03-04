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
	nodeID := ctx.GetNodeID()
	loopVarName := fmt.Sprintf("_loop_%s", nodeID)

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

	// Get access to a direct execution capability, if available
	directExecutionCtx, hasDirectExec := ctx.(interface {
		ExecuteConnectedNodes(pinID string) error
	})

	// Loop initialization
	currentIndex := startValue

	logger.Info("Starting loop execution", map[string]interface{}{
		"iterations": maxIterations,
		"startIndex": startValue,
	})

	debugData["execution"] = "started"
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Loop Started",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Execute each iteration
	for i := 0; i < maxIterations; i++ {
		// Update debug data for this iteration
		iterDebugData := map[string]interface{}{
			"iteration":    i,
			"currentIndex": currentIndex,
			"timestamp":    time.Now(),
		}

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: fmt.Sprintf("Loop Iteration %d", i),
			Value:       iterDebugData,
			Timestamp:   time.Now(),
		})

		// Store the current index in a loop variable
		ctx.SetVariable(loopVarName, types.NewValue(types.PinTypes.Number, currentIndex))

		// Set output value for this iteration
		ctx.SetOutputValue("index", types.NewValue(types.PinTypes.Number, currentIndex))

		logger.Info(fmt.Sprintf("Loop iteration %d", i), map[string]interface{}{
			"index": currentIndex,
		})

		// If we have direct execution capability, use it
		if hasDirectExec {
			// Execute connected nodes immediately and synchronously
			err = directExecutionCtx.ExecuteConnectedNodes("loop")
			if err != nil {
				logger.Error("Failed to execute loop body", map[string]interface{}{
					"error":     err.Error(),
					"iteration": i,
				})
				break
			}
		} else {
			// Fall back to standard activation (might not work as expected)
			err = ctx.ActivateOutputFlow("loop")
			if err != nil {
				logger.Error("Failed to activate loop body", map[string]interface{}{
					"error":     err.Error(),
					"iteration": i,
				})
				break
			}
		}

		// Increment index for next iteration
		currentIndex++
	}

	// Record loop completion
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Loop Completed",
		Value: map[string]interface{}{
			"iterationsDone": maxIterations,
			"finalIndex":     currentIndex - 1,
		},
		Timestamp: time.Now(),
	})

	logger.Info("Loop execution completed", map[string]interface{}{
		"iterations": maxIterations,
	})

	// Activate the completed flow
	return ctx.ActivateOutputFlow("completed")
}
