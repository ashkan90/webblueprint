package data

import (
	"fmt"
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
				Description: "Perform operations on objects",
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
					Description: "Object operation: create, get, set, delete, has, keys",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "object",
					Name:        "Object",
					Description: "Object to operate on",
					Type:        types.PinTypes.Object,
					Optional:    true,
				},
				{
					ID:          "key",
					Name:        "Key",
					Description: "Property key",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value to set (for set operation)",
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
					Description: "Result of the object operation",
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
func (n *ObjectNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Object node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get operation value
	operationValue, operationExists := ctx.GetInputValue("operation")
	if !operationExists {
		err := fmt.Errorf("missing required input: operation")
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

	// Process based on operation
	switch operation {
	case "create":
		return n.handleCreateOperation(ctx, debugData)
	case "get":
		return n.handleGetOperation(ctx, debugData)
	case "set":
		return n.handleSetOperation(ctx, debugData)
	case "delete":
		return n.handleDeleteOperation(ctx, debugData)
	case "has":
		return n.handleHasOperation(ctx, debugData)
	case "keys":
		return n.handleKeysOperation(ctx, debugData)
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
}

func (n *ObjectNode) handleCreateOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	// No logger needed in this function since it's a simple operation

	// Create a new empty object
	result := make(map[string]interface{})

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Object, result))

	debugData["operation"] = "create"
	debugData["result"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Create Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ObjectNode) handleGetOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get object and key values
	objectValue, objectExists := ctx.GetInputValue("object")
	if !objectExists {
		err := fmt.Errorf("missing required input: object")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	keyValue, keyExists := ctx.GetInputValue("key")
	if !keyExists {
		err := fmt.Errorf("missing required input: key")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse object and key
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	key, err := keyValue.AsString()
	if err != nil {
		logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get value from object
	value, exists := obj[key]
	if !exists {
		// Return nil if key doesn't exist
		value = nil
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, value))

	debugData["operation"] = "get"
	debugData["object"] = obj
	debugData["key"] = key
	debugData["keyExists"] = exists
	debugData["result"] = value

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Get Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ObjectNode) handleSetOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get object, key, and value
	objectValue, objectExists := ctx.GetInputValue("object")
	if !objectExists {
		err := fmt.Errorf("missing required input: object")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	keyValue, keyExists := ctx.GetInputValue("key")
	if !keyExists {
		err := fmt.Errorf("missing required input: key")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	valueValue, valueExists := ctx.GetInputValue("value")
	if !valueExists {
		err := fmt.Errorf("missing required input: value")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse object and key
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	key, err := keyValue.AsString()
	if err != nil {
		logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Create a new object with the value set
	newObj := make(map[string]interface{})
	for k, v := range obj {
		newObj[k] = v
	}
	newObj[key] = valueValue.RawValue

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Object, newObj))

	debugData["operation"] = "set"
	debugData["object"] = obj
	debugData["key"] = key
	debugData["value"] = valueValue.RawValue
	debugData["result"] = newObj

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Set Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ObjectNode) handleDeleteOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get object and key
	objectValue, objectExists := ctx.GetInputValue("object")
	if !objectExists {
		err := fmt.Errorf("missing required input: object")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	keyValue, keyExists := ctx.GetInputValue("key")
	if !keyExists {
		err := fmt.Errorf("missing required input: key")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse object and key
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	key, err := keyValue.AsString()
	if err != nil {
		logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Create a new object without the key
	newObj := make(map[string]interface{})
	for k, v := range obj {
		if k != key {
			newObj[k] = v
		}
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Object, newObj))

	debugData["operation"] = "delete"
	debugData["object"] = obj
	debugData["key"] = key
	debugData["result"] = newObj

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Delete Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ObjectNode) handleHasOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get object and key
	objectValue, objectExists := ctx.GetInputValue("object")
	if !objectExists {
		err := fmt.Errorf("missing required input: object")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	keyValue, keyExists := ctx.GetInputValue("key")
	if !keyExists {
		err := fmt.Errorf("missing required input: key")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse object and key
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	key, err := keyValue.AsString()
	if err != nil {
		logger.Error("Invalid key", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid key: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Check if key exists in object
	_, exists := obj[key]

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Boolean, exists))

	debugData["operation"] = "has"
	debugData["object"] = obj
	debugData["key"] = key
	debugData["result"] = exists

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Has Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ObjectNode) handleKeysOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get object
	objectValue, objectExists := ctx.GetInputValue("object")
	if !objectExists {
		err := fmt.Errorf("missing required input: object")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse object
	obj, err := objectValue.AsObject()
	if err != nil {
		logger.Error("Invalid object", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid object: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get all keys from object
	keys := make([]interface{}, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Array, keys))

	debugData["operation"] = "keys"
	debugData["object"] = obj
	debugData["result"] = keys

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Object Keys Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
