package data

import (
	"encoding/json"
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// JSONNode implements a node that provides JSON operations
type JSONNode struct {
	node.BaseNode
}

// NewJSONNode creates a new JSON node
func NewJSONNode() node.Node {
	return &JSONNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "json-processor",
				Name:        "JSON Processor",
				Description: "Parse or stringify JSON data",
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
					ID:          "operation",
					Name:        "Operation",
					Description: "JSON operation: parse or stringify",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "data",
					Name:        "Data",
					Description: "JSON string or object to process",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "indentOutput",
					Name:        "Indent Output",
					Description: "Whether to indent the JSON output (stringify operation)",
					Type:        types.PinTypes.Boolean,
					Optional:    true,
					Default:     false,
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
					ID:          "error",
					Name:        "Error",
					Description: "Executed if an error occurs",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "result",
					Name:        "Result",
					Description: "Result of the JSON operation",
					Type:        types.PinTypes.Object,
				},
				{
					ID:          "errorMessage",
					Name:        "Error Message",
					Description: "Error message if operation fails",
					Type:        types.PinTypes.String,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *JSONNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing JSON node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	operationValue, operationExists := ctx.GetInputValue("operation")
	dataValue, dataExists := ctx.GetInputValue("data")
	indentOutputValue, indentOutputExists := ctx.GetInputValue("indentOutput")

	// Check required inputs
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !dataExists {
		err := fmt.Errorf("missing required input: data")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse operation
	operation, err := operationValue.AsString()
	if err != nil {
		logger.Error("Invalid operation", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid operation: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Default values
	indentOutput := false
	if indentOutputExists {
		indentOutputBool, err := indentOutputValue.AsBoolean()
		if err == nil {
			indentOutput = indentOutputBool
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"operation":    operation,
		"indentOutput": indentOutput,
	}

	var result interface{}

	// Process based on operation
	switch operation {
	case "parse":
		// Parse JSON string to object
		if inputStr, err := dataValue.AsString(); err == nil {
			var parsedData interface{}
			if err := json.Unmarshal([]byte(inputStr), &parsedData); err != nil {
				logger.Error("JSON parse error", map[string]interface{}{"error": err.Error()})
				debugData["error"] = err.Error()
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Parse error: "+err.Error()))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "JSON Parse Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}
			result = parsedData
			debugData["operation"] = "parse"
			debugData["successful"] = true
		} else {
			logger.Error("Input is not a string", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Input is not a string"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Input is not a string"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "JSON Parse Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

	case "stringify":
		// Convert object to JSON string
		var jsonBytes []byte
		var err error

		if indentOutput {
			jsonBytes, err = json.MarshalIndent(dataValue.RawValue, "", "  ")
		} else {
			jsonBytes, err = json.Marshal(dataValue.RawValue)
		}

		if err != nil {
			logger.Error("JSON stringify error", map[string]interface{}{"error": err.Error()})
			debugData["error"] = err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Stringify error: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "JSON Stringify Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		result = string(jsonBytes)
		debugData["operation"] = "stringify"
		debugData["successful"] = true

	default:
		logger.Error("Invalid operation", map[string]interface{}{"operation": operation})
		debugData["error"] = fmt.Sprintf("Invalid operation: %s", operation)
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid operation: %s", operation)))

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "JSON Operation Error",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return ctx.ActivateOutputFlow("error")
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, result))

	debugData["result"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "JSON Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("JSON operation completed", map[string]interface{}{
		"operation": operation,
		"success":   true,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
