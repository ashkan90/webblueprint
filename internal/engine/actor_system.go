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
	"webblueprint/internal/nodes/data"
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

		// inputs := s.preprocessInputs(bp, nodeConfig.ID, s.executionID, s.variables) // Removed call
		inputs := make(map[string]types.Value) // Initialize empty inputs, handled by messages now

		// Simple activateFlow placeholder
		activateFlow := func(ctx *engineext.DefaultExecutionContext, nodeID, pinID string) error {
			// In the actor model, the actual triggering happens in followConnections.
			// This callback might be less critical here.
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

		// Removed incorrect attempt to create LoopContext here.
		// LoopContext creation should be handled within the LoopNode's execution.
		// if nodeConfig.Type == "loop" {
		// 	actorCtx = s.ctxManager.CreateLoopContext() // This was incorrect
		// }

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
			s.variables, // Pass ActorSystem's shared variables map
			&s.mutex,    // Pass ActorSystem's mutex for shared variables
			s,           // Pass ActorSystem reference
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

	// TODO try to delete those codes ?
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

	// s.preprocessActorNodeInputs(actor) // Ensure this call is removed

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

	// Use the explicit flow signal from the response, if available
	flowToActivate := response.FlowToActivate
	if flowToActivate == "" {
		// Fallback to context's activated flows if response doesn't specify one
		// This maintains compatibility with nodes not using the new response field
		if extCtx := engineext.GetExtendedContext(actor.ctx); extCtx != nil {
			activatedFlows := extCtx.GetActivatedOutputFlows()
			if len(activatedFlows) > 0 {
				flowToActivate = activatedFlows[0] // Use the first activated flow as default? Or handle multiple?
				if len(activatedFlows) > 1 {
					s.logger.Warn("Multiple flows activated in context, but only following the first", map[string]interface{}{"nodeId": actor.NodeID, "flows": activatedFlows})
				}
			}
		} else {
			s.logger.Warn("Could not retrieve ExtendedExecutionContext to get activated flows", map[string]interface{}{"nodeId": actor.NodeID, "contextType": fmt.Sprintf("%T", actor.decoratedCtx)})
		}
	}

	// Process data connections first to ensure data is available before execution
	for _, conn := range connections {
		if conn.ConnectionType == "data" {
			// Check if we have a value for this pin
			value, exists := response.OutputPins[conn.SourcePinID]
			if !exists {
				//s.logger.Debug("No value for data connection from response, fallback actor outputs", map[string]interface{}{})
				//value, exists = actor.GetOutput(conn.SourcePinID)
				//if !exists {
				s.logger.Debug("No value for data connection", map[string]interface{}{
					"sourceNodeId": conn.SourceNodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
					"targetPinId":  conn.TargetPinID,
				})
				continue
				//}
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

				if strings.Contains(targetActor.NodeType, "variable-set") {
					varName, varValue := data.NewVariableDefinition(targetActor.node, inputMsg.Value)
					targetActor.ctx.SetVariable(varName, varValue)

					varId, _ := targetActor.GetProperty("variableId")
					if varId != nil {
						varGetNode := targetActor.bp.FindNodeByVar(varId.(string))
						if varGetNode != nil {
							s.mutex.RLock()
							getActor := s.actors[varGetNode.ID]
							s.mutex.RUnlock()

							getActor.ctx.SetVariable(varName, varValue)

							getActor.SendAsync(inputMsg)
						}
					}
				}
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

	// Now follow execution connections (single pass)
	for _, conn := range connections {
		// Check if the connection's source pin matches the flow we need to activate
		if conn.ConnectionType == "execution" {

			// Get the target actor
			s.mutex.RLock()
			targetActor, targetExists := s.actors[conn.TargetNodeID]
			s.mutex.RUnlock()
			if !targetExists {
				continue // Target actor doesn't exist
			}

			// --- Special Handling for Loop Node ---
			if actor.NodeType == "loop" && conn.SourcePinID == flowToActivate {
				indexValue, indexExists := response.OutputPins["index"]
				if !indexExists {
					s.logger.Warn("Loop node activated 'loop' pin but 'index' output is missing", map[string]interface{}{"loopNodeId": actor.NodeID})
					continue // Skip this iteration if index is missing
				}

				loopIterationPayload := map[string]interface{}{
					"_loop_index": indexValue.RawValue,
				}
				execMsg := NodeMessage{
					Type:     "execute",
					Value:    types.NewValue(types.PinTypes.Object, loopIterationPayload),
					Response: make(chan NodeResponse, 1),
				}

				s.waitGroup.Add(1)
				go func(sourceActor, targetActor *NodeActor, msg NodeMessage) {
					defer s.waitGroup.Done()
					s.logger.Debug("Executing loop body node", map[string]interface{}{"targetNodeId": targetActor.NodeID, "indexPayload": msg.Value.RawValue})
					execResponse := targetActor.Send(msg) // Execute the loop body node
					if !execResponse.Success {
						s.logger.Error("Loop body node execution failed", map[string]interface{}{"nodeId": targetActor.NodeID, "error": execResponse.Error})
						// If body fails, should we stop the loop? Send error back to loop actor?
						// For now, we just log and don't proceed with this path or signal loop actor.
						return
					}
					// Follow connections *from* the loop body node
					s.followConnections(targetActor, execResponse) // Removed recursive call

					// After body execution (and its downstream effects) are done *for this iteration*,
					// signal the original LoopNode actor to proceed to the next iteration.
					sourceActor.Send(NodeMessage{Type: "loop_next"})

					//s.logger.Debug("execution response", map[string]interface{}{
					//	"response": execResponse,
					//})
				}(actor, targetActor, execMsg)
			} else {
				// --- Standard Execution Flow ---
				s.waitGroup.Add(1)
				go func(targetNodeID string, triggerPinID string) {
					defer s.waitGroup.Done()
					//s.executeNode(targetNodeID)
					s.logger.Debug("standard execution", map[string]interface{}{})
					s.executeNodeTriggered(targetNodeID, triggerPinID)
				}(conn.TargetNodeID, conn.TargetPinID)

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

// Removed preprocessInputs and preprocessActorNodeInputs functions
// executeNodeTriggered executes a node when triggered by a specific input pin
func (s *ActorSystem) executeNodeTriggered(nodeID, triggerPinID string) { // Simplified signature
	s.mutex.RLock()
	actor, exists := s.actors[nodeID]
	s.mutex.RUnlock()
	if !exists {
		s.logger.Error("Node not found for triggered execution", map[string]interface{}{"nodeId": nodeID})
		return
	}

	s.logger.Debug("Executing node (triggered)", map[string]interface{}{"nodeId": nodeID, "triggerPin": triggerPinID})

	// Create execute message
	// Note: Removed TriggerPin field as it doesn't exist in NodeMessage
	msg := NodeMessage{
		Type:     "execute",
		Response: make(chan NodeResponse, 1),
		// If triggerPinID needs to be passed, use FlowData:
		// FlowData: map[string]interface{}{"triggerPin": triggerPinID},
	}

	if triggerPinID != "" {
		msg.TriggerPin = triggerPinID
	}

	// Send the message to the actor
	response := actor.Send(msg)

	if !response.Success {
		s.logger.Error("Node execution failed (triggered)", map[string]interface{}{"nodeId": nodeID, "error": response.Error})
		return
	}

	// Store debug outputs
	for pinID, value := range response.OutputPins {
		if s.debugMgr != nil {
			s.debugMgr.StoreNodeOutputValue(s.executionID, nodeID, pinID, value.RawValue)
		}
	}

	// Follow connections from this node's execution
	s.followConnections(actor, response)
}
