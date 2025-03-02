package logic

import (
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// IfConditionNode implements a conditional branch
type IfConditionNode struct {
	node.BaseNode
}

// NewIfConditionNode creates a new If condition node
func NewIfConditionNode() node.Node {
	return &IfConditionNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "if-condition",
				Name:        "If Condition",
				Description: "Executes one of two branches based on a condition",
				Category:    "Logic",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "exec",
					Name:        "Execution",
					Description: "Execution input",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "condition",
					Name:        "Condition",
					Description: "Boolean condition to evaluate",
					Type:        types.PinTypes.Boolean,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "true",
					Name:        "True",
					Description: "Executed if condition is true",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "false",
					Name:        "False",
					Description: "Executed if condition is false",
					Type:        types.PinTypes.Execution,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *IfConditionNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing If Condition node", nil)

	// Get the condition value
	conditionValue, exists := ctx.GetInputValue("condition")
	if !exists {
		err := fmt.Errorf("missing required input: condition")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		return err
	}

	// Convert to boolean
	condition, err := conditionValue.AsBoolean()
	if err != nil {
		logger.Error("Invalid condition value", map[string]interface{}{"error": err.Error()})
		return err
	}

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Condition evaluation",
		Value: map[string]interface{}{
			"condition": condition,
		},
		Timestamp: time.Now(),
	})

	// Activate the appropriate output flow
	if condition {
		logger.Info("Condition is true, taking true branch", nil)
		return ctx.ActivateOutputFlow("true")
	} else {
		logger.Info("Condition is false, taking false branch", nil)
		return ctx.ActivateOutputFlow("false")
	}
}
