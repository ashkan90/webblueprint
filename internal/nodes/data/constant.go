package data

import (
	"fmt"
	"strconv"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ConstantNode implements a node that outputs a constant value
type ConstantNode struct {
	node.BaseNode
	ValueType *types.PinType
	Value     interface{}
}

// NewConstantNode creates a new Constant node
func NewConstantNode() node.Node {
	return &ConstantNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "constant",
				Name:        "Constant",
				Description: "Outputs a constant value",
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
					Description: "Constant value output",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
			},
		},
		ValueType: types.PinTypes.String, // Default to string
		Value:     "",
	}
}

// NewStringConstantNode creates a new Constant node for string values
func NewStringConstantNode() node.Node {
	node := NewConstantNode().(*ConstantNode)
	node.Metadata.TypeID = "constant-string"
	node.Metadata.Name = "String Constant"
	node.Metadata.Description = "Outputs a constant string value"
	node.ValueType = types.PinTypes.String
	node.Value = ""
	node.Outputs[1].Type = types.PinTypes.String

	// Add an input for setting the value via property
	node.Inputs = append(node.Inputs, types.Pin{
		ID:          "constantValue",
		Name:        "Value",
		Description: "The constant string value to output",
		Type:        types.PinTypes.String,
		Optional:    true,
		Default:     "",
	})

	return node
}

// NewNumberConstantNode creates a new Constant node for number values
func NewNumberConstantNode() node.Node {
	node := NewConstantNode().(*ConstantNode)
	node.Metadata.TypeID = "constant-number"
	node.Metadata.Name = "Number Constant"
	node.Metadata.Description = "Outputs a constant numeric value"
	node.ValueType = types.PinTypes.Number
	node.Value = 0.0
	node.Outputs[1].Type = types.PinTypes.Number

	// Add an input for setting the value via property
	node.Inputs = append(node.Inputs, types.Pin{
		ID:          "constantValue",
		Name:        "Value",
		Description: "The constant numeric value to output",
		Type:        types.PinTypes.Number,
		Optional:    true,
		Default:     0.0,
	})

	return node
}

// NewBooleanConstantNode creates a new Constant node for boolean values
func NewBooleanConstantNode() node.Node {
	node := NewConstantNode().(*ConstantNode)
	node.Metadata.TypeID = "constant-boolean"
	node.Metadata.Name = "Boolean Constant"
	node.Metadata.Description = "Outputs a constant boolean value"
	node.ValueType = types.PinTypes.Boolean
	node.Value = false
	node.Outputs[1].Type = types.PinTypes.Boolean

	// Add an input for setting the value via property
	node.Inputs = append(node.Inputs, types.Pin{
		ID:          "constantValue",
		Name:        "Value",
		Description: "The constant boolean value to output",
		Type:        types.PinTypes.Boolean,
		Optional:    true,
		Default:     false,
	})

	return node
}

// Execute runs the node logic
func (n *ConstantNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Constant node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	bp, _ := db.Blueprints.GetBlueprint(ctx.GetBlueprintID())
	_ = bp

	// Try to get value from node properties
	var value = n.Value

	// First, check if there's a constantValue input
	if constInput, exists := ctx.GetInputValue("constantValue"); exists {
		// Use the input value directly since it's already properly typed
		value = constInput.RawValue
		debugData["valueSource"] = "input_pin"
	} else {
		// Check if the value is stored in the blueprint node properties
		// This is the preferred way as it allows the value to be edited in the UI
		nodeProperties := getNodePropertiesFromContext(ctx)
		if nodeProperties != nil {
			if constValue, exists := nodeProperties["constantValue"]; exists && constValue != nil {
				// Convert value based on type if needed
				typedValue, err := convertValueToType(constValue, n.ValueType)
				if err == nil {
					value = typedValue
					debugData["valueSource"] = "node_properties"
				} else {
					logger.Warn("Failed to convert property value to correct type", map[string]interface{}{
						"error": err.Error(),
					})
					debugData["conversionError"] = err.Error()
				}
			}
		}
	}

	debugData["initialValue"] = value

	// Set output value
	ctx.SetOutputValue("value", types.NewValue(n.ValueType, value))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Constant Value Output",
		Value: map[string]interface{}{
			"valueType": n.ValueType.Name,
			"value":     value,
			"debug":     debugData,
		},
		Timestamp: time.Now(),
	})

	logger.Info("Output constant value", map[string]interface{}{
		"valueType": n.ValueType.Name,
		"value":     value,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// Helper function to get node properties from execution context
func getNodePropertiesFromContext(ctx node.ExecutionContext) map[string]interface{} {
	// The debug data might contain node properties
	debugData := ctx.GetDebugData()
	if debugData == nil {
		return nil
	}

	// Check if we have a node configuration
	if nodeConfig, exists := debugData["nodeConfig"]; exists {
		if configMap, ok := nodeConfig.(map[string]interface{}); ok {
			if properties, exists := configMap["properties"]; exists {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					return propsMap
				}
			}
		}
	}

	return nil
}

// Helper function to convert a value to the expected type
func convertValueToType(value interface{}, targetType *types.PinType) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch targetType.ID {
	case "string":
		// Convert to string
		return fmt.Sprintf("%v", value), nil
	case "number":
		// Convert to number
		switch v := value.(type) {
		case float64:
			return v, nil
		case float32:
			return float64(v), nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return 0.0, fmt.Errorf("cannot convert %T to number", value)
		}
	case "boolean":
		// Convert to boolean
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		case float64:
			return v != 0, nil
		case int:
			return v != 0, nil
		default:
			return false, fmt.Errorf("cannot convert %T to boolean", value)
		}
	default:
		// Just return the value as is
		return value, nil
	}
}
