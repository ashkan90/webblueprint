package events

import (
	"webblueprint/internal/registry"
)

// RegisterEntryPointNodes registers all entry point event nodes with the global registry
func RegisterEntryPointNodes() {
	registry := registry.GetInstance()

	// Register entry point event nodes
	registry.RegisterNodeType("event-on-created", NewOnCreatedEventNode)
	registry.RegisterNodeType("event-on-tick", NewOnTickEventNode)
	registry.RegisterNodeType("event-on-input", NewOnInputEventNode)
}
