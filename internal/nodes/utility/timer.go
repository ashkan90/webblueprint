package utility

import (
	"fmt"
	"strconv"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// TimerNode implements a node for delay, timing, and scheduling operations
type TimerNode struct {
	node.BaseNode
}

// NewTimerNode creates a new Timer node
func NewTimerNode() node.Node {
	return &TimerNode{
		BaseNode: node.BaseNode{
			Metadata: node.NodeMetadata{
				TypeID:      "timer",
				Name:        "Timer",
				Description: "Delay execution, measure elapsed time, or schedule operations",
				Category:    "Utility",
				Version:     "1.0.0",
			},
			Inputs: []types.Pin{
				{
					ID:          "exec",
					Name:        "Execute",
					Description: "Execution input",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "operation",
					Name:        "Operation",
					Description: "Timer operation: delay, elapsed, timestamp",
					Type:        types.PinTypes.String,
					Default:     "delay",
				},
				{
					ID:          "duration",
					Name:        "Duration",
					Description: "Delay duration in milliseconds (for delay operation)",
					Type:        types.PinTypes.Number,
					Optional:    true,
					Default:     1000,
				},
				{
					ID:          "startTime",
					Name:        "Start Time",
					Description: "Start timestamp for elapsed time calculation",
					Type:        types.PinTypes.Any,
					Optional:    true,
				},
			},
			Outputs: []types.Pin{
				{
					ID:          "then",
					Name:        "Then",
					Description: "Execution continues after timer completes",
					Type:        types.PinTypes.Execution,
				},
				{
					ID:          "timestamp",
					Name:        "Timestamp",
					Description: "Current timestamp (Unix milliseconds)",
					Type:        types.PinTypes.Number,
				},
				{
					ID:          "formatted",
					Name:        "Formatted Time",
					Description: "Formatted current time (ISO 8601)",
					Type:        types.PinTypes.String,
				},
				{
					ID:          "elapsed",
					Name:        "Elapsed",
					Description: "Elapsed time in milliseconds",
					Type:        types.PinTypes.Number,
				},
			},
		},
	}
}

// Execute runs the node logic
func (n *TimerNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Timer node", nil)

	// Collect debug data
	debugData := make(map[string]interface{})

	// Get input values
	operationValue, operationExists := ctx.GetInputValue("operation")
	durationValue, durationExists := ctx.GetInputValue("duration")
	startTimeValue, startTimeExists := ctx.GetInputValue("startTime")

	// Default to "delay" operation if not specified
	operation := "delay"
	if operationExists {
		if opStr, err := operationValue.AsString(); err == nil {
			operation = opStr
		}
	}

	// Record input values for debugging
	debugData["inputs"] = map[string]interface{}{
		"operation":    operation,
		"hasDuration":  durationExists,
		"hasStartTime": startTimeExists,
	}

	// Current time for all operations
	now := time.Now()
	nowUnix := float64(now.UnixNano()) / 1e6 // Convert to milliseconds
	formatted := now.Format(time.RFC3339)

	// Set common outputs
	ctx.SetOutputValue("timestamp", types.NewValue(types.PinTypes.Number, nowUnix))
	ctx.SetOutputValue("formatted", types.NewValue(types.PinTypes.String, formatted))

	// Process based on operation
	switch operation {
	case "delay":
		// Delay execution for the specified duration
		duration := 1000.0 // Default to 1 second
		if durationExists {
			if durationNum, err := durationValue.AsNumber(); err == nil {
				duration = durationNum
			}
		}

		// Minimum delay of 1ms to avoid blocking
		if duration < 1 {
			duration = 1
		}

		logger.Info("Starting delay", map[string]interface{}{
			"duration": duration,
		})

		debugData["operation"] = "delay"
		debugData["duration"] = duration
		debugData["startTime"] = nowUnix

		// Record the start of the delay
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Timer Delay Started",
			Value:       debugData,
			Timestamp:   now,
		})

		// Sleep for the specified duration
		time.Sleep(time.Duration(duration) * time.Millisecond)

		// Update time values after the delay
		afterDelay := time.Now()
		afterDelayUnix := float64(afterDelay.UnixNano()) / 1e6
		afterDelayFormatted := afterDelay.Format(time.RFC3339)

		elapsed := afterDelayUnix - nowUnix

		// Update outputs after delay
		ctx.SetOutputValue("timestamp", types.NewValue(types.PinTypes.Number, afterDelayUnix))
		ctx.SetOutputValue("formatted", types.NewValue(types.PinTypes.String, afterDelayFormatted))
		ctx.SetOutputValue("elapsed", types.NewValue(types.PinTypes.Number, elapsed))

		// Record the end of the delay
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Timer Delay Completed",
			Value: map[string]interface{}{
				"duration":       duration,
				"actualDuration": elapsed,
				"endTime":        afterDelayUnix,
			},
			Timestamp: afterDelay,
		})

		logger.Info("Delay completed", map[string]interface{}{
			"duration":       duration,
			"actualDuration": elapsed,
		})

	case "elapsed":
		// Calculate elapsed time since the start time
		if !startTimeExists {
			// If no start time provided, use current time and set elapsed to 0
			ctx.SetOutputValue("elapsed", types.NewValue(types.PinTypes.Number, 0.0))

			logger.Warn("No start time provided for elapsed operation", nil)

			debugData["operation"] = "elapsed"
			debugData["warning"] = "No start time provided"
			debugData["elapsed"] = 0

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Timer Elapsed (No Start Time)",
				Value:       debugData,
				Timestamp:   now,
			})
		} else {
			// Parse the start time from the provided value
			var startTime time.Time
			var startTimeMs float64

			// Try different formats for the start time
			if startTimeStr, err := startTimeValue.AsString(); err == nil {
				// Try to parse as timestamp string in ISO format
				if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					startTime = parsed
					startTimeMs = float64(startTime.UnixNano()) / 1e6
				} else {
					// Try to parse as Unix timestamp (seconds or milliseconds)
					if unixSec, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
						// Determine if seconds or milliseconds based on length
						if len(startTimeStr) <= 10 { // seconds (10 digits or fewer)
							startTime = time.Unix(unixSec, 0)
							startTimeMs = float64(startTime.UnixNano()) / 1e6
						} else { // milliseconds
							nanos := unixSec * 1e6
							startTime = time.Unix(0, nanos)
							startTimeMs = float64(unixSec)
						}
					} else {
						// Fallback: Use current time
						startTime = now
						startTimeMs = nowUnix
						logger.Warn("Could not parse start time, using current time", map[string]interface{}{
							"startTime": startTimeStr,
						})
					}
				}
			} else if startTimeNum, err := startTimeValue.AsNumber(); err == nil {
				// Numeric timestamp (assume milliseconds)
				nanos := int64(startTimeNum * 1e6)
				startTime = time.Unix(0, nanos)
				startTimeMs = startTimeNum
			} else {
				// Fallback: Use current time
				startTime = now
				startTimeMs = nowUnix
				logger.Warn("Could not parse start time, using current time", map[string]interface{}{
					"startTime": startTimeValue.RawValue,
				})
			}

			// Calculate elapsed time in milliseconds
			elapsed := nowUnix - startTimeMs

			// Handle negative elapsed time (future date)
			if elapsed < 0 {
				elapsed = 0
				logger.Warn("Start time is in the future, elapsed time set to 0", map[string]interface{}{
					"startTime": startTime,
					"now":       now,
				})
			}

			ctx.SetOutputValue("elapsed", types.NewValue(types.PinTypes.Number, elapsed))

			logger.Info("Elapsed time calculated", map[string]interface{}{
				"startTime": startTime,
				"now":       now,
				"elapsed":   elapsed,
			})

			debugData["operation"] = "elapsed"
			debugData["startTime"] = startTimeMs
			debugData["currentTime"] = nowUnix
			debugData["elapsed"] = elapsed

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.GetNodeID(),
				Description: "Timer Elapsed Time",
				Value:       debugData,
				Timestamp:   now,
			})
		}

	case "timestamp":
		// Just return the current timestamp (already set above)
		logger.Info("Timestamp generated", map[string]interface{}{
			"timestamp": nowUnix,
			"formatted": formatted,
		})

		debugData["operation"] = "timestamp"
		debugData["timestamp"] = nowUnix
		debugData["formatted"] = formatted

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Timer Timestamp",
			Value:       debugData,
			Timestamp:   now,
		})

		ctx.SetOutputValue("elapsed", types.NewValue(types.PinTypes.Number, 0.0))

	default:
		// Unknown operation
		err := fmt.Errorf("unknown timer operation: %s", operation)
		logger.Error("Execution failed", map[string]interface{}{"error": err.Error()})

		debugData["error"] = err.Error()

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.GetNodeID(),
			Description: "Timer Error",
			Value:       debugData,
			Timestamp:   now,
		})

		// Set default output values
		ctx.SetOutputValue("elapsed", types.NewValue(types.PinTypes.Number, 0.0))
	}

	// Continue execution
	return ctx.ActivateOutputFlow("then")
}
