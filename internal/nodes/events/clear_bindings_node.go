package events

import (
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ClearBindingsNode removes all event bindings for a blueprint
type ClearBindingsNode struct {
	node.BaseNode
}

// NewClearBindingsNode creates a new clear bindings node
func NewClearBindingsNode() node.Node {
	return &ClearBindingsNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "clear-event-bindings",
				Name:        "Clear Event Bindings",
				Description: "Removes all event bindings for a blueprint",
				Category:    "Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "execute",
					Name:        "Execute",
					Description: "Triggers the clearing of bindings",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "blueprintID",
					Name:        "Blueprint ID",
					Description: "ID of the blueprint to clear bindings for (empty for current blueprint)",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "",
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Executed after the bindings are cleared",
					Type:        types.PinTypes.Execution,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *ClearBindingsNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing ClearBindingsNode", nil)

	// Get input value
	blueprintIDValue, blueprintIDExists := ctx.GetInputValue("blueprintID")

	// Default value - current blueprint
	blueprintID := ctx.GetBlueprintID()
	if blueprintIDExists {
		blueprintIDStr, err := blueprintIDValue.AsString()
		if err == nil && blueprintIDStr != "" {
			blueprintID = blueprintIDStr
		}
	}

	// Get event manager from context
	var eventManager event.EventManagerInterface
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		eventManager = evtCtx.GetEventManager()
	} else {
		logger.Error("Event manager not available in context", nil)
		return ctx.ActivateOutputFlow("then")
	}

	// Clear all bindings for the blueprint
	eventManager.ClearBindings(blueprintID)

	logger.Info("Event bindings cleared", map[string]interface{}{
		"blueprintID": blueprintID,
	})

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
