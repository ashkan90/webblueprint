package utils

import (
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// UpdateNodePinsForEvent updates node input/output pins based on event parameters
// This allows for dynamic pin generation when binding to or dispatching events
func UpdateNodePinsForEvent(
	n node.Node,
	eventDef event.EventDefinition,
	forDispatcher bool,
) {
	// Create new pins collection
	var updatedPins []types.Pin

	// Keep existing non-parameter pins
	for _, pin := range n.GetInputPins() {
		// Skip parameter pins for dispatcher node, or keep all pins for bind node
		if forDispatcher && isParameterPin(pin.ID, eventDef) {
			continue
		}
		updatedPins = append(updatedPins, pin)
	}

	// Add parameter pins based on event definition
	if forDispatcher {
		// For dispatcher, add parameters as input pins
		for _, param := range eventDef.Parameters {
			paramPin := types.Pin{
				ID:          param.Name,
				Name:        param.Name,
				Description: param.Description,
				Type:        param.Type,
				Optional:    param.Optional,
			}

			// Set default value if present
			if param.Default != nil {
				paramPin.Default = param.Default
			}

			updatedPins = append(updatedPins, paramPin)
		}

		// Set the updated input pins
		n.SetInputPins(updatedPins)
	} else {
		// Keep existing output pins for bind nodes
		updatedOutputPins := []types.Pin{}
		for _, pin := range n.GetOutputPins() {
			// Skip parameter pins
			if !isParameterPin(pin.ID, eventDef) {
				updatedOutputPins = append(updatedOutputPins, pin)
			}
		}

		// For bind nodes, add parameters as output pins
		for _, param := range eventDef.Parameters {
			paramPin := types.Pin{
				ID:          param.Name,
				Name:        param.Name,
				Description: param.Description,
				Type:        param.Type,
			}

			updatedOutputPins = append(updatedOutputPins, paramPin)
		}

		// Set the updated output pins
		n.SetOutputPins(updatedOutputPins)
	}
}

// isParameterPin checks if a pin ID matches an event parameter
func isParameterPin(pinID string, eventDef event.EventDefinition) bool {
	for _, param := range eventDef.Parameters {
		if param.Name == pinID {
			return true
		}
	}
	return false
}
