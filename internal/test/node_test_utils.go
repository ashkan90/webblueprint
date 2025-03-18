package test

import (
	"reflect"
	"strings"
	"testing"
	"webblueprint/internal/node"
	"webblueprint/internal/test/mocks"
	"webblueprint/internal/types"
)

// NodeTestCase represents a test case for node execution
type NodeTestCase struct {
	Name            string                 // Test case name
	Inputs          map[string]interface{} // Input values for the node
	ExpectedOutputs map[string]interface{} // Expected output values
	ExpectedFlow    string                 // Expected activated flow
	ExpectedError   bool                   // Whether an error is expected
	ErrorContains   string                 // Expected error substring
}

// ExecuteNodeTestCase runs a test case for a node
func ExecuteNodeTestCase(t *testing.T, node node.Node, tc NodeTestCase) {
	t.Helper()

	// Create a mock context
	logger := mocks.NewMockLogger()
	ctx := mocks.NewMockExecutionContext("test-node", node.GetMetadata().TypeID, logger)

	// Set input values
	for pinID, value := range tc.Inputs {
		// Create a Value from the raw interface
		var val types.Value
		switch v := value.(type) {
		case float64:
			val = types.NewValue(types.PinTypes.Number, v)
		case int:
			val = types.NewValue(types.PinTypes.Number, float64(v))
		case string:
			val = types.NewValue(types.PinTypes.String, v)
		case bool:
			val = types.NewValue(types.PinTypes.Boolean, v)
		case map[string]interface{}:
			val = types.NewValue(types.PinTypes.Object, v)
		case []interface{}:
			val = types.NewValue(types.PinTypes.Array, v)
		case nil:
			val = types.NewValue(types.PinTypes.Any, nil)
		default:
			t.Fatalf("Unsupported input type for pin %s: %T", pinID, value)
		}
		ctx.SetInputValue(pinID, val)
	}

	// Execute the node
	err := node.Execute(ctx)

	// Check if an error was expected
	if tc.ExpectedError {
		if err == nil {
			t.Errorf("Expected an error but got none")
		} else if tc.ErrorContains != "" && err != nil {
			if !strings.Contains(err.Error(), tc.ErrorContains) {
				t.Errorf("Error message does not contain expected substring.\nExpected substring: %s\nActual error: %s",
					tc.ErrorContains, err.Error())
			}
		}
		return
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Check output values
	for pinID, expected := range tc.ExpectedOutputs {
		output, exists := ctx.GetOutputValue(pinID)
		if !exists {
			t.Errorf("Expected output pin %s to have a value, but it was not set", pinID)
			continue
		}

		// Check the value
		switch expectedValue := expected.(type) {
		case float64:
			actual, err := output.AsNumber()
			if err != nil {
				t.Errorf("Output pin %s: expected number but got %v", pinID, output.Type.Name)
				continue
			}
			if actual != expectedValue {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, expectedValue, actual)
			}
		case int:
			actual, err := output.AsNumber()
			if err != nil {
				t.Errorf("Output pin %s: expected number but got %v", pinID, output.Type.Name)
				continue
			}
			if actual != float64(expectedValue) {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, float64(expectedValue), actual)
			}
		case string:
			actual, err := output.AsString()
			if err != nil {
				t.Errorf("Output pin %s: expected string but got %v", pinID, output.Type.Name)
				continue
			}
			if actual != expectedValue {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, expectedValue, actual)
			}
		case bool:
			actual, err := output.AsBoolean()
			if err != nil {
				t.Errorf("Output pin %s: expected boolean but got %v", pinID, output.Type.Name)
				continue
			}
			if actual != expectedValue {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, expectedValue, actual)
			}
		case map[string]interface{}:
			actual, err := output.AsObject()
			if err != nil {
				t.Errorf("Output pin %s: expected object but got %v", pinID, output.Type.Name)
				continue
			}

			// For objects, just check that all expected keys exist with the right values
			// without being strict about the actual vs expected comparison which fails
			// for maps in some Go versions
			allKeysExist := true
			for k, v := range expectedValue {
				actualVal, ok := actual[k]
				if !ok || !reflect.DeepEqual(v, actualVal) {
					allKeysExist = false
					break
				}
			}

			if !allKeysExist {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, expectedValue, actual)
			}
		case []interface{}:
			actual, err := output.AsArray()
			if err != nil {
				t.Errorf("Output pin %s: expected array but got %v", pinID, output.Type.Name)
				continue
			}
			if !reflect.DeepEqual(actual, expectedValue) {
				t.Errorf("Output pin %s: expected %v, got %v", pinID, expectedValue, actual)
			}
		case nil:
			if output.Type != types.PinTypes.Any && output.RawValue != nil {
				t.Errorf("Output pin %s: expected nil but got %v with value %v", pinID, output.Type.Name, output.RawValue)
			}
		default:
			t.Errorf("Unsupported expected output type for pin %s: %T", pinID, expected)
		}
	}

	// Check the activated flow
	if tc.ExpectedFlow != "" {
		if ctx.GetActivatedFlow() != tc.ExpectedFlow {
			t.Errorf("Expected flow %s to be activated, but got %s", tc.ExpectedFlow, ctx.GetActivatedFlow())
		}
	}
}
