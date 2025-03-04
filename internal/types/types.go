package types

import (
	"fmt"
	"strconv"
	"time"
)

// PinType defines the type information for node pins
type PinType struct {
	ID          string                                       // Unique identifier for the type
	Name        string                                       // Human-readable name
	Description string                                       // Description of the type
	Validator   func(value interface{}) error                // Function to validate values of this type
	Converter   func(value interface{}) (interface{}, error) // Function to convert values to this type
}

// Value represents a strongly-typed value that can flow between nodes
type Value struct {
	Type     *PinType
	RawValue interface{}
}

// NewValue creates a new Value with the specified type and raw value
func NewValue(pinType *PinType, rawValue interface{}) Value {
	return Value{
		Type:     pinType,
		RawValue: rawValue,
	}
}

// AsString converts the value to string
func (v Value) AsString() (string, error) {
	if v.Type == PinTypes.String {
		return v.RawValue.(string), nil
	}

	if v.RawValue == nil {
		return "", nil
	}

	if v.Type.Converter != nil {
		conv, err := v.Type.Converter(v.RawValue)
		if err != nil {
			return "", fmt.Errorf("cannot convert to string: %w", err)
		}
		if str, ok := conv.(string); ok {
			return str, nil
		}
	}

	return fmt.Sprintf("%v", v.RawValue), nil
}

// AsNumber converts the value to a float64
func (v Value) AsNumber() (float64, error) {
	if v.Type == PinTypes.Number {
		return v.RawValue.(float64), nil
	}

	if v.RawValue == nil {
		return 0, nil
	}

	if v.Type.Converter != nil {
		conv, err := v.Type.Converter(v.RawValue)
		if err != nil {
			return 0, fmt.Errorf("cannot convert to number: %w", err)
		}
		if num, ok := conv.(float64); ok {
			return num, nil
		}
	}

	// Try to convert directly based on type
	switch val := v.RawValue.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to number: %w", val, err)
		}
		return num, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v.RawValue)
	}
}

// AsBoolean converts the value to a boolean
func (v Value) AsBoolean() (bool, error) {
	if v.Type == PinTypes.Boolean {
		return v.RawValue.(bool), nil
	}

	if v.RawValue == nil {
		return false, nil
	}

	if v.Type.Converter != nil {
		conv, err := v.Type.Converter(v.RawValue)
		if err != nil {
			return false, fmt.Errorf("cannot convert to boolean: %w", err)
		}
		if b, ok := conv.(bool); ok {
			return b, nil
		}
	}

	// Try to convert directly based on type
	switch val := v.RawValue.(type) {
	case bool:
		return val, nil
	case int:
		return val != 0, nil
	case float64:
		return val != 0, nil
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			// Non-boolean strings: empty string is false, anything else is true
			return val != "", nil
		}
		return b, nil
	default:
		// Non-nil values are generally considered true
		return true, nil
	}
}

// AsObject converts the value to a map
func (v Value) AsObject() (map[string]interface{}, error) {
	if v.Type == PinTypes.Object {
		if v.RawValue == nil {
			return make(map[string]interface{}), nil
		}
		return v.RawValue.(map[string]interface{}), nil
	}

	if v.RawValue == nil {
		return make(map[string]interface{}), nil
	}

	if obj, ok := v.RawValue.(map[string]interface{}); ok {
		return obj, nil
	}

	return nil, fmt.Errorf("cannot convert %T to object", v.RawValue)
}

// AsArray converts the value to a slice
func (v Value) AsArray() ([]interface{}, error) {
	if v.Type == PinTypes.Array {
		if v.RawValue == nil {
			return make([]interface{}, 0), nil
		}
		return v.RawValue.([]interface{}), nil
	}

	if v.RawValue == nil {
		return make([]interface{}, 0), nil
	}

	if arr, ok := v.RawValue.([]interface{}); ok {
		return arr, nil
	}

	return nil, fmt.Errorf("cannot convert %T to array", v.RawValue)
}

// Type validation functions
func validateString(value interface{}) error {
	if value == nil {
		return nil
	}

	if _, ok := value.(string); !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	return nil
}

func validateNumber(value interface{}) error {
	if value == nil {
		return nil
	}

	switch value.(type) {
	case int, int64, float32, float64:
		return nil
	default:
		return fmt.Errorf("expected number, got %T", value)
	}
}

func validateBoolean(value interface{}) error {
	if value == nil {
		return nil
	}

	if _, ok := value.(bool); !ok {
		return fmt.Errorf("expected boolean, got %T", value)
	}
	return nil
}

func validateObject(value interface{}) error {
	if value == nil {
		return nil
	}

	if _, ok := value.(map[string]interface{}); !ok {
		return fmt.Errorf("expected object, got %T", value)
	}
	return nil
}

func validateArray(value interface{}) error {
	if value == nil {
		return nil
	}

	if _, ok := value.([]interface{}); !ok {
		return fmt.Errorf("expected array, got %T", value)
	}
	return nil
}

// Type conversion functions
func convertToString(value interface{}) (interface{}, error) {
	if value == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", value), nil
}

func convertToNumber(value interface{}) (interface{}, error) {
	if value == nil {
		return float64(0), nil
	}

	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return float64(0), fmt.Errorf("cannot convert string '%s' to number: %w", v, err)
		}
		return num, nil
	case bool:
		if v {
			return float64(1), nil
		}
		return float64(0), nil
	default:
		return float64(0), fmt.Errorf("cannot convert %T to number", value)
	}
}

func convertToBoolean(value interface{}) (interface{}, error) {
	if value == nil {
		return false, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			// Non-boolean strings: empty string is false, anything else is true
			return v != "", nil
		}
		return b, nil
	default:
		// Non-nil values are generally considered true
		return true, nil
	}
}

// PinTypes defines built-in pin types
var PinTypes = struct {
	Execution *PinType
	String    *PinType
	Number    *PinType
	Boolean   *PinType
	Object    *PinType
	Array     *PinType
	Any       *PinType
}{
	Execution: &PinType{
		ID:          "execution",
		Name:        "Execution",
		Description: "Controls execution flow",
		Validator:   func(v interface{}) error { return nil }, // Execution pins don't carry data
	},
	String: &PinType{
		ID:          "string",
		Name:        "String",
		Description: "Text value",
		Validator:   validateString,
		Converter:   convertToString,
	},
	Number: &PinType{
		ID:          "number",
		Name:        "Number",
		Description: "Numeric value",
		Validator:   validateNumber,
		Converter:   convertToNumber,
	},
	Boolean: &PinType{
		ID:          "boolean",
		Name:        "Boolean",
		Description: "True or false value",
		Validator:   validateBoolean,
		Converter:   convertToBoolean,
	},
	Object: &PinType{
		ID:          "object",
		Name:        "Object",
		Description: "Key-value data structure",
		Validator:   validateObject,
	},
	Array: &PinType{
		ID:          "array",
		Name:        "Array",
		Description: "List of values",
		Validator:   validateArray,
	},
	Any: &PinType{
		ID:          "any",
		Name:        "Any",
		Description: "Any type of value",
		Validator:   func(v interface{}) error { return nil }, // Accept any value
	},
}

// Pin represents an input or output connection point on a node
type Pin struct {
	ID          string      // Unique identifier
	Name        string      // Human-readable name
	Description string      // Description of what the pin does
	Type        *PinType    // Type of data for this pin
	Optional    bool        // Whether this pin is required
	Default     interface{} // Default value if not connected
}

// Property ...
type Property struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// DebugInfo stores debug information during execution
type DebugInfo struct {
	NodeID      string
	PinID       string
	Description string
	Value       interface{}
	Timestamp   time.Time
}

// ValidateConnection checks if a pin can connect to another pin
func (p *Pin) ValidateConnection(targetPin *Pin) error {
	// Execution pins can only connect to execution pins
	if p.Type == PinTypes.Execution && targetPin.Type != PinTypes.Execution {
		return fmt.Errorf("cannot connect execution pin to data pin")
	}

	if p.Type != PinTypes.Execution && targetPin.Type == PinTypes.Execution {
		return fmt.Errorf("cannot connect data pin to execution pin")
	}

	// Any type pins can connect to any data pin
	if p.Type == PinTypes.Any || targetPin.Type == PinTypes.Any {
		return nil
	}

	// Check if the types are compatible
	if p.Type != targetPin.Type {
		// Check if we can convert between the types
		if targetPin.Type.Converter != nil {
			return nil
		}
		return fmt.Errorf("incompatible pin types: %s -> %s", p.Type.Name, targetPin.Type.Name)
	}

	return nil
}
