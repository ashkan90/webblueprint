package api

import (
	"fmt"
	"time"
	"webblueprint/internal/event"
)

// EventListener listens for events and sends notifications via WebSockets
type EventListener struct {
	wsManager *WebSocketManager
	logger    *WebSocketLogger
}

// NewEventListener creates a new event listener that sends events via WebSockets
func NewEventListener(wsManager *WebSocketManager, logger *WebSocketLogger) *EventListener {
	return &EventListener{
		wsManager: wsManager,
		logger:    logger,
	}
}

// OnEventDispatched is called when an event is dispatched
func (el *EventListener) OnEventDispatched(eventID string, request event.EventDispatchRequest) {
	// Create event data
	data := map[string]interface{}{
		"eventID":     eventID,
		"parameters":  request.Parameters,
		"sourceID":    request.SourceID,
		"blueprintID": request.BlueprintID,
		"executionID": request.ExecutionID,
		"timestamp":   request.Timestamp.Format(time.RFC3339),
	}

	// Send to all clients
	el.wsManager.BroadcastMessage("event.dispatched", data)

	// Log the event
	el.logger.Info(fmt.Sprintf("Event dispatched: %s", eventID), map[string]interface{}{
		"eventID":     eventID,
		"blueprintID": request.BlueprintID,
		"executionID": request.ExecutionID,
	})
}

// OnEventBound is called when an event binding is created
func (el *EventListener) OnEventBound(binding event.EventBinding) {
	// Create binding data
	data := map[string]interface{}{
		"bindingID":   binding.ID,
		"eventID":     binding.EventID,
		"handlerID":   binding.HandlerID,
		"handlerType": binding.HandlerType,
		"blueprintID": binding.BlueprintID,
		"priority":    binding.Priority,
		"enabled":     binding.Enabled,
		"timestamp":   binding.CreatedAt.Format(time.RFC3339),
	}

	// Send to all clients
	el.wsManager.BroadcastMessage("event.bound", data)

	// Log the binding
	el.logger.Info(fmt.Sprintf("Event bound: %s -> %s", binding.EventID, binding.HandlerID), map[string]interface{}{
		"bindingID":   binding.ID,
		"blueprintID": binding.BlueprintID,
	})
}

// OnEventUnbound is called when an event binding is removed
func (el *EventListener) OnEventUnbound(bindingID string) {
	// Create unbind data
	data := map[string]interface{}{
		"bindingID": bindingID,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Send to all clients
	el.wsManager.BroadcastMessage("event.unbound", data)

	// Log the unbinding
	el.logger.Info(fmt.Sprintf("Event unbound: %s", bindingID), nil)
}
