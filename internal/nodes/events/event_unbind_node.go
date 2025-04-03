package events

import (
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventUnbindNode removes an event binding
type EventUnbindNode struct {
	node.BaseNode
}

// NewEventUnbindNode creates a new event unbind node
func NewEventUnbindNode() node.Node {
	return &EventUnbindNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-unbind",
				Name:        "Unbind Event",
				Description: "Removes an event binding",
				Category:    "Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "execute",
					Name:        "Execute",
					Description: "Triggers the unbinding",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "bindingID",
					Name:        "Binding ID",
					Description: "ID of the binding to remove",
					Type:        types.PinTypes.String,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Executed after the binding is removed",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "success",
					Name:        "Success",
					Description: "Whether the binding was successfully removed",
					Type:        types.PinTypes.Boolean,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *EventUnbindNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing EventUnbindNode", nil)

	// Get input value
	bindingIDValue, bindingIDExists := ctx.GetInputValue("bindingID")

	// Default value
	bindingID := ""
	if bindingIDExists {
		bindingIDStr, err := bindingIDValue.AsString()
		if err == nil && bindingIDStr != "" {
			bindingID = bindingIDStr
		}
	}

	// Validate binding ID
	if bindingID == "" {
		logger.Error("No binding ID provided", nil)

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

	// Remove the binding
	eventManager.RemoveBinding(bindingID)

	// Set outputs
	ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, true))

	logger.Info("Event binding removed", map[string]interface{}{
		"bindingID": bindingID,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
