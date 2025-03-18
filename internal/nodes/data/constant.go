package data

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// StringConstantNode implements a node that outputs a constant string value
type StringConstantNode struct {
	node.BaseNode
}

// NewStringConstantNode creates a new String Constant node
func NewStringConstantNode() node.Node {
	return &StringConstantNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "constant-string",
				Name:        "String Constant",
				Description: "Outputs a constant string value",
				Category:    "Data",
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
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Constant string value",
					Type:        types.PinTypes.String,
				},
			},
			Properties: []types.Property{
				{
					Name:        "value",
					Description: "String value",
					Value:       "",
					Type:        types.PinTypes.String,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *StringConstantNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing String Constant node", nil)

	// Get the string value from properties, default to empty string
	value := ""
	if propValue, exists := ctx.GetInputValue("value"); exists {
		if strValue, err := propValue.AsString(); err == nil {
			value = strValue
		}
	}

	// Set the output value
	ctx.SetOutputValue("value", types.NewValue(types.PinTypes.String, value))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "String Constant",
		Value: map[string]interface{}{
			"value": value,
		},
		Timestamp: time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// NumberConstantNode implements a node that outputs a constant number value
type NumberConstantNode struct {
	node.BaseNode
}

// NewNumberConstantNode creates a new Number Constant node
func NewNumberConstantNode() node.Node {
	return &NumberConstantNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "constant-number",
				Name:        "Number Constant",
				Description: "Outputs a constant number value",
				Category:    "Data",
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
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Constant number value",
					Type:        types.PinTypes.Number,
				},
			},
			Properties: []types.Property{
				{
					Name:        "value",
					Description: "Number value",
					Value:       0.0,
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *NumberConstantNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Number Constant node", nil)

	// Get the number value from properties, default to 0
	value := 0.0
	if propValue, exists := ctx.GetInputValue("value"); exists {
		if numValue, err := propValue.AsNumber(); err == nil {
			value = numValue
		}
	}

	// Set the output value
	ctx.SetOutputValue("value", types.NewValue(types.PinTypes.Number, value))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Number Constant",
		Value: map[string]interface{}{
			"value": value,
		},
		Timestamp: time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// BooleanConstantNode implements a node that outputs a constant boolean value
type BooleanConstantNode struct {
	node.BaseNode
}

// NewBooleanConstantNode creates a new Boolean Constant node
func NewBooleanConstantNode() node.Node {
	return &BooleanConstantNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "constant-boolean",
				Name:        "Boolean Constant",
				Description: "Outputs a constant boolean value",
				Category:    "Data",
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
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Constant boolean value",
					Type:        types.PinTypes.Boolean,
				},
			},
			Properties: []types.Property{
				{
					Name:        "value",
					Description: "Boolean value",
					Value:       false,
					Type:        types.PinTypes.Boolean,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *BooleanConstantNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Boolean Constant node", nil)

	// Get the boolean value from properties, default to false
	value := false
	if propValue, exists := ctx.GetInputValue("value"); exists {
		if boolValue, err := propValue.AsBoolean(); err == nil {
			value = boolValue
		}
	}

	// Set the output value
	ctx.SetOutputValue("value", types.NewValue(types.PinTypes.Boolean, value))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Boolean Constant",
		Value: map[string]interface{}{
			"value": value,
		},
		Timestamp: time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
