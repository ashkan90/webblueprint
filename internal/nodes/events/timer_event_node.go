package events

import (
	"fmt"
	"time"
	"webblueprint/internal/event"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// TimerEventNode sets up a timer event
type TimerEventNode struct {
	node.BaseNode
	timerID  string
	interval time.Duration
	count    int
}

// NewTimerEventNode creates a new timer event node
func NewTimerEventNode() node.Node {
	return &TimerEventNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "timer-event",
				Name:        "Timer Event",
				Description: "Dispatches events at specified intervals",
				Category:    "Events",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "execute",
					Name:        "Execute",
					Description: "Starts the timer",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "interval",
					Name:        "Interval (ms)",
					Description: "Time between events in milliseconds",
					Type:        types.PinTypes.Number,
					Default:     1000, // Default to 1 second
				},
				{
					ID:          "repeat",
					Name:        "Repeat",
					Description: "Number of times to trigger (0 for infinite)",
					Type:        types.PinTypes.Number,
					Optional:    true,
					Default:     0,
				},
				{
					ID:          "stop",
					Name:        "Stop",
					Description: "Stops the timer when triggered",
					Type:        types.PinTypes.Execution,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "started",
					Name:        "Started",
					Description: "Executed when the timer starts",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "stopped",
					Name:        "Stopped",
					Description: "Executed when the timer stops",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "timerID",
					Name:        "Timer ID",
					Description: "Unique identifier for the timer",
					Type:        types.PinTypes.String,
				},
			},
		},
		timerID:  "",
		interval: 1000 * time.Millisecond,
		count:    0,
	}
}

// Execute runs the node logic
func (n *TimerEventNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing TimerEventNode", nil)

	// Check which input pin triggered execution
	if ctx.IsInputPinActive("stop") {
		return n.stopTimer(ctx)
	}

	// Default is to start the timer
	return n.startTimer(ctx)
}

// startTimer starts the timer
func (n *TimerEventNode) startTimer(ctx node.ExecutionContext) error {
	logger := ctx.Logger()

	// Get input values
	intervalValue, intervalExists := ctx.GetInputValue("interval")
	repeatValue, repeatExists := ctx.GetInputValue("repeat")

	// Default values
	interval := float64(1000) // 1 second
	if intervalExists {
		intervalNum, err := intervalValue.AsNumber()
		if err == nil && intervalNum > 0 {
			interval = intervalNum
		}
	}

	repeat := 0 // Infinite
	if repeatExists {
		repeatNum, err := repeatValue.AsNumber()
		if err == nil && repeatNum >= 0 {
			repeat = int(repeatNum)
		}
	}

	// Convert to duration
	n.interval = time.Duration(interval) * time.Millisecond

	// Create a unique timer ID
	blueprintID := ctx.GetBlueprintID()
	nodeID := ctx.GetNodeID()
	n.timerID = fmt.Sprintf("%s.%s.timer.%d", blueprintID, nodeID, time.Now().UnixNano())

	// Get event manager
	var eventManager event.EventManagerInterface
	if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
		eventManager = evtCtx.GetEventManager()
	} else {
		logger.Error("Event manager not available in context", nil)
		return fmt.Errorf("event manager not available in context")
	}

	// Get the system timer event ID
	timerEventID, exists := eventManager.GetSystemEventID(event.EventTypeTimer)
	if !exists {
		logger.Error("Timer event not defined in system", nil)
		return fmt.Errorf("timer event not defined in system")
	}

	// Reset count
	n.count = 0

	// Set output
	ctx.SetOutputValue("timerID", types.NewValue(types.PinTypes.String, n.timerID))

	// Start timer in a goroutine
	go func() {
		ticker := time.NewTicker(n.interval)
		defer ticker.Stop()

		// Check if we should stop
		shouldStop := func() bool {
			if repeat > 0 && n.count >= repeat {
				return true
			}
			// If timerID is empty, it means the timer was stopped
			return n.timerID == ""
		}

		for {
			select {
			case <-ticker.C:
				// Increment count
				n.count++

				// Create parameters for the event
				params := map[string]types.Value{
					"blueprintID": types.NewValue(types.PinTypes.String, blueprintID),
					"executionID": types.NewValue(types.PinTypes.String, ctx.GetExecutionID()),
					"interval":    types.NewValue(types.PinTypes.Number, float64(n.interval/time.Millisecond)),
					"count":       types.NewValue(types.PinTypes.Number, float64(n.count)),
					"timerID":     types.NewValue(types.PinTypes.String, n.timerID),
				}

				// Dispatch the timer event
				if evtCtx, ok := ctx.(event.ExecutionContextWithEvents); ok {
					err := evtCtx.DispatchEvent(timerEventID, params)
					if err != nil {
						logger.Error("Failed to dispatch timer event", map[string]interface{}{
							"error":    err.Error(),
							"timerID":  n.timerID,
							"interval": n.interval,
							"count":    n.count,
						})
					}
				}

				// Check if we should stop
				if shouldStop() {
					// Activate the stopped output
					ctx.SetOutputValue("timerID", types.NewValue(types.PinTypes.String, n.timerID))
					ctx.ActivateOutputFlow("stopped")
					return
				}
			}
		}
	}()

	logger.Info("Timer started", map[string]interface{}{
		"timerID":  n.timerID,
		"interval": n.interval,
		"repeat":   repeat,
	})

	// Activate the started output
	return ctx.ActivateOutputFlow("started")
}

// stopTimer stops the timer
func (n *TimerEventNode) stopTimer(ctx node.ExecutionContext) error {
	logger := ctx.Logger()

	// Check if timer is running
	if n.timerID == "" {
		logger.Warn("No timer running", nil)
		return ctx.ActivateOutputFlow("stopped")
	}

	// Set timerID to empty to signal the timer to stop
	oldTimerID := n.timerID
	n.timerID = ""

	// Set output
	ctx.SetOutputValue("timerID", types.NewValue(types.PinTypes.String, oldTimerID))

	logger.Info("Timer stopped", map[string]interface{}{
		"timerID": oldTimerID,
		"count":   n.count,
	})

	// Activate the stopped output
	return ctx.ActivateOutputFlow("stopped")
}
