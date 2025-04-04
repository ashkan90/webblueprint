package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"webblueprint/internal/common"
	"webblueprint/internal/core"
	"webblueprint/internal/engineext"
	"webblueprint/internal/event" // Add event import
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
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

	// Extensions for context management
	extensions *engineext.ExecutionEngineExtensions

	// Add the hook for node execution recording
	OnNodeExecutionHook func(
		ctx context.Context,
		executionID, nodeID, nodeType, execState string,
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
	hooks           *node.ExecutionHooks // Keep track of hooks for the current execution
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

// LoadBlueprint registers a blueprint with the engine and auto-binds event listeners
func (e *ExecutionEngine) LoadBlueprint(bp *blueprint.Blueprint) error {
	// --- Step 1: Update Engine State (Requires Lock) ---
	e.mutex.Lock()
	// Initialize variables for this blueprint if not already present
	if _, exists := e.variables[bp.ID]; !exists {
		e.variables[bp.ID] = make(map[string]types.Value)
	}
	// Store the blueprint
	e.blueprints[bp.ID] = bp
	// Get extensions reference while holding lock
	extensions := e.extensions
	logger := e.logger // Get logger reference
	e.mutex.Unlock()   // --- Release Lock ---

	// --- Step 2: Get Event Manager (Does not require engine lock) ---
	var concreteEventManager *event.EventManager
	if extensions != nil {
		concreteEventManager = extensions.GetConcreteEventManager()
	}

	if concreteEventManager == nil {
		logger.Warn("Cannot register custom events or bind listeners: EventManager not available.", map[string]interface{}{"blueprintId": bp.ID})
		return nil // Allow loading blueprint even if event manager isn't ready
	}

	for _, definition := range bp.Events {
		eventParams := make([]event.EventParameter, len(definition.Parameters))
		for _, parameter := range definition.Parameters {
			pinType, _ := types.GetPinTypeByID(parameter.TypeID)
			eventParams = append(eventParams, event.EventParameter{
				Name:        parameter.Name,
				Type:        pinType,
				Description: parameter.Description,
				Optional:    parameter.Optional,
				Default:     parameter.Default,
			})
		}
		rErr := concreteEventManager.RegisterEvent(event.EventDefinition{
			ID:          definition.ID,
			Name:        definition.Name,
			Description: definition.Description,
			Parameters:  eventParams,
			Category:    definition.Category,
			BlueprintID: bp.ID,
			CreatedAt:   time.Now(),
		})
		if rErr != nil {
			logger.Error("An error occurred while event registration", map[string]interface{}{
				"blueprintId": bp.ID,
				"name":        definition.Name,
				"description": definition.Description,
				"parameters":  eventParams,
				"category":    definition.Category,
				"error":       rErr.Error(),
			})
			continue
		}
		logger.Info("Registered custom event", map[string]interface{}{
			"blueprintId": bp.ID,
			"name":        definition.Name,
			"description": definition.Description,
			"parameters":  eventParams,
			"category":    definition.Category,
		})
	}

	for _, binding := range bp.EventBindings {
		bErr := concreteEventManager.BindEvent(event.EventBinding{
			ID:          binding.ID,
			EventID:     binding.EventID,
			HandlerID:   binding.HandlerID,
			HandlerType: binding.HandlerType,
			BlueprintID: bp.ID,
			Priority:    binding.Priority,
			CreatedAt:   time.Now(),
			Enabled:     binding.Enabled,
		})
		if bErr != nil {
			logger.Error("An error occurred while event binding", map[string]interface{}{
				"blueprintId": bp.ID,
				"eventID":     binding.EventID,
				"handlerID":   binding.HandlerID,
				"handlerType": binding.HandlerType,
				"error":       bErr.Error(),
			})
			continue
		}

		logger.Info("EventBind successful", map[string]interface{}{
			"blueprintId": bp.ID,
			"eventID":     binding.EventID,
			"handlerID":   binding.HandlerID,
			"handlerType": binding.HandlerType,
		})
	}

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

// SetExtensions sets the engine extensions for context management
func (e *ExecutionEngine) SetExtensions(extensions *engineext.ExecutionEngineExtensions) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.extensions = extensions
}

// GetExtensions returns the engine extensions
func (e *ExecutionEngine) GetExtensions() *engineext.ExecutionEngineExtensions {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.extensions
}

// BasicExecutionContext is a minimal implementation of the node.ExecutionContext interface
// This is used as a fallback when no extension context is available
type BasicExecutionContext struct {
	nodeID       string
	nodeType     string
	blueprintID  string
	executionID  string
	inputs       map[string]types.Value
	outputs      map[string]types.Value
	variables    map[string]types.Value
	logger       node.Logger
	hooks        *node.ExecutionHooks
	activateFlow func(ctx *BasicExecutionContext, nodeID, pinID string) error
}

// NewBasicExecutionContext creates a new basic execution context
func NewBasicExecutionContext(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *BasicExecutionContext, nodeID, pinID string) error,
) *BasicExecutionContext {
	return &BasicExecutionContext{
		nodeID:       nodeID,
		nodeType:     nodeType,
		blueprintID:  blueprintID,
		executionID:  executionID,
		inputs:       inputs,
		outputs:      make(map[string]types.Value),
		variables:    variables,
		logger:       logger,
		hooks:        hooks,
		activateFlow: activateFlow,
	}
}

// GetNodeID returns the ID of the node
func (c *BasicExecutionContext) GetNodeID() string {
	return c.nodeID
}

// GetNodeType returns the type of the node
func (c *BasicExecutionContext) GetNodeType() string {
	return c.nodeType
}

// GetBlueprintID returns the ID of the blueprint
func (c *BasicExecutionContext) GetBlueprintID() string {
	return c.blueprintID
}

// GetExecutionID returns the ID of the execution
func (c *BasicExecutionContext) GetExecutionID() string {
	return c.executionID
}

// GetInputValue gets an input value by pin ID
func (c *BasicExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := c.inputs[pinID]
	if exists && c.hooks != nil && c.hooks.OnPinValue != nil {
		c.hooks.OnPinValue(c.nodeID, pinID, value.RawValue)
	}
	return value, exists
}

// IsInputPinActive checks if an input pin is active (assuming the default pin is always active)
func (c *BasicExecutionContext) IsInputPinActive(pinID string) bool {
	// In basic context, we assume the default pin is always active
	return pinID == "execute"
}

// SetOutputValue sets an output value by pin ID
func (c *BasicExecutionContext) SetOutputValue(pinID string, value types.Value) {
	c.outputs[pinID] = value
	if c.hooks != nil && c.hooks.OnPinValue != nil {
		c.hooks.OnPinValue(c.nodeID, pinID, value.RawValue)
	}
}

// GetVariable gets a variable by name
func (c *BasicExecutionContext) GetVariable(name string) (types.Value, bool) {
	value, exists := c.variables[name]
	return value, exists
}

// SetVariable sets a variable by name
func (c *BasicExecutionContext) SetVariable(name string, value types.Value) {
	c.variables[name] = value
}

// ActivateOutputFlow activates an output execution flow
func (c *BasicExecutionContext) ActivateOutputFlow(pinID string) error {
	return c.activateFlow(c, c.nodeID, pinID)
}

// ExecuteConnectedNodes executes nodes connected to an output pin
func (c *BasicExecutionContext) ExecuteConnectedNodes(pinID string) error {
	return c.activateFlow(c, c.nodeID, pinID)
}

// Logger returns the logger
func (c *BasicExecutionContext) Logger() node.Logger {
	return c.logger
}

// RecordDebugInfo records debug information
func (c *BasicExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// In the basic implementation, we just log the debug info
	c.logger.Debug("Debug info", map[string]interface{}{
		"nodeID":      info.NodeID,
		"pinID":       info.PinID,
		"description": info.Description,
		"value":       info.Value,
	})
}

// GetDebugData returns debug data
func (c *BasicExecutionContext) GetDebugData() map[string]interface{} {
	// Basic context doesn't store debug data
	return map[string]interface{}{}
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

// TriggerNodeExecution starts the execution of a specific node, typically an event handler.
// This method implements the core.EngineController interface.
func (e *ExecutionEngine) TriggerNodeExecution(blueprintID string, nodeID string, triggerContext core.EventHandlerContext) error {
	e.mutex.RLock()
	bp, bpExists := e.blueprints[blueprintID]
	// extensions := e.extensions // No longer needed here
	e.mutex.RUnlock()

	if !bpExists {
		return fmt.Errorf("TriggerNodeExecution: blueprint %s not loaded", blueprintID)
	}

	// We need variables, but hooks might not apply directly when triggering
	// Get variables associated with the blueprint instance if possible
	e.mutex.RLock()
	variables := make(map[string]types.Value)
	if blueprintVars, ok := e.variables[blueprintID]; ok {
		for k, v := range blueprintVars {
			variables[k] = v // Copy variables
		}
	}
	e.mutex.RUnlock()

	executionID := triggerContext.ExecutionID
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nID, nodeType string) {
			e.mutex.Lock()
			status, ok := e.executionStatus[executionID]
			if ok {
				status.NodeStatuses[nID] = NodeStatus{
					NodeID:    nID,
					Status:    "executing",
					StartTime: time.Now(),
				}
			}
			e.mutex.Unlock()
			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeStarted,
				Timestamp: time.Now(),
				NodeID:    nID,
				Data:      map[string]interface{}{"nodeType": nodeType},
			})
		},
		OnNodeComplete: func(nID, nodeType string) {
			e.mutex.Lock()
			status, ok := e.executionStatus[executionID]
			if ok {
				nodeStatus, exists := status.NodeStatuses[nID]
				if exists {
					nodeStatus.Status = "completed"
					nodeStatus.EndTime = time.Now()
					status.NodeStatuses[nID] = nodeStatus
				}
			}
			e.mutex.Unlock()
			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeCompleted,
				Timestamp: time.Now(),
				NodeID:    nID,
				Data:      map[string]interface{}{"nodeType": nodeType},
			})
		},
		OnNodeError: func(nID string, err error) {
			e.mutex.Lock()
			status, ok := e.executionStatus[executionID]
			if ok {
				nodeStatus, exists := status.NodeStatuses[nID]
				if exists {
					nodeStatus.Status = "error"
					nodeStatus.Error = err
					nodeStatus.EndTime = time.Now()
					status.NodeStatuses[nID] = nodeStatus
				}
			}
			e.mutex.Unlock()
			e.EmitEvent(ExecutionEvent{
				Type:      EventNodeError,
				Timestamp: time.Now(),
				NodeID:    nID,
				Data:      map[string]interface{}{"error": err.Error()},
			})
		},
		OnPinValue: func(nID, pinName string, value interface{}) {
			e.debugManager.StoreNodeOutputValue(executionID, nID, pinName, value)
			// Optionally emit EventValueProduced here if needed for event-triggered flows
		},
		OnLog: func(nID, message string) {
			e.EmitEvent(ExecutionEvent{
				Type:      EventDebugData,
				Timestamp: time.Now(),
				NodeID:    nID,
				Data:      map[string]interface{}{"message": message},
			})
		},
	}

	e.logger.Info("Triggering event handler node execution", map[string]interface{}{"nodeId": nodeID, "eventId": triggerContext.EventID})

	//entryPoints := []string{triggerContext.HandlerID}
	//err := e.executeWithActorSystem(bp, executionID, entryPoints, variables)
	// Call executeNode, passing the triggerContext and the newly defined hooks.
	err := e.executeNode(triggerContext.HandlerID, bp, blueprintID, executionID, variables, hooks, &triggerContext) // Pass hooks
	if err != nil {
		e.logger.Error("Error executing triggered event handler node", map[string]interface{}{"nodeId": nodeID, "error": err.Error()})
		// Potentially call OnNodeError hook here as well if executeNode fails immediately
		if hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, err)
		}
		return err
	}

	return nil
}

// GetNodeDebugData returns debug data for a specific node
func (e *ExecutionEngine) GetNodeDebugData(executionID, nodeID string) (map[string]interface{}, bool) {
	return e.debugManager.GetNodeDebugData(executionID, nodeID)
}

// Execute runs a blueprint
func (e *ExecutionEngine) Execute(bp *blueprint.Blueprint, executionID string, initialData map[string]types.Value) (common.ExecutionResult, error) {
	e.nodeRegistry = registry.GetInstance().GetAllNodeFactories()

	// Load the blueprint (this will register event bindings)
	if err := e.LoadBlueprint(bp); err != nil {
		// Create minimal error result
		return common.ExecutionResult{
			ExecutionID: executionID,
			Success:     false,
			Error:       fmt.Errorf("failed to load blueprint: %w", err),
			StartTime:   time.Now(), // Or get from somewhere?
			EndTime:     time.Now(),
		}, err
	}
	// Keep mutex locked for status/variable initialization? No, LoadBlueprint unlocks. Lock again.
	blueprintID := bp.ID

	// Initialize execution status
	status := &ExecutionStatus{
		ExecutionID:  executionID,
		Status:       "running",
		StartTime:    time.Now(),
		NodeStatuses: make(map[string]NodeStatus),
	}
	e.executionStatus[executionID] = status

	// Initialize result
	result := common.ExecutionResult{
		ExecutionID: executionID,
		StartTime:   status.StartTime,
		NodeResults: make(map[string]map[string]interface{}),
	}

	// Create execution context variables
	variables := make(map[string]types.Value)

	// Copy blueprint variables
	// Need lock? Assume variables are read-only after LoadBlueprint or copied safely
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

	// Define hooks for this execution
	e.hooks = &node.ExecutionHooks{
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

				err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeType, "executing", inputMap, nil)
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

				err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeType, "completed", nil, outputMap)
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
	dummyActivateFlow := func(ctx *engineext.DefaultExecutionContext, nodeID, pinID string) error {
		return nil
	}

	// Create execution context
	ctx := engineext.NewExecutionContext(
		nodeID,
		nodeConfig.Type,
		bp.ID,
		executionID,
		inputValues,
		variables,
		nodeLogger,
		hooks,
		dummyActivateFlow,
		context.Background(), // Add context.Context argument
	)

	// Execute the node (this will set or get variables as needed)
	hooks.OnNodeStart(nodeID, nodeConfig.Type)
	err := nodeInstance.Execute(ctx)

	// Store debug data
	debugData := ctx.GetDebugData()
	e.debugManager.StoreNodeDebugData(executionID, nodeID, debugData)

	// Store outputs in debug manager using the helper function
	if extCtx := engineext.GetExtendedContext(ctx); extCtx != nil {
		outputs := extCtx.GetAllOutputs()
		for pinID, outValue := range outputs {
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinID, outValue.RawValue)
		}
	} else {
		e.logger.Warn("Could not retrieve ExtendedExecutionContext to store outputs", map[string]interface{}{"nodeId": nodeID, "contextType": fmt.Sprintf("%T", ctx)})
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
		e.GetExtensions().GetContextManager(),
		executionID,
		bp,
		e.nodeRegistry,
		e.logger,
		e.listeners,
		e.debugManager,
		variables,
		e.hooks,
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

	for _, nodeID := range entryPoints { // Use nodeID directly in loop variable
		wg.Add(1)

		go func(currentNodeID string) { // Pass nodeID as argument
			defer wg.Done()
			// Pass e.hooks and nil for triggerCtx in standard execution flow
			if err := e.executeNode(currentNodeID, bp, bp.ID, executionID, variables, e.hooks, nil); err != nil { // Pass e.hooks and nil triggerCtx
				errors <- err
			}
		}(nodeID) // Pass nodeID to the goroutine
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
// createGenericActivateFlow creates a function that can work with any context type
func (e *ExecutionEngine) createGenericActivateFlow(
	activateFlowFn func(ctx *engineext.DefaultExecutionContext, nodeID string, pinID string) error,
) func(ctx node.ExecutionContext, nodeID string, pinID string) error {
	return func(ctx node.ExecutionContext, nodeID string, pinID string) error {
		// Try to convert to DefaultExecutionContext
		if defaultCtx, ok := ctx.(*engineext.DefaultExecutionContext); ok {
			return activateFlowFn(defaultCtx, nodeID, pinID)
		}
		// If not possible, just log and continue
		e.logger.Warn("Cannot activate flow due to incompatible context type", map[string]interface{}{
			"nodeID": nodeID,
			"pinID":  pinID,
		})
		return nil
	}
}

func (e *ExecutionEngine) executeNode(nodeID string, bp *blueprint.Blueprint, blueprintID, executionID string, variables map[string]types.Value, hooks *node.ExecutionHooks, triggerCtx *core.EventHandlerContext) error { // Re-added hooks, Added triggerCtx
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

		err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeConfig.Type, "executing", inputMap, nil)
		if err != nil {
			slog.Debug("[DEBUG] Error recording node execution", slog.Any("error", err))
		}
	}

	// Create a function to activate output flows
	activateFlowFn := func(ctx *engineext.DefaultExecutionContext, nodeID, pinID string) error {
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
				// Pass the correct 'hooks' variable down
				// Pass nil for triggerCtx when activating flow normally
				if err := e.executeNode(targetNodeID, bp, blueprintID, executionID, variables, hooks, nil); err != nil { // Pass hooks and nil triggerCtx
					return err
				}
			}
		}
		return nil
	}

	// --- Create Execution Context using ContextManager ---
	var ctx node.ExecutionContext
	extensions := e.GetExtensions()
	contextManager := (*engineext.ContextManager)(nil) // Initialize to nil

	// Safely get the ContextManager from extensions
	// Check if extensions is not nil
	if extensions != nil {
		contextManager = extensions.GetContextManager()
	}

	if contextManager == nil {
		// If context manager is unavailable, we cannot proceed correctly.
		// Log an error and potentially return, or use a very basic context.
		e.logger.Error("ContextManager not available in engine extensions, cannot create proper context", map[string]interface{}{"nodeId": nodeID})
		// Fallback to a very basic context that likely won't work for complex nodes
		// Ensure NewExecutionContext exists and has the correct signature
		// Pass the 'hooks' parameter from executeNode and context.Background()
		ctx = engineext.NewExecutionContext(nodeID, nodeConfig.Type, blueprintID, executionID, inputValues, variables, e.logger, hooks, activateFlowFn, context.Background()) // Pass hooks and context.Background()
	} else {
		// Use the ContextManager to create the appropriate context
		if triggerCtx != nil {
			// Create context for an event handler trigger
			// Pass event parameters as initial inputs?
			eventInputs := make(map[string]types.Value)
			for k, v := range inputValues { // Start with regular inputs
				eventInputs[k] = v
			}
			for k, v := range triggerCtx.Parameters { // Add/overwrite with event params
				eventInputs[k] = v
			}

			// Add bp argument
			ctx = contextManager.CreateEventHandlerContext(
				bp, // Pass blueprint
				nodeID,
				nodeConfig.Type,
				blueprintID,
				executionID, // Use execution ID from trigger? Or main execution? Needs clarification. Using main for now.
				eventInputs, // Pass combined inputs
				variables,
				e.logger,
				hooks,          // Pass the 'hooks' parameter from executeNode
				activateFlowFn, // Pass the correct activateFlowFn
				triggerCtx,
			)
		} else {
			// Create standard context for normal execution flow
			// Add bp argument
			ctx = contextManager.CreateStandardContext(
				bp, // Pass blueprint
				nodeID,
				nodeConfig.Type,
				blueprintID,
				executionID,
				inputValues,
				variables,
				e.logger,
				hooks,          // Pass the 'hooks' parameter from executeNode
				activateFlowFn, // Pass the correct activateFlowFn
			)
		}
	}
	// --- End Context Creation ---

	// Notify node start
	if e.hooks != nil && e.hooks.OnNodeStart != nil {
		e.hooks.OnNodeStart(nodeID, nodeConfig.Type)
	}

	//ctx.SaveData("node.properties", actor.properties)
	ctx.SaveData("node.inputPins", nodeInstance.GetInputPins())

	// Execute the node
	err := nodeInstance.Execute(ctx)

	// Collect output values
	outputMap := make(map[string]interface{})
	// Use helper function to find the ExtendedExecutionContext
	if extCtx := engineext.GetExtendedContext(ctx); extCtx != nil {
		outputs := extCtx.GetAllOutputs()
		for pinID, outValue := range outputs {
			outputMap[pinID] = outValue.RawValue
		}
	} else {
		e.logger.Warn("Could not retrieve ExtendedExecutionContext to store outputs", map[string]interface{}{"nodeId": nodeID, "contextType": fmt.Sprintf("%T", ctx)})
	}

	// Record node execution with outputs
	if e.OnNodeExecutionHook != nil {
		err := e.OnNodeExecutionHook(context.Background(), executionID, nodeID, nodeConfig.Type, "completed", nil, outputMap)
		if err != nil {
			slog.Debug("[DEBUG] Error recording node execution completion", slog.Any("error", err))
		}
	}

	// Handle errors
	if err != nil {
		if e.hooks != nil && e.hooks.OnNodeError != nil {
			e.hooks.OnNodeError(nodeID, err)
		}
		return err
	}

	// Store debug data
	e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

	// Follow execution flows
	var activatedFlows []string
	// Use helper function to find the ExtendedExecutionContext
	if extCtx := engineext.GetExtendedContext(ctx); extCtx != nil {
		activatedFlows = extCtx.GetActivatedOutputFlows()
	} else {
		// Log if we couldn't get activated flows (might indicate context issue)
		e.logger.Warn("Could not retrieve ExtendedExecutionContext to get activated flows", map[string]interface{}{"nodeId": nodeID, "contextType": fmt.Sprintf("%T", ctx)})
		// activatedFlows remains empty
	}

	for _, outputPin := range activatedFlows {
		// Skip flow activation for non-DefaultExecutionContext
		if defaultCtx, ok := ctx.(*engineext.DefaultExecutionContext); ok {
			if err := activateFlowFn(defaultCtx, nodeID, outputPin); err != nil {
				return err
			}
		}
	}

	// Notify node completion
	if e.hooks != nil && e.hooks.OnNodeComplete != nil {
		e.hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	return nil
}
