package node

import (
	"webblueprint/internal/types"
)

// BaseNode provides common implementation for all nodes
type BaseNode struct {
	Metadata   NodeMetadata
	Inputs     []types.Pin
	Outputs    []types.Pin
	Properties []types.Property
}

// GetMetadata returns the node's metadata
func (n *BaseNode) GetMetadata() NodeMetadata {
	return n.Metadata
}

// GetInputPins returns the node's input pins
func (n *BaseNode) GetInputPins() []types.Pin {
	return n.Inputs
}

// SetInputPins sets the node's input pins
func (n *BaseNode) SetInputPins(pins []types.Pin) {
	n.Inputs = pins
}

// GetOutputPins returns the node's output pins
func (n *BaseNode) GetOutputPins() []types.Pin {
	return n.Outputs
}

// SetOutputPins sets the node's output pins
func (n *BaseNode) SetOutputPins(pins []types.Pin) {
	n.Outputs = pins
}

// AddOutputPin adds a new output pin to the node
func (n *BaseNode) AddOutputPin(pin types.Pin) {
	n.Outputs = append(n.Outputs, pin)
}

// GetProperties returns the node's properties
func (n *BaseNode) GetProperties() []types.Property {
	return n.Properties
}
