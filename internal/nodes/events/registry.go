package events

import (
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
)

// RegisterWithGlobalRegistry registers all event nodes with the global registry
func RegisterWithGlobalRegistry() {
	registry := registry.GetInstance()

	// Register event nodes
	registry.RegisterNodeType("event-definition", NewEventDefinitionNode)
	registry.RegisterNodeType("event-dispatcher", NewEventDispatcherNode)
	registry.RegisterNodeType("improved-event-dispatcher", NewImprovedEventDispatcherNode)
	registry.RegisterNodeType("event-bind", NewEventBindNode)
	registry.RegisterNodeType("event-unbind", NewEventUnbindNode)
	registry.RegisterNodeType("clear-event-bindings", NewClearBindingsNode)
	registry.RegisterNodeType("timer-event", NewTimerEventNode)
	registry.RegisterNodeType("event-with-payload", NewEventWithPayloadNode)

	// Register entry point event nodes
	registry.RegisterNodeType("event-on-created", NewOnCreatedEventNode)
	registry.RegisterNodeType("event-on-tick", NewOnTickEventNode)
	registry.RegisterNodeType("event-on-input", NewOnInputEventNode)

	// Register the entry point nodes using our helper function
	RegisterEventEntryPointNodes()
}

// RegisterWithEngine registers all event nodes with the execution engine
func RegisterWithEngine(engine interface{}) {
	// Use type assertion to register nodes with the engine
	if engineRegistry, ok := engine.(node.NodeTypeRegistry); ok {
		for typeID, factory := range GetNodeTypes() {
			engineRegistry.RegisterNodeType(typeID, factory)
		}
	}
}

// GetNodeTypes returns all event node types
func GetNodeTypes() map[string]node.NodeFactory {
	return map[string]node.NodeFactory{
		// Event management nodes
		"event-definition":          NewEventDefinitionNode,
		"event-dispatcher":          NewEventDispatcherNode,
		"improved-event-dispatcher": NewImprovedEventDispatcherNode,
		"event-bind":                NewEventBindNode,
		"event-unbind":              NewEventUnbindNode,
		"clear-event-bindings":      NewClearBindingsNode,
		"timer-event":               NewTimerEventNode,
		"event-with-payload":        NewEventWithPayloadNode,

		// Entry point event nodes
		"event-on-created": NewOnCreatedEventNode,
		"event-on-tick":    NewOnTickEventNode,
		"event-on-input":   NewOnInputEventNode,
	}
}
