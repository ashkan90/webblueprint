package events

import (
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// OnCreatedEventNode represents the entry point when a blueprint is created
type OnCreatedEventNode struct {
	node.BaseNode
}

// NewOnCreatedEventNode creates a new OnCreated entry point node
func NewOnCreatedEventNode() node.Node {
	return &OnCreatedEventNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-on-created",
				Name:        "On Created",
				Description: "Entry point called when a blueprint is created",
				Category:    "Constructive Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Then flow when the blueprint is created",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "blueprintID",
					Name:        "Blueprint ID",
					Description: "ID of the blueprint",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "timestamp",
					Name:        "Timestamp",
					Description: "Creation timestamp",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the OnCreated entry point node's logic
func (n *OnCreatedEventNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing OnCreatedEventNode", nil)

	// Set output values
	ctx.SetOutputValue("blueprintID", types.NewValue(types.PinTypes.String, ctx.GetBlueprintID()))
	ctx.SetOutputValue("timestamp", types.NewValue(types.PinTypes.Number, float64(time.Now().UnixNano()/1e6)))

	// Activate the execution flow
	return ctx.ActivateOutputFlow("then")
}

// OnTickEventNode represents the periodic tick event
type OnTickEventNode struct {
	node.BaseNode
}

// NewOnTickEventNode creates a new OnTick entry point node
func NewOnTickEventNode() node.Node {
	return &OnTickEventNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-on-tick",
				Name:        "On Tick",
				Description: "Entry point called periodically during blueprint execution",
				Category:    "Constructive Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "interval",
					Name:        "Interval (ms)",
					Description: "Tick interval in milliseconds",
					Type:        types.PinTypes.Number,
					Default:     float64(1000), // 1 second default
				},
				{
					ID:          "enabled",
					Name:        "Enabled",
					Description: "Whether the tick is enabled",
					Type:        types.PinTypes.Boolean,
					Default:     true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "execution",
					Name:        "Execution",
					Description: "Execution flow on each tick",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "deltaTime",
					Name:        "Delta Time",
					Description: "Time since last tick in milliseconds",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "tickCount",
					Name:        "Tick Count",
					Description: "The number of ticks since start",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the OnTick entry point node's logic
func (n *OnTickEventNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing OnTickEventNode", nil)

	// Get tick interval - we don't use it in this simplified implementation
	// but we fetch it to demonstrate how it would be used
	_, exists := ctx.GetInputValue("interval")
	if !exists {
		logger.Debug("No tick interval specified, using default", nil)
	}

	// Check if enabled
	enabledVal, exists := ctx.GetInputValue("enabled")
	enabled := true
	if exists {
		if boolVal, err := enabledVal.AsBoolean(); err == nil {
			enabled = boolVal
		}
	}

	// Skip execution if disabled
	if !enabled {
		logger.Debug("Tick is disabled, skipping", nil)
		return nil
	}

	// Store in variables for persistence between ticks
	tickCountVar, exists := ctx.GetVariable("tickCount_" + ctx.GetNodeID())
	tickCount := float64(0)
	if exists {
		if count, err := tickCountVar.AsNumber(); err == nil {
			tickCount = count
		}
	}

	lastTickVar, exists := ctx.GetVariable("lastTick_" + ctx.GetNodeID())
	lastTick := float64(0)
	if exists {
		if last, err := lastTickVar.AsNumber(); err == nil {
			lastTick = last
		}
	}

	// Current time in milliseconds
	now := float64(time.Now().UnixNano() / 1e6)

	// Calculate delta time
	deltaTime := float64(0)
	if lastTick > 0 {
		deltaTime = now - lastTick
	}

	// Update persistent variables
	tickCount++
	ctx.SetVariable("tickCount_"+ctx.GetNodeID(), types.NewValue(types.PinTypes.Number, tickCount))
	ctx.SetVariable("lastTick_"+ctx.GetNodeID(), types.NewValue(types.PinTypes.Number, now))

	// Set output values
	ctx.SetOutputValue("deltaTime", types.NewValue(types.PinTypes.Number, deltaTime))
	ctx.SetOutputValue("tickCount", types.NewValue(types.PinTypes.Number, tickCount))

	// Activate the execution flow
	return ctx.ActivateOutputFlow("execution")
}

// OnInputEventNode represents an entry point triggered by external input
type OnInputEventNode struct {
	node.BaseNode
}

// NewOnInputEventNode creates a new OnInput entry point node
func NewOnInputEventNode() node.Node {
	return &OnInputEventNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "event-on-input",
				Name:        "On Input",
				Description: "Entry point called when input is received",
				Category:    "Constructive Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "inputName",
					Name:        "Input Name",
					Description: "Name of the input to listen for",
					Type:        types.PinTypes.String,
					Default:     "",
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "execution",
					Name:        "Execution",
					Description: "Execution flow when input is received",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "inputValue",
					Name:        "Input Value",
					Description: "Value of the received input",
					Type:        types.PinTypes.Any,
				},
				{
					ID:          "timestamp",
					Name:        "Timestamp",
					Description: "Timestamp when input was received",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the OnInput entry point node's logic
func (n *OnInputEventNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing OnInputEventNode", nil)

	// In a real implementation, we would check if this is the right input event
	// and extract the input value from the event data

	// Set output values - using default/dummy values for now
	ctx.SetOutputValue("inputValue", types.NewValue(types.PinTypes.String, "example input"))
	ctx.SetOutputValue("timestamp", types.NewValue(types.PinTypes.Number, float64(time.Now().UnixNano()/1e6)))

	// Activate execution flow
	return ctx.ActivateOutputFlow("execution")
}
