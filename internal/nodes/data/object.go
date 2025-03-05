package data

import (
	"fmt"
	"strings"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ObjectNode implements a node that provides object operations
type ObjectNode struct {
	node.BaseNode
}

// NewObjectNode creates a new Object operations node
func NewObjectNode() node.Node {
	return &ObjectNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "object-operations",
				Name:        "Object Operations",
				Description: "Perform operations on objects (JavaScript objects/JSON)",
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
					Description: "Object operation: get, set, has, keys, values, entries, merge, create",
					Type:        types.PinTypes.String,
					Default:     "get",
				},
				{
					ID:          "object",
					Name:        "Object",
					Description: "Object to operate on",
					Type:        types.PinTypes.Object,
				},
				{
					ID:          "key",
					Name:        "Key",
					Description: "Property key/path (e.g., 'user.name')",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value to set/add (for set operations)",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "secondObject",
					Name:        "Second Object",
					Description: "Second object for merge operation",
					Type:        types.PinTypes.Object,
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
					Description: "Result of the object operation",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "exists",
					Name:        "Exists",
					Description: "Whether the key exists (for has operation)",
					Type:        types.PinTypes.Boolean,
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
func (n *ObjectNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Object node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	operationValue, operationExists := ctx.GetInputValue("operation")
	objectValue, objectExists := ctx.GetInputValue("object")
	keyValue, keyExists := ctx.GetInputValue("key")
	valueInput, valueExists := ctx.GetInputValue("value")
	secondObjectValue, secondObjectExists := ctx.GetInputValue("secondObject")

	// Check required inputs
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !objectExists {
		err := fmt.Errorf("missing required input: object")
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

	// Parse object
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"operation":       operation,
		"hasKey":          keyExists,
		"hasValue":        valueExists,
		"hasSecondObject": secondObjectExists,
	}

	var result interface{}
	exists := false

	// Process based on operation
	switch strings.ToLower(operation) {
	case "get":
		// Get a property value by key/path
		if !keyExists {
			logger.Error("Missing key for get operation", nil)
			debugData["error"] = "Missing key for get operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing key for get operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Get Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get the key/path
		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid key: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Get Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Handle nested path with dot notation
		if strings.Contains(key, ".") {
			// Split the path
			parts := strings.Split(key, ".")
			current := obj

			// Navigate to the final object
			for i := 0; i < len(parts)-1; i++ {
				part := parts[i]
				if val, ok := current[part]; ok {
					if nextObj, ok := val.(map[string]interface{}); ok {
						current = nextObj
					} else {
						// If the intermediate path is not an object, return nil
						result = nil
						exists = false
						break
					}
				} else {
					// Path doesn't exist
					result = nil
					exists = false
					break
				}
			}

			// Get the final property
			if exists != false { // Only check if we haven't failed already
				lastPart := parts[len(parts)-1]
				if val, ok := current[lastPart]; ok {
					result = val
					exists = true
				} else {
					result = nil
					exists = false
				}
			}
		} else {
			// Simple direct key
			if val, ok := obj[key]; ok {
				result = val
				exists = true
			} else {
				result = nil
				exists = false
			}
		}

		debugData["operation"] = "get"
		debugData["key"] = key
		debugData["exists"] = exists
		debugData["successful"] = true

	case "set":
		// Set a property value by key/path
		if !keyExists {
			logger.Error("Missing key for set operation", nil)
			debugData["error"] = "Missing key for set operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing key for set operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Set Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		if !valueExists {
			logger.Error("Missing value for set operation", nil)
			debugData["error"] = "Missing value for set operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing value for set operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Set Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get the key/path
		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid key: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Set Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Create a copy of the object to modify
		newObj := make(map[string]interface{})
		for k, v := range obj {
			newObj[k] = v
		}

		// Handle nested path with dot notation
		if strings.Contains(key, ".") {
			// Split the path
			parts := strings.Split(key, ".")
			current := newObj

			// Navigate or create intermediate objects
			for i := 0; i < len(parts)-1; i++ {
				part := parts[i]
				if val, ok := current[part]; ok {
					if nextObj, ok := val.(map[string]interface{}); ok {
						current = nextObj
					} else {
						// Replace with a new object
						nextObj = make(map[string]interface{})
						current[part] = nextObj
						current = nextObj
					}
				} else {
					// Create new intermediate object
					nextObj := make(map[string]interface{})
					current[part] = nextObj
					current = nextObj
				}
			}

			// Set the final property
			lastPart := parts[len(parts)-1]
			current[lastPart] = valueInput.RawValue
		} else {
			// Simple direct key
			newObj[key] = valueInput.RawValue
		}

		result = newObj
		exists = true
		debugData["operation"] = "set"
		debugData["key"] = key
		debugData["value"] = valueInput.RawValue
		debugData["successful"] = true

	case "has":
		// Check if a property exists
		if !keyExists {
			logger.Error("Missing key for has operation", nil)
			debugData["error"] = "Missing key for has operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing key for has operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Has Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get the key/path
		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid key: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Has Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Handle nested path with dot notation
		if strings.Contains(key, ".") {
			// Split the path
			parts := strings.Split(key, ".")
			current := obj

			// Navigate through the path
			for i := 0; i < len(parts)-1; i++ {
				part := parts[i]
				if val, ok := current[part]; ok {
					if nextObj, ok := val.(map[string]interface{}); ok {
						current = nextObj
					} else {
						// Path doesn't exist
						exists = false
						break
					}
				} else {
					// Path doesn't exist
					exists = false
					break
				}
			}

			// Check the final property
			if exists != false { // Only check if we haven't failed already
				lastPart := parts[len(parts)-1]
				_, exists = current[lastPart]
			}
		} else {
			// Simple direct key
			_, exists = obj[key]
		}

		result = exists
		debugData["operation"] = "has"
		debugData["key"] = key
		debugData["exists"] = exists
		debugData["successful"] = true

	case "keys":
		// Get all keys of the object
		keys := make([]interface{}, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		result = keys
		exists = len(keys) > 0
		debugData["operation"] = "keys"
		debugData["keyCount"] = len(keys)
		debugData["successful"] = true

	case "values":
		// Get all values of the object
		values := make([]interface{}, 0, len(obj))
		for _, v := range obj {
			values = append(values, v)
		}
		result = values
		exists = len(values) > 0
		debugData["operation"] = "values"
		debugData["valueCount"] = len(values)
		debugData["successful"] = true

	case "entries":
		// Get all key-value pairs of the object
		entries := make([]interface{}, 0, len(obj))
		for k, v := range obj {
			entry := map[string]interface{}{
				"key":   k,
				"value": v,
			}
			entries = append(entries, entry)
		}
		result = entries
		exists = len(entries) > 0
		debugData["operation"] = "entries"
		debugData["entryCount"] = len(entries)
		debugData["successful"] = true

	case "merge":
		// Merge two objects
		if !secondObjectExists {
			logger.Error("Missing second object for merge operation", nil)
			debugData["error"] = "Missing second object for merge operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing second object for merge operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Merge Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Parse second object
		secondObj, err := secondObjectValue.AsObject()
		if err != nil {
			logger.Error("Invalid second object", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid second object: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid second object: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Merge Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Create a new object with properties from both objects
		mergedObj := make(map[string]interface{})

		// Copy properties from first object
		for k, v := range obj {
			mergedObj[k] = v
		}

		// Add or overwrite properties from second object
		for k, v := range secondObj {
			mergedObj[k] = v
		}

		result = mergedObj
		exists = true
		debugData["operation"] = "merge"
		debugData["firstObjectSize"] = len(obj)
		debugData["secondObjectSize"] = len(secondObj)
		debugData["mergedSize"] = len(mergedObj)
		debugData["successful"] = true

	case "create":
		// Create a new object with key-value pair
		if keyExists && valueExists {
			// Create object with a single property
			key, _ := keyValue.AsString()
			newObj := map[string]interface{}{
				key: valueInput.RawValue,
			}
			result = newObj
		} else {
			// Return empty object
			result = make(map[string]interface{})
		}
		exists = true
		debugData["operation"] = "create"
		debugData["successful"] = true

	case "remove":
		// Remove a property from the object
		if !keyExists {
			logger.Error("Missing key for remove operation", nil)
			debugData["error"] = "Missing key for remove operation"
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Missing key for remove operation"))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Remove Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Get the key/path
		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			debugData["error"] = "Invalid key: " + err.Error()
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Object Remove Error",
				Value:       debugData,
				Timestamp:   time.Now(),
			})

			return ctx.ActivateOutputFlow("error")
		}

		// Create a copy of the object
		newObj := make(map[string]interface{})
		for k, v := range obj {
			newObj[k] = v
		}

		// Handle nested path with dot notation
		if strings.Contains(key, ".") {
			// Split the path
			parts := strings.Split(key, ".")
			current := newObj
			success := true

			// Navigate to the containing object
			for i := 0; i < len(parts)-1; i++ {
				part := parts[i]
				if val, ok := current[part]; ok {
					if nextObj, ok := val.(map[string]interface{}); ok {
						current = nextObj
					} else {
						// Path doesn't lead to an object
						success = false
						break
					}
				} else {
					// Path doesn't exist
					success = false
					break
				}
			}

			// Delete the property if path exists
			if success {
				lastPart := parts[len(parts)-1]
				delete(current, lastPart)
				exists = true
			} else {
				exists = false
			}
		} else {
			// Simple direct key
			_, exists = newObj[key]
			delete(newObj, key)
		}

		result = newObj
		debugData["operation"] = "remove"
		debugData["key"] = key
		debugData["keyExisted"] = exists
		debugData["successful"] = true

	case "size":
		// Get the number of properties in the object
		result = float64(len(obj))
		exists = true
		debugData["operation"] = "size"
		debugData["successful"] = true

	default:
		logger.Error("Invalid operation", map[string]interface{}{"operation": operation})
		debugData["error"] = fmt.Sprintf("Invalid operation: %s", operation)
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, fmt.Sprintf("Invalid operation: %s", operation)))

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Object Operation Error",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return ctx.ActivateOutputFlow("error")
	}

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, result))
	ctx.SetOutputValue("exists", types.NewValue(types.PinTypes.Boolean, exists))

	debugData["result"] = result
	debugData["exists"] = exists

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Object operation completed", map[string]interface{}{
		"operation": operation,
		"success":   true,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
