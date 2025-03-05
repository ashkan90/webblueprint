package logic

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// SequenceNode implements a node that executes a sequence of operations
type SequenceNode struct {
	node.BaseNode
}

// NewSequenceNode creates a new Sequence node
func NewSequenceNode() node.Node {
	return &SequenceNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "sequence",
				Name:        "Sequence",
				Description: "Executes multiple operations in sequence",
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
			},
			Outputs: []types.Pin{
				{
					ID:          "then1",
					Name:        "Then 1",
					Description: "First operation in sequence",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "then2",
					Name:        "Then 2",
					Description: "Second operation in sequence",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "then3",
					Name:        "Then 3",
					Description: "Third operation in sequence",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "then4",
					Name:        "Then 4",
					Description: "Fourth operation in sequence",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "completed",
					Name:        "Completed",
					Description: "Executed after all operations complete",
					Type:        types.PinTypes.Execution,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *SequenceNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Sequence node", nil)

	// Get direct execution capability if available
	directExecutionCtx, hasDirectExec := ctx.(interface {
		ExecuteConnectedNodes(pinID string) error
	})

	// Record start of sequence
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Sequence Started",
		Value: map[string]interface{}{
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("Starting sequence execution", nil)

	// We need to execute the steps in sequence
	if hasDirectExec {
		// Step 1
		logger.Debug("Executing sequence step 1", nil)
		if err := directExecutionCtx.ExecuteConnectedNodes("then1"); err != nil {
			logger.Error("Error in sequence step 1", map[string]interface{}{"error": err.Error()})
			return err
		}

		// Step 2
		logger.Debug("Executing sequence step 2", nil)
		if err := directExecutionCtx.ExecuteConnectedNodes("then2"); err != nil {
			logger.Error("Error in sequence step 2", map[string]interface{}{"error": err.Error()})
			return err
		}

		// Step 3
		logger.Debug("Executing sequence step 3", nil)
		if err := directExecutionCtx.ExecuteConnectedNodes("then3"); err != nil {
			logger.Error("Error in sequence step 3", map[string]interface{}{"error": err.Error()})
			return err
		}

		// Step 4
		logger.Debug("Executing sequence step 4", nil)
		if err := directExecutionCtx.ExecuteConnectedNodes("then4"); err != nil {
			logger.Error("Error in sequence step 4", map[string]interface{}{"error": err.Error()})
			return err
		}

		// Record completion of sequence
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Sequence Completed",
			Value: map[string]interface{}{
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
		})

		logger.Info("Sequence execution completed", nil)

		// Complete the sequence
		return directExecutionCtx.ExecuteConnectedNodes("completed")
	} else {
		// If we don't have direct execution capability, we can only activate the first step
		// and rely on each step activating the next
		logger.Warn("Direct execution not available, activating first step only", nil)

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Sequence Warning",
			Value: map[string]interface{}{
				"message":   "Direct execution not available, activating first step only",
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
		})

		return ctx.ActivateOutputFlow("then1")
	}
}
