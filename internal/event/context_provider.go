package event

import (
	"webblueprint/internal/core"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ExecutionContextWithEvents defines an execution context with event capabilities
type ExecutionContextWithEvents interface {
	node.ExecutionContext
	GetEventManager() EventManagerInterface
	DispatchEvent(eventID string, params map[string]types.Value) error
	IsEventHandlerActive() bool
	GetEventHandlerContext() *EventHandlerContext
	GetEventParameter(paramName string) (types.Value, bool)
	GetEventID() string
	GetEventSourceID() string
}

// ContextProvider provides execution contexts with event capabilities
type ContextProvider struct {
	eventManager     *EventManager // Use concrete type
	engineController core.EngineController
}

// NewContextProvider creates a new context provider
func NewContextProvider(eventManager *EventManager, engineController core.EngineController) *ContextProvider {
	if engineController == nil {
		panic("ContextProvider requires a non-nil EngineController")
	}
	if eventManager == nil {
		// If no manager is provided, create one
		eventManager = NewEventManager(engineController)
	}
	return &ContextProvider{
		eventManager:     eventManager,
		engineController: engineController,
	}
}

// CreateEventAwareContext creates a new event-aware execution context
func (p *ContextProvider) CreateEventAwareContext(
	baseCtx node.ExecutionContext,
	isEventHandler bool,
	eventHandlerContext *core.EventHandlerContext,
) node.ExecutionContext {
	// Sanity check to ensure we have valid event data
	if eventHandlerContext != nil && eventHandlerContext.EventID == "" {
		// Set a default ID if missing
		eventHandlerContext.EventID = "event-" + baseCtx.GetExecutionID()
	}

	// Use the event manager stored in the provider
	eventMgr := p.eventManager // Already concrete type *EventManager

	// Create and return a new event-aware context
	// Pass the interface adapter to NewEventAwareContext if it expects the core interface
	return NewEventAwareContext(baseCtx, eventMgr, isEventHandler, eventHandlerContext)
}

// AsContextProvider converts this context provider to the expected interface type
//func (p *ContextProvider) AsContextProvider() core.ContextProvider {
//	return p
//}
