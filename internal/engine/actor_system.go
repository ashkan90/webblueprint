package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"webblueprint/internal/common"
	"webblueprint/internal/engineext"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// ActorSystem manages all actors for a blueprint execution
type ActorSystem struct {
	ctxManager    *engineext.ContextManager
	actors        map[string]*NodeActor
	connections   map[string][]Connection
	executionID   string
	blueprintID   string
	nodeRegistry  map[string]node.NodeFactory
	logger        node.Logger
	listeners     []ExecutionListener
	debugMgr      *DebugManager
	variables     map[string]types.Value
	mutex         sync.RWMutex
	executionDone chan struct{}
	waitGroup     sync.WaitGroup

	// Add hooks
	hooks             *node.ExecutionHooks
	nodeExecutionHook func(ctx context.Context, executionID, nodeID, nodeType, execState string,
		inputs, outputs map[string]interface{}) error
	anyHook func(ctx context.Context, executionID, nodeID, level, message string,
		details map[string]interface{}) error
}

// Connection represents a connection between nodes
type Connection struct {
	ID             string
	SourceNodeID   string
	SourcePinID    string
	TargetNodeID   string
	TargetPinID    string
	ConnectionType string // "execution" or "data"
}

// NewActorSystem creates a new actor system for a blueprint execution
func NewActorSystem(
	ctxManager *engineext.ContextManager,
	executionID string,
	bp *blueprint.Blueprint,
	nodeRegistry map[string]node.NodeFactory,
	logger node.Logger,
	listeners []ExecutionListener,
	debugMgr *DebugManager,
	initialVariables map[string]types.Value,
	hooks *node.ExecutionHooks,
	nodeExecutionHook func(ctx context.Context, executionID, nodeID, nodeType, execState string,
		inputs, outputs map[string]interface{}) error,
	anyHook func(ctx context.Context, executionID, nodeID, level, message string,
		details map[string]interface{}) error,
) (*ActorSystem, error) {
	// Convert blueprint connections to internal format
	connections := make(map[string][]Connection)
	for _, conn := range bp.Connections {
		sourceNodeID := conn.SourceNodeID
		connections[sourceNodeID] = append(connections[sourceNodeID], Connection{
			ID:             conn.ID,
			SourceNodeID:   conn.SourceNodeID,
			SourcePinID:    conn.SourcePinID,
			TargetNodeID:   conn.TargetNodeID,
			TargetPinID:    conn.TargetPinID,
			ConnectionType: conn.ConnectionType,
		})
	}

	// Initialize variables
	variables := make(map[string]types.Value)
	for k, v := range initialVariables {
		variables[k] = v
	}

	return &ActorSystem{
		ctxManager:        ctxManager,
		actors:            make(map[string]*NodeActor),
		connections:       connections,
		executionID:       executionID,
		blueprintID:       bp.ID,
		nodeRegistry:      nodeRegistry,
		logger:            logger,
		listeners:         listeners,
		debugMgr:          debugMgr,
		variables:         variables,
		executionDone:     make(chan struct{}),
		hooks:             hooks,
		nodeExecutionHook: nodeExecutionHook,
		anyHook:           anyHook,
	}, nil
}

// Start initializes all actors and prepares the system for execution
func (s *ActorSystem) Start(bp *blueprint.Blueprint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// First pass: create all actors
	for _, nodeConfig := range bp.Nodes {
		// Get the node factory
		factory, exists := s.nodeRegistry[nodeConfig.Type]
		if !exists {
			return fmt.Errorf("node type not registered: %s", nodeConfig.Type)
		}

		// Create the node instance
		nodeInstance := factory()

		// Create a node-specific logger with the nodeId
		nodeLogger := s.logger
		nodeLogger.Opts(map[string]interface{}{
			"nodeId": nodeConfig.ID,
		})

		inputs := s.preprocessInputs(bp, nodeConfig.ID, s.executionID, s.variables)

		activateFlow := func(ctx *engineext.DefaultExecutionContext, nodeID, pinID string) error {
			if nodeConfig.Type == "loop" {
				connections := bp.GetNodeConnections(nodeID)
				for _, connection := range connections {
					if connection.ConnectionType != "execution" {
						continue
					}

					if connection.SourceNodeID == nodeID {
						s.mutex.RLock()
						loopActor, exists := s.actors[nodeID]
						s.mutex.RUnlock()
						if !exists {
							return fmt.Errorf("loop actor not registered")
						}

						_ = loopActor

						s.waitGroup.Add(1)
						go func() {
							defer s.waitGroup.Done()
							s.executeNode(connection.TargetNodeID)
						}()
					}
				}
			}
			return nil
		}

		actorCtx := s.ctxManager.CreateActorContext(
			bp,
			nodeConfig.ID,
			nodeConfig.Type,
			bp.ID,
			s.executionID,
			inputs,
			s.variables,
			s.logger,
			s.hooks,
			activateFlow,
		)

		if nodeConfig.Type == "loop" {
			actorCtx = s.ctxManager.CreateLoopContext()
		}

		// Create the actor
		actor := NewNodeActor(
			nodeConfig.ID,
			nodeConfig.Type,
			bp,
			s.executionID,
			nodeInstance,
			nodeLogger,
			s.listeners,
			s.debugMgr,
			s.variables,
			s.nodeExecutionHook,
			s.anyHook,
		)

		actorCtx.SaveData("node.properties", actor.properties)
		actorCtx.SaveData("node.inputPins", nodeInstance.GetInputPins())

		// Store the actor
		s.actors[nodeConfig.ID] = actor

		// Start the actor
		actor.Start(actorCtx)
	}

	// Second pass: execute data nodes immediately
	// This ensures that constant values are available right away
	for _, nodeConfig := range bp.Nodes {
		actor := s.actors[nodeConfig.ID]

		// Check if this is a data node (like constant-*)
		// Data nodes are identified by not having execution inputs
		if isDataNode(nodeConfig.Type) {
			// Execute the node immediately
			s.logger.Debug("Executing data node during initialization", map[string]interface{}{
				"nodeId":   nodeConfig.ID,
				"nodeType": nodeConfig.Type,
			})

			// Create execute message
			msg := NodeMessage{
				Type:     "execute",
				Response: make(chan NodeResponse, 1),
			}

			// Send the message to the actor
			response := actor.Send(msg)

			if !response.Success {
				s.logger.Error("Failed to initialize data node", map[string]interface{}{
					"nodeId": nodeConfig.ID,
					"error":  response.Error.Error(),
				})
			}
		}
	}

	return nil
}

// isDataNode determines if a node is a "data node" that doesn't require execution flow
func isDataNode(nodeType string) bool {
	// List of node types that are considered data nodes
	dataNodeTypes := []string{
		"constant-string",
		"constant-number",
		"constant-boolean",
		"variable-get",
	}

	for _, dataType := range dataNodeTypes {
		if nodeType == dataType {
			return true
		}
	}

	return false
}

// Execute executes the blueprint starting from the specified entry points
func (s *ActorSystem) Execute(entryPoints []string) error {
	// Emit execution start event
	for _, listener := range s.listeners {
		listener.OnExecutionEvent(ExecutionEvent{
			Type:      EventExecutionStart,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"executionId": s.executionID,
				"blueprintId": s.blueprintID,
			},
		})
	}

	// Execute each entry point
	for _, nodeID := range entryPoints {
		s.waitGroup.Add(1)
		go func(nodeID string) {
			defer s.waitGroup.Done()
			s.executeNode(nodeID)
		}(nodeID)
	}

	// Wait for execution to complete in a separate goroutine
	go func() {
		s.waitGroup.Wait()
		close(s.executionDone)

		// Emit execution end event
		for _, listener := range s.listeners {
			listener.OnExecutionEvent(ExecutionEvent{
				Type:      EventExecutionEnd,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"executionId": s.executionID,
					"blueprintId": s.blueprintID,
					"success":     true,
				},
			})
		}
	}()

	return nil
}

// executeNode executes a single node and follows its output connections
func (s *ActorSystem) executeNode(nodeID string) {
	// Get the actor
	s.mutex.RLock()
	actor, exists := s.actors[nodeID]
	s.mutex.RUnlock()

	if !exists {
		s.logger.Error("Node not found", map[string]interface{}{
			"nodeId": nodeID,
		})
		return
	}

	// Log node execution
	s.logger.Debug("Executing node", map[string]interface{}{
		"nodeId":   nodeID,
		"nodeType": actor.NodeType,
	})

	s.preprocessActorNodeInputs(actor)

	// Create execute message
	msg := NodeMessage{
		Type:     "execute",
		Response: make(chan NodeResponse, 1),
	}

	// Send the message to the actor
	response := actor.Send(msg)

	if !response.Success {
		s.logger.Error("Node execution failed", map[string]interface{}{
			"nodeId": nodeID,
			"error":  response.Error.Error(),
		})
		return
	}

	// Store all output values in the debug manager
	for pinID, value := range response.OutputPins {
		if s.debugMgr != nil {
			s.debugMgr.StoreNodeOutputValue(s.executionID, nodeID, pinID, value.RawValue)
			s.logger.Debug("Stored output value in debug manager", map[string]interface{}{
				"nodeId": nodeID,
				"pinId":  pinID,
				// "value":  value.RawValue,
			})
		}
	}

	// Log number of outputs
	s.logger.Debug("Node execution succeeded", map[string]interface{}{
		"nodeId":      nodeID,
		"outputCount": len(response.OutputPins),
	})

	// Follow output connections
	s.followConnections(actor, response)
}

// followConnections follows outgoing connections from a node
func (s *ActorSystem) followConnections(actor *NodeActor, response NodeResponse) {
	// Get the node's outgoing connections
	s.mutex.RLock()
	connections, exists := s.connections[actor.NodeID]
	s.mutex.RUnlock()

	if !exists {
		return
	}

	// Get activated flows using the helper function
	var activatedFlows []string
	if extCtx := engineext.GetExtendedContext(actor.decoratedCtx); extCtx != nil {
		activatedFlows = extCtx.GetActivatedOutputFlows()
	} else {
		// Fallback or log error if ExtendedExecutionContext was not found
		s.logger.Warn("Could not retrieve ExtendedExecutionContext to get activated flows", map[string]interface{}{"nodeId": actor.NodeID, "contextType": fmt.Sprintf("%T", actor.decoratedCtx)})
		activatedFlows = []string{} // Initialize to empty slice
	}

	// Process data connections first to ensure data is available before execution
	for _, conn := range connections {
		if conn.ConnectionType == "data" {
			// Check if we have a value for this pin
			value, exists := response.OutputPins[conn.SourcePinID]
			if !exists {
				s.logger.Debug("No value for data connection", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
				})
				continue
			}

			// Get the target actor
			s.mutex.RLock()
			targetActor, exists := s.actors[conn.TargetNodeID]
			s.mutex.RUnlock()

			if !exists {
				s.logger.Warn("Target actor not found for data connection", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"targetNodeId": conn.TargetNodeID,
				})
				continue
			}

			// Send the value to the target actor
			inputMsg := NodeMessage{
				Type:  "input",
				PinID: conn.TargetPinID,
				Value: value,
			}
			if sent := targetActor.SendAsync(inputMsg); !sent {
				s.logger.Warn("Failed to send input value to target actor", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
				})
			} else {
				s.logger.Debug("Sent input value to target actor", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
					//"value":        value.RawValue,
				})
			}

			// Emit value produced event
			for _, listener := range s.listeners {
				listener.OnExecutionEvent(ExecutionEvent{
					Type:      EventValueProduced,
					Timestamp: time.Now(),
					NodeID:    conn.SourceNodeID,
					Data: map[string]interface{}{
						"sourceNodeId": conn.SourceNodeID,
						"sourcePinId":  conn.SourcePinID,
						"targetNodeId": conn.TargetNodeID,
						"targetPinId":  conn.TargetPinID,
						"value":        value.RawValue,
					},
				})
			}
		}
	}

	// Now follow execution connections
	for _, conn := range connections {
		if conn.ConnectionType == "execution" {
			// Check if this flow was activated
			flowActivated := false
			for _, flow := range activatedFlows {
				if flow == conn.SourcePinID {
					flowActivated = true
					break
				}
			}

			if flowActivated {
				// Execute the target node
				s.waitGroup.Add(1)
				go func(conn Connection) {
					defer s.waitGroup.Done()
					s.executeNode(conn.TargetNodeID)
				}(conn)
			}
		}
	}
}

// Wait waits for all execution to complete with a timeout
func (s *ActorSystem) Wait(timeout time.Duration) bool {
	select {
	case <-s.executionDone:
		// Execution completed
		return true
	case <-time.After(timeout):
		// Timeout
		return false
	}
}

// Stop stops all actors and cleans up resources
func (s *ActorSystem) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, actor := range s.actors {
		actor.Stop()
	}
}

// GetResult returns the execution result
func (s *ActorSystem) GetResult() common.ExecutionResult {
	// Collect output values from all nodes
	nodeResults := make(map[string]map[string]interface{})

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for nodeID, actor := range s.actors {
		// Get all output values
		outputs := make(map[string]interface{})

		// For each output pin, get the value
		for _, pin := range actor.node.GetOutputPins() {
			if value, exists := actor.GetOutput(pin.ID); exists {
				outputs[pin.ID] = value.RawValue
			}
		}

		nodeResults[nodeID] = outputs
	}

	return common.ExecutionResult{
		Success:     true,
		ExecutionID: s.executionID,
		StartTime:   time.Time{}, // Will be set by the caller
		EndTime:     time.Now(),
		NodeResults: nodeResults,
	}
}

// GetNodesStatus returns the status of all nodes
func (s *ActorSystem) GetNodesStatus() map[string]NodeStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	statuses := make(map[string]NodeStatus)
	for nodeID, actor := range s.actors {
		statuses[nodeID] = actor.GetStatus()
	}

	return statuses
}

func (s *ActorSystem) preprocessInputs(bp *blueprint.Blueprint, nodeID, executionID string, variables map[string]types.Value) map[string]types.Value {
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
					s.debugMgr.StoreNodeOutputValue(executionID, conn.SourceNodeID, conn.SourcePinID, varValue.RawValue)
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
			if nodeResults, ok := s.debugMgr.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); ok {
				// Convert to Value
				inputValues[targetPinID] = types.NewValue(types.PinTypes.Any, nodeResults)

				// Emit value consumed event
				//s..EmitEvent(ExecutionEvent{
				//	Type:      EventValueConsumed,
				//	Timestamp: time.Now(),
				//	NodeID:    nodeID,
				//	Data: map[string]interface{}{
				//		"sourceNodeID": sourceNodeID,
				//		"sourcePinID":  sourcePinID,
				//		"targetPinID":  targetPinID,
				//		"value":        nodeResults,
				//	},
				//})
			}
		}
	}

	return inputValues
}

func (s *ActorSystem) preprocessActorNodeInputs(actor *NodeActor) {
	connections := actor.bp.GetNodeInputConnections(actor.NodeID)
	for _, conn := range connections {
		if conn.ConnectionType != "data" {
			continue
		}

		s.mutex.RLock()
		sourceActor, exists := s.actors[conn.SourceNodeID]
		s.mutex.RUnlock()
		if !exists {
			continue
		}

		if strings.Contains(sourceActor.node.GetMetadata().TypeID, "variable-get") {
			// TODO improve this variable naming approach
			varId := types.GetProperty(sourceActor.properties, "variableId")
			if varId == nil {
				continue
			}

			setVar := actor.bp.FindNodeByVar(varId.Value.(string))

			if setVar == nil {
				return
			}

			s.mutex.Lock()
			targetActor, exists := s.actors[setVar.ID]
			s.mutex.Unlock()

			if !exists {
				s.logger.Warn("Target actor not found for data connection", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"targetNodeId": conn.TargetNodeID,
				})
				continue
			}

			value, vExists := engineext.GetExtendedContext(targetActor.decoratedCtx).GetInputValue(conn.SourcePinID)
			if !vExists {
				s.logger.Warn("Target actor has no input to feed", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
				})
				continue
			}

			// Send the value to the target actor
			inputMsg := NodeMessage{
				Type:  "input",
				PinID: conn.TargetPinID,
				Value: value,
			}
			if sent := actor.SendAsync(inputMsg); !sent {
				s.logger.Warn("Failed to send input value to target actor", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
				})
			} else {
				// send same message to current actor so it can now recognize variable
				//actor.SendAsync(inputMsg)
				s.logger.Debug("Sent input value to target actor", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
					//"value":        value.RawValue,
				})
			}

			// Emit value produced event
			for _, listener := range s.listeners {
				listener.OnExecutionEvent(ExecutionEvent{
					Type:      EventValueProduced,
					Timestamp: time.Now(),
					NodeID:    conn.SourceNodeID,
					Data: map[string]interface{}{
						"sourceNodeId": conn.SourceNodeID,
						"sourcePinId":  conn.SourcePinID,
						"targetNodeId": conn.TargetNodeID,
						"targetPinId":  conn.TargetPinID,
						"value":        value.RawValue,
					},
				})
			}
		}
	}
}
