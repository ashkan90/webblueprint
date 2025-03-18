package logic

import (
	"fmt"
	"reflect"
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

	// Log types for debugging
	debugData["valueType"] = reflect.TypeOf(valueInput.RawValue)

	if case1Exists {
		debugData["case1Type"] = reflect.TypeOf(case1Value.RawValue)
	}
	if case2Exists {
		debugData["case2Type"] = reflect.TypeOf(case2Value.RawValue)
	}
	if case3Exists {
		debugData["case3Type"] = reflect.TypeOf(case3Value.RawValue)
	}
	if case4Exists {
		debugData["case4Type"] = reflect.TypeOf(case4Value.RawValue)
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

	// Get type information
	typeA := reflect.TypeOf(valueA)
	typeB := reflect.TypeOf(valueB)

	// If types don't match, we need to be more careful about comparisons
	if typeA != typeB {
		// Special case for numeric types - they can be compared if both are numeric
		switch valueA.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// A is numeric, check if B is also numeric
			switch valueB.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				// Both are numeric types, convert to float64 for comparison
				var numA, numB float64

				switch a := valueA.(type) {
				case int:
					numA = float64(a)
				case int8:
					numA = float64(a)
				case int16:
					numA = float64(a)
				case int32:
					numA = float64(a)
				case int64:
					numA = float64(a)
				case uint:
					numA = float64(a)
				case uint8:
					numA = float64(a)
				case uint16:
					numA = float64(a)
				case uint32:
					numA = float64(a)
				case uint64:
					numA = float64(a)
				case float32:
					numA = float64(a)
				case float64:
					numA = a
				}

				switch b := valueB.(type) {
				case int:
					numB = float64(b)
				case int8:
					numB = float64(b)
				case int16:
					numB = float64(b)
				case int32:
					numB = float64(b)
				case int64:
					numB = float64(b)
				case uint:
					numB = float64(b)
				case uint8:
					numB = float64(b)
				case uint16:
					numB = float64(b)
				case uint32:
					numB = float64(b)
				case uint64:
					numB = float64(b)
				case float32:
					numB = float64(b)
				case float64:
					numB = b
				}

				return numA == numB
			}
		}

		// For non-numeric types with different types, we should generally return false
		// This is especially important for mixed type comparison test case
		return false
	}

	// At this point, we know both values are of the same type
	switch a := valueA.(type) {
	case string:
		b := valueB.(string)
		return a == b
	case float64:
		b := valueB.(float64)
		return a == b
	case bool:
		b := valueB.(bool)
		return a == b
	case int:
		b := valueB.(int)
		return a == b
	case map[string]interface{}:
		// Deep comparison for objects is complex
		// For now, we'll do a simple string representation comparison
		b := valueB.(map[string]interface{})
		if len(a) != len(b) {
			return false
		}
		// Simple check: if keys and values match
		for k, aVal := range a {
			if bVal, exists := b[k]; !exists || !compareValues(aVal, bVal) {
				return false
			}
		}
		return true
	case []interface{}:
		// Deep comparison for arrays
		b := valueB.([]interface{})
		if len(a) != len(b) {
			return false
		}
		// Check if each element matches
		for i, aVal := range a {
			if !compareValues(aVal, b[i]) {
				return false
			}
		}
		return true
	}

	// For any other types, use string representation as a fallback
	// But we should be more cautious and avoid matching different types
	return fmt.Sprintf("%v", valueA) == fmt.Sprintf("%v", valueB)
}
