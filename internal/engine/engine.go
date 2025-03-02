package engine

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// ExecutionEvent represents an event during blueprint execution
type ExecutionEvent struct {
	Type      ExecutionEventType
	Timestamp time.Time
	NodeID    string
	Data      map[string]interface{}
}

// ExecutionEventType defines types of execution events
type ExecutionEventType string

const (
	EventNodeStarted    ExecutionEventType = "node.started"
	EventNodeCompleted  ExecutionEventType = "node.completed"
	EventNodeError      ExecutionEventType = "node.error"
	EventValueProduced  ExecutionEventType = "value.produced"
	EventValueConsumed  ExecutionEventType = "value.consumed"
	EventExecutionStart ExecutionEventType = "execution.start"
	EventExecutionEnd   ExecutionEventType = "execution.end"
	EventDebugData      ExecutionEventType = "debug.data"
)

// ExecutionListener listens for execution events
type ExecutionListener interface {
	OnExecutionEvent(event ExecutionEvent)
}

// ExecutionStatus represents the current state of a blueprint execution
type ExecutionStatus struct {
	ExecutionID  string
	Status       string // "running", "completed", "failed"
	StartTime    time.Time
	EndTime      time.Time
	NodeStatuses map[string]NodeStatus
}

// NodeStatus represents the execution status of a single node
type NodeStatus struct {
	NodeID    string
	Status    string // "idle", "executing", "completed", "error"
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// ExecutionResult represents the result of executing a blueprint
type ExecutionResult struct {
	Success     bool
	ExecutionID string
	StartTime   time.Time
	EndTime     time.Time
	Error       error
	NodeResults map[string]map[string]interface{} // NodeID -> PinID -> Value
}

// ExecutionEngine manages blueprint execution
type ExecutionEngine struct {
	nodeRegistry    map[string]node.NodeFactory
	blueprints      map[string]*blueprint.Blueprint
	executionStatus map[string]*ExecutionStatus
	variables       map[string]map[string]types.Value // BlueprintID -> VariableName -> Value
	listeners       []ExecutionListener
	debugManager    *DebugManager
	mutex           sync.RWMutex
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(debugManager *DebugManager) *ExecutionEngine {
	return &ExecutionEngine{
		nodeRegistry:    make(map[string]node.NodeFactory),
		blueprints:      make(map[string]*blueprint.Blueprint),
		executionStatus: make(map[string]*ExecutionStatus),
		variables:       make(map[string]map[string]types.Value),
		listeners:       make([]ExecutionListener, 0),
		debugManager:    debugManager,
	}
}

// RegisterNodeType registers a node type with the engine
func (e *ExecutionEngine) RegisterNodeType(typeID string, factory node.NodeFactory) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.nodeRegistry[typeID] = factory
}

// LoadBlueprint registers a blueprint with the engine
func (e *ExecutionEngine) LoadBlueprint(bp *blueprint.Blueprint) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Initialize variables for this blueprint
	e.variables[bp.ID] = make(map[string]types.Value)

	// Store the blueprint
	e.blueprints[bp.ID] = bp

	return nil
}

// AddExecutionListener adds a listener for execution events
func (e *ExecutionEngine) AddExecutionListener(listener ExecutionListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.listeners = append(e.listeners, listener)
}

// EmitEvent sends an event to all listeners
func (e *ExecutionEngine) EmitEvent(event ExecutionEvent) {
	e.mutex.RLock()
	listeners := make([]ExecutionListener, len(e.listeners))
	copy(listeners, e.listeners)
	e.mutex.RUnlock()

	for _, listener := range listeners {
		listener.OnExecutionEvent(event)
	}
}

// GetNodeStatus returns the current execution status of a node
func (e *ExecutionEngine) GetNodeStatus(executionID, nodeID string) (NodeStatus, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	status, exists := e.executionStatus[executionID]
	if !exists {
		return NodeStatus{}, false
	}

	nodeStatus, exists := status.NodeStatuses[nodeID]
	return nodeStatus, exists
}

// GetExecutionStatus returns the current execution status
func (e *ExecutionEngine) GetExecutionStatus(executionID string) (ExecutionStatus, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	status, exists := e.executionStatus[executionID]
	if !exists {
		return ExecutionStatus{}, false
	}

	return *status, true
}

// GetNodeDebugData returns debug data for a specific node
func (e *ExecutionEngine) GetNodeDebugData(executionID, nodeID string) (map[string]interface{}, bool) {
	return e.debugManager.GetNodeDebugData(executionID, nodeID)
}

// Execute runs a blueprint from the specified start node
func (e *ExecutionEngine) Execute(blueprintID string, initialData map[string]types.Value) (ExecutionResult, error) {
	e.mutex.Lock()

	// Check if blueprint exists
	bp, exists := e.blueprints[blueprintID]
	if !exists {
		e.mutex.Unlock()
		return ExecutionResult{}, fmt.Errorf("blueprint not found: %s", blueprintID)
	}

	// Create a unique execution ID
	executionID := fmt.Sprintf("%s-%d", blueprintID, time.Now().UnixNano())

	// Initialize execution status
	status := &ExecutionStatus{
		ExecutionID:  executionID,
		Status:       "running",
		StartTime:    time.Now(),
		NodeStatuses: make(map[string]NodeStatus),
	}
	e.executionStatus[executionID] = status
	e.mutex.Unlock()

	// Initialize result
	result := ExecutionResult{
		ExecutionID: executionID,
		StartTime:   status.StartTime,
		NodeResults: make(map[string]map[string]interface{}),
	}

	// Create execution context variables
	variables := make(map[string]types.Value)

	// Copy blueprint variables
	if vars, exists := e.variables[blueprintID]; exists {
		for k, v := range vars {
			variables[k] = v
		}
	}

	// Add any initial data
	for k, v := range initialData {
		variables[k] = v
	}

	// Emit execution start event
	e.EmitEvent(ExecutionEvent{
		Type:      EventExecutionStart,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"blueprintID": blueprintID,
			"executionID": executionID,
		},
	})

	// Find entry points
	entryPoints := bp.FindEntryPoints()
	if len(entryPoints) == 0 {
		err := fmt.Errorf("no entry points found in blueprint")
		// Update execution status
		e.mutex.Lock()
		status.Status = "failed"
		status.EndTime = time.Now()
		e.mutex.Unlock()

		// Emit execution end event
		e.EmitEvent(ExecutionEvent{
			Type:      EventExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"blueprintID":  blueprintID,
				"executionID":  executionID,
				"success":      false,
				"errorMessage": err.Error(),
			},
		})

		result.Success = false
		result.Error = err
		result.EndTime = status.EndTime

		return result, err
	}

	// Create hooks for tracking execution
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			e.mutex.Lock()
			status.NodeStatuses[nodeID] = NodeStatus{
				NodeID:    nodeID,
				Status:    "executing",
				StartTime: time.Now(),
			}
			e.mutex.Unlock()

			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeStarted,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"nodeType": nodeType,
				},
			})
		},
		OnNodeComplete: func(nodeID, nodeType string) {
			e.mutex.Lock()
			nodeStatus, exists := status.NodeStatuses[nodeID]
			if exists {
				nodeStatus.Status = "completed"
				nodeStatus.EndTime = time.Now()
				status.NodeStatuses[nodeID] = nodeStatus
			}
			e.mutex.Unlock()

			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeCompleted,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"nodeType": nodeType,
					"duration": nodeStatus.EndTime.Sub(nodeStatus.StartTime).String(),
				},
			})
		},
		OnNodeError: func(nodeID string, err error) {
			e.mutex.Lock()
			nodeStatus, exists := status.NodeStatuses[nodeID]
			if exists {
				nodeStatus.Status = "error"
				nodeStatus.Error = err
				nodeStatus.EndTime = time.Now()
				status.NodeStatuses[nodeID] = nodeStatus
			}
			e.mutex.Unlock()

			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeError,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			})
		},
		OnPinValue: func(nodeID, pinName string, value interface{}) {
			// Store value for result
			if _, exists := result.NodeResults[nodeID]; !exists {
				result.NodeResults[nodeID] = make(map[string]interface{})
			}
			result.NodeResults[nodeID][pinName] = value

			// Emit value produced event
			e.EmitEvent(ExecutionEvent{
				Type:      EventValueProduced,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"pinName": pinName,
					"value":   value,
				},
			})
		},
		OnLog: func(nodeID, message string) {
			// Emit log event as debug data
			e.EmitEvent(ExecutionEvent{
				Type:      EventDebugData,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"message": message,
				},
			})
		},
	}

	// Process each entry point
	wg := sync.WaitGroup{}
	errors := make(chan error, len(entryPoints))

	for _, entryNodeID := range entryPoints {
		wg.Add(1)

		go func(nodeID string) {
			defer wg.Done()

			if err := e.executeNode(nodeID, blueprintID, executionID, variables, hooks); err != nil {
				errors <- err
			}
		}(entryNodeID)
	}

	// Wait for all entry points to complete
	wg.Wait()
	close(errors)

	// Check for errors
	var lastError error
	errorCount := 0
	for err := range errors {
		errorCount++
		lastError = err
	}

	// Update execution status
	e.mutex.Lock()
	if errorCount > 0 {
		status.Status = "failed"
	} else {
		status.Status = "completed"
	}
	status.EndTime = time.Now()
	e.mutex.Unlock()

	// Set result
	result.Success = errorCount == 0
	result.EndTime = status.EndTime
	if errorCount > 0 {
		result.Error = lastError
	}

	// Emit execution end event
	e.EmitEvent(ExecutionEvent{
		Type:      EventExecutionEnd,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"blueprintID": blueprintID,
			"executionID": executionID,
			"success":     result.Success,
			"errorCount":  errorCount,
			"duration":    result.EndTime.Sub(result.StartTime).String(),
		},
	})

	return result, lastError
}

// executeNode executes a single node
func (e *ExecutionEngine) executeNode(nodeID, blueprintID, executionID string, variables map[string]types.Value, hooks *node.ExecutionHooks) error {
	e.mutex.RLock()
	bp, exists := e.blueprints[blueprintID]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("blueprint not found: %s", blueprintID)
	}

	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Get the node factory
	e.mutex.RLock()
	factory, exists := e.nodeRegistry[nodeConfig.Type]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("node type not registered: %s", nodeConfig.Type)
	}

	// Create the node instance
	nodeInstance := factory()

	// Create execution context
	logger := NewDefaultLogger(nodeID)

	// Collect input values from connected nodes
	inputValues := make(map[string]types.Value)

	// Get input connections for this node
	inputConnections := bp.GetNodeInputConnections(nodeID)

	// Process data connections
	for _, conn := range inputConnections {
		if conn.ConnectionType == "data" {
			// Get the source node's output value
			sourceNodeID := conn.SourceNodeID
			sourcePinID := conn.SourcePinID
			targetPinID := conn.TargetPinID

			// Check if we have a result for this pin
			if nodeResults, ok := e.debugManager.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); ok {
				// Convert to Value
				// This is simplified - we'd need to determine the type properly
				inputValues[targetPinID] = types.NewValue(types.PinTypes.Any, nodeResults)

				// Emit value consumed event
				e.EmitEvent(ExecutionEvent{
					Type:      EventValueConsumed,
					Timestamp: time.Now(),
					NodeID:    nodeID,
					Data: map[string]interface{}{
						"sourceNodeID": sourceNodeID,
						"sourcePinID":  sourcePinID,
						"targetPinID":  targetPinID,
						"value":        nodeResults,
					},
				})
			}
		}
	}

	// Create a function to activate output flows
	activateFlow := func(nodeID, pinID string) error {
		// Find connections from this output pin
		for _, conn := range bp.GetNodeOutputConnections(nodeID) {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID

				// Execute the target node
				if err := e.executeNode(targetNodeID, blueprintID, executionID, variables, hooks); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Create execution context
	ctx := NewExecutionContext(
		nodeID,
		nodeConfig.Type,
		blueprintID,
		executionID,
		inputValues,
		variables,
		logger,
		hooks,
		activateFlow,
	)

	// Notify node start
	if hooks != nil && hooks.OnNodeStart != nil {
		hooks.OnNodeStart(nodeID, nodeConfig.Type)
	}

	// Execute the node
	err := nodeInstance.Execute(ctx)

	// Store debug data
	e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

	// Store output values
	for _, pin := range nodeInstance.GetOutputPins() {
		if value, exists := ctx.GetOutputValue(pin.ID); exists {
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pin.ID, value.RawValue)
		}
	}

	// Notify node completion or error
	if err != nil {
		if hooks != nil && hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, err)
		}
		return err
	}

	if hooks != nil && hooks.OnNodeComplete != nil {
		hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	return nil
}
