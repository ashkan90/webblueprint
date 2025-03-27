package events

import (
	"fmt"
	"strings"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/nodes/events/utils"
	"webblueprint/internal/types"
)

// ImprovedEventDispatcherNode dispatches an event by name or ID
type ImprovedEventDispatcherNode struct {
	node.BaseNode
}

// NewImprovedEventDispatcherNode creates a new event dispatcher node with improved name/ID handling
func NewImprovedEventDispatcherNode() node.Node {
	return &ImprovedEventDispatcherNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "improved-event-dispatcher",
				Name:        "Dispatch Event",
				Description: "Dispatches an event by name or ID",
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
					ID:          "eventName",
					Name:        "Event Name",
					Description: "Name of the event to dispatch (optional if using ID)",
					Type:        types.PinTypes.String,
					Optional:    true,
				},
				{
					ID:          "eventID",
					Name:        "Event ID",
					Description: "ID of the event to dispatch (optional if using name)",
					Type:        types.PinTypes.String,
					Optional:    true,
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
					Name:        "eventName",
					Description: "Name of the event to dispatch",
					Type:        types.PinTypes.String,
					Value:       "",
				},
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
func (n *ImprovedEventDispatcherNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing ImprovedEventDispatcherNode", nil)

	// Get input values for event name and ID
	eventNameValue, eventNameExists := ctx.GetInputValue("eventName")
	eventIDValue, eventIDExists := ctx.GetInputValue("eventID")

	// Default values
	eventName := ""
	eventID := ""

	if eventNameExists {
		nameStr, err := eventNameValue.AsString()
		if err == nil && nameStr != "" {
			eventName = nameStr
		}
	}

	if eventIDExists {
		idStr, err := eventIDValue.AsString()
		if err == nil && idStr != "" {
			eventID = idStr
		}
	}

	// If no values were provided via inputs, check the properties
	if eventName == "" && eventID == "" {
		for _, prop := range n.GetProperties() {
			if prop.Name == "eventName" {
				if propVal, ok := prop.Value.(string); ok && propVal != "" {
					eventName = propVal
				}
			} else if prop.Name == "eventID" {
				if propVal, ok := prop.Value.(string); ok && propVal != "" {
					eventID = propVal
				}
			}
		}
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

	// If we have a name but no ID, try to generate or find the ID
	if eventName != "" && eventID == "" {
		// Try to find the event by name in registered events
		events := eventManager.GetAllEvents()
		for _, evt := range events {
			if evt.Name == eventName {
				eventID = evt.ID
				break
			}
		}

		// If still not found, generate a standardized ID
		if eventID == "" {
			// Generate a standardized ID from the name
			eventID = generateEventID(eventName)
		}
	}

	// Validate event ID - it's required for dispatching
	if eventID == "" {
		logger.Error("No event ID or name provided", nil)

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Check if the event exists
	eventDef, exists := eventManager.GetEventDefinition(eventID)
	if !exists {
		// If using a name-derived ID that doesn't exist, we could create the event
		// This would allow for dynamic event creation
		if eventName != "" && strings.HasPrefix(eventID, "custom.") {
			// For now, we'll just log this as an error
			logger.Error("Event does not exist", map[string]interface{}{
				"eventID":   eventID,
				"eventName": eventName,
			})
		} else {
			logger.Error("Event does not exist", map[string]interface{}{
				"eventID": eventID,
			})
		}

		// Set outputs
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	// Update inputs based on event parameters if requested
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
			"eventID":   eventID,
			"eventName": eventName,
			"params":    params,
		})

		// Continue execution
		return ctx.ActivateOutputFlow("then")
	}

	return fmt.Errorf("event manager not available in context")
}

// generateEventID creates a standardized event ID from a name
func generateEventID(name string) string {
	// Convert to lowercase
	id := strings.ToLower(name)

	// Replace spaces and special chars with hyphens
	id = strings.ReplaceAll(id, " ", "-")

	// Remove any characters that aren't alphanumeric or hyphen
	cleanID := ""
	for _, char := range id {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			cleanID += string(char)
		}
	}

	// Ensure it has the custom prefix
	if !strings.HasPrefix(cleanID, "custom.") {
		cleanID = "custom." + cleanID
	}

	return cleanID
}
