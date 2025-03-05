package web

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// DOMEventNode implements a node that listens for DOM events
type DOMEventNode struct {
	node.BaseNode
}

// NewDOMEventNode creates a new DOM Event node
func NewDOMEventNode() node.Node {
	return &DOMEventNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "dom-event",
				Name:        "DOM Event",
				Description: "Listens for events on DOM elements",
				Category:    "Web",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "selector",
					Name:        "Selector",
					Description: "CSS selector for element(s) to listen on",
					Type:        types.PinTypes.String,
					Default:     "body",
				},
				{
					ID:          "eventType",
					Name:        "Event Type",
					Description: "Type of event to listen for (e.g., 'click', 'submit')",
					Type:        types.PinTypes.String,
					Default:     "click",
				},
				{
					ID:          "useCapture",
					Name:        "Use Capture",
					Description: "Whether to use capture phase instead of bubbling",
					Type:        types.PinTypes.Boolean,
					Optional:    true,
					Default:     false,
				},
				{
					ID:          "preventDefault",
					Name:        "Prevent Default",
					Description: "Automatically prevent default behavior",
					Type:        types.PinTypes.Boolean,
					Optional:    true,
					Default:     false,
				},
				{
					ID:          "stopPropagation",
					Name:        "Stop Propagation",
					Description: "Automatically stop event propagation",
					Type:        types.PinTypes.Boolean,
					Optional:    true,
					Default:     false,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "triggered",
					Name:        "Triggered",
					Description: "Executed when the event occurs",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "event",
					Name:        "Event",
					Description: "The event object with details",
					Type:        types.PinTypes.Object,
				},
				{
					ID:          "target",
					Name:        "Target",
					Description: "The element that triggered the event",
					Type:        types.PinTypes.Object,
				},
			},
		},
	}
}

// Execute runs the node logic - for DOM event nodes, this sets up the event handler
// The actual execution happens when the event is triggered on the client side
func (n *DOMEventNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Setting up DOM Event node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	selectorValue, selectorExists := ctx.GetInputValue("selector")
	eventTypeValue, eventTypeExists := ctx.GetInputValue("eventType")
	useCaptureValue, useCaptureExists := ctx.GetInputValue("useCapture")
	preventDefaultValue, preventDefaultExists := ctx.GetInputValue("preventDefault")
	stopPropagationValue, stopPropagationExists := ctx.GetInputValue("stopPropagation")

	// Default values
	selector := "body"
	if selectorExists {
		if selectorStr, err := selectorValue.AsString(); err == nil && selectorStr != "" {
			selector = selectorStr
		}
	}

	eventType := "click"
	if eventTypeExists {
		if eventTypeStr, err := eventTypeValue.AsString(); err == nil && eventTypeStr != "" {
			eventType = eventTypeStr
		}
	}

	// Get boolean settings with defaults
	useCapture := false
	if useCaptureExists {
		if useCaptureVal, err := useCaptureValue.AsBoolean(); err == nil {
			useCapture = useCaptureVal
		}
	}

	preventDefault := false
	if preventDefaultExists {
		if preventDefaultVal, err := preventDefaultValue.AsBoolean(); err == nil {
			preventDefault = preventDefaultVal
		}
	}

	stopPropagation := false
	if stopPropagationExists {
		if stopPropagationVal, err := stopPropagationValue.AsBoolean(); err == nil {
			stopPropagation = stopPropagationVal
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"selector":        selector,
		"eventType":       eventType,
		"useCapture":      useCapture,
		"preventDefault":  preventDefault,
		"stopPropagation": stopPropagation,
	}

	// Create the listener configuration
	listenerConfig := map[string]interface{}{
		"selector":        selector,
		"eventType":       eventType,
		"useCapture":      useCapture,
		"preventDefault":  preventDefault,
		"stopPropagation": stopPropagation,
		"nodeId":          ctx.GetNodeID(),
		"executionId":     ctx.GetExecutionID(),
		"blueprintId":     ctx.GetBlueprintID(),
	}

	// In a real implementation, we would register this event with a client-side handler
	// For this backend implementation, we'll just log and prepare the output

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "DOM Event Listener Registration",
		Value: map[string]interface{}{
			"listenerConfig": listenerConfig,
			"timestamp":      time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("DOM event listener configured", map[string]interface{}{
		"selector":  selector,
		"eventType": eventType,
	})

	// We don't call ActivateOutputFlow here since this node doesn't activate immediately
	// It will be triggered by client-side events
	return nil
}

// HandleClientEvent is called when an event is triggered from the client side
// This would be called via an API endpoint or WebSocket message
func (n *DOMEventNode) HandleClientEvent(ctx node.ExecutionContext, eventData map[string]interface{}) error {
	logger := ctx.Logger()
	logger.Info("DOM event triggered", map[string]interface{}{
		"eventType": eventData["type"],
		"target":    eventData["target"],
	})

	// Set output values
	ctx.SetOutputValue("event", types.NewValue(types.PinTypes.Object, eventData))

	// Extract target element if available
	if target, ok := eventData["target"]; ok {
		ctx.SetOutputValue("target", types.NewValue(types.PinTypes.Object, target))
	}

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "DOM Event Triggered",
		Value: map[string]interface{}{
			"eventData": eventData,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	// Activate the execution flow
	return ctx.ActivateOutputFlow("triggered")
}
