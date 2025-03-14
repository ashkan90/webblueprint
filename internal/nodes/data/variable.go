// internal/nodes/data/variable.go
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
	VariableName string // Store the variable name directly
}

// NewVariableGetNode creates a new generic Variable Get node
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
					ID:          "name",
					Name:        "Variable Name",
					Description: "Name of the variable to get",
					Type:        types.PinTypes.String,
				},
			},
			Outputs: []types.Pin{
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

// NewVariableGetNodeFor creates a specialized VariableGetNode for a specific variable
func NewVariableGetNodeFor(varName, varType string) func() node.Node {
	return func() node.Node {
		// Create a new node based on the generic VariableGetNode
		node := NewVariableGetNode().(*VariableGetNode)

		// Store the variable name directly in the node
		node.VariableName = varName

		// Customize metadata for this variable
		node.Metadata.TypeID = "get-variable-" + varName
		node.Metadata.Name = "Get " + varName
		node.Metadata.Description = "Gets the value of variable '" + varName + "'"

		// No inputs needed for specialized variable nodes
		node.Inputs = []types.Pin{}

		// Determine output type based on variable type
		pinType := types.PinTypes.Any
		switch strings.ToLower(varType) {
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

		// Create a single output with the variable name and appropriate type
		node.Outputs = []types.Pin{
			{
				ID:          strings.ToLower(varName),
				Name:        varName,
				Description: fmt.Sprintf("Value of variable '%s'", varName),
				Type:        pinType,
			},
		}

		return node
	}
}

// Execute runs the node logic
func (n *VariableGetNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Get Variable node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Determine the variable name (either from node property or input)
	var varName string
	if n.VariableName != "" {
		// Use the pre-configured variable name for specialized nodes
		varName = n.VariableName
		logger.Debug("Using pre-configured variable name", map[string]interface{}{
			"name": varName,
		})
	} else {
		// Get the variable name from input for generic variable nodes
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
		var err error
		varName, err = nameValue.AsString()
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
	}

	// Get the variable value from the execution context
	varValue, exists := ctx.GetVariable(varName)

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"name": varName,
	}

	// Set output value
	if exists {
		if n.VariableName != "" {
			// For specialized nodes, use the variable name as the output pin ID
			ctx.SetOutputValue(strings.ToLower(n.VariableName), varValue)
		} else {
			// For generic nodes, use "value" as the output pin ID
			ctx.SetOutputValue("value", varValue)
		}

		debugData["output"] = map[string]interface{}{
			"exists": true,
			"type":   varValue.Type.Name,
			"value":  varValue.RawValue,
		}

		logger.Info("Got variable value", map[string]interface{}{
			"name":   varName,
			"exists": exists,
			"value":  varValue.RawValue,
		})
	} else {
		// Variable not found, output null
		if n.VariableName != "" {
			ctx.SetOutputValue(strings.ToLower(n.VariableName), types.NewValue(types.PinTypes.Any, nil))
		} else {
			ctx.SetOutputValue("value", types.NewValue(types.PinTypes.Any, nil))
		}

		debugData["output"] = map[string]interface{}{
			"exists": false,
			"value":  nil,
		}

		logger.Warn("Variable not found", map[string]interface{}{
			"name": varName,
		})
	}

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Get Variable",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// No execution flow activation needed - this is a pure data node
	return nil
}

// VariableSetNode implements a node that sets a variable value
type VariableSetNode struct {
	node.BaseNode
	VariableName string // Store the variable name directly
}

// NewVariableSetNode creates a new generic Variable Set node
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
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value to set",
					Type:        types.PinTypes.Any,
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

		// Store the variable name directly in the node
		node.VariableName = varName

		// Customize metadata for this variable
		node.Metadata.TypeID = "set-variable-" + varName
		node.Metadata.Name = "Set " + varName
		node.Metadata.Description = "Sets the value of variable '" + varName + "'"

		// Determine input type based on variable type
		pinType := types.PinTypes.Any
		switch strings.ToLower(varType) {
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

		// For specialized nodes, we still need the execution pin
		node.Inputs = []types.Pin{
			{
				ID:          "exec",
				Name:        "Execute",
				Description: "Execution input",
				Type:        types.PinTypes.Execution,
			},
			{
				ID:          strings.ToLower(varName),
				Name:        varName,
				Description: fmt.Sprintf("Value for variable '%s'", varName),
				Type:        pinType,
			},
		}

		// Keep the execution output
		node.Outputs = []types.Pin{
			{
				ID:          "then",
				Name:        "Then",
				Description: "Execution continues",
				Type:        types.PinTypes.Execution,
			},
			{
				ID:          "result",
				Name:        "Result",
				Description: "True if variable was set successfully",
				Type:        types.PinTypes.Boolean,
			},
		}

		return node
	}
}

// Execute runs the node logic
func (n *VariableSetNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Set Variable node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Determine the variable name and value
	var varName string
	var value types.Value
	var valueExists bool

	if n.VariableName != "" {
		// Specialized node: use the pre-configured variable name
		varName = n.VariableName
		logger.Debug("Using pre-configured variable name", map[string]interface{}{
			"name": varName,
		})

		// Look for the input with the variable name
		value, valueExists = ctx.GetInputValue(strings.ToLower(varName))
	} else {
		// Generic node: get variable name from inputs
		nameValue, nameExists := ctx.GetInputValue("name")
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
		var err error
		varName, err = nameValue.AsString()
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

		// Get value from the value input
		value, valueExists = ctx.GetInputValue("value")
	}

	// Update debug data with actual values
	if valueExists {
		debugData["inputs"] = map[string]interface{}{
			"name":  varName,
			"value": value.RawValue,
			"type":  value.Type.Name,
		}

		// Set the variable in the execution context
		ctx.SetVariable(varName, value)

		logger.Info("Set variable value", map[string]interface{}{
			"name":  varName,
			"value": value.RawValue,
			"type":  value.Type.Name,
		})
	} else {
		debugData["inputs"] = map[string]interface{}{
			"name":  varName,
			"value": nil,
		}

		// Set the variable to nil
		ctx.SetVariable(varName, types.NewValue(types.PinTypes.Any, nil))

		logger.Warn("Set variable to nil", map[string]interface{}{
			"name": varName,
		})
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

	// Continue execution flow
	return ctx.ActivateOutputFlow("then")
}
