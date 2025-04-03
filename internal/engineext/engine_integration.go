package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/event"
)

// Initializes the ExecutionEngineExtensions defined in context_integration.go

// InitializeExtensions creates a new extension manager
func InitializeExtensions(
	engine interface{},
	contextManager *ContextManager,
	errorManager *bperrors.ErrorManager,
	recoveryManager *bperrors.RecoveryManager,
	// Expect the concrete EventManager
	concreteEventManager *event.EventManager,
) *ExecutionEngineExtensions {
	// Get the core interface adapter from the concrete manager
	eventManagerCoreAdapter := concreteEventManager.AsEventManagerInterface()

	return &ExecutionEngineExtensions{
		Engine:               engine,
		ContextManager:       contextManager,
		ErrorManager:         errorManager,
		RecoveryManager:      recoveryManager,
		EventManager:         eventManagerCoreAdapter, // Store the core interface
		ConcreteEventManager: concreteEventManager,    // Store the concrete manager
	}
}
