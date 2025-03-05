package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// TypeConversionNode implements a node that converts between different data types
type TypeConversionNode struct {
	node.BaseNode
}

// NewTypeConversionNode creates a new Type Conversion node
func NewTypeConversionNode() node.Node {
	return &TypeConversionNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "type-conversion",
				Name:        "Type Conversion",
				Description: "Convert values between different data types",
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
					ID:          "input",
					Name:        "Input",
					Description: "Value to convert",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "targetType",
					Name:        "Target Type",
					Description: "Type to convert to: string, number, boolean, array, object",
					Type:        types.PinTypes.String,
					Default:     "string",
				},
				{
					ID:          "parseFormat",
					Name:        "Parse Format",
					Description: "Format for parsing (e.g. date format)",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "radix",
					Name:        "Radix",
					Description: "Base for number conversion (default: 10)",
					Type:        types.PinTypes.Number,
					Optional:    true,
					Default:     10,
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
					Description: "Converted value",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "errorMessage",
					Name:        "Error Message",
					Description: "Error message if conversion fails",
					Type:        types.PinTypes.String,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *TypeConversionNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Type Conversion node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	inputValue, inputExists := ctx.GetInputValue("input")
	targetTypeValue, targetTypeExists := ctx.GetInputValue("targetType")
	parseFormatValue, parseFormatExists := ctx.GetInputValue("parseFormat")
	radixValue, radixExists := ctx.GetInputValue("radix")

	// Check required inputs
	if !inputExists {
		err := fmt.Errorf("missing required input: input")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !targetTypeExists {
		err := fmt.Errorf("missing required input: targetType")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse target type
	targetType, err := targetTypeValue.AsString()
	if err != nil {
		logger.Error("Invalid target type", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid target type: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get parse format if provided
	parseFormat := ""
	if parseFormatExists {
		if formatStr, err := parseFormatValue.AsString(); err == nil {
			parseFormat = formatStr
		}
	}

	// Get radix if provided
	radix := 10
	if radixExists {
		if radixNum, err := radixValue.AsNumber(); err == nil {
			radix = int(radixNum)
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"inputType":   fmt.Sprintf("%T", inputValue.RawValue),
		"targetType":  targetType,
		"hasFormat":   parseFormatExists,
		"parseFormat": parseFormat,
		"radix":       radix,
	}

	var result interface{}

	// Process based on target type
	switch strings.ToLower(targetType) {
	case "string":
		// Convert to string
		if strValue, err := inputValue.AsString(); err == nil {
			result = strValue
		} else {
			// If AsString fails, just use fmt
			result = fmt.Sprintf("%v", inputValue.RawValue)
		}
		debugData["operation"] = "tostring"
		debugData["successful"] = true

	case "number":
		// Convert to number
		if numValue, err := inputValue.AsNumber(); err == nil {
			result = numValue
		} else {
			// Try to parse from string
			if strValue, err := inputValue.AsString(); err == nil {
				// Try to parse using the specified radix
				if num, err := strconv.ParseInt(strValue, radix, 64); err == nil {
					result = float64(num)
				} else if num, err := strconv.ParseFloat(strValue, 64); err == nil {
					result = num
				} else {
					logger.Error("Could not convert to number", map[string]interface{}{"error": err.Error()})
					debugData["error"] = "Could not convert to number: " + err.Error()
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not convert to number: "+err.Error()))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "Type Conversion Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			} else {
				logger.Error("Could not convert to number", map[string]interface{}{"error": err.Error()})
				debugData["error"] = "Could not convert to number: " + err.Error()
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not convert to number: "+err.Error()))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "Type Conversion Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}
		}
		debugData["operation"] = "tonumber"
		debugData["successful"] = true

	case "boolean":
		// Convert to boolean
		if boolValue, err := inputValue.AsBoolean(); err == nil {
			result = boolValue
		} else {
			// Try additional boolean conversions
			if strValue, err := inputValue.AsString(); err == nil {
				lowerStr := strings.ToLower(strValue)
				if lowerStr == "true" || lowerStr == "yes" || lowerStr == "1" || lowerStr == "on" {
					result = true
				} else if lowerStr == "false" || lowerStr == "no" || lowerStr == "0" || lowerStr == "off" || lowerStr == "" {
					result = false
				} else {
					logger.Error("Could not convert to boolean", nil)
					debugData["error"] = "Could not convert to boolean"
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not convert to boolean"))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "Type Conversion Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			} else if numValue, err := inputValue.AsNumber(); err == nil {
				result = numValue != 0
			} else {
				logger.Error("Could not convert to boolean", map[string]interface{}{"error": err.Error()})
				debugData["error"] = "Could not convert to boolean: " + err.Error()
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not convert to boolean: "+err.Error()))

				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					Description: "Type Conversion Error",
					Value:       debugData,
					Timestamp:   time.Now(),
				})

				return ctx.ActivateOutputFlow("error")
			}
		}
		debugData["operation"] = "toboolean"
		debugData["successful"] = true

	case "array":
		// Convert to array
		if arrayValue, err := inputValue.AsArray(); err == nil {
			result = arrayValue
		} else {
			// Try to convert string to array by splitting
			if strValue, err := inputValue.AsString(); err == nil {
				separator := ","
				if parseFormatExists {
					separator = parseFormat
				}

				parts := strings.Split(strValue, separator)
				resultArray := make([]interface{}, len(parts))
				for i, part := range parts {
					resultArray[i] = strings.TrimSpace(part)
				}
				result = resultArray
			} else if objValue, err := inputValue.AsObject(); err == nil {
				// Convert object to array of values
				resultArray := make([]interface{}, 0, len(objValue))
				for _, v := range objValue {
					resultArray = append(resultArray, v)
				}
				result = resultArray
			} else {
				// Convert single value to array with one element
				result = []interface{}{inputValue.RawValue}
			}
		}
		debugData["operation"] = "toarray"
		debugData["successful"] = true

	case "object":
		// Convert to object
		if objValue, err := inputValue.AsObject(); err == nil {
			result = objValue
		} else if arrayValue, err := inputValue.AsArray(); err == nil {
			// Convert array to object with numeric keys
			resultObject := make(map[string]interface{})
			for i, v := range arrayValue {
				resultObject[fmt.Sprintf("%d", i)] = v
			}
			result = resultObject
		} else {
			// For other types, create an object with a default key
			defaultKey := "value"
			if parseFormatExists {
				defaultKey = parseFormat
			}
			result = map[string]interface{}{
				defaultKey: inputValue.RawValue,
			}
		}
		debugData["operation"] = "toobject"
		debugData["successful"] = true

	case "date":
		// Convert to date (as timestamp)
		var timestamp time.Time

		if strValue, err := inputValue.AsString(); err == nil {
			// Try to parse using the specified format if provided
			if parseFormatExists {
				if parsedTime, err := time.Parse(parseFormat, strValue); err == nil {
					timestamp = parsedTime
				} else {
					logger.Error("Could not parse date", map[string]interface{}{"error": err.Error()})
					debugData["error"] = "Could not parse date: " + err.Error()
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not parse date: "+err.Error()))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "Type Conversion Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			} else {
				// Try common formats
				layouts := []string{
					time.RFC3339,
					"2006-01-02T15:04:05",
					"2006-01-02 15:04:05",
					"2006-01-02",
					"01/02/2006",
					"01-02-2006",
					"January 2, 2006",
					"Jan 2, 2006",
				}

				parsed := false
				for _, layout := range layouts {
					if parsedTime, err := time.Parse(layout, strValue); err == nil {
						timestamp = parsedTime
						parsed = true
						break
					}
				}

				if !parsed {
					logger.Error("Could not parse date with any known format", nil)
					debugData["error"] = "Could not parse date with any known format"
					ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not parse date with any known format"))

					ctx.RecordDebugInfo(types.DebugInfo{
						NodeID:      ctx.GetNodeID(),
						Description: "Type Conversion Error",
						Value:       debugData,
						Timestamp:   time.Now(),
					})

					return ctx.ActivateOutputFlow("error")
				}
			}
		} else if numValue, err := inputValue.AsNumber(); err == nil {
			// Treat as Unix timestamp
			timestamp = time.Unix(int64(numValue), 0)
		} else {
			logger.Error("Could not convert to date", nil)
			debugData["error"] = "Could not convert to date"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Could not convert to date"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Type Conversion Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Format the result as a string in RFC3339 format (ISO standard)
		result = timestamp.Format(time.RFC3339)
		debugData["operation"] = "todate"
		debugData["successful"] = true

	default:
		logger.Error("Invalid target type", map[string]interface{}{"targetType": targetType})
		debugData["error"] = fmt.Sprintf("Invalid target type: %s", targetType)
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid target type: %s", targetType)))

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Type Conversion Error",
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
		Description: "Type Conversion",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Type conversion completed", map[string]interface{}{
		"targetType": targetType,
		"success":    true,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
