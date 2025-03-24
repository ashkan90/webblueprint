package utility

import (
	"encoding/json"
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// PrintNode implements a node that prints values
type PrintNode struct {
	node.BaseNode
}

// NewPrintNode creates a new Print node
func NewPrintNode() node.Node {
	return &PrintNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "print",
				Name:        "Print",
				Description: "Prints a value to the console",
				Category:    "Utility",
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
					ID:          "message",
					Name:        "Message",
					Description: "Value to print",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "prefix",
					Name:        "Prefix",
					Description: "Prefix to add before the message",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Executed after printing",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "output",
					Name:        "Output",
					Description: "The same value that was printed",
					Type:        types.PinTypes.Any,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *PrintNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Print node", nil)

	// Get message to print
	messageValue, messageExists := ctx.GetInputValue("message")

	// Log what we received
	logger.Debug("Print node input received", map[string]interface{}{
		"messageExists": messageExists,
		"messageType":   fmt.Sprintf("%T", messageValue.RawValue),
	})

	if messageExists {
		// Try to log the value in a structured way
		jsonData, err := json.Marshal(messageValue.RawValue)
		if err == nil {
			logger.Debug("Message value", map[string]interface{}{
				"json": string(jsonData),
			})
		} else {
			logger.Debug("Non-serializable message value", map[string]interface{}{
				"error":  err.Error(),
				"string": fmt.Sprintf("%v", messageValue.RawValue),
			})
		}
	}

	// Collect debug data
	debugData := make(map[string]interface{})
	debugData["inputs"] = map[string]interface{}{
		"message": messageExists,
	}

	if !messageExists {
		logger.Warn("No message provided to print", nil)
		debugData["output"] = "[undefined]"

		// Output empty string
		ctx.SetOutputValue("output", types.NewValue(types.PinTypes.String, "[undefined]"))
	} else {
		// Get prefix if available
		var prefix string
		if prefixValue, exists := ctx.GetInputValue("prefix"); exists {
			if prefixStr, err := prefixValue.AsString(); err == nil {
				prefix = prefixStr
			}
		}

		// Format message for display
		var displayValue string

		switch messageValue.Type {
		case types.PinTypes.String:
			if messageStr, err := messageValue.AsString(); err == nil {
				displayValue = messageStr
			} else {
				displayValue = fmt.Sprintf("%v", messageValue.RawValue)
			}
		case types.PinTypes.Number, types.PinTypes.Boolean:
			displayValue = fmt.Sprintf("%v", messageValue.RawValue)
		default:
			// For objects and arrays, format as JSON
			jsonData, err := json.MarshalIndent(messageValue.RawValue, "", "  ")
			if err == nil {
				displayValue = string(jsonData)
			} else {
				displayValue = fmt.Sprintf("%v", messageValue.RawValue)
			}
		}

		// Add prefix if provided
		if prefix != "" {
			displayValue = prefix + " " + displayValue
		}

		// Record for debugging
		debugData["output"] = displayValue
		logger.Info("Printed message", map[string]interface{}{
			"message": displayValue,
		})

		// Pass through the value
		ctx.SetOutputValue("output", messageValue)
	}

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Print Output",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
