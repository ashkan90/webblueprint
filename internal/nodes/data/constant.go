package data

import (
	"fmt"
	"strconv"
	"time"
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
			// No inputs - constants don't need execution inputs
			Outputs: []types.Pin{
				{
					ID:          "value",
					Name:        "Value",
					Description: "Constant value output",
					Type:        types.PinTypes.Any,
				},
			},
			Properties: []types.Property{
				{
					Name:        "constantValue",
					Description: "Default value",
					Value:       "",
					Type:        types.PinTypes.Any,
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
	node.Outputs[0].Type = types.PinTypes.String

	// Set default property value
	node.Properties[0] = types.Property{
		Name:        "constantValue",
		Description: "Default value",
		Value:       "",
		Type:        types.PinTypes.String,
	}

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
	node.Outputs[0].Type = types.PinTypes.Number

	// Set default property value
	node.Properties[0] = types.Property{
		Name:        "constantValue",
		Description: "Default value",
		Value:       0.0,
		Type:        types.PinTypes.Number,
	}

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
	node.Outputs[0].Type = types.PinTypes.Boolean

	// Set default property value
	node.Properties[0] = types.Property{
		Name:        "constantValue",
		Description: "Default value",
		Value:       false,
		Type:        types.PinTypes.Boolean,
	}

	return node
}

// Execute runs the node logic
// For constant nodes, we'll immediately set the output value on execution
// and won't activate any flows (since it doesn't have execution outputs)
func (n *ConstantNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Constant node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Start with the default value from the node definition
	var value = n.Value

	// Try to get the constantValue property directly from properties
	// This is the main path we expect to use for constant nodes
	if constValue, exists := getPropertyValue(ctx, "constantValue"); exists && constValue != nil {
		// Convert value based on type if needed
		typedValue, err := convertValueToType(constValue, n.ValueType)
		if err == nil {
			value = typedValue
			debugData["valueSource"] = "constantValue_property"
			debugData["propertyValue"] = constValue
		} else {
			logger.Warn("Failed to convert property value to correct type", map[string]interface{}{
				"error": err.Error(),
			})
			debugData["conversionError"] = err.Error()
		}
	} else {
		debugData["valueSource"] = "node_default"
	}

	debugData["finalValue"] = value
	debugData["valueType"] = fmt.Sprintf("%T", value)

	// Create a properly typed value to output
	outputValue := types.NewValue(n.ValueType, value)

	// Set output value
	ctx.SetOutputValue("value", outputValue)

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

	// No execution flow to activate - just return success
	return nil
}

// Helper function to get property value from execution context
func getPropertyValue(ctx node.ExecutionContext, name string) (interface{}, bool) {
	// Try to get the property from debug data first (for backwards compatibility)
	debugData := ctx.GetDebugData()
	if nodeConfig, exists := debugData["nodeConfig"]; exists {
		if configMap, ok := nodeConfig.(map[string]interface{}); ok {
			if properties, exists := configMap["properties"]; exists {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					if value, exists := propsMap[name]; exists {
						return value, true
					}
				}
			}
		}
	}

	// As an alternative, check for the property directly in the input value
	// This matches how the execution context now looks for properties
	if value, exists := ctx.GetInputValue(name); exists {
		return value.RawValue, true
	}

	return nil, false
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
