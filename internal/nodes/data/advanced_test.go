package data_test

import (
	"testing"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/test/mocks"
)

func TestTypeConversionNodeGetters(t *testing.T) {
	node := data.NewTypeConversionNode()

	// Test GetMetadata
	metadata := node.GetMetadata()
	if metadata.TypeID != "type-conversion" {
		t.Errorf("Expected TypeID to be 'type-conversion', got '%s'", metadata.TypeID)
	}

	// Test GetInputPins
	inputPins := node.GetInputPins()
	if len(inputPins) < 2 {
		t.Errorf("Expected at least 2 input pins, got %d", len(inputPins))
	}

	// Test GetOutputPins
	outputPins := node.GetOutputPins()
	if len(outputPins) < 2 {
		t.Errorf("Expected at least 2 output pins, got %d", len(outputPins))
	}
}

func TestConstantNodesCustom(t *testing.T) {
	// Test the string constant node
	t.Run("StringConstant", func(t *testing.T) {
		node := data.NewStringConstantNode()
		logger := mocks.NewMockLogger()
		ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)

		// Execute the node
		err := node.Execute(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check the output
		value, exists := ctx.GetOutputValue("value")
		if !exists {
			t.Errorf("Expected output value to be set")
		} else {
			str, err := value.AsString()
			if err != nil {
				t.Errorf("Expected string output, got %v", value.Type)
			} else if str != "" {
				t.Errorf("Expected empty string by default, got '%s'", str)
			}
		}
	})

	// Test the number constant node
	t.Run("NumberConstant", func(t *testing.T) {
		node := data.NewNumberConstantNode()
		logger := mocks.NewMockLogger()
		ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)

		// Execute the node
		err := node.Execute(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Check the output
		value, exists := ctx.GetOutputValue("value")
		if !exists {
			t.Errorf("Expected output value to be set")
		} else {
			num, err := value.AsNumber()
			if err != nil {
				t.Errorf("Expected number output, got %v", value.Type)
			} else if num != 0 {
				t.Errorf("Expected 0 by default, got %f", num)
			}
		}
	})
}
