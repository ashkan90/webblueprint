package nodes_test

import (
	"strings"
	"testing"
	node2 "webblueprint/internal/node"
	"webblueprint/internal/nodes"
)

// TestNodeFactories verifies that all registered node factories can create nodes
func TestNodeFactories(t *testing.T) {
	for nodeType, factory := range nodes.Core {
		t.Run(nodeType, func(t *testing.T) {
			// Create the node
			node := factory()
			if node == nil {
				t.Errorf("Factory for node type %s returned nil", nodeType)
				return
			}

			// Check that the metadata is properly set
			metadata := node.GetMetadata()
			if metadata.TypeID != nodeType {
				t.Errorf("Node type mismatch: expected %s, got %s", nodeType, metadata.TypeID)
			}

			// Check that the node has input pins (except for constant nodes which may not have input pins)
			isConstant := strings.HasPrefix(nodeType, "constant-")
			inputs := node.GetInputPins()
			if len(inputs) == 0 && !isConstant {
				t.Errorf("Node has no input pins")
			}

			// Check that the node has output pins
			outputs := node.GetOutputPins()
			if len(outputs) == 0 {
				t.Errorf("Node has no output pins")
			}

			// Check that the node implements the Execute method
			if _, ok := node.(interface {
				Execute(ctx node2.ExecutionContext) error
			}); !ok {
				t.Errorf("Node does not implement Execute method")
			}
		})
	}
}
