package data

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ArrayNode implements a node that provides array operations
type ArrayNode struct {
	node.BaseNode
}

// NewArrayNode creates a new Array operations node
func NewArrayNode() node.Node {
	return &ArrayNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "array-operations",
				Name:        "Array Operations",
				Description: "Perform operations on arrays",
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
					Description: "Array operation: map, filter, forEach, sort, reverse, join, slice, concat",
					Type:        types.PinTypes.String,
					Default:     "map",
				},
				{
					ID:          "array",
					Name:        "Array",
					Description: "Array to operate on",
					Type:        types.PinTypes.Array,
				},
				{
					ID:          "propertyPath",
					Name:        "Property Path",
					Description: "Path to property for map/filter (e.g., 'user.name')",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "filterValue",
					Name:        "Filter Value",
					Description: "Value to filter by (for filter operation)",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "separator",
					Name:        "Separator",
					Description: "Separator for join operation",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     ",",
				},
				{
					ID:          "startIndex",
					Name:        "Start Index",
					Description: "Start index for slice operation",
					Type:        types.PinTypes.Number,
					Optional:    true,
					Default:     0,
				},
				{
					ID:          "endIndex",
					Name:        "End Index",
					Description: "End index for slice operation",
					Type:        types.PinTypes.Number,
					Optional:    true,
				},
				{
					ID:          "arrayToConcat",
					Name:        "Array to Concat",
					Description: "Second array for concat operation",
					Type:        types.PinTypes.Array,
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
				{
					ID:          "error",
					Name:        "Error",
					Description: "Executed if an error occurs",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "result",
					Name:        "Result",
					Description: "Result of the array operation",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "length",
					Name:        "Length",
					Description: "Length of the resulting array",
					Type:        types.PinTypes.Number,
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
func (n *ArrayNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Array node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	operationValue, operationExists := ctx.GetInputValue("operation")
	arrayValue, arrayExists := ctx.GetInputValue("array")
	propertyPathValue, propertyPathExists := ctx.GetInputValue("propertyPath")
	filterValue, filterValueExists := ctx.GetInputValue("filterValue")
	separatorValue, separatorExists := ctx.GetInputValue("separator")
	startIndexValue, startIndexExists := ctx.GetInputValue("startIndex")
	endIndexValue, endIndexExists := ctx.GetInputValue("endIndex")
	arrayConcatValue, arrayConcatExists := ctx.GetInputValue("arrayToConcat")

	// Check required inputs
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
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

	// Parse array
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"operation":       operation,
		"arrayLength":     len(array),
		"hasPropertyPath": propertyPathExists,
		"hasFilterValue":  filterValueExists,
	}

	var result interface{}
	var resultArray []interface{}
	var resultLength int

	// Process based on operation
	switch strings.ToLower(operation) {
	case "map":
		// Map array elements to a property
		if !propertyPathExists {
			logger.Error("Missing property path for map operation", nil)
			debugData["error"] = "Missing property path for map operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing property path for map operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Array Map Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get property path
		propertyPath, err := propertyPathValue.AsString()
		if err != nil {
			logger.Error("Invalid property path", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid property path: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid property path: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Array Map Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Map elements to the specified property
		resultArray = make([]interface{}, 0, len(array))
		pathParts := strings.Split(propertyPath, ".")

		for _, item := range array {
			if item == nil {
				resultArray = append(resultArray, nil)
				continue
			}

			// Navigate to the property
			currentValue := item
			found := true

			for _, part := range pathParts {
				if currentMap, ok := currentValue.(map[string]interface{}); ok {
					if value, exists := currentMap[part]; exists {
						currentValue = value
					} else {
						found = false
						break
					}
				} else {
					found = false
					break
				}
			}

			if found {
				resultArray = append(resultArray, currentValue)
			} else {
				resultArray = append(resultArray, nil)
			}
		}

		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "map"
		debugData["propertyPath"] = propertyPath
		debugData["successful"] = true

	case "filter":
		// Filter array elements by a property and value
		if !propertyPathExists {
			logger.Error("Missing property path for filter operation", nil)
			debugData["error"] = "Missing property path for filter operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing property path for filter operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Array Filter Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		if !filterValueExists {
			logger.Error("Missing filter value for filter operation", nil)
			debugData["error"] = "Missing filter value for filter operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing filter value for filter operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Array Filter Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get property path and filter value
		propertyPath, err := propertyPathValue.AsString()
		if err != nil {
			logger.Error("Invalid property path", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid property path: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid property path: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Array Filter Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Filter elements by the specified property and value
		resultArray = make([]interface{}, 0)
		pathParts := strings.Split(propertyPath, ".")
		filterVal := filterValue.RawValue

		for _, item := range array {
			if item == nil {
				continue
			}

			// Navigate to the property
			currentValue := item
			found := true

			for _, part := range pathParts {
				if currentMap, ok := currentValue.(map[string]interface{}); ok {
					if value, exists := currentMap[part]; exists {
						currentValue = value
					} else {
						found = false
						break
					}
				} else {
					found = false
					break
				}
			}

			if found && currentValue == filterVal {
				resultArray = append(resultArray, item)
			}
		}

		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "filter"
		debugData["propertyPath"] = propertyPath
		debugData["filterValue"] = filterVal
		debugData["successful"] = true

	case "foreach":
		// Pass through array but provide access to each element
		// This is mostly for looping through the array in the UI
		result = array
		resultLength = len(array)
		debugData["operation"] = "forEach"
		debugData["successful"] = true

	case "sort":
		// Sort the array (only works for arrays of strings or numbers)
		resultArray = make([]interface{}, len(array))
		copy(resultArray, array)

		// Try to determine array type and sort accordingly
		if len(array) > 0 {
			firstItem := array[0]
			if _, isString := firstItem.(string); isString {
				// Sort string array
				stringArray := make([]string, len(array))
				allStrings := true

				for i, item := range array {
					if str, ok := item.(string); ok {
						stringArray[i] = str
					} else {
						allStrings = false
						break
					}
				}

				if allStrings {
					sort.Strings(stringArray)
					for i, str := range stringArray {
						resultArray[i] = str
					}
				} else {
					logger.Warn("Array contains mixed types, sorting may not be accurate", nil)
				}
			} else if _, isNumber := firstItem.(float64); isNumber {
				// Sort number array
				numberArray := make([]float64, len(array))
				allNumbers := true

				for i, item := range array {
					if num, ok := item.(float64); ok {
						numberArray[i] = num
					} else {
						allNumbers = false
						break
					}
				}

				if allNumbers {
					sort.Float64s(numberArray)
					for i, num := range numberArray {
						resultArray[i] = num
					}
				} else {
					logger.Warn("Array contains mixed types, sorting may not be accurate", nil)
				}
			} else {
				logger.Warn("Array contains complex types that cannot be automatically sorted", nil)
			}
		}

		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "sort"
		debugData["successful"] = true

	case "reverse":
		// Reverse the array
		resultArray = make([]interface{}, len(array))
		for i, item := range array {
			resultArray[len(array)-1-i] = item
		}

		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "reverse"
		debugData["successful"] = true

	case "join":
		// Join array elements into a string
		separator := ","
		if separatorExists {
			if sepStr, err := separatorValue.AsString(); err == nil {
				separator = sepStr
			}
		}

		// Convert all elements to strings and join
		stringArray := make([]string, len(array))
		for i, item := range array {
			stringArray[i] = fmt.Sprintf("%v", item)
		}

		result = strings.Join(stringArray, separator)
		resultLength = len(array)
		debugData["operation"] = "join"
		debugData["separator"] = separator
		debugData["successful"] = true

	case "slice":
		// Slice the array (similar to array.slice in JavaScript)
		startIndex := 0
		if startIndexExists {
			if startIdx, err := startIndexValue.AsNumber(); err == nil {
				startIndex = int(startIdx)
			}
		}

		endIndex := len(array)
		if endIndexExists {
			if endIdx, err := endIndexValue.AsNumber(); err == nil {
				endIndex = int(endIdx)
			}
		}

		// Bound check
		if startIndex < 0 {
			startIndex = 0
		}
		if endIndex > len(array) {
			endIndex = len(array)
		}
		if startIndex > endIndex {
			startIndex = endIndex
		}

		// Create the slice
		resultArray = array[startIndex:endIndex]
		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "slice"
		debugData["startIndex"] = startIndex
		debugData["endIndex"] = endIndex
		debugData["successful"] = true

	case "concat":
		// Concatenate two arrays
		resultArray = make([]interface{}, len(array))
		copy(resultArray, array)

		if arrayConcatExists {
			if arrayToConcat, err := arrayConcatValue.AsArray(); err == nil {
				resultArray = append(resultArray, arrayToConcat...)
			} else {
				logger.Warn("Second array is not valid, using only first array", nil)
			}
		}

		result = resultArray
		resultLength = len(resultArray)
		debugData["operation"] = "concat"
		debugData["successful"] = true

	case "length":
		// Get the length of the array
		result = float64(len(array))
		resultLength = len(array)
		debugData["operation"] = "length"
		debugData["successful"] = true

	default:
		logger.Error("Invalid operation", map[string]interface{}{"operation": operation})
		debugData["error"] = fmt.Sprintf("Invalid operation: %s", operation)
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid operation: %s", operation)))

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Array Operation Error",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return ctx.ActivateOutputFlow("error")
	}

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, result))
	ctx.SetOutputValue("length", types.NewValue(types.PinTypes.Number, float64(resultLength)))

	debugData["result"] = result
	debugData["resultLength"] = resultLength

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Array operation completed", map[string]interface{}{
		"operation": operation,
		"success":   true,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
