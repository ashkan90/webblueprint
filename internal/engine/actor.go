package engine

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// NodeActor represents a single node in the actor model system
type NodeActor struct {
	NodeID      string
	NodeType    string
	BlueprintID string
	ExecutionID string
	node        node.Node
	mailbox     chan NodeMessage
	outputs     map[string]types.Value
	ctx         *ActorExecutionContext
	logger      node.Logger
	listeners   []ExecutionListener
	debugMgr    *DebugManager
	done        chan struct{}
	mutex       sync.RWMutex
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
	nodeID, nodeType, blueprintID, executionID string,
	nodeInstance node.Node,
	logger node.Logger,
	listeners []ExecutionListener,
	debugMgr *DebugManager,
) *NodeActor {
	return &NodeActor{
		NodeID:      nodeID,
		NodeType:    nodeType,
		BlueprintID: blueprintID,
		ExecutionID: executionID,
		node:        nodeInstance,
		mailbox:     make(chan NodeMessage, 50), // Buffer for handling multiple messages
		outputs:     make(map[string]types.Value),
		logger:      logger,
		listeners:   listeners,
		debugMgr:    debugMgr,
		done:        make(chan struct{}),
	}
}

// Start begins processing messages from the mailbox
func (a *NodeActor) Start() {
	// Initialize execution context
	a.ctx = NewActorExecutionContext(
		a.NodeID,
		a.NodeType,
		a.BlueprintID,
		a.ExecutionID,
		a.logger,
		a,
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
		defer close(responseChan)
	}

	// Send the message
	select {
	case a.mailbox <- msg:
		// Message sent, wait for response
		select {
		case response := <-msg.Response:
			return response
		case <-time.After(5 * time.Second):
			// Timeout waiting for response
			return NodeResponse{
				Success: false,
				Error:   fmt.Errorf("timeout waiting for node response"),
			}
		}
	case <-time.After(1 * time.Second):
		// Timeout sending message
		return NodeResponse{
			Success: false,
			Error:   fmt.Errorf("mailbox full, node is not responding"),
		}
	}
}

// SendAsync sends a message to the actor without waiting for a response
func (a *NodeActor) SendAsync(msg NodeMessage) {
	select {
	case a.mailbox <- msg:
		// Message sent
	default:
		// Mailbox full, log an error
		a.logger.Error("Failed to send message to node actor, mailbox full", map[string]interface{}{
			"nodeId":  a.NodeID,
			"msgType": msg.Type,
		})
	}
}

// processMessages handles messages from the mailbox
func (a *NodeActor) processMessages() {
	for {
		select {
		case <-a.done:
			// Actor is being stopped
			return
		case msg := <-a.mailbox:
			// Process the message
			response := a.handleMessage(msg)

			// Send the response if a response channel was provided
			if msg.Response != nil {
				resp := msg.Response
				select {
				case resp <- response:
					// Response sent
				default:
					// Response channel is full or closed
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
	// Mark node as executing
	a.emitNodeStartedEvent()

	// Execute the node
	err := a.node.Execute(a.ctx)

	// Get the outputs
	a.mutex.RLock()
	outputs := make(map[string]types.Value)
	for k, v := range a.outputs {
		outputs[k] = v
	}
	a.mutex.RUnlock()

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
	// Store the input value in the context
	a.ctx.SetInput(msg.PinID, msg.Value)

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
	return &ActorExecutionContext{
		nodeID:         nodeID,
		nodeType:       nodeType,
		blueprintID:    blueprintID,
		executionID:    executionID,
		inputs:         make(map[string]types.Value),
		variables:      make(map[string]types.Value),
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

	return ctx.activatedFlows
}

// Implementation of ExecutionContext interface for ActorExecutionContext

// GetInputValue retrieves an input value by pin ID
func (ctx *ActorExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	value, exists := ctx.inputs[pinID]
	return value, exists
}

// SetOutputValue sets an output value by pin ID
func (ctx *ActorExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.actor.SetOutput(pinID, value)
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
func (ctx *ActorExecutionContext) ExecuteConnectedNodes(pinID string) error {
	// This is a placeholder - the actual implementation would depend on the actor system
	// Directly executing connected nodes violates the actor model
	// Instead, we'd queue messages for connected actors

	// For now, just record that the flow was activated
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
	ctx.debugData[key] = info

	// Store in debug manager if available
	if ctx.actor.debugMgr != nil {
		ctx.actor.debugMgr.StoreNodeDebugData(ctx.executionID, ctx.nodeID, map[string]interface{}{
			key: info,
		})
	}
}

// GetDebugData returns all debug data
func (ctx *ActorExecutionContext) GetDebugData() map[string]interface{} {
	return ctx.debugData
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
