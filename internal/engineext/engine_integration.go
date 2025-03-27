package engineext

import (
	"fmt"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/event"
)

// Initializes the ExecutionEngineExtensions defined in context_integration.go

// InitializeExtensions creates a new extension manager
func InitializeExtensions(
	engine interface{},
	contextManager *ContextManager,
	errorManager *bperrors.ErrorManager,
	recoveryManager *bperrors.RecoveryManager,
	eventManager core.EventManagerInterface,
) *ExecutionEngineExtensions {
	// Convert core.EventManagerInterface to event.EventManagerInterface if needed
	var adaptedEventManager core.EventManagerInterface = eventManager

	return &ExecutionEngineExtensions{
		Engine:          engine,
		ContextManager:  contextManager,
		ErrorManager:    errorManager,
		RecoveryManager: recoveryManager,
		EventManager:    adaptedEventManager,
	}
}

// eventManagerAdapter adapts event.EventManagerInterface to core.EventManagerInterface
type eventManagerAdapter struct {
	manager event.EventManagerInterface
}

// DispatchEvent dispatches an event
func (a *eventManagerAdapter) DispatchEvent(request interface{}) []error {
	// Convert the request to event.EventDispatchRequest if needed
	if req, ok := request.(event.EventDispatchRequest); ok {
		return a.manager.DispatchEvent(req)
	}

	// If it's a core.EventDispatchRequest, convert to event.EventDispatchRequest
	if req, ok := request.(core.EventDispatchRequest); ok {
		eventReq := event.EventDispatchRequest{
			EventID:     req.EventID,
			Parameters:  req.Parameters,
			SourceID:    req.SourceID,
			BlueprintID: "",
			ExecutionID: "",
			Timestamp:   req.Timestamp,
		}
		return a.manager.DispatchEvent(eventReq)
	}

	return []error{fmt.Errorf("invalid event dispatch request type")}
}

// RegisterEventHandler registers a handler for an event
func (a *eventManagerAdapter) RegisterEventHandler(eventID string, handler interface{}) error {
	// Not fully implemented - would need to adapt between core and event handler types
	return fmt.Errorf("not implemented")
}

// UnregisterEventHandler unregisters a handler
func (a *eventManagerAdapter) UnregisterEventHandler(eventID string, handlerID string) error {
	a.manager.RemoveBinding(handlerID)
	return nil
}

// GetEventHandlers gets all handlers for an event
func (a *eventManagerAdapter) GetEventHandlers(eventID string) []interface{} {
	// Not fully implemented
	return []interface{}{}
}
