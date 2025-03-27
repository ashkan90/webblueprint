package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/event"
)

type ExecutionEngineExtensions struct {
	// The core execution engine
	engine interface{}

	// Context management
	contextManager *ContextManager

	// Error handling
	errorManager    bperrors.ErrorManager
	recoveryManager bperrors.RecoveryManager

	// Event system
	eventManager event.EventManagerInterface
}

func InitializeExtensions(
	engine interface{},
	contextManager *ContextManager,
	errorManager bperrors.ErrorManager,
	recoveryManager bperrors.RecoveryManager,
	eventManager event.EventManagerInterface,
) *ExecutionEngineExtensions {
	return &ExecutionEngineExtensions{
		engine:          engine,
		contextManager:  contextManager,
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
		eventManager:    eventManager,
	}
}
