package logic

import (
	"fmt"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// BranchNode implements a multi-way branch node (similar to a switch statement)
type BranchNode struct {
	node.BaseNode
}

// NewBranchNode creates a new Branch node
func NewBranchNode() node.Node {
	return &BranchNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "branch",
				Name:        "Branch",
				Description: "Routes execution based on a value (similar to switch/case)",
				Category:    "Logic",
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
					ID:          "value",
					Name:        "Value",
					Description: "Value to branch on",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "case1",
					Name:        "Case 1",
					Description: "First case value to compare",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "case2",
					Name:        "Case 2",
					Description: "Second case value to compare",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "case3",
					Name:        "Case 3",
					Description: "Third case value to compare",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
				{
					ID:          "case4",
					Name:        "Case 4",
					Description: "Fourth case value to compare",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "case1_out",
					Name:        "Case 1",
					Description: "Executed if value matches Case 1",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "case2_out",
					Name:        "Case 2",
					Description: "Executed if value matches Case 2",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "case3_out",
					Name:        "Case 3",
					Description: "Executed if value matches Case 3",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "case4_out",
					Name:        "Case 4",
					Description: "Executed if value matches Case 4",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "default",
					Name:        "Default",
					Description: "Executed if no cases match",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "matched_case",
					Name:        "Matched Case",
					Description: "The case that matched (1-4 or 0 for default)",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *BranchNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Branch node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get the value to switch on
	valueInput, valueExists := ctx.GetInputValue("value")
	if !valueExists {
		err := fmt.Errorf("missing required input: value")
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = map[string]string{
			"type":    "missing_input",
			"message": err.Error(),
		}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Error: Missing value input",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		// Set default output for error case
		ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 0.0))
		return ctx.ActivateOutputFlow("default")
	}

	// Get case values
	case1Value, case1Exists := ctx.GetInputValue("case1")
	case2Value, case2Exists := ctx.GetInputValue("case2")
	case3Value, case3Exists := ctx.GetInputValue("case3")
	case4Value, case4Exists := ctx.GetInputValue("case4")

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"value":    valueInput.RawValue,
		"hasCase1": case1Exists,
		"hasCase2": case2Exists,
		"hasCase3": case3Exists,
		"hasCase4": case4Exists,
	}

	// Try to compare values
	// Case 1
	if case1Exists && compareValues(valueInput.RawValue, case1Value.RawValue) {
		logger.Info("Value matched Case 1", map[string]interface{}{
			"value": valueInput.RawValue,
			"case":  case1Value.RawValue,
		})

		debugData["match"] = 1
		debugData["matchedValue"] = case1Value.RawValue

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Branch: Case 1 Matched",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 1.0))
		return ctx.ActivateOutputFlow("case1_out")
	}

	// Case 2
	if case2Exists && compareValues(valueInput.RawValue, case2Value.RawValue) {
		logger.Info("Value matched Case 2", map[string]interface{}{
			"value": valueInput.RawValue,
			"case":  case2Value.RawValue,
		})

		debugData["match"] = 2
		debugData["matchedValue"] = case2Value.RawValue

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Branch: Case 2 Matched",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 2.0))
		return ctx.ActivateOutputFlow("case2_out")
	}

	// Case 3
	if case3Exists && compareValues(valueInput.RawValue, case3Value.RawValue) {
		logger.Info("Value matched Case 3", map[string]interface{}{
			"value": valueInput.RawValue,
			"case":  case3Value.RawValue,
		})

		debugData["match"] = 3
		debugData["matchedValue"] = case3Value.RawValue

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Branch: Case 3 Matched",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 3.0))
		return ctx.ActivateOutputFlow("case3_out")
	}

	// Case 4
	if case4Exists && compareValues(valueInput.RawValue, case4Value.RawValue) {
		logger.Info("Value matched Case 4", map[string]interface{}{
			"value": valueInput.RawValue,
			"case":  case4Value.RawValue,
		})

		debugData["match"] = 4
		debugData["matchedValue"] = case4Value.RawValue

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Branch: Case 4 Matched",
			Value:       debugData,
			Timestamp:   time.Now(),
		})

		ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 4.0))
		return ctx.ActivateOutputFlow("case4_out")
	}

	// No matches, use default
	logger.Info("No cases matched, taking default path", map[string]interface{}{
		"value": valueInput.RawValue,
	})

	debugData["match"] = 0
	debugData["matchedValue"] = nil

	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Branch: No Match (Default)",
		Value:       debugData,
		Timestamp:   time.Now(),
	})

	ctx.SetOutputValue("matched_case", types.NewValue(types.PinTypes.Number, 0.0))
	return ctx.ActivateOutputFlow("default")
}

// Helper function to compare values of potentially different types
func compareValues(valueA, valueB interface{}) bool {
	// Handle nil cases
	if valueA == nil && valueB == nil {
		return true
	}
	if valueA == nil || valueB == nil {
		return false
	}

	// Check if types match
	switch a := valueA.(type) {
	case string:
		if b, ok := valueB.(string); ok {
			return a == b
		}
	case float64:
		if b, ok := valueB.(float64); ok {
			return a == b
		}
		// Special case for integers
		if b, ok := valueB.(int); ok {
			return a == float64(b)
		}
	case bool:
		if b, ok := valueB.(bool); ok {
			return a == b
		}
	case map[string]interface{}:
		// For objects, we'll just check equality of string representations
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", valueB)
	case []interface{}:
		// For arrays, we'll just check equality of string representations
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", valueB)
	}

	// Fallback to string comparison
	return fmt.Sprintf("%v", valueA) == fmt.Sprintf("%v", valueB)
}
