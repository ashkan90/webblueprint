package engine

import (
	"fmt"
	"log"
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
	logger          node.Logger
	mutex           sync.RWMutex
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(logger node.Logger, debugManager *DebugManager) *ExecutionEngine {
	return &ExecutionEngine{
		nodeRegistry:    make(map[string]node.NodeFactory),
		blueprints:      make(map[string]*blueprint.Blueprint),
		executionStatus: make(map[string]*ExecutionStatus),
		variables:       make(map[string]map[string]types.Value),
		listeners:       make([]ExecutionListener, 0),
		logger:          logger,
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
			defer e.mutex.Unlock()
			nodeStatus, exists := status.NodeStatuses[nodeID]
			if !exists {
				return
				//nodeStatus.Status = "error"
				//nodeStatus.Error = err
				//nodeStatus.EndTime = time.Now()
				//status.NodeStatuses[nodeID] = nodeStatus
			}

			if nodeStatus.Status == "failed" {
				nodeStatus.NodeID = nodeID
				nodeStatus.Error = err
				nodeStatus.EndTime = time.Now()
				status.NodeStatuses[nodeID] = nodeStatus
			}

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

			_nodeConnections := bp.GetNodeOutputConnections(nodeID)
			for _, connection := range _nodeConnections {
				if connection.ConnectionType == "data" {
					// Emit value produced event
					e.EmitEvent(ExecutionEvent{
						Type:      EventValueProduced,
						Timestamp: time.Now(),
						NodeID:    connection.ID,
						Data: map[string]interface{}{
							"sourceNodeId": connection.SourceNodeID,
							"sourcePinId":  connection.SourcePinID,
							"targetNodeId": connection.TargetNodeID,
							"targetPinId":  connection.TargetPinID,
							"timestamp":    time.Now(),
							"value":        value,
						},
					})
				}
			}
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
	log.Printf("[ENGINE] Processing input connections for node %s", nodeID)

	// Collect input values from connected nodes
	inputValues := make(map[string]types.Value)

	// Get input connections for this node
	inputConnections := bp.GetNodeInputConnections(nodeID)

	log.Printf("[ENGINE] Node %s has %d input connections", nodeID, len(inputConnections))
	for i, conn := range inputConnections {
		log.Printf("[ENGINE] Connection %d: %s.%s -> %s.%s (type: %s)",
			i, conn.SourceNodeID, conn.SourcePinID, conn.TargetNodeID, conn.TargetPinID, conn.ConnectionType)
	}

	// Process data connections
	for _, conn := range inputConnections {
		if conn.ConnectionType == "data" {
			// Get the source node's output value
			sourceNodeID := conn.SourceNodeID
			sourcePinID := conn.SourcePinID
			targetPinID := conn.TargetPinID

			log.Printf("[ENGINE] Processing data connection from %s.%s to %s.%s",
				sourceNodeID, sourcePinID, nodeID, targetPinID)

			// Check if we have a result for this pin
			if nodeResults, ok := e.debugManager.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); ok {
				log.Printf("[ENGINE] Found output value from %s.%s: %v (type: %T)",
					sourceNodeID, sourcePinID, nodeResults, nodeResults)

				// Convert to Value
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
			} else {
				log.Printf("[ENGINE] No output value found for %s.%s", sourceNodeID, sourcePinID)
				log.Printf("[ENGINE] Lookup data from %s.%s %v %%!d(bool=%v)",
					sourceNodeID, sourcePinID, nodeResults, ok)
			}
		}
	}

	// Create a function to activate output flows that maintains proper execution order
	// This will be passed to the execution context but used later after outputs are stored
	activateFlowFn := func(nodeID, pinID string) error {
		log.Printf("[ENGINE] Activating output flow for %s.%s", nodeID, pinID)

		// Find connections from this output pin
		outputConnections := bp.GetNodeOutputConnections(nodeID)

		log.Printf("[ENGINE] Node %s has %d output connections", nodeID, len(outputConnections))

		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID
				log.Printf("[ENGINE] Following execution connection to node %s", targetNodeID)

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
		e.logger,
		hooks,
		activateFlowFn,
	)

	// Notify node start
	if hooks != nil && hooks.OnNodeStart != nil {
		hooks.OnNodeStart(nodeID, nodeConfig.Type)
	}

	log.Printf("[ENGINE] Executing node %s (type: %s)", nodeID, nodeConfig.Type)

	// Execute the node
	err := nodeInstance.Execute(ctx)

	// Collect all output values that need to be stored and activated
	outputValues := make(map[string]map[string]types.Value)
	activateOutputPins := make([]string, 0)

	// Store debug data
	e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

	// CRITICAL: Store output values BEFORE following execution paths
	log.Printf("[ENGINE] Storing output values for node %s", nodeID)
	for _, pin := range nodeInstance.GetOutputPins() {
		if value, exists := ctx.GetOutputValue(pin.ID); exists {
			log.Printf("[ENGINE] Storing output value for %s.%s: %v (type: %T)",
				nodeID, pin.ID, value.RawValue, value.RawValue)

			// Store the value
			outputValues[nodeID] = map[string]types.Value{
				pin.ID: value,
			}

			// Store in debug manager for data flow
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pin.ID, value.RawValue)
		} else {
			log.Printf("[ENGINE] No output value set for pin %s on node %s", pin.ID, nodeID)
		}
	}

	// Handle execution paths that were stored during node.Execute
	// Get these from the custom ExecutionContext implementation
	activateOutputPins = ctx.GetActivatedOutputFlows()
	log.Printf("[ENGINE] Node %s has %d activated output flows", nodeID, len(activateOutputPins))

	// Notify node completion or error
	if err != nil {
		if hooks != nil && hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, err)
		}
		log.Printf("[ENGINE] Node %s execution failed: %v", nodeID, err)
		return err
	}

	if hooks != nil && hooks.OnNodeComplete != nil {
		hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	// AFTER storing all outputs, now follow execution flows
	for _, outputPin := range activateOutputPins {
		if err := activateFlowFn(nodeID, outputPin); err != nil {
			return err
		}
	}

	log.Printf("[ENGINE] Node %s execution completed successfully", nodeID)
	return nil
}
