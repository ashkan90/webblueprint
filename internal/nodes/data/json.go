package data

import (
	"encoding/json"
	"fmt"
	"strings"
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
				Description: "Parse, stringify, or access JSON data",
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
					Description: "JSON operation: parse, stringify, or access",
					Type:        types.PinTypes.String,
					Default:     "parse",
				},
				{
					ID:          "input",
					Name:        "Input",
					Description: "JSON string or object to process",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "path",
					Name:        "Path",
					Description: "JSON path for 'access' operation (e.g., 'user.name')",
					Type:        types.PinTypes.String,
					Optional:    true,
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
					Type:        types.PinTypes.Any,
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
	inputValue, inputExists := ctx.GetInputValue("input")
	pathValue, pathExists := ctx.GetInputValue("path")
	indentOutputValue, indentOutputExists := ctx.GetInputValue("indentOutput")

	// Check required inputs
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !inputExists {
		err := fmt.Errorf("missing required input: input")
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
		"inputType":    fmt.Sprintf("%T", inputValue.RawValue),
		"hasPath":      pathExists,
		"indentOutput": indentOutput,
	}

	var result interface{}

	// Process based on operation
	switch strings.ToLower(operation) {
	case "parse":
		// Parse JSON string to object
		if inputStr, err := inputValue.AsString(); err == nil {
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
			jsonBytes, err = json.MarshalIndent(inputValue.RawValue, "", "  ")
		} else {
			jsonBytes, err = json.Marshal(inputValue.RawValue)
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

	case "access":
		// Access a property in a JSON object using path
		if !pathExists {
			logger.Error("Missing path for access operation", nil)
			debugData["error"] = "Missing path for access operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing path for access operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "JSON Access Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get the path
		path, err := pathValue.AsString()
		if err != nil {
			logger.Error("Invalid path", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid path: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid path: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "JSON Access Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// If input is a string, try to parse it first
		var dataObj map[string]interface{}
		if inputStr, err := inputValue.AsString(); err == nil {
			// Check if the input is a JSON string
			if strings.HasPrefix(strings.TrimSpace(inputStr), "{") {
				if err := json.Unmarshal([]byte(inputStr), &dataObj); err != nil {
					logger.Error("Input string is not valid JSON", map[string]interface{}{"error": err.Error()})
					debugData["error"] = "Input string is not valid JSON: " + err.Error()
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Input string is not valid JSON"))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "JSON Access Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			} else {
				logger.Error("Input is not a JSON object", nil)
				debugData["error"] = "Input is not a JSON object"
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Input is not a JSON object"))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "JSON Access Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}
		} else if inputObj, err := inputValue.AsObject(); err == nil {
			// Input is already an object
			dataObj = inputObj
		} else {
			logger.Error("Input is not a JSON object or string", nil)
			debugData["error"] = "Input is not a JSON object or string"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Input is not a JSON object or string"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "JSON Access Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Access the path
		pathParts := strings.Split(path, ".")
		var current interface{} = dataObj

		for _, part := range pathParts {
			if current == nil {
				logger.Error("Path element is null", map[string]interface{}{"part": part})
				debugData["error"] = fmt.Sprintf("Path element '%s' is null", part)
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Path element '%s' is null", part)))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "JSON Access Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}

			// If current is an object
			if currentObj, ok := current.(map[string]interface{}); ok {
				current = currentObj[part]
			} else if currentArr, ok := current.([]interface{}); ok {
				// If current is an array, try to parse the part as an index
				var index int
				if _, err := fmt.Sscanf(part, "%d", &index); err == nil && index >= 0 && index < len(currentArr) {
					current = currentArr[index]
				} else {
					logger.Error("Invalid array index", map[string]interface{}{"part": part})
					debugData["error"] = fmt.Sprintf("Invalid array index '%s'", part)
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid array index '%s'", part)))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "JSON Access Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			} else {
				logger.Error("Invalid path element", map[string]interface{}{"part": part})
				debugData["error"] = fmt.Sprintf("Invalid path element '%s'", part)
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid path element '%s'", part)))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "JSON Access Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}
		}

		result = current
		debugData["operation"] = "access"
		debugData["path"] = path
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
