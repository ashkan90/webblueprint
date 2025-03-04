package math

import (
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// AddNode implements a node that adds two numbers
type AddNode struct {
	node.BaseNode
}

// NewAddNode creates a new Add node
func NewAddNode() node.Node {
	return &AddNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "math-add",
				Name:        "Add",
				Description: "Adds two numbers",
				Category:    "Math",
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
					ID:          "a",
					Name:        "A",
					Description: "First number",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "b",
					Name:        "B",
					Description: "Second number",
					Type:        types.PinTypes.Number,
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
					Description: "Sum of A and B",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *AddNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Add node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	aValue, aExists := ctx.GetInputValue("a")
	bValue, bExists := ctx.GetInputValue("b")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"aExists": aExists,
		"bExists": bExists,
	}

	if !aExists {
		err := fmt.Errorf("missing required input: a")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing input A",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	if !bExists {
		err := fmt.Errorf("missing required input: b")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing input B",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Convert to numbers
	a, err := aValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input A", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input A: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input A",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	b, err := bValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input B", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input B: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input B",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Perform addition
	result := a + b

	// Update debug data with actual values
	debugData["inputs"] = map[string]interface{}{
		"a": a,
		"b": b,
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

	debugData["output"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Add Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Add result", map[string]interface{}{
		"a":      a,
		"b":      b,
		"result": result,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// SubtractNode implements a node that subtracts one number from another
type SubtractNode struct {
	node.BaseNode
}

// NewSubtractNode creates a new Subtract node
func NewSubtractNode() node.Node {
	return &SubtractNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "math-subtract",
				Name:        "Subtract",
				Description: "Subtracts B from A",
				Category:    "Math",
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
					ID:          "a",
					Name:        "A",
					Description: "Number to subtract from",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "b",
					Name:        "B",
					Description: "Number to subtract",
					Type:        types.PinTypes.Number,
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
					Description: "A - B",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *SubtractNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Subtract node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	aValue, aExists := ctx.GetInputValue("a")
	bValue, bExists := ctx.GetInputValue("b")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"aExists": aExists,
		"bExists": bExists,
	}

	if !aExists || !bExists {
		errorMsg := ""
		if !aExists {
			errorMsg = "missing required input: a"
		} else {
			errorMsg = "missing required input: b"
		}

		err := fmt.Errorf(errorMsg)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing input",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Convert to numbers
	a, err := aValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input A", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input A: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input A",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	b, err := bValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input B", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input B: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input B",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Perform subtraction
	result := a - b

	// Update debug data with actual values
	debugData["inputs"] = map[string]interface{}{
		"a": a,
		"b": b,
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

	debugData["output"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Subtract Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Subtraction result", map[string]interface{}{
		"a":      a,
		"b":      b,
		"result": result,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// MultiplyNode implements a node that multiplies two numbers
type MultiplyNode struct {
	node.BaseNode
}

// NewMultiplyNode creates a new Multiply node
func NewMultiplyNode() node.Node {
	return &MultiplyNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "math-multiply",
				Name:        "Multiply",
				Description: "Multiplies two numbers",
				Category:    "Math",
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
					ID:          "a",
					Name:        "A",
					Description: "First number",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "b",
					Name:        "B",
					Description: "Second number",
					Type:        types.PinTypes.Number,
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
					Description: "A × B",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *MultiplyNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Multiply node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	aValue, aExists := ctx.GetInputValue("a")
	bValue, bExists := ctx.GetInputValue("b")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"aExists": aExists,
		"bExists": bExists,
	}

	if !aExists || !bExists {
		errorMsg := ""
		if !aExists {
			errorMsg = "missing required input: a"
		} else {
			errorMsg = "missing required input: b"
		}

		err := fmt.Errorf(errorMsg)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing input",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Convert to numbers
	a, err := aValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input A", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input A: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input A",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	b, err := bValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input B", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input B: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input B",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Perform multiplication
	result := a * b

	// Update debug data with actual values
	debugData["inputs"] = map[string]interface{}{
		"a": a,
		"b": b,
	}

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

	debugData["output"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Multiply Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Multiplication result", map[string]interface{}{
		"a":      a,
		"b":      b,
		"result": result,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}

// DivideNode implements a node that divides one number by another
type DivideNode struct {
	node.BaseNode
}

// NewDivideNode creates a new Divide node
func NewDivideNode() node.Node {
	return &DivideNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "math-divide",
				Name:        "Divide",
				Description: "Divides A by B",
				Category:    "Math",
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
					ID:          "a",
					Name:        "A",
					Description: "Dividend (number to divide)",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "b",
					Name:        "B",
					Description: "Divisor (number to divide by)",
					Type:        types.PinTypes.Number,
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
					Description: "A ÷ B",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "error",
					Name:        "Error",
					Description: "Execution on error (e.g., division by zero)",
					Type:        types.PinTypes.Execution,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *DivideNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Divide node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	aValue, aExists := ctx.GetInputValue("a")
	bValue, bExists := ctx.GetInputValue("b")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"aExists": aExists,
		"bExists": bExists,
	}

	if !aExists || !bExists {
		errorMsg := ""
		if !aExists {
			errorMsg = "missing required input: a"
		} else {
			errorMsg = "missing required input: b"
		}

		err := fmt.Errorf(errorMsg)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing input",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Convert to numbers
	a, err := aValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input A", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input A: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input A",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	b, err := bValue.AsNumber()
	if err != nil {
		logger.Error("Invalid input B", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "invalid_input",
			"message": "Invalid input B: " + err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Invalid input B",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		return err
	}

	// Update debug data with actual values
	debugData["inputs"] = map[string]interface{}{
		"a": a,
		"b": b,
	}

	// Check for division by zero
	if b == 0 {
		err := fmt.Errorf("division by zero")
		logger.Error("Division by zero", nil)

		debugData["error"] = map[string]string{
			"type":    "division_by_zero",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Division by zero",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		// Set error message as output
		ctx.SetOutputValue("result", types.NewValue(types.PinTypes.String, "Division by zero"))

		// Activate error flow
		return ctx.ActivateOutputFlow("error")
	}

	// Perform division
	result := a / b

	// Set output value
	ctx.SetOutputValue("result", types.NewValue(types.PinTypes.Number, result))

	debugData["output"] = result

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Divide Operation",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	logger.Info("Division result", map[string]interface{}{
		"a":      a,
		"b":      b,
		"result": result,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
