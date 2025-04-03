package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// ExecutionEngineExtensions adds context management capabilities to the execution engine
type ExecutionEngineExtensions struct {
	// The core execution engine - using interface{} to avoid import cycle
	Engine interface{}

	// Context management
	ContextManager *ContextManager

	// Dependencies
	ErrorManager         *bperrors.ErrorManager
	RecoveryManager      *bperrors.RecoveryManager
	EventManager         core.EventManagerInterface // Core interface for general use
	ConcreteEventManager *event.EventManager        // Concrete type for specific needs

	// Logger from engine
	logger node.Logger
}

// CreateContext creates an execution context with the appropriate capabilities
func (ext *ExecutionEngineExtensions) CreateContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	// Determine what kind of context to create based on execution mode and other factors
	executionMode := "standard"

	// Try to get execution mode using interface assertion
	if engine, ok := ext.Engine.(interface{ GetExecutionMode() string }); ok {
		executionMode = engine.GetExecutionMode()
	}

	// Create the appropriate context using the context manager
	if executionMode == "actor" {
		return ext.ContextManager.CreateActorContext(
			bp,
			nodeID,
			nodeType,
			blueprintID,
			executionID,
			inputs,
			variables,
			ext.logger,
			hooks,
			activateFlow,
		)
	}

	// For standard mode, determine if event or error handling is required
	return ext.ContextManager.CreateStandardContext(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		ext.logger,
		hooks,
		activateFlow,
	)
}

// CreateEventHandlerContext creates a context for event handlers
func (ext *ExecutionEngineExtensions) CreateEventHandlerContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
	eventHandlerContext *core.EventHandlerContext,
) node.ExecutionContext {
	return ext.ContextManager.CreateEventHandlerContext(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		ext.logger,
		hooks,
		activateFlow,
		eventHandlerContext,
	)
}

func (ext *ExecutionEngineExtensions) CreateFunctionContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	functionID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	return ext.ContextManager.CreateFunctionContext(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		functionID,
		executionID,
		inputs,
		variables,
		ext.logger,
		hooks,
		activateFlow,
	)
}

// GetContextManager returns the context manager
func (ext *ExecutionEngineExtensions) GetContextManager() *ContextManager {
	return ext.ContextManager
}

// GetErrorManager returns the error manager
func (ext *ExecutionEngineExtensions) GetErrorManager() *bperrors.ErrorManager {
	return ext.ErrorManager
}

// GetRecoveryManager returns the recovery manager
func (ext *ExecutionEngineExtensions) GetRecoveryManager() *bperrors.RecoveryManager {
	return ext.RecoveryManager
}

// GetEventManager returns the event manager
func (ext *ExecutionEngineExtensions) GetEventManager() core.EventManagerInterface {
	return ext.EventManager
}

// GetConcreteEventManager returns the concrete event manager instance
func (ext *ExecutionEngineExtensions) GetConcreteEventManager() *event.EventManager {
	return ext.ConcreteEventManager
}

// EventAwareExecutionEngine backwards compatibility type
type EventAwareExecutionEngine struct {
	*ExecutionEngineExtensions
	contextProvider *event.ContextProvider
}

// SetContextProvider sets the context provider
func (e *EventAwareExecutionEngine) SetContextProvider(provider *event.ContextProvider) {
	e.contextProvider = provider
}

// CreateErrorAwareContext creates an error-aware context (backwards compatibility)
func (e *ExecutionEngineExtensions) CreateErrorAwareContext(ctx node.ExecutionContext) *ErrorAwareContext {
	return NewErrorAwareContext(ctx, e.ErrorManager, e.RecoveryManager)
}
