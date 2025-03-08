package data

import (
	"fmt"
	"strings"
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
			Inputs:  []types.Pin{},
			Outputs: []types.Pin{
				//{
				//	ID:          "value",
				//	Name:        "Value",
				//	Description: "Variable value",
				//	Type:        types.PinTypes.Any,
				//},
			},
		},
	}
}

// NewVariableGetNodeFor creates a specialized VariableGetNode for a specific variable
func NewVariableGetNodeFor(varName, varType string) func() node.Node {
	return func() node.Node {
		// Create a new node based on the generic VariableGetNode
		node := NewVariableGetNode().(*VariableGetNode)

		// Customize metadata for this variable
		node.Metadata.TypeID = "get-variable-" + varName
		node.Metadata.Name = "Get " + varName
		node.Metadata.Description = "Gets the value of variable '" + varName + "'"

		// Set default input for "name" to this variable's name
		//for i, pin := range node.Inputs {
		//	if pin.ID == "name" {
		//		// Create a copy of the pin
		//		updatedPin := pin
		//		updatedPin.Default = varName
		//		node.Inputs[i] = updatedPin
		//	}
		//}

		// Determine output type based on variable type
		pinType := types.PinTypes.Any
		switch varType {
		case "string":
			pinType = types.PinTypes.String
		case "number":
			pinType = types.PinTypes.Number
		case "boolean":
			pinType = types.PinTypes.Boolean
		case "object":
			pinType = types.PinTypes.Object
		case "array":
			pinType = types.PinTypes.Array
		}
		node.Outputs = append(node.Outputs, types.Pin{
			ID:          strings.ToLower(varName),
			Name:        varName,
			Description: fmt.Sprintf("Gets the value of variable '%s'", varName),
			Type:        pinType,
		})

		// Update output pin type
		//for i, pin := range node.Outputs {
		//	if pin.ID == "value" {
		//		// Create a copy of the pin
		//		updatedPin := pin
		//		updatedPin.Type = pinType
		//		node.Outputs[i] = updatedPin
		//	}
		//}

		return node
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

	// No need to continue execution since we removed the execution pins
	return nil
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
				//{
				//	ID:          "name",
				//	Name:        "Variable Name",
				//	Description: "Name of the variable to set",
				//	Type:        types.PinTypes.String,
				//	Optional:    true,
				//},
				//{
				//	ID:          "value",
				//	Name:        "Value",
				//	Description: "Value to set",
				//	Type:        types.PinTypes.Any,
				//	Optional:    true,
				//},
			},
			Outputs: []types.Pin{
				{
					ID:          "result",
					Name:        "Result",
					Description: "True if variable was set successfully",
					Type:        types.PinTypes.Boolean,
				},
			},
		},
	}
}

// NewVariableSetNodeFor creates a specialized VariableSetNode for a specific variable
func NewVariableSetNodeFor(varName, varType string) func() node.Node {
	return func() node.Node {
		// Create a new node based on the generic VariableSetNode
		node := NewVariableSetNode().(*VariableSetNode)

		// Customize metadata for this variable
		node.Metadata.TypeID = "set-variable-" + varName
		node.Metadata.Name = "Set " + varName
		node.Metadata.Description = "Sets the value of variable '" + varName + "'"

		// Determine input type based on variable type
		pinType := types.PinTypes.Any
		switch varType {
		case "string":
			pinType = types.PinTypes.String
		case "number":
			pinType = types.PinTypes.Number
		case "boolean":
			pinType = types.PinTypes.Boolean
		case "object":
			pinType = types.PinTypes.Object
		case "array":
			pinType = types.PinTypes.Array
		}

		// Update input value pin type
		node.Inputs = append(node.Inputs, types.Pin{
			ID:          strings.ToLower(varName),
			Name:        varName,
			Description: fmt.Sprintf("Sets the value of variable '%s'", varName),
			Type:        pinType,
			Optional:    true,
		})

		return node
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

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"nameExists":  nameExists,
		"valueExists": valueExists,
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

		// Set result output to false
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, false))
		return err
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

		// Set result output to false
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, false))
		return err
	}

	// Update debug data with actual values
	if valueExists {
		debugData["inputs"] = map[string]interface{}{
			"name":  varName,
			"value": value.RawValue,
			"type":  value.Type.Name,
		}

		// Set the variable
		ctx.SetVariable(varName, value)
	} else {
		debugData["inputs"] = map[string]interface{}{
			"name":  varName,
			"value": nil,
		}

		// Set the variable to nil
		ctx.SetVariable(varName, types.NewValue(types.PinTypes.Any, nil))
	}

	// Set result output to true
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, true))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Set Variable",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Set variable value", map[string]interface{}{
		"name":        varName,
		"valueExists": valueExists,
	})

	// No need to continue execution since we removed the execution pins
	return nil
}
