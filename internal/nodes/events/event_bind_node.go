package events

import (
	"fmt"
	"webblueprint/internal/bperrors" // Keep bperrors for Validate
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	// "webblueprint/internal/core" // No longer needed? Check usage
)

// EventBindNodeMetadata defines metadata for the event bind node
var EventBindNodeMetadata = node.NodeMetadata{
	TypeID:      "event-bind",
	Name:        "On Event Received",                                                   // Updated Name to reflect behavior
	Description: "Listens for a specific event and triggers execution when it occurs.", // Updated description
	Category:    "Events",
	Version:     "1.1.0", // Bump version due to significant change
	Properties: []types.Property{
		// Keep description and priority properties
		{
			Name:        "description",
			DisplayName: "Description",
			Description: "Optional description for this listener",
			Value:       "",
		},
		{
			Name:        "priority",
			DisplayName: "Priority",
			Description: "Execution priority (higher numbers execute first)",
			Value:       0,
		},
		// Add Event ID as a property instead of an input pin
		{
			Name:        "eventID",
			DisplayName: "Event ID",
			Description: "ID of the event to listen for",
			Value:       "", // Default to empty, user must configure
		},
	},
	OutputPins: []types.Pin{
		{
			ID:          "onEventReceived",   // Changed ID for clarity
			Name:        "On Event Received", // Changed Name for clarity
			Description: "Triggered when the specified event is received",
			Type:        types.PinTypes.Execution,
		},
		// Dynamic output pins for event parameters will be added by AddDynamicOutputPins
	},
}

// EventBindNode listens for a specific event and triggers execution when it occurs.
type EventBindNode struct {
	node.BaseNode
}

// NewEventBindNode creates a new event bind node
func NewEventBindNode() node.Node {
	// Initialize BaseNode directly using metadata
	return &EventBindNode{
		BaseNode: node.BaseNode{
			Metadata:   EventBindNodeMetadata,
			Inputs:     EventBindNodeMetadata.InputPins,  // Get pins from metadata
			Outputs:    EventBindNodeMetadata.OutputPins, // Get pins from metadata
			Properties: EventBindNodeMetadata.Properties, // Get properties from metadata
		},
	}
}

// AddDynamicOutputPins adds output pins based on event parameters
// This might be called during blueprint load/validation or when the event definition is known
func (n *EventBindNode) AddDynamicOutputPins(eventDef event.EventDefinition) {
	// Get existing pins
	pins := n.GetOutputPins()
	pinMap := make(map[string]bool)
	for _, p := range pins {
		pinMap[p.ID] = true
	}

	// Add pins for parameters if they don't already exist
	for _, param := range eventDef.Parameters {
		if !pinMap[param.Name] {
			n.AddOutputPin(types.Pin{
				ID:          param.Name,
				Name:        param.Name, // Use parameter name as pin name
				Description: param.Description,
				Type:        param.Type,
				Optional:    true, // Event parameters are treated as optional outputs here
			})
			pinMap[param.Name] = true // Mark as added
		}
	}
}

// Execute handles the event trigger from the EventManager
func (n *EventBindNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()

	// This node should ONLY execute when triggered by an event.
	evtCtx, ok := ctx.(event.ExecutionContextWithEvents)
	if !ok || !evtCtx.IsEventHandlerActive() {
		// If not triggered by an event, this node shouldn't do anything.
		logger.Warn("EventBindNode executed but not via an event trigger. Ignoring.", map[string]interface{}{"nodeID": ctx.GetNodeID()})
		return nil // Not an error, just unexpected activation
	}

	handlerCtx := evtCtx.GetEventHandlerContext()
	if handlerCtx == nil {
		// This shouldn't happen if IsEventHandlerActive is true, but check defensively.
		err := fmt.Errorf("event handler context is nil despite being active for node %s", ctx.GetNodeID())
		logger.Error(err.Error(), nil)
		return err
	}

	// Optional: Verify the received eventID matches the configured one?
	// configuredEventID := n.GetProperty("eventID").Value.(string) // Assuming property access method
	// if configuredEventID != "" && handlerCtx.EventID != configuredEventID {
	//     logger.Warn("EventBindNode received event it wasn't configured for?", map[string]interface{}{"configured": configuredEventID, "received": handlerCtx.EventID})
	//     return nil // Ignore if mismatch? Or should binding prevent this?
	// }

	logger.Debug("Executing EventBindNode as event handler", map[string]interface{}{
		"eventID":   handlerCtx.EventID,
		"bindingID": handlerCtx.BindingID, // Still useful for debugging maybe
		"sourceID":  handlerCtx.SourceID,
	})

	// Get the Event Definition to potentially add dynamic pins
	// This might happen too late if pins are needed immediately.
	// Consider if AddDynamicOutputPins should be called earlier (e.g., during binding/load).
	eventManager := evtCtx.GetEventManager()
	eventDef, exists := eventManager.GetEventDefinition(handlerCtx.EventID)
	if exists {
		n.AddDynamicOutputPins(eventDef)
	} else {
		logger.Warn("Event definition not found while handling event", map[string]interface{}{"eventID": handlerCtx.EventID, "nodeID": ctx.GetNodeID()})
	}

	// Set output values based on event parameters
	for paramName, paramValue := range handlerCtx.Parameters {
		pinExists := false
		for _, pin := range n.GetOutputPins() {
			if pin.ID == paramName {
				pinExists = true
				break
			}
		}
		if pinExists {
			// Use the context passed into Execute, which is already event-aware
			ctx.SetOutputValue(paramName, paramValue)
		} else {
			// This might happen if AddDynamicOutputPins failed or wasn't called yet
			logger.Warn("Output pin not found for event parameter", map[string]interface{}{
				"paramName": paramName,
				"eventID":   handlerCtx.EventID,
				"nodeID":    ctx.GetNodeID(),
			})
		}
	}

	// Trigger the output execution flow
	return ctx.ActivateOutputFlow("onEventReceived")
}

// Validate validates the node (basic implementation)
func (n *EventBindNode) Validate() []bperrors.BlueprintError {
	// TODO: Add validation logic - e.g., check if Event ID property is set?
	var errors []bperrors.BlueprintError
	// Example validation: Check if eventID property is configured
	// eventIDProp := n.GetProperty("eventID")
	// if eventIDProp == nil || eventIDProp.Value == "" {
	//     errors = append(errors, bperrors.NewValidationError("EventBindNode requires 'eventID' property to be configured.", n.Metadata.TypeID, "MISSING_EVENT_ID"))
	// }
	return errors
}

// Removed GetBindingID and SetBindingID methods
// Removed eventAwareExecutionAdapter struct and methods
