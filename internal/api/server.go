package api

import (
	"sync"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
)

// APIServer handles HTTP API requests
type APIServer struct {
	executionEngine *engine.ExecutionEngine
	nodeRegistry    map[string]node.NodeFactory
	wsManager       *WebSocketManager
	debugManager    *engine.DebugManager
	rw              *sync.RWMutex
}

// NewAPIServer creates a new API server
func NewAPIServer(executionEngine *engine.ExecutionEngine, wsManager *WebSocketManager, debugManager *engine.DebugManager) *APIServer {
	return &APIServer{
		executionEngine: executionEngine,
		nodeRegistry:    make(map[string]node.NodeFactory),
		wsManager:       wsManager,
		debugManager:    debugManager,
		rw:              &sync.RWMutex{},
	}
}

// RegisterNodeType registers a node type with both the execution engine and API server
func (s *APIServer) RegisterNodeType(typeID string, factory node.NodeFactory) {
	// Store locally
	s.nodeRegistry[typeID] = factory

	// Register with execution engine
	s.executionEngine.RegisterNodeType(typeID, factory)

	// Register with global registry
	registry.GetInstance().RegisterNodeType(typeID, factory)

	// Create a node instance to get metadata
	nodeInstance := factory()
	metadata := nodeInstance.GetMetadata()

	// Broadcast node type to connected clients
	s.wsManager.BroadcastMessage(MsgTypeNodeIntro, map[string]interface{}{
		"typeId":      metadata.TypeID,
		"name":        metadata.Name,
		"description": metadata.Description,
		"category":    metadata.Category,
		"version":     metadata.Version,
		"inputs":      convertPinsToInfo(nodeInstance.GetInputPins()),
		"outputs":     convertPinsToInfo(nodeInstance.GetOutputPins()),
	})
}
