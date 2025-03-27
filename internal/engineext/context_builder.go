package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ContextBuilder creates execution contexts with specific capabilities
type ContextBuilder struct {
	nodeID       string
	nodeType     string
	blueprintID  string
	executionID  string
	inputs       map[string]types.Value
	variables    map[string]types.Value
	logger       node.Logger
	hooks        *node.ExecutionHooks
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error

	// Feature flags
	withErrorHandling bool
	errorManager      *bperrors.ErrorManager
	recoveryManager   *bperrors.RecoveryManager

	withEventSupport    bool
	eventManager        core.EventManagerInterface
	isEventHandler      bool
	eventHandlerContext *core.EventHandlerContext

	withActorMode bool

	withFunction bool
	functionID   string
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(
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
	return &ContextBuilder{
		nodeID:       nodeID,
		nodeType:     nodeType,
		blueprintID:  blueprintID,
		executionID:  executionID,
		inputs:       inputs,
		variables:    variables,
		logger:       logger,
		hooks:        hooks,
		activateFlow: activateFlow,

		// Default all features to off
		withErrorHandling: false,
		withEventSupport:  false,
		withActorMode:     false,
	}
}

// WithErrorHandling adds error handling capabilities
func (b *ContextBuilder) WithErrorHandling(errorManager *bperrors.ErrorManager, recoveryManager *bperrors.RecoveryManager) *ContextBuilder {
	b.withErrorHandling = true
	b.errorManager = errorManager
	b.recoveryManager = recoveryManager
	return b
}

// WithEventSupport adds event handling capabilities
func (b *ContextBuilder) WithEventSupport(eventManager core.EventManagerInterface, isEventHandler bool, eventHandlerContext *core.EventHandlerContext) *ContextBuilder {
	b.withEventSupport = true
	b.eventManager = eventManager
	b.isEventHandler = isEventHandler
	b.eventHandlerContext = eventHandlerContext
	return b
}

// WithActorMode enables actor-based execution
func (b *ContextBuilder) WithActorMode() *ContextBuilder {
	b.withActorMode = true
	return b
}

func (b *ContextBuilder) WithFunction(fnID string) *ContextBuilder {
	b.withFunction = true
	b.functionID = fnID
	return b
}

// Build creates the execution context with all requested capabilities
func (b *ContextBuilder) Build() node.ExecutionContext {
	ctxFactory := NewContextFactory(b.errorManager, b.recoveryManager, b.eventManager)
	// Create the base execution context
	baseCtx := NewExecutionContext(
		b.nodeID,
		b.nodeType,
		b.blueprintID,
		b.executionID,
		b.inputs,
		b.variables,
		b.logger,
		b.hooks,
		b.activateFlow,
	)

	// Start with the base context
	var ctx node.ExecutionContext = baseCtx

	// Apply decorators in a consistent order
	// Note: The order matters! Some decorators may depend on others.

	// 1. First, apply actor mode if needed (most fundamental change)
	if b.withActorMode {
		// Make the context actor-aware
		ctx = ctxFactory.CreateActorContext(baseCtx)
		//ctx = NewActorExecutionContext(baseCtx)
	}

	// 2. Next, apply error handling
	if b.withErrorHandling {
		ctx = ctxFactory.CreateErrorAwareContext(baseCtx)
		//ctx = NewErrorAwareContext(ctx, b.errorManager, b.recoveryManager)
	}

	// 3. Then, apply event support
	if b.withEventSupport {
		//ctx = event.NewEventAwareContext()
		ctx = ctxFactory.CreateEventAwareContext(ctx, b.isEventHandler, b.eventHandlerContext)
	}

	if b.withFunction {
		ctx = ctxFactory.CreateFunctionContext(ctx, b.functionID)
	}

	return ctx
}
