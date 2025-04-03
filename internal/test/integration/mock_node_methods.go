package integration

import "webblueprint/internal/types"

// SetInputPins sets the input pins for the mock node
func (n *MockNode) SetInputPins(pins []types.Pin) {
	// In the mock node we don't actually need to update anything
	// since the GetInputPins already returns hardcoded pins based on node type
}

// SetOutputPins sets the output pins for the mock node
func (n *MockNode) SetOutputPins(pins []types.Pin) {
	// In the mock node we don't actually need to update anything
	// since the GetOutputPins already returns hardcoded pins based on node type
}

// SetInputPins sets the input pins for the sequence check node
func (n *SequenceCheckNode) SetInputPins(pins []types.Pin) {
	// For testing, we don't need to modify the pins
}

// SetOutputPins sets the output pins for the sequence check node
func (n *SequenceCheckNode) SetOutputPins(pins []types.Pin) {
	// For testing, we don't need to modify the pins
}

// SetInputPins sets the input pins for the recoverable error node
func (n *RecoverableErrorNode) SetInputPins(pins []types.Pin) {
	// For testing, we don't need to modify the pins
}

// SetOutputPins sets the output pins for the recoverable error node
func (n *RecoverableErrorNode) SetOutputPins(pins []types.Pin) {
	// For testing, we don't need to modify the pins
}
