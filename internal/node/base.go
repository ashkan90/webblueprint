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

// GetOutputPins returns the node's output pins
func (n *BaseNode) GetOutputPins() []types.Pin {
	return n.Outputs
}

func (n *BaseNode) GetProperties() []types.Property {
	return n.Properties
}
