package integration

import (
	"fmt"
	"webblueprint/internal/core"
	"webblueprint/internal/engineext"
)

// Bridge connects components that can't import each other directly

// CreateEventHandler creates an event handler that can be registered with the event system
func CreateEventHandler(id string, handler func(core.EventDispatchRequest) error) core.EventHandler {
	return &bridgeEventHandler{
		id:      id,
		handler: handler,
	}
}

// Implementation of the EventHandler interface
type bridgeEventHandler struct {
	id      string
	handler func(core.EventDispatchRequest) error
}

func (h *bridgeEventHandler) HandleEvent(event core.EventDispatchRequest) error {
	return h.handler(event)
}

func (h *bridgeEventHandler) GetHandlerID() string {
	return h.id
}

// RegisterEventHandler registers an event handler with an engine extension
// DEPRECATED: The underlying core.EventManagerInterface method was removed. Use BindEvent on concrete EventManager.
func RegisterEventHandler(extensions *engineext.ExecutionEngineExtensions, eventID string, handler core.EventHandler) error {
	// eventManager := extensions.GetEventManager()
	// return eventManager.RegisterEventHandler(eventID, handler)
	return fmt.Errorf("RegisterEventHandler via bridge is deprecated; use BindEvent directly")
}

// UnregisterEventHandler unregisters an event handler
// DEPRECATED: The underlying core.EventManagerInterface method was removed. Use RemoveBinding on concrete EventManager.
func UnregisterEventHandler(extensions *engineext.ExecutionEngineExtensions, eventID string, handlerID string) error {
	// eventManager := extensions.GetEventManager()
	// return eventManager.UnregisterEventHandler(eventID, handlerID)
	return fmt.Errorf("UnregisterEventHandler via bridge is deprecated; use RemoveBinding directly")
}
