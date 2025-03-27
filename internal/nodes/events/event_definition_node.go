package events

import (
	"fmt"
	"time"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventDefinitionNode defines a custom event
type EventDefinitionNode struct {
	node.BaseNode
}

// NewEventDefinitionNode creates a new event definition node
func NewEventDefinitionNode() node.Node {
	return &EventDefinitionNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-definition",
				Name:        "Define Event",
				Description: "Defines a custom event that can be dispatched and handled",
				Category:    "Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "name",
					Name:        "Event Name",
					Description: "Name of the event",
					Type:        types.PinTypes.String,
					Default:     "CustomEvent",
				},
				{
					ID:          "description",
					Name:        "Description",
					Description: "Description of the event",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "",
				},
				{
					ID:          "category",
					Name:        "Category",
					Description: "Category for organizing events",
					Type:        types.PinTypes.String,
					Optional:    true,
					Default:     "Custom",
				},
				// Dynamic parameters are handled separately
			},
			Outputs: []types.Pin{
				{
					ID:          "eventID",
					Name:        "Event ID",
					Description: "Unique identifier for the defined event",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "success",
					Name:        "Success",
					Description: "Whether the event was successfully defined",
					Type:        types.PinTypes.Boolean,
				},
			},
			Properties: []types.Property{
				{
					Name:        "parameters",
					Type:        types.PinTypes.Array,
					Description: "Parameters for the event",
					Value:       []interface{}{},
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *EventDefinitionNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing EventDefinitionNode", nil)

	// Get input values
	nameValue, nameExists := ctx.GetInputValue("name")
	descValue, descExists := ctx.GetInputValue("description")
	categoryValue, categoryExists := ctx.GetInputValue("category")

	// Default values
	name := "CustomEvent"
	if nameExists {
		nameStr, err := nameValue.AsString()
		if err == nil && nameStr != "" {
			name = nameStr
		}
	}

	description := ""
	if descExists {
		descStr, err := descValue.AsString()
		if err == nil {
			description = descStr
		}
	}

	category := "Custom"
	if categoryExists {
		categoryStr, err := categoryValue.AsString()
		if err == nil && categoryStr != "" {
			category = categoryStr
		}
	}

	// Get parameters from properties
	var parameters []event.EventParameter

	// Find parameters property
	for _, prop := range n.GetProperties() {
		if prop.Name == "parameters" {
			if paramsArray, ok := prop.Value.([]interface{}); ok {
				for _, paramItem := range paramsArray {
					if paramMap, ok := paramItem.(map[string]interface{}); ok {
						// Determine the pin type based on type ID
						var paramType *types.PinType

						if typeIDValue, exists := paramMap["typeID"]; exists {
							if typeID, ok := typeIDValue.(string); ok {
								switch typeID {
								case "string":
									paramType = types.PinTypes.String
								case "number":
									paramType = types.PinTypes.Number
								case "boolean":
									paramType = types.PinTypes.Boolean
								case "object":
									paramType = types.PinTypes.Object
								case "array":
									paramType = types.PinTypes.Array
								default:
									paramType = types.PinTypes.Any
								}
							}
						} else {
							paramType = types.PinTypes.Any
						}

						// Create the parameter
						param := event.EventParameter{
							Name:        paramMap["name"].(string),
							Description: paramMap["description"].(string),
							Type:        paramType,
							Optional:    paramMap["optional"].(bool),
						}

						// Get default value if present
						if defaultVal, ok := paramMap["default"]; ok {
							param.Default = defaultVal
						}

						parameters = append(parameters, param)
					}
				}
			}
		}
	}

	// Create a unique event ID
	blueprintID := ctx.GetBlueprintID()
	nodeID := ctx.GetNodeID()
	eventID := fmt.Sprintf("%s.%s.%s", blueprintID, nodeID, name)

	// Create event definition
	eventDef := event.EventDefinition{
		ID:          eventID,
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Category:    category,
		BlueprintID: blueprintID,
		CreatedAt:   time.Now(),
	}

	// Get event manager from context
	var eventManager event.EventManagerInterface
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		eventManager = evtCtx.GetEventManager()
	} else {
		logger.Error("Event manager not available in context", nil)

		// Set outputs
		ctx.SetOutputValue("eventID", types.NewValue(types.PinTypes.String, ""))
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		return fmt.Errorf("event manager not available in context")
	}

	// Register the event
	err := eventManager.RegisterEvent(eventDef)
	if err != nil {
		logger.Error("Failed to register event", map[string]interface{}{
			"error": err.Error(),
		})

		// Set outputs
		ctx.SetOutputValue("eventID", types.NewValue(types.PinTypes.String, eventID))
		ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, false))

		return err
	}

	// Set outputs
	ctx.SetOutputValue("eventID", types.NewValue(types.PinTypes.String, eventID))
	ctx.SetOutputValue("success", types.NewValue(types.PinTypes.Boolean, true))

	logger.Info("Event defined successfully", map[string]interface{}{
		"eventID": eventID,
	})

	return nil
}
