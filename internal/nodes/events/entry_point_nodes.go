package events

import (
	"webblueprint/internal/registry"
)

// RegisterEventEntryPointNodes registers all the entry point event nodes with the global registry
func RegisterEventEntryPointNodes() {
	reg := registry.GetInstance()

	// Register the entry point event nodes
	reg.RegisterNodeType("event-on-created", NewOnCreatedEventNode)
	reg.RegisterNodeType("event-on-tick", NewOnTickEventNode)
	reg.RegisterNodeType("event-on-input", NewOnInputEventNode)

	// The EventDispatcherNode is already registered elsewhere, so we don't register it here
}
