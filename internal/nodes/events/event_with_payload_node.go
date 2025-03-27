package events

import (
	"fmt"
	"time"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventWithPayloadNode dispatches an event with a payload of arbitrary data
type EventWithPayloadNode struct {
	node.BaseNode
}

// NewEventWithPayloadNode creates a new event with payload node
func NewEventWithPayloadNode() node.Node {
	return &EventWithPayloadNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-with-payload",
				Name:        "Event With Payload",
				Description: "Dispatches an event with a payload of data",
				Category:    "Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "execute",
					Name:        "Execute",
					Description: "Triggers the event dispatch",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "eventID",
					Name:        "Event ID",
					Description: "ID of the event to dispatch",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "payload",
					Name:        "Payload",
					Description: "Data to include with the event",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "payloadName",
					Name:        "Payload Name",
					Description: "Name for the payload parameter",
					Type:        types.PinTypes.String,
					Default:     "payload",
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Executed after the event is dispatched",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "success",
					Name:        "Success",
					Description: "Whether the event was successfully dispatched",
					Type:        types.PinTypes.Boolean,
				},
			},
			Properties: []types.Property{
				{
					Name:        "eventID",
					Description: "ID of the event to dispatch",
					Type:        types.PinTypes.String,
					Value:       "",
				},
				{
					Name:        "payloadName",
					Description: "Name for the payload parameter",
					Type:        types.PinTypes.String,
					Value:       "payload",
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *EventWithPayloadNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing EventWithPayloadNode", nil)

	// Get input values
	eventIDValue, eventIDExists := ctx.GetInputValue("eventID")
	payloadValue, payloadExists := ctx.GetInputValue("payload")
	payloadNameValue, payloadNameExists := ctx.GetInputValue("payloadName")

	// Default values
	eventID := ""
	if eventIDExists {
		eventIDStr, err := eventIDValue.AsString()
		if err == nil && eventIDStr != "" {
			eventID = eventIDStr
		}
	}

	// If no event ID was provided via input, check the properties
	if eventID == "" {
		for _, prop := range n.GetProperties() {
			if prop.Name == "eventID" {
				if propVal, ok := prop.Value.(string); ok && propVal != "" {
					eventID = propVal
				}
			}
		}
	}

	// Validate event ID
	if eventID == "" {
		logger.Error("No event ID provided", nil)

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Get the payload name
	payloadName := "payload"
	if payloadNameExists {
		payloadNameStr, err := payloadNameValue.AsString()
		if err == nil && payloadNameStr != "" {
			payloadName = payloadNameStr
		}
	} else {
		// Check properties
		for _, prop := range n.GetProperties() {
			if prop.Name == "payloadName" {
				if propVal, ok := prop.Value.(string); ok && propVal != "" {
					payloadName = propVal
				}
			}
		}
	}

	// Validate payload
	if !payloadExists {
		logger.Error("No payload provided", nil)

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Get event manager from context
	var eventManager event.EventManagerInterface
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		eventManager = evtCtx.GetEventManager()
	} else {
		logger.Error("Event manager not available in context", nil)

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Check if the event exists
	_, exists := eventManager.GetEventDefinition(eventID)
	if !exists {
		logger.Error("Event does not exist", map[string]interface{}{
			"eventID": eventID,
		})

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Create parameters for the event
	params := map[string]types.Value{
		"blueprintID": types.NewValue(types.PinTypes.String, ctx.GetBlueprintID()),
		"executionID": types.NewValue(types.PinTypes.String, ctx.GetExecutionID()),
		"timestamp":   types.NewValue(types.PinTypes.Number, float64(time.Now().UnixNano())),
		payloadName:   payloadValue,
	}

	// Dispatch the event
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		err := evtCtx.DispatchEvent(eventID, params)
		if err != nil {
			logger.Error("Failed to dispatch event", map[string]interface{}{
				"error":   err.Error(),
				"eventID": eventID,
			})

			// Set outputs
			ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

			// Continue execution
			return ctx.ActivateOutputFlow("then")
		}

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, true))

		logger.Info("Event dispatched successfully", map[string]interface{}{
			"eventID":     eventID,
			"payloadName": payloadName,
		})

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	return fmt.Errorf("event manager not available in context")
}
