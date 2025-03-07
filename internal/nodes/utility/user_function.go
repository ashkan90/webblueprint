package utility

import (
	"fmt"
	"strings"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// UserFunctionNode implements a node that prints values
type UserFunctionNode struct {
	node.BaseNode
}

// NewUserFunctionNode creates a new Print node
func NewUserFunctionNode(function blueprint.Function) func() node.Node {
	return func() node.Node {
		var fnNode = &UserFunctionNode{
			BaseNode: node.BaseNode{
				Metadata: node.NodeMetadata{
					TypeID:      strings.ToLower(function.Name),
					Name:        function.Name,
					Description: function.Description,
					Category:    "Function",
					Version:     "1.0.0",
				},
			},
		}

		for _, input := range function.NodeType.Inputs {
			fnNode.Inputs = append(fnNode.Inputs, types.Pin{
				ID:          input.ID,
				Name:        input.Name,
				Description: input.Description,
				Type: &types.PinType{
					ID:          input.Type.ID,
					Name:        input.Type.Name,
					Description: input.Type.Description,
				},
				Optional: input.Optional,
				Default:  input.Default,
			})
		}

		for _, output := range function.NodeType.Outputs {
			fnNode.Outputs = append(fnNode.Outputs, types.Pin{
				ID:          output.ID,
				Name:        output.Name,
				Description: output.Description,
				Type: &types.PinType{
					ID:          output.Type.ID,
					Name:        output.Type.Name,
					Description: output.Type.Description,
				},
				Optional: output.Optional,
				Default:  output.Default,
			})
		}

		for _, property := range function.NodeType.Properties {
			fnNode.Properties = append(fnNode.Properties, types.Property{
				Name:  property.Name,
				Value: property.Value,
			})
		}

		return fnNode
	}
}

// Execute runs the node logic
func (n *UserFunctionNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Function node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})
	debugData["inputs"] = make(map[string]interface{})
	debugData["outputs"] = make(map[string]interface{})

	for _, input := range n.GetInputPins() {
		inputValue, inputValueExist := ctx.GetInputValue(input.ID)
		if inputValueExist && inputValue.Type.ID != "exec" {
			logger.Debug(fmt.Sprintf("Function Received Input Value Named: %s", input.ID), map[string]interface{}{
				input.ID: inputValue.RawValue,
			})

			debugData["inputs"].(map[string]interface{})[input.ID] = inputValue.RawValue

			for _, output := range n.GetOutputPins() {
				if output.ID == input.ID {
					ctx.SetOutputValue(output.ID, inputValue)
					logger.Debug(fmt.Sprintf("Function Received Output Named: %s", output.ID), map[string]interface{}{
						output.ID: inputValue.RawValue,
					})

					debugData["inputs"].(map[string]interface{})[input.ID] = inputValue.RawValue
				}
			}
		}
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
