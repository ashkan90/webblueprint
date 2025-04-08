package engine

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"
	"webblueprint/internal/engineext"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// NodeActor represents a single node in the actor model system
type NodeActor struct {
	NodeID       string
	NodeType     string
	ExecutionID  string
	node         node.Node
	bp           *blueprint.Blueprint
	mailbox      chan NodeMessage
	inputs       map[string]types.Value
	outputs      map[string]types.Value
	variables    map[string]types.Value
	status       NodeStatus
	ctx          *ActorExecutionContext
	logger       node.Logger
	listeners    []ExecutionListener
	debugMgr     *DebugManager
	done         chan struct{}
	mutex        sync.RWMutex
	properties   []types.Property      // Store node properties from the blueprint
	decoratedCtx node.ExecutionContext // Store the potentially decorated context passed to Start

	// Add hooks
	nodeExecutionHook func(ctx context.Context, executionID, nodeID, nodeType, execState string,
		inputs, outputs map[string]interface{}) error
	anyHook func(ctx context.Context, executionID, nodeID, level, message string,
		details map[string]interface{}) error
}

// NodeMessage represents a message that can be sent to a NodeActor
type NodeMessage struct {
	Type     string                 // Message type: "execute", "input", "stop"
	PinID    string                 // Target pin ID for input messages
	Value    types.Value            // Value for input messages
	Response chan NodeResponse      // Channel for the response
	FlowData map[string]interface{} // Additional flow data
}

// NodeResponse is the response to a NodeMessage
type NodeResponse struct {
	Success    bool
	Error      error
	OutputPins map[string]types.Value
}

// NewNodeActor creates a new actor for a node
func NewNodeActor(
	nodeID, nodeType string,
	bp *blueprint.Blueprint,
	executionID string,
	nodeInstance node.Node,
	logger node.Logger,
	listeners []ExecutionListener,
	debugMgr *DebugManager,
	variables map[string]types.Value,
	nodeExecutionHook func(ctx context.Context, executionID, nodeID, nodeType, execState string,
		inputs, outputs map[string]interface{}) error,
	anyHook func(ctx context.Context, executionID, nodeID, level, message string,
		details map[string]interface{}) error,
) *NodeActor {
	// Create a deep copy of the variable map to avoid concurrency issues
	varsCopy := make(map[string]types.Value)
	for k, v := range variables {
		varsCopy[k] = v
	}

	// Load node properties from blueprint
	var properties []types.Property
	if node := bp.FindNode(nodeID); node != nil {
		propertyLog := make(map[string]interface{})
		for _, property := range node.Properties {
			nodeInstance.SetProperty(property.Name, property.Value)
			properties = append(properties, types.Property{
				Name:  property.Name,
				Value: property.Value,
			})
			propertyLog["property."+property.Name] = property.Value
		}
		maps.Copy(propertyLog, map[string]interface{}{
			"nodeId":           nodeID,
			"properties.count": len(properties),
		})
		logger.Debug("Loaded node properties", propertyLog)
	}

	return &NodeActor{
		NodeID:      nodeID,
		NodeType:    nodeType,
		ExecutionID: executionID,
		node:        nodeInstance,
		bp:          bp,
		mailbox:     make(chan NodeMessage, 1024), // Buffer for handling multiple messages
		inputs:      make(map[string]types.Value),
		outputs:     make(map[string]types.Value),
		variables:   varsCopy,
		status: NodeStatus{
			NodeID: nodeID,
			Status: "idle",
		},
		logger:            logger,
		listeners:         listeners,
		debugMgr:          debugMgr,
		done:              make(chan struct{}),
		properties:        properties,
		nodeExecutionHook: nodeExecutionHook,
		anyHook:           anyHook,
	}
}

// Helper interface to check if a context wraps another
type contextWrapper interface {
	Unwrap() node.ExecutionContext
}

// Start begins processing messages from the mailbox
func (a *NodeActor) Start(ctx node.ExecutionContext) {
	if ctx == nil {
		return
	}

	// Store the potentially decorated context if provided
	a.decoratedCtx = ctx

	// Always create a specific ActorExecutionContext for internal actor use
	a.ctx = NewActorExecutionContext(
		a.NodeID,
		a.NodeType,
		a.bp.ID,
		a.ExecutionID,
		a.logger,
		a, // Pass the actor itself
	)

	// Start processing messages
	go a.processMessages()
}

// Stop gracefully stops the actor
func (a *NodeActor) Stop() {
	close(a.done)
}

// Send sends a message to the actor and waits for a response
func (a *NodeActor) Send(msg NodeMessage) NodeResponse {
	// Create a response channel if not provided
	if msg.Response == nil {
		responseChan := make(chan NodeResponse, 1)
		msg.Response = responseChan
	}

	select {
	case a.mailbox <- msg:
		// Message sent, wait for response
		select {
		case response, ok := <-msg.Response:
			if !ok {
				// Channel was closed
				return NodeResponse{
					Success: false,
					Error:   fmt.Errorf("response channel closed"),
				}
			}
			return response
		case <-time.After(10 * time.Second):
			// Timeout waiting for response
			return NodeResponse{
				Success: false,
				Error:   fmt.Errorf("timeout waiting for node response"),
			}
		}
	case <-time.After(2 * time.Second):
		// Timeout sending message
		return NodeResponse{
			Success: false,
			Error:   fmt.Errorf("mailbox full, node is not responding"),
		}
	}
}

// SendAsync sends a message to the actor without waiting for a response
func (a *NodeActor) SendAsync(msg NodeMessage) bool {
	select {
	case a.mailbox <- msg:
		// Message sent
		return true
	case <-time.After(1 * time.Second):
		// Mailbox full, log an error
		a.logger.Error("Failed to send message to node actor, mailbox full", map[string]interface{}{
			"nodeId":  a.NodeID,
			"msgType": msg.Type,
		})
		return false
	}
}

// GetProperty retrieves a property value by name
func (a *NodeActor) GetProperty(name string) (interface{}, bool) {
	for _, prop := range a.properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}
	return nil, false
}

// processMessages handles messages from the mailbox
func (a *NodeActor) processMessages() {
	for {
		select {
		case <-a.done:
			// Actor is being stopped
			return
		case msg, ok := <-a.mailbox:
			if !ok {
				// Mailbox was closed
				return
			}
			// Process the message
			response := a.handleMessage(msg)

			// Send the response if a response channel was provided
			if msg.Response != nil {
				select {
				case msg.Response <- response:
					// Response sent
				default:
					// Response channel is full or closed
					a.logger.Warn("Could not send response, channel may be full or closed", map[string]interface{}{
						"nodeId":  a.NodeID,
						"msgType": msg.Type,
					})
				}
			}
		}
	}
}

// handleMessage processes a single message
func (a *NodeActor) handleMessage(msg NodeMessage) NodeResponse {
	switch msg.Type {
	case "execute":
		return a.handleExecuteMessage(msg)
	case "input":
		return a.handleInputMessage(msg)
	case "stop":
		return a.handleStopMessage(msg)
	default:
		return NodeResponse{
			Success: false,
			Error:   fmt.Errorf("unknown message type: %s", msg.Type),
		}
	}
}

// handleExecuteMessage handles an execute message
func (a *NodeActor) handleExecuteMessage(msg NodeMessage) NodeResponse {
	// Update node status
	a.mutex.Lock()
	a.status.Status = "executing"
	a.status.StartTime = time.Now()
	a.mutex.Unlock()

	// Mark node as executing
	a.emitNodeStartedEvent()

	// Record node execution with inputs
	if a.nodeExecutionHook != nil {
		// Collect input values
		a.mutex.RLock()
		inputMap := make(map[string]interface{})
		for pinID, value := range a.inputs {
			inputMap[pinID] = value.RawValue
		}
		a.mutex.RUnlock()

		// Call the hook
		err := a.nodeExecutionHook(context.Background(), a.ExecutionID, a.NodeID, a.NodeType, "executing", inputMap, nil)
		if err != nil {
			a.logger.Debug("Error in node execution hook", map[string]interface{}{
				"nodeId": a.NodeID,
				"error":  err.Error(),
			})
		}
	}

	// Execute the node using the potentially decorated context stored from Start()
	err := a.node.Execute(a.decoratedCtx)

	a.outputs = engineext.GetExtendedContext(a.decoratedCtx).GetAllOutputs()

	// Get the outputs
	a.mutex.RLock()
	outputs := make(map[string]types.Value)
	for k, v := range a.outputs {
		outputs[k] = v
	}
	a.mutex.RUnlock()

	// Record node execution with outputs
	if a.nodeExecutionHook != nil {
		// Collect output values
		outputMap := make(map[string]interface{})
		for pinID, value := range outputs {
			outputMap[pinID] = value.RawValue
		}

		// Call the hook
		err := a.nodeExecutionHook(context.Background(), a.ExecutionID, a.NodeID, a.NodeType, "completed", nil, outputMap)
		if err != nil {
			a.logger.Debug("Error in node execution hook", map[string]interface{}{
				"nodeId": a.NodeID,
				"error":  err.Error(),
			})
		}
	}

	// Store outputs in debug manager
	if a.debugMgr != nil {
		// Store all outputs for this node
		for pinID, value := range outputs {
			a.debugMgr.StoreNodeOutputValue(a.ExecutionID, a.NodeID, pinID, value.RawValue)
		}
	}

	// Update node status based on execution result
	a.mutex.Lock()
	a.status.EndTime = time.Now()
	if err != nil {
		a.status.Status = "error"
		a.status.Error = err
	} else {
		a.status.Status = "completed"
	}
	a.mutex.Unlock()

	// Mark node as completed or error
	if err != nil {
		a.emitNodeErrorEvent(err)
		return NodeResponse{
			Success:    false,
			Error:      err,
			OutputPins: outputs,
		}
	}

	a.emitNodeCompletedEvent()
	return NodeResponse{
		Success:    true,
		OutputPins: outputs,
	}
}

// handleInputMessage handles an input message
func (a *NodeActor) handleInputMessage(msg NodeMessage) NodeResponse {
	if msg.PinID == "" {
		a.logger.Error("Input message missing pin ID", nil)
		return NodeResponse{
			Success: false,
			Error:   fmt.Errorf("input message missing pin ID"),
		}
	}

	if msg.Value.Type == nil {
		a.logger.Warn("Received nil type for pin", map[string]interface{}{
			"pinId": msg.PinID,
		})
	}

	// Store the input value - even if it's an empty string, as long as the type is valid
	a.mutex.Lock()
	a.inputs[msg.PinID] = msg.Value
	a.mutex.Unlock()

	// Store input in the execution context
	a.ctx.SetInput(msg.PinID, msg.Value)
	engineext.GetExtendedContext(a.decoratedCtx).SetInput(msg.PinID, msg.Value)

	// Log the received input for debugging
	a.logger.Debug("Received input value", map[string]interface{}{
		"pinId": msg.PinID,
		//"valueType": fmt.Sprintf("%T", msg.Value.RawValue),
		//"value":     msg.Value.RawValue,
	})

	// Return success
	return NodeResponse{
		Success: true,
	}
}

// handleStopMessage handles a stop message
func (a *NodeActor) handleStopMessage(msg NodeMessage) NodeResponse {
	// Stop the actor
	a.Stop()

	// Return success
	return NodeResponse{
		Success: true,
	}
}

// emitNodeStartedEvent emits an event when the node starts executing
func (a *NodeActor) emitNodeStartedEvent() {
	for _, listener := range a.listeners {
		listener.OnExecutionEvent(ExecutionEvent{
			Type:      EventNodeStarted,
			Timestamp: time.Now(),
			NodeID:    a.NodeID,
			Data: map[string]interface{}{
				"nodeType": a.NodeType,
				"status":   "executing",
			},
		})
	}

	// Use the anyHook if provided
	if a.anyHook != nil {
		a.anyHook(context.Background(), a.ExecutionID, a.NodeID, "info", string(EventNodeStarted), map[string]interface{}{
			"nodeType":  a.NodeType,
			"timestamp": time.Now(),
		})
	}
}

// emitNodeCompletedEvent emits an event when the node completes execution
func (a *NodeActor) emitNodeCompletedEvent() {
	for _, listener := range a.listeners {
		listener.OnExecutionEvent(ExecutionEvent{
			Type:      EventNodeCompleted,
			Timestamp: time.Now(),
			NodeID:    a.NodeID,
			Data: map[string]interface{}{
				"nodeType": a.NodeType,
				"status":   "completed",
			},
		})
	}

	// Use the anyHook if provided
	if a.anyHook != nil {
		a.anyHook(context.Background(), a.ExecutionID, a.NodeID, "info", string(EventNodeCompleted), map[string]interface{}{
			"nodeType":  a.NodeType,
			"timestamp": time.Now(),
		})
	}
}

// emitNodeErrorEvent emits an event when the node encounters an error
func (a *NodeActor) emitNodeErrorEvent(err error) {
	for _, listener := range a.listeners {
		listener.OnExecutionEvent(ExecutionEvent{
			Type:      EventNodeError,
			Timestamp: time.Now(),
			NodeID:    a.NodeID,
			Data: map[string]interface{}{
				"nodeType": a.NodeType,
				"status":   "error",
				"error":    err.Error(),
			},
		})
	}

	// Use the anyHook if provided
	if a.anyHook != nil {
		a.anyHook(context.Background(), a.ExecutionID, a.NodeID, "error", string(EventNodeError), map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// emitValueProducedEvent emits an event when a value is produced on an output pin
func (a *NodeActor) emitValueProducedEvent(pinID string, value interface{}) {
	for _, listener := range a.listeners {
		listener.OnExecutionEvent(ExecutionEvent{
			Type:      EventValueProduced,
			Timestamp: time.Now(),
			NodeID:    a.NodeID,
			Data: map[string]interface{}{
				"pinId": pinID,
				"value": value,
			},
		})
	}

	// Use the anyHook if provided
	if a.anyHook != nil {
		a.anyHook(context.Background(), a.ExecutionID, a.NodeID, "debug", string(EventValueProduced), map[string]interface{}{
			"pinId": pinID,
			"value": value,
		})
	}
}

// SetOutput sets an output value
func (a *NodeActor) SetOutput(pinID string, value types.Value) {
	a.mutex.Lock()
	a.outputs[pinID] = value
	a.mutex.Unlock()

	// Store in debug manager
	if a.debugMgr != nil {
		a.debugMgr.StoreNodeOutputValue(a.ExecutionID, a.NodeID, pinID, value.RawValue)
	}

	// Emit value produced event
	a.emitValueProducedEvent(pinID, value.RawValue)
}

// GetOutput gets an output value
func (a *NodeActor) GetOutput(pinID string) (types.Value, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	value, exists := a.outputs[pinID]
	return value, exists
}

// GetStatus returns the current status of the node
func (a *NodeActor) GetStatus() NodeStatus {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.status
}

// ActorExecutionContext is the execution context for a node actor
type ActorExecutionContext struct {
	nodeID         string
	nodeType       string
	blueprintID    string
	executionID    string
	inputs         map[string]types.Value
	variables      map[string]types.Value
	debugData      map[string]interface{}
	logger         node.Logger
	actor          *NodeActor
	activatedFlows []string
	activePin      string // The pin that triggered execution
	mutex          sync.RWMutex
}

// NewActorExecutionContext creates a new execution context for a node actor
func NewActorExecutionContext(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	logger node.Logger,
	actor *NodeActor,
) *ActorExecutionContext {
	// Copy the actor's inputs to initialize the context
	inputs := make(map[string]types.Value)
	actor.mutex.RLock()
	for k, v := range actor.inputs {
		inputs[k] = v
	}
	actor.mutex.RUnlock()

	return &ActorExecutionContext{
		nodeID:         nodeID,
		nodeType:       nodeType,
		blueprintID:    blueprintID,
		executionID:    executionID,
		inputs:         inputs,
		variables:      actor.variables, // Share variables with the actor
		debugData:      make(map[string]interface{}),
		logger:         logger,
		actor:          actor,
		activatedFlows: make([]string, 0),
	}
}

// SetInput sets an input value
func (ctx *ActorExecutionContext) SetInput(pinID string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.inputs[pinID] = value
}

// GetActivatedOutputFlows returns the list of activated output flows
func (ctx *ActorExecutionContext) GetActivatedOutputFlows() []string {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Return a copy to avoid concurrent modification issues
	flows := make([]string, len(ctx.activatedFlows))
	copy(flows, ctx.activatedFlows)
	return flows
}

func (ctx *ActorExecutionContext) GetBlueprint() *blueprint.Blueprint {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	return ctx.actor.bp
}

// Implementation of node.ExecutionContext interface for ActorExecutionContext

// IsInputPinActive checks if the input pin triggered execution
func (ctx *ActorExecutionContext) IsInputPinActive(pinID string) bool {
	// If a specific active pin is set, check it
	if ctx.activePin != "" {
		return ctx.activePin == pinID
	}

	// Otherwise, for backward compatibility, assume the default execute pin is active
	return pinID == "execute"
}

// GetInputValue retrieves an input value by pin ID
func (ctx *ActorExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	ctx.mutex.RLock()
	value, exists := ctx.inputs[pinID]
	ctx.mutex.RUnlock()

	// If the value exists, return it
	if exists {
		//// Log the input access
		//if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
		//	ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
		//}
		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.nodeID,
			PinID:       pinID,
			Description: "Existing value used",
			Value: map[string]interface{}{
				"default": value.RawValue,
				"source":  "node property",
			},
			Timestamp: time.Now(),
		})
		return value, true
	}

	// If the value doesn't exist in direct inputs, try to find it from connected variable nodes
	// Get the blueprint
	bp := ctx.GetBlueprint()
	// Get input connections for this node
	inputConnections := bp.GetNodeInputConnections(ctx.GetNodeID())

	// Look for connections to this pin from variable nodes
	for _, conn := range inputConnections {
		if conn.TargetPinID == pinID && conn.ConnectionType == "data" {
			// Check if the source node is a variable getter
			sourceNode := bp.FindNode(conn.SourceNodeID)
			if sourceNode != nil && strings.HasPrefix(sourceNode.Type, "get-variable-") {
				// Extract variable name from the node type
				varName := strings.TrimPrefix(sourceNode.Type, "get-variable-")

				// Try to get the variable value from the execution context
				if varValue, varExists := ctx.GetVariable(varName); varExists {
					return varValue, true
				}
			}

			if sourceNode != nil {
				for _, property := range sourceNode.Properties {
					var propType = ctx.resolveConstantType(ctx.nodeType)
					return types.NewValue(propType, property.Value), true
				}
			}
		}
	}

	// Check for property-based defaults

	// First, check if there's a node property for this input
	// Check for "input_[pinID]" property (used for print's message input)
	if propValue, exists := ctx.getPropertyValue(fmt.Sprintf("input_%s", pinID)); exists {
		// Get appropriate pin type based on the node and pin
		pinType := ctx.getPinTypeForInput(pinID)
		defaultValue := types.NewValue(pinType, propValue)

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.nodeID,
			PinID:       pinID,
			Description: "Property default value used",
			Value: map[string]interface{}{
				"default": propValue,
				"source":  "node property input_" + pinID,
			},
			Timestamp: time.Now(),
		})

		return defaultValue, true
	}

	// For constant nodes, check "constantValue" property
	if strings.HasPrefix(ctx.nodeType, "constant-") && (pinID == "value" || pinID == "constantValue") {
		if propValue, exists := ctx.getPropertyValue("constantValue"); exists {
			// Determine the type based on the node type
			var pinType = ctx.resolveConstantType(ctx.nodeType)
			// şş

			defaultValue := types.NewValue(pinType, propValue)

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.nodeID,
				PinID:       pinID,
				Description: "Constant default value used",
				Value: map[string]interface{}{
					"default": propValue,
					"source":  "constantValue property",
				},
				Timestamp: time.Now(),
			})

			return defaultValue, true
		}
	}

	// Check for a loop-specific property
	if propValue, exists := ctx.getPropertyValue(fmt.Sprintf("_loop_%s", pinID)); exists {
		defaultValue := types.NewValue(types.PinTypes.Any, propValue)

		ctx.RecordDebugInfo(types.DebugInfo{
			NodeID:      ctx.nodeID,
			PinID:       pinID,
			Description: "Loop default value used",
			Value: map[string]interface{}{
				"default": propValue,
				"source":  "loop property",
			},
			Timestamp: time.Now(),
		})

		return defaultValue, true
	}

	// Check for defaults from node definition if we still haven't found a value
	for _, pin := range ctx.actor.node.GetInputPins() {
		if pin.ID == pinID && pin.Default != nil {
			defaultValue := types.NewValue(pin.Type, pin.Default)

			ctx.RecordDebugInfo(types.DebugInfo{
				NodeID:      ctx.nodeID,
				PinID:       pinID,
				Description: "Pin default value used",
				Value: map[string]interface{}{
					"default": pin.Default,
					"source":  "pin definition",
				},
				Timestamp: time.Now(),
			})

			return defaultValue, true
		}
	}

	// No value or default found
	return types.Value{}, false
}

func (ctx *ActorExecutionContext) resolveConstantType(nodeType string) *types.PinType {
	switch nodeType {
	case "constant-string":
		return types.PinTypes.String
	case "constant-number":
		return types.PinTypes.Number
	case "constant-boolean":
		return types.PinTypes.Boolean
	default:
		return types.PinTypes.Any
	}
}

// getPropertyValue retrieves a property value from the actor's properties
func (ctx *ActorExecutionContext) getPropertyValue(name string) (interface{}, bool) {
	return ctx.actor.GetProperty(name)
}

// getPinTypeForInput determines the appropriate type for an input pin
func (ctx *ActorExecutionContext) getPinTypeForInput(pinID string) *types.PinType {
	// Try to get the type from pin definitions
	for _, pin := range ctx.actor.node.GetInputPins() {
		if pin.ID == pinID {
			return pin.Type
		}
	}

	// Default to appropriate type based on node type and pin ID
	if ctx.nodeType == "print" && pinID == "message" {
		return types.PinTypes.Any // Print accepts any type
	} else if strings.HasPrefix(ctx.nodeType, "constant-") {
		// Set type based on constant type
		switch ctx.nodeType {
		case "constant-string":
			return types.PinTypes.String
		case "constant-number":
			return types.PinTypes.Number
		case "constant-boolean":
			return types.PinTypes.Boolean
		}
	}

	// Default to Any if we can't determine a specific type
	return types.PinTypes.Any
}

// SetOutputValue sets an output value by pin ID
func (ctx *ActorExecutionContext) SetOutputValue(pinID string, value types.Value) {
	// Store the output value in the actor
	ctx.actor.SetOutput(pinID, value)

	// Log the output for debugging
	ctx.logger.Debug("Set output value", map[string]interface{}{
		"pinId":     pinID,
		"valueType": fmt.Sprintf("%T", value.RawValue),
		"value":     value.RawValue,
	})
}

// ActivateOutputFlow activates an output execution flow
func (ctx *ActorExecutionContext) ActivateOutputFlow(pinID string) error {
	ctx.mutex.Lock()
	ctx.activatedFlows = append(ctx.activatedFlows, pinID)
	ctx.mutex.Unlock()

	// We'll handle actually activating the flow in the actor system
	return nil
}

// ExecuteConnectedNodes executes nodes connected to the given output pin
// This is needed for some node types that require synchronous execution
func (ctx *ActorExecutionContext) ExecuteConnectedNodes(pinID string) error {
	// In the actor model, we can't directly execute connected nodes
	// Instead, we just mark the flow as activated and let the system handle it
	return ctx.ActivateOutputFlow(pinID)
}

// GetVariable retrieves a variable by name
func (ctx *ActorExecutionContext) GetVariable(name string) (types.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	value, exists := ctx.variables[name]
	return value, exists
}

// SetVariable sets a variable by name
func (ctx *ActorExecutionContext) SetVariable(name string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.variables[name] = value
}

// Logger returns the execution logger
func (ctx *ActorExecutionContext) Logger() node.Logger {
	return ctx.logger
}

// RecordDebugInfo stores debug information
func (ctx *ActorExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// Add the debug info to our collection
	key := fmt.Sprintf("debug_%d", time.Now().UnixNano())

	ctx.mutex.Lock()
	ctx.debugData[key] = info
	ctx.mutex.Unlock()

	// Store in debug manager if available
	if ctx.actor.debugMgr != nil {
		ctx.actor.debugMgr.StoreNodeDebugData(ctx.executionID, ctx.nodeID, map[string]interface{}{
			key: info,
		})
	}
}

// GetDebugData returns all debug data
func (ctx *ActorExecutionContext) GetDebugData() map[string]interface{} {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Create a copy to avoid concurrency issues
	dataCopy := make(map[string]interface{})
	for k, v := range ctx.debugData {
		dataCopy[k] = v
	}

	return dataCopy
}

// GetNodeID returns the ID of the executing node
func (ctx *ActorExecutionContext) GetNodeID() string {
	return ctx.nodeID
}

// GetNodeType returns the type of the executing node
func (ctx *ActorExecutionContext) GetNodeType() string {
	return ctx.nodeType
}

// GetBlueprintID returns the ID of the executing blueprint
func (ctx *ActorExecutionContext) GetBlueprintID() string {
	return ctx.blueprintID
}

// GetExecutionID returns the current execution ID
func (ctx *ActorExecutionContext) GetExecutionID() string {
	return ctx.executionID
}
