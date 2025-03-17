package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"webblueprint/internal/common"
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

// ExecutionMode defines the execution engine to use
type ExecutionMode string

const (
	ModeStandard ExecutionMode = "standard" // Original sequential execution
	ModeActor    ExecutionMode = "actor"    // Actor-based concurrent execution
)

// ExecutionEngine manages blueprint execution
type ExecutionEngine struct {
	OnAnyHook func(
		ctx context.Context,
		executionID, nodeID, level, message string,
		details map[string]interface{},
	) error

	// Add the hook for node execution recording
	OnNodeExecutionHook func(
		ctx context.Context,
		executionID, nodeID, nodeType string,
		inputs, outputs map[string]interface{},
	) error

	nodeRegistry    map[string]node.NodeFactory
	blueprints      map[string]*blueprint.Blueprint
	executionStatus map[string]*ExecutionStatus
	variables       map[string]map[string]types.Value // BlueprintID -> VariableName -> Value
	listeners       []ExecutionListener
	debugManager    *DebugManager
	logger          node.Logger
	executionMode   ExecutionMode
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
		executionMode:   ModeStandard, // Default to standard mode
	}
}

// SetExecutionMode sets the execution mode
func (e *ExecutionEngine) SetExecutionMode(mode ExecutionMode) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.executionMode = mode
}

func (e *ExecutionEngine) GetExecutionMode() ExecutionMode {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.executionMode
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

func (e *ExecutionEngine) GetLogger() node.Logger {
	return e.logger
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

// Execute runs a blueprint
func (e *ExecutionEngine) Execute(bp *blueprint.Blueprint, executionID string, initialData map[string]types.Value) (common.ExecutionResult, error) {
	e.mutex.Lock()

	blueprintID := bp.ID

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
	result := common.ExecutionResult{
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

	// Process variables first to ensure they're available to all nodes
	if err := e.processVariableNodes(bp, executionID, variables); err != nil {
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
		result.EndTime = time.Now()
		return result, err
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

	// Select execution method based on mode
	var err error
	executionMode := e.GetExecutionMode()

	if executionMode == ModeActor {
		err = e.executeWithActorSystem(bp, executionID, entryPoints, variables)
	} else {
		err = e.executeWithStandardEngine(bp, executionID, entryPoints, variables)
	}

	// Handle execution result
	if err != nil {
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
		result.EndTime = time.Now()
	} else {
		// Update execution status
		e.mutex.Lock()
		status.Status = "completed"
		status.EndTime = time.Now()
		e.mutex.Unlock()

		// Get node results from debug manager
		nodeResults := make(map[string]map[string]interface{})
		if nodeOutputs := e.debugManager.GetExecutionOutputValues(executionID); nodeOutputs != nil {
			for nodeID, outputs := range nodeOutputs {
				nodeResults[nodeID] = outputs
			}
		}

		// Emit execution end event
		e.EmitEvent(ExecutionEvent{
			Type:      EventExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"blueprintID": blueprintID,
				"executionID": executionID,
				"success":     true,
				"duration":    time.Since(status.StartTime).String(),
			},
		})

		result.Success = true
		result.EndTime = time.Now()
		result.NodeResults = nodeResults
	}

	return result, err
}

// Pre-process variable nodes before execution
func (e *ExecutionEngine) processVariableNodes(bp *blueprint.Blueprint, executionID string, variables map[string]types.Value) error {
	// Find all variable get and set nodes
	var setterNodes []string
	var getterNodes []string

	for _, nodeConfig := range bp.Nodes {
		// Check if this is a variable node (get or set)
		if strings.HasPrefix(nodeConfig.Type, "set-variable-") {
			setterNodes = append(setterNodes, nodeConfig.ID)
		} else if strings.HasPrefix(nodeConfig.Type, "get-variable-") {
			getterNodes = append(getterNodes, nodeConfig.ID)
		}
	}

	// Process set nodes first - to ensure variables are initialized
	for _, nodeID := range setterNodes {
		if err := e.executeVariableNode(nodeID, bp, executionID, variables); err != nil {
			return fmt.Errorf("error processing setter node %s: %w", nodeID, err)
		}
	}

	// Process get nodes second - to ensure they have the latest values
	for _, nodeID := range getterNodes {
		if err := e.executeVariableNode(nodeID, bp, executionID, variables); err != nil {
			return fmt.Errorf("error processing getter node %s: %w", nodeID, err)
		}
	}

	return nil
}

// Special execution for variable nodes that don't participate in execution flow
func (e *ExecutionEngine) executeVariableNode(nodeID string, bp *blueprint.Blueprint, executionID string, variables map[string]types.Value) error {
	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Get the node factory
	factory, exists := e.nodeRegistry[nodeConfig.Type]
	if !exists {
		return fmt.Errorf("node type not registered: %s", nodeConfig.Type)
	}

	// Create node instance
	nodeInstance := factory()

	// Create logger for this node
	nodeLogger := e.logger
	nodeLogger.Opts(map[string]interface{}{"nodeId": nodeID})

	// Get all input connections for this node
	inputConnections := bp.GetNodeInputConnections(nodeID)

	// Prepare input values
	inputValues := make(map[string]types.Value)

	// Process data connections to get input values
	for _, conn := range inputConnections {
		if conn.ConnectionType == "data" {
			// Get source node's output values (if already processed)
			sourceNodeID := conn.SourceNodeID
			sourcePinID := conn.SourcePinID
			targetPinID := conn.TargetPinID

			// Try to get the value from debug manager
			if outputValue, exists := e.debugManager.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); exists {
				// Convert to Value type
				var pinType *types.PinType
				switch outputValue.(type) {
				case string:
					pinType = types.PinTypes.String
				case float64, int:
					pinType = types.PinTypes.Number
				case bool:
					pinType = types.PinTypes.Boolean
				case map[string]interface{}:
					pinType = types.PinTypes.Object
				case []interface{}:
					pinType = types.PinTypes.Array
				default:
					pinType = types.PinTypes.Any
				}
				inputValues[targetPinID] = types.NewValue(pinType, outputValue)
			} else {
				// Try to execute the source node if it's a constant or another data node
				sourceNodeConfig := bp.FindNode(sourceNodeID)
				if sourceNodeConfig != nil {
					// Check if this is a constant or another non-executable node
					isDataNode := false
					nodeType := sourceNodeConfig.Type
					if strings.HasPrefix(nodeType, "constant-") ||
						strings.HasPrefix(nodeType, "get-variable-") ||
						strings.HasPrefix(nodeType, "set-variable-") {
						isDataNode = true
					}

					if isDataNode {
						// Execute this node first
						if err := e.executeVariableNode(sourceNodeID, bp, executionID, variables); err == nil {
							// Now try to get the output value again
							if outputValue, exists := e.debugManager.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); exists {
								var pinType *types.PinType
								switch outputValue.(type) {
								case string:
									pinType = types.PinTypes.String
								case float64, int:
									pinType = types.PinTypes.Number
								case bool:
									pinType = types.PinTypes.Boolean
								case map[string]interface{}:
									pinType = types.PinTypes.Object
								case []interface{}:
									pinType = types.PinTypes.Array
								default:
									pinType = types.PinTypes.Any
								}
								inputValues[targetPinID] = types.NewValue(pinType, outputValue)
							}
						}
					}
				}
			}
		}
	}

	// Create execution hooks
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			if e.OnAnyHook != nil {
				anyErr := e.OnAnyHook(context.Background(), executionID, nodeID, "info", string(EventNodeStarted), map[string]interface{}{
					"nodeType":  nodeType,
					"timestamp": time.Now(),
				})
				slog.Debug("[DEBUG] An error caught on any hook", slog.Any("error", anyErr))
			}

			// Record node execution with inputs
			if e.OnNodeExecutionHook != nil {
				inputMap := make(map[string]interface{})
				for pinID, value := range inputValues {
					inputMap[pinID] = value.RawValue
				}

				err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeType, inputMap, nil)
				if err != nil {
					slog.Debug("[DEBUG] Error recording node execution", slog.Any("error", err))
				}
			}

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
			if e.OnAnyHook != nil {
				anyErr := e.OnAnyHook(context.Background(), executionID, nodeID, "info", string(EventNodeCompleted), map[string]interface{}{
					"nodeType":  nodeType,
					"timestamp": time.Now(),
				})
				slog.Debug("[DEBUG] An error caught on any hook", slog.Any("error", anyErr))
			}

			// Record node execution with outputs
			if e.OnNodeExecutionHook != nil {
				// Collect output values
				outputMap, _ := e.debugManager.GetNodeDebugData(executionID, nodeID)

				err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeType, nil, outputMap)
				if err != nil {
					slog.Debug("[DEBUG] Error recording node execution completion", slog.Any("error", err))
				}
			}

			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeCompleted,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"nodeType": nodeType,
				},
			})
		},
		OnNodeError: func(nodeID string, err error) {
			if e.OnAnyHook != nil {
				anyErr := e.OnAnyHook(context.Background(), executionID, nodeID, "error", string(EventNodeError), map[string]interface{}{
					"error": err.Error(),
				})
				slog.Debug("[DEBUG] An error caught on any hook", slog.Any("error", anyErr))
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
			// Store the output value in debug manager
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinName, value)

			if e.OnAnyHook != nil {
				anyErr := e.OnAnyHook(context.Background(), executionID, nodeID, "debug", string(EventValueProduced), map[string]interface{}{
					"pinId": pinName,
					"value": value,
				})
				slog.Debug("[DEBUG] An error caught on any hook", slog.Any("error", anyErr))
			}

			// Emit event
			e.EmitEvent(ExecutionEvent{
				Type:      EventValueProduced,
				Timestamp: time.Now(),
				NodeID:    nodeID,
				Data: map[string]interface{}{
					"pinId": pinName,
					"value": value,
				},
			})
		},
	}

	// Custom function that does nothing since these nodes don't activate flow
	dummyActivateFlow := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		return nil
	}

	// Create execution context
	ctx := NewExecutionContext(
		nodeID,
		nodeConfig.Type,
		bp.ID,
		executionID,
		inputValues,
		variables,
		nodeLogger,
		hooks,
		dummyActivateFlow,
	)

	// Execute the node (this will set or get variables as needed)
	hooks.OnNodeStart(nodeID, nodeConfig.Type)
	err := nodeInstance.Execute(ctx)

	// Store debug data
	debugData := ctx.GetDebugData()
	e.debugManager.StoreNodeDebugData(executionID, nodeID, debugData)

	// Store outputs in debug manager
	for pinID, outValue := range ctx.outputs {
		e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinID, outValue.RawValue)
	}

	if err != nil {
		hooks.OnNodeError(nodeID, err)
		return err
	}

	hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	return nil
}

// executeWithActorSystem executes a blueprint using the actor system
func (e *ExecutionEngine) executeWithActorSystem(bp *blueprint.Blueprint, executionID string, entryPoints []string, variables map[string]types.Value) error {
	// Create an actor system for this execution
	actorSystem, err := NewActorSystem(
		executionID,
		bp,
		e.nodeRegistry,
		e.logger,
		e.listeners,
		e.debugManager,
		variables,
		e.OnNodeExecutionHook, // Pass the node execution hook
		e.OnAnyHook,           // Pass the any hook
	)
	if err != nil {
		return fmt.Errorf("failed to create actor system: %w", err)
	}

	// Initialize actor system
	if err := actorSystem.Start(bp); err != nil {
		return fmt.Errorf("failed to start actor system: %w", err)
	}

	if err := e.processVariableNodes(bp, executionID, variables); err != nil {
		// Emit execution end event
		e.EmitEvent(ExecutionEvent{
			Type:      EventExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"blueprintID":  bp.ID,
				"executionID":  executionID,
				"success":      false,
				"errorMessage": err.Error(),
			},
		})

		return err
	}

	// Execute the blueprint
	if err := actorSystem.Execute(entryPoints); err != nil {
		return fmt.Errorf("actor system execution failed: %w", err)
	}

	// Wait for completion with timeout (30 seconds)
	if !actorSystem.Wait(30 * time.Second) {
		actorSystem.Stop()
		return fmt.Errorf("actor system execution timed out")
	}

	// Get node statuses from actor system
	e.mutex.Lock()
	status := e.executionStatus[executionID]
	status.NodeStatuses = actorSystem.GetNodesStatus()
	e.mutex.Unlock()

	// Clean up resources
	actorSystem.Stop()

	return nil
}

// executeWithStandardEngine executes a blueprint using the standard engine
func (e *ExecutionEngine) executeWithStandardEngine(bp *blueprint.Blueprint, executionID string, entryPoints []string, variables map[string]types.Value) error {
	// Create hooks for tracking execution
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			e.mutex.Lock()
			status := e.executionStatus[executionID]
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
			status := e.executionStatus[executionID]
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
					"duration": time.Since(nodeStatus.StartTime).String(),
				},
			})
		},
		OnNodeError: func(nodeID string, err error) {
			e.mutex.Lock()
			status := e.executionStatus[executionID]
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
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinName, value)

			// Find connections from this output pin
			for _, conn := range bp.GetNodeOutputConnections(nodeID) {
				if conn.SourcePinID == pinName && conn.ConnectionType == "data" {
					// Emit value produced event
					e.EmitEvent(ExecutionEvent{
						Type:      EventValueProduced,
						Timestamp: time.Now(),
						NodeID:    nodeID,
						Data: map[string]interface{}{
							"sourceNodeId": conn.SourceNodeID,
							"sourcePinId":  conn.SourcePinID,
							"targetNodeId": conn.TargetNodeID,
							"targetPinId":  conn.TargetPinID,
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

	// Process variables first to ensure they're available to all nodes
	if err := e.processVariableNodes(bp, executionID, variables); err != nil {
		// Emit execution end event
		e.EmitEvent(ExecutionEvent{
			Type:      EventExecutionEnd,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"blueprintID":  bp.ID,
				"executionID":  executionID,
				"success":      false,
				"errorMessage": err.Error(),
			},
		})

		return err
	}

	// Process each entry point
	wg := sync.WaitGroup{}
	errors := make(chan error, len(entryPoints))

	for _, entryNodeID := range entryPoints {
		wg.Add(1)

		go func(nodeID string) {
			defer wg.Done()

			if err := e.executeNode(nodeID, bp, bp.ID, executionID, variables, hooks); err != nil {
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

	if errorCount > 0 {
		return lastError
	}

	return nil
}

// executeNode executes a single node in the standard engine
func (e *ExecutionEngine) executeNode(nodeID string, bp *blueprint.Blueprint, blueprintID, executionID string, variables map[string]types.Value, hooks *node.ExecutionHooks) error {
	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	e.mutex.RLock()
	factory, exists := e.nodeRegistry[nodeConfig.Type]
	e.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("node type not registered: %s", nodeConfig.Type)
	}

	// Create the node instance
	nodeInstance := factory()

	// Create execution context
	// Collect input values from connected nodes
	inputValues := make(map[string]types.Value)

	// Get input connections for this node
	inputConnections := bp.GetNodeInputConnections(nodeID)

	for _, conn := range inputConnections {
		if conn.ConnectionType == "data" {
			sourceNode := bp.FindNode(conn.SourceNodeID)
			if sourceNode != nil && strings.HasPrefix(sourceNode.Type, "get-variable-") {
				// This is a connection from a variable getter node
				// Extract variable name from the node type
				varName := strings.TrimPrefix(sourceNode.Type, "get-variable-")

				// Try to get the variable value
				if varValue, exists := variables[varName]; exists {
					// Add it directly to input values
					inputValues[conn.TargetPinID] = varValue

					// Also ensure the value is stored in debug manager for the source node
					e.debugManager.StoreNodeOutputValue(executionID, conn.SourceNodeID, conn.SourcePinID, varValue.RawValue)
				}
			}
		}
	}

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

	// Record node execution with inputs
	if e.OnNodeExecutionHook != nil {
		inputMap := make(map[string]interface{})
		for pinID, value := range inputValues {
			inputMap[pinID] = value.RawValue
		}

		err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeConfig.Type, inputMap, nil)
		if err != nil {
			slog.Debug("[DEBUG] Error recording node execution", slog.Any("error", err))
		}
	}

	// Create a function to activate output flows
	activateFlowFn := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		// Store all outputs before activating flows
		for _, pin := range nodeInstance.GetOutputPins() {
			if value, exists := ctx.GetOutputValue(pin.ID); exists {
				e.debugManager.StoreNodeOutputValue(executionID, nodeID, pin.ID, value.RawValue)
			}
		}

		// Find connections from this output pin
		outputConnections := bp.GetNodeOutputConnections(nodeID)
		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID

				// Execute the target node
				if err := e.executeNode(targetNodeID, bp, blueprintID, executionID, variables, hooks); err != nil {
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

	// Execute the node
	err := nodeInstance.Execute(ctx)

	// Collect output values
	outputMap := make(map[string]interface{})
	for _, pin := range nodeInstance.GetOutputPins() {
		if value, exists := ctx.GetOutputValue(pin.ID); exists {
			outputMap[pin.ID] = value.RawValue
		}
	}

	// Record node execution with outputs
	if e.OnNodeExecutionHook != nil {
		err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeConfig.Type, nil, outputMap)
		if err != nil {
			slog.Debug("[DEBUG] Error recording node execution completion", slog.Any("error", err))
		}
	}

	// Handle errors
	if err != nil {
		if hooks != nil && hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, err)
		}
		return err
	}

	// Store debug data
	e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

	// Follow execution flows
	activatedFlows := ctx.GetActivatedOutputFlows()
	for _, outputPin := range activatedFlows {
		if err := activateFlowFn(ctx, nodeID, outputPin); err != nil {
			return err
		}
	}

	// Notify node completion
	if hooks != nil && hooks.OnNodeComplete != nil {
		hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	return nil
}
