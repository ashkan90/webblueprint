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

// ActorSystem manages all actors for a blueprint execution
type ActorSystem struct {
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
	executionID string,
	bp *blueprint.Blueprint,
	nodeRegistry map[string]node.NodeFactory,
	logger node.Logger,
	listeners []ExecutionListener,
	debugMgr *DebugManager,
	initialVariables map[string]types.Value,
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
		actors:        make(map[string]*NodeActor),
		connections:   connections,
		executionID:   executionID,
		blueprintID:   bp.ID,
		nodeRegistry:  nodeRegistry,
		logger:        logger,
		listeners:     listeners,
		debugMgr:      debugMgr,
		variables:     variables,
		executionDone: make(chan struct{}),
	}, nil
}

// Start initializes all actors and prepares the system for execution
func (s *ActorSystem) Start(bp *blueprint.Blueprint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create actors for all nodes
	for _, nodeConfig := range bp.Nodes {
		// Get the node factory
		log.Println(s.nodeRegistry, nodeConfig.Type)
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

		// Create the actor
		actor := NewNodeActor(
			nodeConfig.ID,
			nodeConfig.Type,
			s.blueprintID,
			s.executionID,
			nodeInstance,
			nodeLogger,
			s.listeners,
			s.debugMgr,
			s.variables,
		)

		// Store the actor
		s.actors[nodeConfig.ID] = actor

		// Start the actor
		actor.Start()
	}

	return nil
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
				"value":  value.RawValue,
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

	// Get activated flows
	activatedFlows := actor.ctx.GetActivatedOutputFlows()

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
					"value":        value.RawValue,
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
func (s *ActorSystem) GetResult() ExecutionResult {
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

	return ExecutionResult{
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
