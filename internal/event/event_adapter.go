package event

import (
	"webblueprint/internal/core"
)

// EventManagerAdapter adapts the EventManager to core.EventManagerInterface
type EventManagerAdapter struct {
	manager *EventManager
}

// NewEventManagerAdapter creates a new event manager adapter
func NewEventManagerAdapter(manager *EventManager) *EventManagerAdapter {
	return &EventManagerAdapter{
		manager: manager,
	}
}

// DispatchEvent dispatches an event
func (a *EventManagerAdapter) DispatchEvent(req core.EventDispatchRequest) []error {
	return a.manager.DispatchEvent(EventDispatchRequest{
		EventID:     req.EventID,
		Parameters:  req.Parameters,
		SourceID:    req.SourceID,
		BlueprintID: req.BlueprintID,
		ExecutionID: req.ExecutionID,
		Timestamp:   req.Timestamp,
	})
}

//// RegisterEventHandler registers an event handler
//func (a *EventManagerAdapter) RegisterEventHandler(eventID string, handler interface{}) error {
//	return a.manager.RegisterHandler(integration.CreateEventHandler())
//}

// UnregisterEventHandler unregisters an event handler
//func (a *EventManagerAdapter) UnregisterEventHandler(eventID string, handlerID string) error {
//	return a.manager.UnregisterEventHandler(eventID, handlerID)
//}

// GetEventHandlers gets all handlers for an event
//func (a *EventManagerAdapter) GetEventHandlers(eventID string) []interface{} {
//	// Convert the handlers to interface{}
//	handlers := a.manager.GetEventHandlers(eventID)
//	result := make([]interface{}, len(handlers))
//	for i, h := range handlers {
//		result[i] = h
//	}
//	return result
//}

// EventHandlerAdapter adapts a core.EventHandler to event.EventHandler
type EventHandlerAdapter struct {
	handler core.EventHandler
}

// HandleEvent handles an event
func (a *EventHandlerAdapter) HandleEvent(event EventDispatchRequest) error {
	// Convert the event to core.EventDispatchRequest
	return a.handler.HandleEvent(core.EventDispatchRequest{
		EventID:     event.EventID,
		Parameters:  event.Parameters,
		SourceID:    event.SourceID,
		BlueprintID: event.BlueprintID,
		ExecutionID: event.ExecutionID,
		Timestamp:   event.Timestamp,
	})
}

// GetHandlerID gets the handler ID
func (a *EventHandlerAdapter) GetHandlerID() string {
	return a.handler.GetHandlerID()
}
