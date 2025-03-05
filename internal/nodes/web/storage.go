package web

import (
	"encoding/json"
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// StorageNode implements a node for working with browser storage (localStorage/sessionStorage)
type StorageNode struct {
	node.BaseNode
}

// NewStorageNode creates a new Storage node
func NewStorageNode() node.Node {
	return &StorageNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "storage",
				Name:        "Storage",
				Description: "Work with browser local and session storage",
				Category:    "Web",
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
					Description: "Storage operation: get, set, remove, clear",
					Type:        types.PinTypes.String,
					Default:     "get",
				},
				{
					ID:          "storageType",
					Name:        "Storage Type",
					Description: "Type of storage: local, session",
					Type:        types.PinTypes.String,
					Default:     "local",
				},
				{
					ID:          "key",
					Name:        "Key",
					Description: "Storage key (for get/set/remove operations)",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value to store (for set operation)",
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
				{
					ID:          "error",
					Name:        "Error",
					Description: "Executed if an error occurs",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "result",
					Name:        "Result",
					Description: "Retrieved value or operation result",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "exists",
					Name:        "Exists",
					Description: "Whether the key exists (for get operation)",
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
func (n *StorageNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Storage node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	operationValue, operationExists := ctx.GetInputValue("operation")
	storageTypeValue, storageTypeExists := ctx.GetInputValue("storageType")
	keyValue, keyExists := ctx.GetInputValue("key")
	valueInput, valueExists := ctx.GetInputValue("value")

	// Check required inputs
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	if !storageTypeExists {
		err := fmt.Errorf("missing required input: storageType")
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

	// Parse storage type
	storageType, err := storageTypeValue.AsString()
	if err != nil {
		logger.Error("Invalid storage type", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid storage type: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Validate storage type
	if storageType != "local" && storageType != "session" {
		err := fmt.Errorf("invalid storage type: %s (must be 'local' or 'session')", storageType)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"operation":   operation,
		"storageType": storageType,
		"hasKey":      keyExists,
		"hasValue":    valueExists,
	}

	// Create a storage operation object to be sent to the client
	storageOperation := map[string]interface{}{
		"operation":   operation,
		"storageType": storageType,
		"nodeId":      ctx.GetNodeID(),
		"executionId": ctx.GetExecutionID(),
		"timestamp":   time.Now().UnixNano() / 1e6,
	}

	// Add operation-specific properties
	switch operation {
	case "get":
		if !keyExists {
			err := fmt.Errorf("missing required input: key (for get operation)")
			logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		storageOperation["key"] = key
		debugData["key"] = key

		// Since we don't have actual browser storage access in the backend,
		// we'll just simulate the operation result

		// For debugging/testing purposes, we'll use a special key format to simulate storage:
		// "test:value" -> returns the string "value"
		// "test:json:value" -> returns the object {value: "value"}
		// "test:null" -> returns null
		// "test:missing" -> simulates a missing key

		var value interface{}
		exists := false

		if key == "test:missing" {
			value = nil
			exists = false
		} else if key == "test:null" {
			value = nil
			exists = true
		} else if key == "test:true" {
			value = true
			exists = true
		} else if key == "test:false" {
			value = false
			exists = true
		} else if key == "test:number" {
			value = 42.5
			exists = true
		} else if key == "test:string" {
			value = "test value"
			exists = true
		} else if key == "test:array" {
			value = []interface{}{1, 2, 3, "test"}
			exists = true
		} else if key == "test:object" {
			value = map[string]interface{}{
				"name": "Test Object",
				"properties": map[string]interface{}{
					"value": 123,
					"flag":  true,
				},
			}
			exists = true
		} else {
			// For any other key, extract value from the key itself if it starts with "test:"
			if len(key) > 5 && key[:5] == "test:" {
				if len(key) > 10 && key[:10] == "test:json:" {
					// Parse as JSON object with a value property
					value = map[string]interface{}{
						"value": key[10:],
					}
				} else {
					// Just return the part after "test:"
					value = key[5:]
				}
				exists = true
			} else {
				// For regular keys, just return a dummy value in production this would be the actual stored value
				dummyVal := fmt.Sprintf("Stored value for %s", key)
				value = dummyVal
				exists = false // In a real implementation, this would depend on whether the key exists
			}
		}

		// Set output values
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, value))
		ctx.SetOutputValue("exists", types.NewValue(types.PinTypes.Boolean, exists))

		debugData["value"] = value
		debugData["exists"] = exists

		logger.Info("Get storage value", map[string]interface{}{
			"storageType": storageType,
			"key":         key,
			"exists":      exists,
		})

	case "set":
		if !keyExists {
			err := fmt.Errorf("missing required input: key (for set operation)")
			logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		if !valueExists {
			err := fmt.Errorf("missing required input: value (for set operation)")
			logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		// For objects and non-primitive values, convert to JSON string
		var valueForStorage interface{}
		valueForStorage = valueInput.RawValue

		// Check if value is object or array
		isComplex := false
		switch valueInput.RawValue.(type) {
		case map[string]interface{}, []interface{}:
			isComplex = true
			// In real browser storage, complex values need to be JSON stringified
			jsonBytes, err := json.Marshal(valueInput.RawValue)
			if err != nil {
				logger.Error("Failed to marshal value to JSON", map[string]interface{}{"error": err.Error()})
				ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Failed to marshal value to JSON: "+err.Error()))
				return ctx.ActivateOutputFlow("error")
			}
			valueForStorage = string(jsonBytes)
		}

		storageOperation["key"] = key
		storageOperation["value"] = valueForStorage
		storageOperation["isComplex"] = isComplex

		debugData["key"] = key
		debugData["value"] = valueInput.RawValue
		debugData["valueForStorage"] = valueForStorage
		debugData["isComplex"] = isComplex

		// Set output values (same as what we stored)
		ctx.SetOutputValue("result", valueInput)
		ctx.SetOutputValue("exists", types.NewValue(types.PinTypes.Boolean, true))

		logger.Info("Set storage value", map[string]interface{}{
			"storageType": storageType,
			"key":         key,
			"isComplex":   isComplex,
		})

	case "remove":
		if !keyExists {
			err := fmt.Errorf("missing required input: key (for remove operation)")
			logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		key, err := keyValue.AsString()
		if err != nil {
			logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
			ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
			return ctx.ActivateOutputFlow("error")
		}

		storageOperation["key"] = key
		debugData["key"] = key

		// Set output values
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, true))
		ctx.SetOutputValue("exists", types.NewValue(types.PinTypes.Boolean, false))

		logger.Info("Remove storage key", map[string]interface{}{
			"storageType": storageType,
			"key":         key,
		})

	case "clear":
		// No additional properties needed

		// Set output values
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, true))
		ctx.SetOutputValue("exists", types.NewValue(types.PinTypes.Boolean, false))

		logger.Info("Clear storage", map[string]interface{}{
			"storageType": storageType,
		})

	default:
		err := fmt.Errorf("invalid operation: %s (must be 'get', 'set', 'remove', or 'clear')", operation)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Result includes the operation that would be performed on the client
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Object, storageOperation))

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: fmt.Sprintf("Storage %s Operation", operation),
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
