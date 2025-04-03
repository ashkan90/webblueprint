package event

import (
	"fmt"
	"time"
	"webblueprint/internal/core"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventAwareContextImpl implements an execution context with event awareness
type EventAwareContextImpl struct {
	node.ExecutionContext
	eventManager        *EventManager // Use concrete type
	isEventHandler      bool
	eventHandlerContext *EventHandlerContext
}

// NewEventAwareContext creates a new event-aware execution context
func NewEventAwareContext(
	baseCtx node.ExecutionContext,
	eventManager *EventManager, // Expect concrete type
	isEventHandler bool,
	coreEventHandlerContext *core.EventHandlerContext,
) node.ExecutionContext {
	var eventHandlerContext *EventHandlerContext

	// Convert core.EventHandlerContext to event.EventHandlerContext if needed
	if coreEventHandlerContext != nil {
		eventHandlerContext = &EventHandlerContext{
			EventID:     coreEventHandlerContext.EventID,
			Parameters:  coreEventHandlerContext.Parameters,
			SourceID:    coreEventHandlerContext.SourceID,
			BlueprintID: coreEventHandlerContext.BlueprintID,
			ExecutionID: coreEventHandlerContext.ExecutionID,
			HandlerID:   coreEventHandlerContext.HandlerID,
			BindingID:   coreEventHandlerContext.BindingID,
			Timestamp:   coreEventHandlerContext.Timestamp,
		}
	}

	return &EventAwareContextImpl{
		ExecutionContext:    baseCtx,
		eventManager:        eventManager,
		isEventHandler:      isEventHandler,
		eventHandlerContext: eventHandlerContext,
	}
}

// GetEventManager returns the event manager
func (ctx *EventAwareContextImpl) GetEventManager() EventManagerInterface {
	return ctx.eventManager
}

// DispatchEvent dispatches an event with the given parameters
func (ctx *EventAwareContextImpl) DispatchEvent(eventID string, params map[string]types.Value) error {
	// Create a dispatch request
	request := EventDispatchRequest{
		EventID:     eventID,
		Parameters:  params,
		SourceID:    ctx.GetNodeID(),
		BlueprintID: ctx.GetBlueprintID(),
		ExecutionID: ctx.GetExecutionID(),
		Timestamp:   time.Now(),
	}

	// Dispatch the event
	errors := ctx.eventManager.DispatchEvent(request)
	if len(errors) > 0 {
		// Log errors
		errorMsg := fmt.Sprintf("Error dispatching event %s: %v", eventID, errors[0])
		ctx.Logger().Error(errorMsg, map[string]interface{}{
			"eventID": eventID,
			"nodeID":  ctx.GetNodeID(),
		})
		return errors[0]
	}

	// Log successful dispatch
	ctx.Logger().Debug(fmt.Sprintf("Event dispatched: %s", eventID), map[string]interface{}{
		"sourceID":    ctx.GetNodeID(),
		"executionID": ctx.GetExecutionID(),
	})

	return nil
}

// IsEventHandlerActive returns true if this context is handling an event
func (ctx *EventAwareContextImpl) IsEventHandlerActive() bool {
	return ctx.isEventHandler
}

// GetEventHandlerContext returns the event handler context if this is an event handler
func (ctx *EventAwareContextImpl) GetEventHandlerContext() *EventHandlerContext {
	return ctx.eventHandlerContext
}

// GetEventParameter gets a parameter from the event being handled
func (ctx *EventAwareContextImpl) GetEventParameter(paramName string) (types.Value, bool) {
	if !ctx.isEventHandler || ctx.eventHandlerContext == nil {
		return types.Value{}, false
	}

	param, exists := ctx.eventHandlerContext.Parameters[paramName]
	return param, exists
}

// GetEventID gets the ID of the event being handled
func (ctx *EventAwareContextImpl) GetEventID() string {
	if !ctx.isEventHandler || ctx.eventHandlerContext == nil {
		return ""
	}

	return ctx.eventHandlerContext.EventID
}

// GetEventSourceID gets the source ID of the event being handled
func (ctx *EventAwareContextImpl) GetEventSourceID() string {
	if !ctx.isEventHandler || ctx.eventHandlerContext == nil {
		return ""
	}

	return ctx.eventHandlerContext.SourceID
}

// AsEventAwareContext casts this context to an ExecutionContextWithEvents
func (ctx *EventAwareContextImpl) AsEventAwareContext() ExecutionContextWithEvents {
	return ctx
}

// Unwrap returns the underlying ExecutionContext that this context decorates.
func (ctx *EventAwareContextImpl) Unwrap() node.ExecutionContext {
	return ctx.ExecutionContext
}

// Ensure EventAwareContextImpl implements ExecutionContextWithEvents
var _ ExecutionContextWithEvents = (*EventAwareContextImpl)(nil)
