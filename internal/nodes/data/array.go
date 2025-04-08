package data

import (
	"fmt"
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
					Description: "Array operation (create, get, set, push, pop, length)",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "array",
					Name:        "Array",
					Description: "Array to operate on",
					Type:        types.PinTypes.Array,
					Optional:    true,
				},
				{
					ID:          "index",
					Name:        "Index",
					Description: "Index for get/set operations",
					Type:        types.PinTypes.Number,
					Optional:    true,
				},
				{
					ID:          "value",
					Name:        "Value",
					Description: "Value for set/push operations",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "size",
					Name:        "Size",
					Description: "Size for create operation",
					Type:        types.PinTypes.Number,
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
					ID:          "popped_item",
					Name:        "Popped Item",
					Description: "Item popped from the array",
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
func (n *ArrayNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Array node", nil)

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
	case "push":
		return n.handlePushOperation(ctx, debugData)
	case "pop":
		return n.handlePopOperation(ctx, debugData)
	case "length":
		return n.handleLengthOperation(ctx, debugData)
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
}

func (n *ArrayNode) handleCreateOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get size value
	sizeValue, sizeExists := ctx.GetInputValue("size")
	if !sizeExists {
		err := fmt.Errorf("missing required input: size")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse size
	size, err := sizeValue.AsNumber()
	if err != nil {
		logger.Error("Invalid size", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid size: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Create array of the specified size
	array := make([]interface{}, int(size))

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Array, array))

	debugData["operation"] = "create"
	debugData["size"] = size
	debugData["result"] = array

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Create Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ArrayNode) handleGetOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get array and index values
	arrayValue, arrayExists := ctx.GetInputValue("array")
	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	indexValue, indexExists := ctx.GetInputValue("index")
	if !indexExists {
		err := fmt.Errorf("missing required input: index")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse array and index
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	index, err := indexValue.AsNumber()
	if err != nil {
		logger.Error("Invalid index", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid index: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Check if index is valid
	intIndex := int(index)
	if intIndex < 0 || intIndex >= len(array) {
		err := fmt.Errorf("index out of bounds: %d", intIndex)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get the element at the specified index
	element := array[intIndex]

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Any, element))

	debugData["operation"] = "get"
	debugData["array"] = array
	debugData["index"] = index
	debugData["result"] = element

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Get Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	ctx.Logger().Debug("Array Get operation succeed", map[string]interface{}{
		"value": index,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ArrayNode) handleSetOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get array, index, and value
	arrayValue, arrayExists := ctx.GetInputValue("array")
	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	indexValue, indexExists := ctx.GetInputValue("index")
	if !indexExists {
		err := fmt.Errorf("missing required input: index")
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

	// Parse array and index
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	index, err := indexValue.AsNumber()
	if err != nil {
		logger.Error("Invalid index", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid index: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Check if index is valid
	intIndex := int(index)
	if intIndex < 0 || intIndex >= len(array) {
		err := fmt.Errorf("index out of bounds: %d", intIndex)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Create a new array with the value set at the specified index
	newArray := make([]interface{}, len(array))
	copy(newArray, array)
	newArray[intIndex] = valueValue.RawValue

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Array, newArray))

	debugData["operation"] = "set"
	debugData["array"] = array
	debugData["index"] = index
	debugData["value"] = valueValue.RawValue
	debugData["result"] = newArray

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Set Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ArrayNode) handlePushOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get array and value
	arrayValue, arrayExists := ctx.GetInputValue("array")
	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
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

	// Parse array
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Create a new array with the value pushed to the end
	newArray := make([]interface{}, len(array)+1)
	copy(newArray, array)
	newArray[len(array)] = valueValue.RawValue

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Array, newArray))

	debugData["operation"] = "push"
	debugData["array"] = array
	debugData["value"] = valueValue.RawValue
	debugData["result"] = newArray

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Push Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ArrayNode) handlePopOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get array
	arrayValue, arrayExists := ctx.GetInputValue("array")
	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse array
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Check if array is not empty
	if len(array) == 0 {
		err := fmt.Errorf("cannot pop from empty array")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get the popped item
	poppedItem := array[len(array)-1]

	// Create a new array without the last element
	newArray := make([]interface{}, len(array)-1)
	copy(newArray, array[:len(array)-1])

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Array, newArray))
	ctx.SetOutputValue("popped_item", types.NewValue(types.PinTypes.Any, poppedItem))

	debugData["operation"] = "pop"
	debugData["array"] = array
	debugData["poppedItem"] = poppedItem
	debugData["result"] = newArray

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Pop Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

func (n *ArrayNode) handleLengthOperation(ctx node.ExecutionContext, debugData map[string]interface{}) error {
	logger := ctx.Logger()

	// Get array
	arrayValue, arrayExists := ctx.GetInputValue("array")
	if !arrayExists {
		err := fmt.Errorf("missing required input: array")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Parse array
	array, err := arrayValue.AsArray()
	if err != nil {
		logger.Error("Invalid array", map[string]interface{}{"error": err.Error()})
		ctx.SetOutputValue("errorMessage", types.NewValue(types.PinTypes.String, "Invalid array: "+err.Error()))
		return ctx.ActivateOutputFlow("error")
	}

	// Get the length of the array
	length := float64(len(array))

	// Set output values
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, length))

	debugData["operation"] = "length"
	debugData["array"] = array
	debugData["result"] = length

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Array Length Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
