package events

import (
	"fmt"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/nodes/events/utils"
	"webblueprint/internal/types"
)

// EventDispatcherNode dispatches an event
type EventDispatcherNode struct {
	node.BaseNode
}

// NewEventDispatcherNode creates a new event dispatcher node
func NewEventDispatcherNode() node.Node {
	return &EventDispatcherNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-dispatcher",
				Name:        "Dispatch Event",
				Description: "Dispatches an event to all bound handlers",
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
				// Dynamic parameter pins will be added based on the event definition
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
					Name:        "dynamicParameters",
					Description: "Whether to include dynamic parameter pins based on event definition",
					Type:        types.PinTypes.Boolean,
					Value:       true,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *EventDispatcherNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing EventDispatcherNode", nil)

	// Get input value for event ID
	eventIDValue, eventIDExists := ctx.GetInputValue("eventID")

	// Default value
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
	eventDef, exists := eventManager.GetEventDefinition(eventID)
	if !exists {
		logger.Error("Event does not exist", map[string]interface{}{
			"eventID": eventID,
		})

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Update inputs based on event parameters
	useDynamicParams := true
	for _, prop := range n.GetProperties() {
		if prop.Name == "dynamicParameters" {
			if val, ok := prop.Value.(bool); ok {
				useDynamicParams = val
			}
		}
	}

	if useDynamicParams {
		utils.UpdateNodePinsForEvent(n, eventDef, true)
	}

	// Collect parameters from inputs
	params := make(map[string]types.Value)
	for _, paramDef := range eventDef.Parameters {
		// Try to get value from input
		if paramValue, exists := ctx.GetInputValue(paramDef.Name); exists {
			params[paramDef.Name] = paramValue
		} else if !paramDef.Optional {
			// Missing required parameter
			logger.Error("Missing required parameter", map[string]interface{}{
				"parameter": paramDef.Name,
				"eventID":   eventID,
			})

			// Set outputs
			ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

			// Continue execution
			return ctx.ActivateOutputFlow("then")
		} else if paramDef.Default != nil {
			// Use default value for optional parameter
			defaultValue := types.NewValue(paramDef.Type, paramDef.Default)
			params[paramDef.Name] = defaultValue
		}
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
			"eventID": eventID,
			"params":  params,
		})

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	return fmt.Errorf("event manager not available in context")
}
