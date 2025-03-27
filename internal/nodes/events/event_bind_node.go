package events

import (
	"fmt"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventBindNodeMetadata defines metadata for the event bind node
var EventBindNodeMetadata = node.NodeMetadata{
	TypeID:      "event-bind",
	Name:        "Event Bind",
	Description: "Binds a function to an event",
	Category:    "Events",
	Version:     "1.0.0",
	Properties: []types.Property{
		{
			Name:        "description",
			DisplayName: "Description",
			Description: "Optional description for this binding",
			Value:       "",
		},
		{
			Name:        "priority",
			DisplayName: "Priority",
			Description: "Execution priority (higher numbers execute first)",
			Value:       0,
		},
	},
	InputPins: []types.Pin{
		{
			ID:          "bind",
			Name:        "Bind",
			Description: "Trigger to bind the event",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "unbind",
			Name:        "Unbind",
			Description: "Trigger to unbind the event",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "eventID",
			Name:        "Event ID",
			Description: "ID of the event to bind to",
			Type:        types.PinTypes.String,
		},
	},
	OutputPins: []types.Pin{
		{
			ID:          "onBound",
			Name:        "On Bound",
			Description: "Triggered when the event is bound",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "onUnbound",
			Name:        "On Unbound",
			Description: "Triggered when the event is unbound",
			Type:        types.PinTypes.Execution,
		},
		{
			ID:          "bindingID",
			Name:        "Binding ID",
			Description: "ID of the binding (for later unbinding)",
			Type:        types.PinTypes.String,
		},
	},
}

// EventBindNode binds an event handler to an event
type EventBindNode struct {
	node.BaseNode
	bindingID string // Track the binding ID
}

// NewEventBindNode creates a new event bind node
func NewEventBindNode() node.Node {
	return &EventBindNode{
		BaseNode: node.BaseNode{
			Metadata: EventBindNodeMetadata,
		},
		bindingID: "",
	}
}

// AddDynamicOutputPins adds output pins based on event parameters
func (n *EventBindNode) AddDynamicOutputPins(eventDef event.EventDefinition) {
	// Get existing pins
	pins := n.GetOutputPins()

	// Check if the pin already exists
	for _, param := range eventDef.Parameters {
		exists := false
		for _, pin := range pins {
			if pin.ID == param.Name {
				exists = true
				break
			}
		}

		// If the pin doesn't exist, add it
		if !exists {
			// Add a pin for each parameter
			n.AddOutputPin(types.Pin{
				ID:          param.Name,
				Name:        param.Name,
				Description: param.Description,
				Type:        param.Type,
				Optional:    true, // All event parameters are optional
			})
		}
	}
}

// Execute binds an event handler
func (n *EventBindNode) Execute(ctx node.ExecutionContext) error {
	// Get the logger
	logger := ctx.Logger()

	// Check which pin was activated
	inputPin := ""
	if activationCtx, ok := ctx.(node.ActivationAwareContext); ok {
		for _, pin := range []string{"bind", "unbind"} {
			if activationCtx.IsInputPinActive(pin) {
				inputPin = pin
				break
			}
		}
	} else {
		// Fallback - assume "bind" was activated
		inputPin = "bind"
	}

	// Get the event ID
	eventIDValue, exists := ctx.GetInputValue("eventID")
	if !exists {
		logger.Error("Event ID not provided", nil)
		return fmt.Errorf("event ID not provided")
	}

	// Convert to string
	eventID, ok := eventIDValue.RawValue.(string)
	if !ok {
		logger.Error("Event ID is not a string", nil)
		return fmt.Errorf("event ID is not a string")
	}

	// Get event manager from context
	var eventManager event.EventManagerInterface
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		eventManager = evtCtx.GetEventManager()
	} else {
		logger.Error("Event manager not available in context", nil)

		// Set outputs
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))

		return fmt.Errorf("event manager not available in context")
	}

	// Check if the event exists
	eventDef, exists := eventManager.GetEventDefinition(eventID)
	if !exists {
		logger.Error("Event does not exist", map[string]interface{}{
			"eventID": eventID,
		})

		// Set outputs
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))

		return fmt.Errorf("event does not exist: %s", eventID)
	}

	// Handle unbind first
	if inputPin == "unbind" {
		if n.bindingID != "" {
			// Unbind the event
			eventManager.RemoveBinding(n.bindingID)
			logger.Info("Event unbound", map[string]interface{}{
				"eventID":   eventID,
				"bindingID": n.bindingID,
			})

			// Trigger the onUnbound output
			ctx.ActivateOutputFlow("onUnbound")
		}

		// Set the binding ID output (empty)
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))
		n.bindingID = ""

		return nil
	}

	// Get binding settings from properties
	var priority int
	var description string // Used for documentation purposes

	for _, prop := range n.GetProperties() {
		if prop.Name == "description" {
			description, _ = prop.Value.(string)
			// Log the description for debugging purposes
			if description != "" {
				logger.Debug("Binding description", map[string]interface{}{
					"description": description,
				})
			}
		} else if prop.Name == "priority" {
			if val, ok := prop.Value.(float64); ok {
				priority = int(val)
			} else if val, ok := prop.Value.(int); ok {
				priority = val
			}
		}
	}

	// Add dynamic output pins for event parameters
	n.AddDynamicOutputPins(eventDef)

	// Create a binding ID
	bindingID := fmt.Sprintf("binding-%s-%s", eventID, ctx.GetNodeID())

	// Create a binding struct later when calling BindEvent
	// binding := event.EventBinding{ ... } // Removed unused declaration

	// Bind the event (this internally registers the handler func now)
	err := eventManager.BindEvent(event.EventBinding{ // Create struct here
		ID:          bindingID,
		EventID:     eventID,
		HandlerID:   ctx.GetNodeID(), // This node itself is the handler
		HandlerType: "event-bind",
		BlueprintID: ctx.GetBlueprintID(),
		Priority:    priority,
		Enabled:     true,
	}) // Close parenthesis for BindEvent call
	if err != nil {
		logger.Error("Failed to bind event", map[string]interface{}{
			"eventID":   eventID,
			"bindingID": bindingID,
			"error":     err.Error(),
		})
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))
		return fmt.Errorf("failed to bind event: %w", err)
	}

	// Check if this execution is triggered by an event being handled
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok && evtCtx.IsEventHandlerActive() {
		handlerCtx := evtCtx.GetEventHandlerContext()
		if handlerCtx == nil {
			return fmt.Errorf("event handler context is nil despite being active")
		}

		logger.Debug("Executing EventBindNode as event handler", map[string]interface{}{
			"eventID":   handlerCtx.EventID,
			"bindingID": handlerCtx.BindingID,
		})

		// Ensure dynamic pins for parameters exist (might be redundant if done reliably elsewhere)
		eventDef, exists := eventManager.GetEventDefinition(handlerCtx.EventID)
		if exists {
			n.AddDynamicOutputPins(eventDef)
		} else {
			logger.Warn("Event definition not found while handling event", map[string]interface{}{"eventID": handlerCtx.EventID})
		}

		// Set output values based on event parameters
		for paramName, paramValue := range handlerCtx.Parameters {
			// Check if output pin exists before setting
			pinExists := false
			for _, pin := range n.GetOutputPins() {
				if pin.ID == paramName {
					pinExists = true
					break
				}
			}
			if pinExists {
				ctx.SetOutputValue(paramName, paramValue)
			} else {
				logger.Warn("Output pin not found for event parameter", map[string]interface{}{
					"paramName": paramName,
					"eventID":   handlerCtx.EventID,
					"nodeID":    ctx.GetNodeID(),
				})
			}
		}

		// Set the binding ID output (redundant if already set during bind?)
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, handlerCtx.BindingID))

		// Trigger the onBound output flow (representing the event handler execution)
		return ctx.ActivateOutputFlow("onBound") // Use the 'onBound' pin to signify event received
	}

	// --- Handle regular execution (Bind/Unbind pins) ---

	// Get the event ID (already done above)

	// Handle unbind first
	if inputPin == "unbind" {
		if n.bindingID != "" {
			// Unbind the event
			eventManager.RemoveBinding(n.bindingID) // Assumes RemoveBinding exists and works
			logger.Info("Event unbound", map[string]interface{}{
				"eventID":   eventID,
				"bindingID": n.bindingID,
			})
			ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, "")) // Clear output
			n.bindingID = ""                                                           // Clear internal state
			return ctx.ActivateOutputFlow("onUnbound")                                 // Activate unbind flow
		}
		// If already unbound or never bound, just pass through
		logger.Debug("Unbind called but no active binding found", map[string]interface{}{"nodeID": ctx.GetNodeID()})
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))
		return ctx.ActivateOutputFlow("onUnbound") // Still activate flow? Or maybe a different 'alreadyUnbound' flow?
	}

	// Handle bind
	if inputPin == "bind" {
		// Get binding settings from properties (already done above)

		// Add dynamic output pins for event parameters (already done above)

		// Create a binding ID
		// Ensure this is unique enough, maybe include blueprint instance ID if relevant?
		bindingID := fmt.Sprintf("binding-%s-%s", eventID, ctx.GetNodeID())

		// Create a binding struct
		binding := event.EventBinding{
			ID:          bindingID,
			EventID:     eventID,
			HandlerID:   ctx.GetNodeID(), // This node itself is the handler
			HandlerType: n.Metadata.TypeID,
			BlueprintID: ctx.GetBlueprintID(),
			Priority:    priority,
			Enabled:     true, // Default to enabled
		}

		// Bind the event (this internally registers the handler func now)
		err := eventManager.BindEvent(binding)
		if err != nil {
			logger.Error("Failed to bind event", map[string]interface{}{
				"eventID":   eventID,
				"bindingID": bindingID,
				"error":     err.Error(),
			})
			ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, ""))
			return fmt.Errorf("failed to bind event: %w", err)
		}

		// Store the binding ID internally
		n.bindingID = bindingID

		// Set the binding ID output
		ctx.SetOutputValue("bindingID", types.NewValue(types.PinTypes.String, bindingID))

		logger.Info("Event bound successfully", map[string]interface{}{
			"eventID":   eventID,
			"bindingID": bindingID,
		})

		// Activate the onBound flow to indicate successful binding
		return ctx.ActivateOutputFlow("onBound")
	}

	// If neither bind nor unbind was activated (shouldn't happen with current logic)
	logger.Warn("EventBindNode executed without bind or unbind activation", nil)
	return nil
}

// GetBindingID returns the current binding ID
func (n *EventBindNode) GetBindingID() string {
	return n.bindingID
}

// SetBindingID sets the binding ID
func (n *EventBindNode) SetBindingID(id string) {
	n.bindingID = id
}

// Validate validates the node
func (n *EventBindNode) Validate() []bperrors.BlueprintError {
	// Validation will be added in a future update
	return nil
}

// The helper functions for event conversions have been moved to event_utils.go

// eventAwareExecutionAdapter is an adapter for execution contexts with event capabilities
type eventAwareExecutionAdapter struct {
	node.ExecutionContext
	eventManager   event.EventManagerInterface
	isEventHandler bool
	handlerContext event.EventHandlerContext
}

// GetEventManager returns the event manager
func (ctx *eventAwareExecutionAdapter) GetEventManager() event.EventManagerInterface {
	return ctx.eventManager
}

// DispatchEvent dispatches an event with the given parameters
func (ctx *eventAwareExecutionAdapter) DispatchEvent(eventID string, params map[string]types.Value) error {
	// Create a dispatch request
	request := event.EventDispatchRequest{
		EventID:     eventID,
		Parameters:  params,
		SourceID:    ctx.GetNodeID(),
		BlueprintID: ctx.GetBlueprintID(),
		ExecutionID: ctx.GetExecutionID(),
		Timestamp:   ctx.handlerContext.Timestamp,
	}

	// Dispatch the event
	errors := ctx.eventManager.DispatchEvent(request)
	if len(errors) > 0 {
		// Log errors
		for _, err := range errors {
			ctx.Logger().Error("Error dispatching event", map[string]interface{}{
				"eventID": eventID,
				"error":   err.Error(),
			})
		}
		return errors[0]
	}

	return nil
}

// IsEventHandlerActive returns true if this context is handling an event
func (ctx *eventAwareExecutionAdapter) IsEventHandlerActive() bool {
	return ctx.isEventHandler
}

// GetEventHandlerContext returns the event handler context
func (ctx *eventAwareExecutionAdapter) GetEventHandlerContext() *event.EventHandlerContext {
	return &ctx.handlerContext
}

// GetEventParameter gets a parameter from the event being handled
func (ctx *eventAwareExecutionAdapter) GetEventParameter(paramName string) (types.Value, bool) {
	if !ctx.isEventHandler {
		return types.Value{}, false
	}

	param, exists := ctx.handlerContext.Parameters[paramName]
	return param, exists
}
