package engineext

import (
	"fmt"
	"time"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ContextFactory provides utility methods for creating different kinds of contexts
// and for migrating from old context types to new ones
type ContextFactory struct {
	errorManager    *bperrors.ErrorManager
	recoveryManager *bperrors.RecoveryManager
	eventManager    core.EventManagerInterface // Use concrete type
	contextManager  *ContextManager
}

// NewContextFactory creates a new context factory
func NewContextFactory(
	errorManager *bperrors.ErrorManager,
	recoveryManager *bperrors.RecoveryManager,
	eventManager core.EventManagerInterface, // Expect concrete type
) *ContextFactory {
	// Create a context manager, passing the concrete event manager's core interface
	contextManager := NewContextManager(
		errorManager,
		recoveryManager,
		eventManager, // Pass core interface to ContextManager
	)

	return &ContextFactory{
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
		eventManager:    eventManager, // Store concrete type
		contextManager:  contextManager,
	}
}

// MigrateContext migrates an old context to the new context system
func (f *ContextFactory) MigrateContext(oldCtx node.ExecutionContext) (node.ExecutionContext, error) {
	// Check if this is already a new-style context
	if _, ok := oldCtx.(*DefaultExecutionContext); ok {
		return oldCtx, nil
	}

	// Get basic information from the old context
	nodeID := oldCtx.GetNodeID()
	nodeType := oldCtx.GetNodeType()
	blueprintID := oldCtx.GetBlueprintID()
	executionID := oldCtx.GetExecutionID()

	// Create a dummy activate flow function
	activateFlow := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		// Try to call the old context's activate flow method
		return oldCtx.ActivateOutputFlow(pinID)
	}

	// Try to extract variables from the old context
	variables := make(map[string]types.Value)

	// Try to extract inputs from the old context
	inputs := make(map[string]types.Value)

	// Check capabilities of the old context
	capabilities := make(map[string]bool)

	// Check for error handling via type assertion
	if _, ok := oldCtx.(*ErrorAwareContext); ok {
		capabilities["error_handling"] = true
	}

	// Check for event handling via type assertion
	if _, ok := oldCtx.(core.EventAwareContext); ok {
		capabilities["events"] = true
	}

	// Check for actor mode
	if _, ok := oldCtx.(*ActorExecutionContext); ok {
		capabilities["actor"] = true
	}

	// Check for function context
	if _, ok := oldCtx.(*FunctionExecutionContext); ok {
		capabilities["function"] = true
	}

	// Create a new context with the same capabilities
	builder := f.contextManager.CreateContextBuilder(
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		oldCtx.Logger(),
		nil, // We don't have access to hooks
		activateFlow,
	)

	// Add capabilities based on what was detected
	if capabilities["error_handling"] {
		builder.WithErrorHandling(f.errorManager, f.recoveryManager)
	}

	if capabilities["events"] {
		builder.WithEventSupport(f.eventManager, false, nil) // Use adapter
	}

	if capabilities["actor"] {
		builder.WithActorMode()
	}

	// Build the new context
	newCtx := builder.Build()

	// For function contexts, create a specialized function context
	if capabilities["function"] {
		// Extract function ID from context - since we can't access the field directly anymore
		functionID := "unknown"
		// Try to extract it from context data if available

		return f.contextManager.CreateFunctionContext(
			nodeID,
			nodeType,
			blueprintID,
			executionID,
			functionID,
			inputs,
			variables,
			oldCtx.Logger(),
			nil,
			activateFlow,
		), nil
	}

	return newCtx, nil
}

// CreateErrorAwareContext creates a context with error handling capabilities
func (f *ContextFactory) CreateErrorAwareContext(
	baseCtx node.ExecutionContext,
) node.ExecutionContext {
	// Check if the context already has error handling
	if _, ok := baseCtx.(*ErrorAwareContext); ok {
		return baseCtx
	}

	// Add error handling to the context
	return NewErrorAwareContext(
		baseCtx,
		f.errorManager,
		f.recoveryManager,
	)
}

// CreateEventAwareContext creates a context with event handling capabilities
func (f *ContextFactory) CreateEventAwareContext(
	baseCtx node.ExecutionContext,
	isEventHandler bool,
	eventHandlerContext *core.EventHandlerContext,
) node.ExecutionContext {
	// Check if the context already has event handling
	if _, ok := baseCtx.(core.EventAwareContext); ok {
		return baseCtx
	}

	// Add event handling to the context
	return event.NewEventAwareContext(
		baseCtx,
		f.eventManager,
		isEventHandler,
		eventHandlerContext,
	)
}

// CreateActorContext creates a context with event handling capabilities
func (f *ContextFactory) CreateActorContext(
	baseCtx node.ExecutionContext,
) node.ExecutionContext {
	// Check if the context already has event handling
	if _, ok := baseCtx.(core.EventAwareContext); ok {
		return baseCtx
	}

	// Add event handling to the context
	return NewActorExecutionContext(baseCtx.(*DefaultExecutionContext))
}

// CreateFunctionContext creates a context with event handling capabilities
func (f *ContextFactory) CreateFunctionContext(
	baseCtx node.ExecutionContext,
	functionID string,
) node.ExecutionContext {
	// Check if the context already has event handling
	if _, ok := baseCtx.(core.EventAwareContext); ok {
		return baseCtx
	}

	// Add event handling to the context
	return NewFunctionExecutionContext(baseCtx.(*DefaultExecutionContext), functionID)
}

// GetOrCreateErrorManager gets the error manager if the context has error handling,
// or creates a wrapper context with error handling
func (f *ContextFactory) GetOrCreateErrorManager(ctx node.ExecutionContext) (node.ExecutionContext, *bperrors.ErrorManager) {
	// Check if the context already has error handling
	if _, ok := ctx.(*ErrorAwareContext); ok {
		return ctx, f.errorManager
	}

	// Add error handling to the context
	newCtx := NewErrorAwareContext(
		ctx,
		f.errorManager,
		f.recoveryManager,
	)

	return newCtx, f.errorManager
}

// GetOrCreateEventManager gets the event manager if the context has event handling,
// or creates a wrapper context with event handling
func (f *ContextFactory) GetOrCreateEventManager(ctx node.ExecutionContext) (node.ExecutionContext, core.EventManagerInterface) {
	// Check if the context already has event handling
	if eventCtx, ok := ctx.(core.EventAwareContext); ok {
		return ctx, eventCtx.GetEventManager()
	}

	// Add event handling to the context
	// event.NewEventAwareContext now expects *event.EventManager
	newCtx := event.NewEventAwareContext(
		ctx,
		f.eventManager, // Pass the concrete manager
		false,
		nil,
	)

	// Return the new context and the core interface adapter
	return newCtx, f.eventManager
}

// GetContextManager returns the context manager
func (f *ContextFactory) GetContextManager() *ContextManager {
	return f.contextManager
}

// CreateContextFromSettings creates a context with the specified settings
func (f *ContextFactory) CreateContextFromSettings(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	settings map[string]interface{},
) (node.ExecutionContext, error) {
	// Get settings
	useErrorHandling, _ := settings["error_handling"].(bool)
	useEventHandling, _ := settings["event_handling"].(bool)
	useActorMode, _ := settings["actor_mode"].(bool)
	_ = settings["sandbox"]
	_ = settings["user_id"]
	functionID, _ := settings["function_id"].(string)
	isFunction, _ := settings["is_function"].(bool)
	isEventHandler, _ := settings["is_event_handler"].(bool)

	// Create a dummy activate flow function
	activateFlow := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		return nil
	}

	// Create a context builder
	builder := f.contextManager.CreateContextBuilder(
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		nil,
		activateFlow,
	)

	// Add capabilities based on settings
	if useErrorHandling {
		builder.WithErrorHandling(f.errorManager, f.recoveryManager)
	}

	if useEventHandling {
		var eventHandlerContext *core.EventHandlerContext
		if isEventHandler {
			eventHandlerContext = &core.EventHandlerContext{
				EventID:    fmt.Sprintf("event-%s", executionID),
				Parameters: make(map[string]types.Value),
				SourceID:   nodeID,
				Timestamp:  time.Now(),
			}
		}

		builder.WithEventSupport(f.eventManager, isEventHandler, eventHandlerContext) // Use adapter
	}

	if useActorMode {
		builder.WithActorMode()
	}

	// For functions, create a function context
	if isFunction && functionID != "" {
		return builder.WithFunction(functionID).Build(), nil
	}

	// Build a regular context
	return builder.Build(), nil
}
