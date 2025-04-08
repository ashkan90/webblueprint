package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/repository" // Added import for repository
)

// ContextManager provides a centralized way to create and manage execution contexts
type ContextManager struct {
	// Global dependencies
	errorManager    *bperrors.ErrorManager
	recoveryManager *bperrors.RecoveryManager
	eventManager    core.EventManagerInterface
	repoFactory     repository.RepositoryFactory // Added field
}

// NewContextManager creates a new context manager
func NewContextManager(
	errorManager *bperrors.ErrorManager,
	recoveryManager *bperrors.RecoveryManager,
	eventManager core.EventManagerInterface,
	repoFactory repository.RepositoryFactory, // Added parameter
) *ContextManager {
	return &ContextManager{
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
		eventManager:    eventManager,
		repoFactory:     repoFactory, // Added assignment
	}
}

// CreateContextBuilder creates a new context builder with default settings
func (cm *ContextManager) CreateContextBuilder(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) *ContextBuilder {
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory, // Pass repoFactory
	)
}

// CreateStandardContext creates a context with the most common settings
func (cm *ContextManager) CreateStandardContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	// Create a context with common settings
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory, // Pass repoFactory
	).
		WithErrorHandling(cm.errorManager, cm.recoveryManager).
		WithEventSupport(cm.eventManager, false, nil).
		Build()
}

// CreateActorContext creates a context optimized for actor-based execution
func (cm *ContextManager) CreateActorContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	// Create a context optimized for actor-based execution
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory, // Pass repoFactory
	).
		WithErrorHandling(cm.errorManager, cm.recoveryManager).
		WithEventSupport(cm.eventManager, true, nil).
		WithActorMode().
		Build()
}

func (cm *ContextManager) CreateLoopContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory,
	).
		WithErrorHandling(cm.errorManager, cm.recoveryManager).
		WithEventSupport(cm.eventManager, true, nil).
		WithLoopSupport().
		Build()
}

// CreateEventHandlerContext creates a context for event handlers
func (cm *ContextManager) CreateEventHandlerContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
	eventHandlerContext *core.EventHandlerContext,
) node.ExecutionContext {
	// Create a context for event handlers
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory, // Pass repoFactory
	).
		WithErrorHandling(cm.errorManager, cm.recoveryManager).
		WithEventSupport(cm.eventManager, true, eventHandlerContext).
		Build()
}

func (cm *ContextManager) CreateFunctionContext(
	bp *blueprint.Blueprint,
	nodeID string,
	nodeType string,
	blueprintID string,
	functionID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
) node.ExecutionContext {
	return NewContextBuilder(
		bp,
		nodeID,
		nodeType,
		blueprintID,
		executionID,
		inputs,
		variables,
		logger,
		hooks,
		activateFlow,
		cm.repoFactory, // Pass repoFactory
	).
		WithErrorHandling(cm.errorManager, cm.recoveryManager).
		WithEventSupport(cm.eventManager, false, nil).
		WithFunction(functionID).
		Build()
}

// GetErrorManager returns the error manager
func (cm *ContextManager) GetErrorManager() *bperrors.ErrorManager {
	return cm.errorManager
}

// GetRecoveryManager returns the recovery manager
func (cm *ContextManager) GetRecoveryManager() *bperrors.RecoveryManager {
	return cm.recoveryManager
}

// GetEventManager returns the event manager
func (cm *ContextManager) GetEventManager() core.EventManagerInterface {
	return cm.eventManager
}

// GetRepoFactory returns the repository factory
func (cm *ContextManager) GetRepoFactory() repository.RepositoryFactory {
	return cm.repoFactory
}
