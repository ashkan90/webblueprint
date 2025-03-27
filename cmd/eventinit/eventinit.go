package eventinit

import (
	"webblueprint/internal/engine"
	"webblueprint/internal/engineext"
	"webblueprint/internal/integration"
	"webblueprint/internal/nodes/events"
)

// Initialize sets up the entire event system
func Initialize(baseEngine *engine.ExecutionEngine) *engineext.ExecutionEngineExtensions {
	// First, register all event nodes with the global registry
	events.RegisterWithGlobalRegistry()

	// Initialize the event system using our integration package
	extensions := integration.InitializeEventSystem(baseEngine)

	// Register event nodes with the engine
	events.RegisterWithEngine(baseEngine)

	return extensions
}
