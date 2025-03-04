package data

import (
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// VariableGetNode implements a node that gets a variable value
type VariableGetNode struct {
	node.BaseNode
}

// NewVariableGetNode creates a new Variable Get node
func NewVariableGetNode() node.Node {
	return &VariableGetNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "variable-get",
				Name:        "Get Variable",
				Description: "Gets the value of a variable",
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
				{
					ID:          "name",
					Name:        "Variable Name",
					Description: "Name of the variable to get",
					Type:        types.PinTypes.String,
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
					Description: "Variable value",
					Type:        types.PinTypes.Any,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *VariableGetNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Get Variable node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get the variable name
	nameValue, exists := ctx.GetInputValue("name")
	if !exists {
		err := fmt.Errorf("missing required input: name")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing variable name",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Convert to string
	varName, err := nameValue.AsString()
	if err != nil {
		logger.Error("Invalid variable name", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid variable name",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Get the variable value
	varValue, exists := ctx.GetVariable(varName)

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"name": varName,
	}

	// Set output value
	if exists {
		ctx.SetOutputValue("value", varValue)
		debugData["output"] = map[string]interface{}{
			"exists": true,
			"type":   varValue.Type.Name,
			"value":  varValue.RawValue,
		}
	} else {
		// Variable not found, output null
		ctx.SetOutputValue("value", types.NewValue(types.PinTypes.Any, nil))
		debugData["output"] = map[string]interface{}{
			"exists": false,
			"value":  nil,
		}
	}

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Get Variable",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Get variable value", map[string]interface{}{
		"name":   varName,
		"exists": exists,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// VariableSetNode implements a node that sets a variable value
type VariableSetNode struct {
	node.BaseNode
}

// NewVariableSetNode creates a new Variable Set node
func NewVariableSetNode() node.Node {
	return &VariableSetNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "variable-set",
				Name:        "Set Variable",
				Description: "Sets the value of a variable",
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
				{
					ID:          "name",
					Name:        "Variable Name",
					Description: "Name of the variable to set",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value to set",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues",
					Type:        types.PinTypes.Execution,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *VariableSetNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Set Variable node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get the variable name
	nameValue, nameExists := ctx.GetInputValue("name")
	// Get the variable value (now optional)
	value, valueExists := ctx.GetInputValue("value")
	// Get default value (if provided)
	defaultValue, defaultValueExists := ctx.GetInputValue("defaultValue")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"nameExists":         nameExists,
		"valueExists":        valueExists,
		"defaultValueExists": defaultValueExists,
	}

	if !nameExists {
		err := fmt.Errorf("missing required input: name")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing variable name",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Use default value if no value is connected
	if !valueExists {
		if defaultValueExists {
			value = defaultValue
			valueExists = true
			debugData["usingDefault"] = true
		} else {
			logger.Warn("No value or default value provided, using null", nil)
			// Create a null value of type Any
			value = types.NewValue(types.PinTypes.Any, nil)
			valueExists = true
			debugData["usingNull"] = true
		}
	}

	// Convert name to string
	varName, err := nameValue.AsString()
	if err != nil {
		logger.Error("Invalid variable name", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid variable name",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Update debug data with actual values
	debugData["inputs"] = map[string]interface{}{
		"name":  varName,
		"value": value.RawValue,
		"type":  value.Type.Name,
	}

	// Set the variable
	ctx.SetVariable(varName, value)

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Set Variable",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Set variable value", map[string]interface{}{
		"name":  varName,
		"type":  value.Type.Name,
		"value": value.RawValue,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
