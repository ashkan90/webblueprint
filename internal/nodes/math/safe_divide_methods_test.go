package math_test

import (
	"reflect"
	"testing"
	"webblueprint/internal/nodes/math"
)

func TestSafeDivideNodeGetters(t *testing.T) {
	node := math.NewSafeDivideNode()

	// Test GetMetadata
	metadata := node.GetMetadata()
	if metadata.TypeID != "safe-divide" {
		t.Errorf("Expected TypeID to be 'safe-divide', got '%s'", metadata.TypeID)
	}

	// Test GetInputPins
	inputPins := node.GetInputPins()
	if len(inputPins) != 4 {
		t.Errorf("Expected 4 input pins, got %d", len(inputPins))
	}

	// Verify essential input pins exist
	inputPinIDs := make(map[string]bool)
	for _, pin := range inputPins {
		inputPinIDs[pin.ID] = true
	}
	requiredPins := []string{"exec", "dividend", "divisor", "default"}
	for _, id := range requiredPins {
		if !inputPinIDs[id] {
			t.Errorf("Input pin '%s' is missing", id)
		}
	}

	// Test GetOutputPins
	outputPins := node.GetOutputPins()
	if len(outputPins) != 4 {
		t.Errorf("Expected 4 output pins, got %d", len(outputPins))
	}

	// Verify essential output pins exist
	outputPinIDs := make(map[string]bool)
	for _, pin := range outputPins {
		outputPinIDs[pin.ID] = true
	}
	requiredOutputs := []string{"then", "catch", "result", "error"}
	for _, id := range requiredOutputs {
		if !outputPinIDs[id] {
			t.Errorf("Output pin '%s' is missing", id)
		}
	}

	// Test GetProperties
	properties := node.GetProperties()
	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}

	// Check property names and default values
	propMap := make(map[string]interface{})
	for _, prop := range properties {
		propMap[prop.Name] = prop.Value
	}

	if _, exists := propMap["defaultValue"]; !exists {
		t.Errorf("Property 'defaultValue' is missing")
	}

	if mode, exists := propMap["errorHandlingMode"]; !exists {
		t.Errorf("Property 'errorHandlingMode' is missing")
	} else if !reflect.DeepEqual(mode, "auto") {
		t.Errorf("Expected 'errorHandlingMode' default to be 'auto', got '%v'", mode)
	}
}
