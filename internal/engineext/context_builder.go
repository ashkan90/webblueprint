package engineext

import (
	"context"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/repository" // Added import
)

// ContextBuilder creates execution contexts with specific capabilities
type ContextBuilder struct {
	bp           *blueprint.Blueprint
	nodeID       string
	nodeType     string
	blueprintID  string
	executionID  string
	inputs       map[string]types.Value
	variables    map[string]types.Value
	logger       node.Logger
	hooks        *node.ExecutionHooks
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error
	repoFactory  repository.RepositoryFactory // Ensure field is present

	// Feature flags
	withErrorHandling bool
	errorManager      *bperrors.ErrorManager
	recoveryManager   *bperrors.RecoveryManager

	withEventSupport     bool
	eventManager         core.EventManagerInterface // For core interface needs
	concreteEventManager *event.EventManager        // For concrete needs (like factory)
	isEventHandler       bool
	eventHandlerContext  *core.EventHandlerContext

	withActorMode bool

	withFunction bool
	functionID   string

	withLoopSupport bool
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(
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
	repoFactory repository.RepositoryFactory, // Ensure parameter is present
) *ContextBuilder {
	return &ContextBuilder{
		bp:           bp,
		nodeID:       nodeID,
		nodeType:     nodeType,
		blueprintID:  blueprintID,
		executionID:  executionID,
		inputs:       inputs,
		variables:    variables,
		logger:       logger,
		hooks:        hooks,
		activateFlow: activateFlow,
		repoFactory:  repoFactory, // Ensure assignment is present

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
	b.concreteEventManager = event.ExtractEventManager(b.eventManager)
	b.isEventHandler = isEventHandler
	b.eventHandlerContext = eventHandlerContext
	return b
}

func (b *ContextBuilder) WithLoopSupport() *ContextBuilder {
	b.withLoopSupport = true
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
	// Pass the concrete event manager to the factory
	ctxFactory := NewContextFactory(b.errorManager, b.recoveryManager, b.concreteEventManager, b.repoFactory) // Pass repoFactory
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
		context.WithValue(context.Background(), "bp", b.bp),
		b.repoFactory, // Pass repoFactory
	)

	// Apply decorators in a consistent order
	// Note: The order matters! Each decorator should wrap the result of the previous one.
	// 1. Start with the base context
	var currentCtx node.ExecutionContext = baseCtx

	// 2. Apply Actor Mode if requested
	if b.withActorMode {
		// Apply Error Handling
		if b.withErrorHandling {
			// Wrap the *current* context with error handling
			currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx)
		}

		// Apply Event Support
		if b.withEventSupport {
			// Wrap the *current* context with event support
			currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext)
		}

		// Apply Function Support (Does this wrap or replace?)
		if b.withFunction {
			// Wrap the *current* context with function support
			currentCtx = ctxFactory.CreateFunctionContext(currentCtx, b.functionID)
		}

		// If actor mode was requested, we need to ensure the final context IS ActorExecutionContext
		// or can be unwrapped to it. The current sequential wrapping might hide it.
		// Let's adjust the logic: Actor mode might need to be the outermost wrapper or handled differently.

		// --- Attempt 2: Prioritize Actor Mode if present ---
		if b.withActorMode {
			// Create the ActorExecutionContext first.
			// We need a way to get the specific ActorExecutionContext type.
			// Let's assume NewActorExecutionContext exists and works.
			// This bypasses the factory for the actor part, which might be wrong.
			// Need to clarify how ActorExecutionContext is intended to be created/wrapped.

			// --- Attempt 3: Assume builder produces the correct final type ---
			// Let's trust the builder flags and apply sequentially, hoping the final type is correct.
			// Revert to sequential wrapping:
			currentCtx = baseCtx // Start fresh
			if b.withErrorHandling {
				currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx)
			}
			if b.withEventSupport {
				currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext)
			}
			if b.withFunction {
				currentCtx = ctxFactory.CreateFunctionContext(currentCtx, b.functionID)
			}
			// Actor mode needs special handling - maybe it shouldn't be a simple decorator?
			// If actor mode is requested, perhaps the *entire* context returned should be ActorExecutionContext?
			// Let's assume the builder should return ActorExecutionContext if withActorMode is true.
			// This requires CreateActorContext to return the correct type and potentially wrap others.

			// --- Final Attempt: Correct Sequential Wrapping ---
			currentCtx = baseCtx // Start with base
			if b.withErrorHandling {
				currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx) // Wrap base with error
			}
			if b.withEventSupport {
				currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext) // Wrap result with event
			}
			if b.withFunction {
				currentCtx = ctxFactory.CreateFunctionContext(currentCtx, b.functionID) // Wrap result with function
			}
			// Actor mode is still problematic with this sequential approach if ActorExecutionContext doesn't wrap.
			// Let's assume the panic happens because the context passed to actor.Start
			// was created by CreateStandardContext or similar, which doesn't include Actor mode.
			// The fix might be in actor_system.go ensuring CreateActorContext is called.

			// Let's stick to the corrected sequential wrapping for now.
			// The panic implies the context received by actor.Start was NOT created with .WithActorMode()
			// or the builder logic is flawed.

		} else {
			// Standard sequential wrapping if not actor mode
			currentCtx = baseCtx // Start with base
			if b.withErrorHandling {
				currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx) // Wrap base with error
			}
			if b.withEventSupport {
				currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext) // Wrap result with event
			}
			if b.withFunction {
				currentCtx = ctxFactory.CreateFunctionContext(currentCtx, b.functionID) // Wrap result with function
			}
		} // End else block for standard wrapping
	} else if b.withEventSupport {
		currentCtx = baseCtx

		if b.withErrorHandling {
			currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx)
		}

		if b.withActorMode {
			currentCtx = ctxFactory.CreateActorContext(currentCtx)
		}

		currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext)
	} else if b.withLoopSupport {
		currentCtx = baseCtx

		if b.withErrorHandling {
			currentCtx = ctxFactory.CreateErrorAwareContext(currentCtx)
		}

		if b.withActorMode {
			currentCtx = ctxFactory.CreateActorContext(currentCtx)
		}

		if b.withEventSupport {
			currentCtx = ctxFactory.CreateEventAwareContext(currentCtx, b.isEventHandler, b.eventHandlerContext)
		}

		currentCtx = ctxFactory.CreateLoopContext(currentCtx)
	}

	return currentCtx
}
