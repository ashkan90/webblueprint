package event

import (
	"fmt"
	"webblueprint/internal/core"
	"webblueprint/internal/types"
)

// InternalEventManagerAdapter adapts an event.EventManager to the core.EventManagerInterface
type InternalEventManagerAdapter struct {
	manager *EventManager
}

// NewInternalEventManagerAdapter creates a new adapter for event.EventManager
func NewInternalEventManagerAdapter(manager *EventManager) *InternalEventManagerAdapter {
	return &InternalEventManagerAdapter{
		manager: manager,
	}
}

// DispatchEvent dispatches an event
func (a *InternalEventManagerAdapter) DispatchEvent(request interface{}) []error {
	// Try to convert the request to various types
	switch req := request.(type) {
	case EventDispatchRequest:
		// Convert engine.EventDispatchRequest to event dispatch request
		return a.manager.DispatchEvent(req)
	case core.EventDispatchRequest:
		// Convert core.EventDispatchRequest to engine.EventDispatchRequest
		engineRequest := EventDispatchRequest{
			EventID:     req.EventID,
			Parameters:  req.Parameters,
			SourceID:    req.SourceID,
			BlueprintID: req.BlueprintID,
			ExecutionID: req.ExecutionID,
			Timestamp:   req.Timestamp,
		}
		return a.manager.DispatchEvent(engineRequest)
	case map[string]interface{}:
		// Try to extract fields from map
		eventID, _ := req["eventID"].(string)
		parameters := make(map[string]types.Value)

		if paramsMap, ok := req["parameters"].(map[string]types.Value); ok {
			parameters = paramsMap
		} else if paramsMap, ok := req["parameters"].(map[string]interface{}); ok {
			// Try to convert to types.Value
			for k, v := range paramsMap {
				parameters[k] = types.NewValue(types.PinTypes.Any, v)
			}
		}

		sourceID, _ := req["sourceID"].(string)
		blueprintID, _ := req["blueprintID"].(string)
		executionID, _ := req["executionID"].(string)

		// Create engine request
		engineRequest := EventDispatchRequest{
			EventID:     eventID,
			Parameters:  parameters,
			SourceID:    sourceID,
			BlueprintID: blueprintID,
			ExecutionID: executionID,
		}

		return a.manager.DispatchEvent(engineRequest)
	}

	return []error{fmt.Errorf("unsupported request type")}
}

// RegisterEventHandler registers an event handler
// DEPRECATED: This adapter method is incompatible with the refactored EventManager.
// Use EventManager.BindEvent directly.
func (a *InternalEventManagerAdapter) RegisterEventHandler(eventID string, handler interface{}) error {
	return fmt.Errorf("RegisterEventHandler via InternalEventManagerAdapter is deprecated; use BindEvent directly")
}

// UnregisterEventHandler unregisters an event handler
// DEPRECATED: This adapter method is incompatible with the refactored EventManager.
// Use EventManager.RemoveBinding directly.
func (a *InternalEventManagerAdapter) UnregisterEventHandler(eventID string, handlerID string) error {
	// The underlying manager method was removed. We might try to find the binding ID
	// based on eventID and handlerID if needed, but the concept is deprecated.
	// For now, just return an error.
	// Example (if needed, but likely wrong approach):
	// bindings, _ := a.manager.GetEventBindings(eventID)
	// for _, b := range bindings {
	//   if b.HandlerID == handlerID {
	//      a.manager.RemoveBinding(b.ID)
	//      return nil
	//   }
	// }
	return fmt.Errorf("UnregisterEventHandler via InternalEventManagerAdapter is deprecated; use RemoveBinding directly")
}

// GetEventHandlers gets all handlers for an event
// DEPRECATED: This adapter method is incompatible with the refactored EventManager.
func (a *InternalEventManagerAdapter) GetEventHandlers(eventID string) []interface{} {
	// The underlying manager method was removed.
	return []interface{}{}
}

// functionHandlerAdapter adapts a function to the EventHandler interface
// DEPRECATED: This adapter is likely no longer needed as the EventHandler interface is deprecated.
type functionHandlerAdapter struct {
	id      string
	handler func(ctx EventHandlerContext) error
}

func (a *functionHandlerAdapter) HandleEvent(event EventDispatchRequest) error {
	ctx := EventHandlerContext{
		EventID:     event.EventID,
		Parameters:  event.Parameters,
		SourceID:    event.SourceID,
		BlueprintID: event.BlueprintID,
		ExecutionID: event.ExecutionID,
		HandlerID:   a.id,
		Timestamp:   event.Timestamp,
	}
	return a.handler(ctx)
}

func (a *functionHandlerAdapter) GetHandlerID() string {
	return a.id
}
