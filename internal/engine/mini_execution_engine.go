package engine

import (
	"fmt"
	"sync"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// MiniExecutionEngine is a simplified execution engine for running function nodes
type MiniExecutionEngine struct {
	nodeRegistry map[string]node.NodeFactory
	logger       node.Logger
	debugMgr     *DebugManager
	outputs      map[string]map[string]types.Value // nodeID -> pinID -> value
	mutex        sync.RWMutex
}

// NewMiniExecutionEngine creates a new mini execution engine
func NewMiniExecutionEngine(logger node.Logger, debugMgr *DebugManager) *MiniExecutionEngine {
	return &MiniExecutionEngine{
		nodeRegistry: make(map[string]node.NodeFactory),
		logger:       logger,
		debugMgr:     debugMgr,
		outputs:      make(map[string]map[string]types.Value),
	}
}

// RegisterNodeType registers a node type with the mini engine
func (e *MiniExecutionEngine) RegisterNodeType(typeID string, factory node.NodeFactory) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.nodeRegistry[typeID] = factory
}

// Execute runs a mini-blueprint for a function
func (e *MiniExecutionEngine) Execute(blueprintID string, initialData map[string]types.Value) (bool, error) {
	e.mutex.Lock()

	// Check if blueprint exists
	bp, err := db.Blueprints.GetBlueprint(blueprintID)
	if err != nil {
		e.mutex.Unlock()
		return false, fmt.Errorf("mini-engine: blueprint not found: %s", blueprintID)
	}

	// Initialize variables
	variables := make(map[string]types.Value)
	for k, v := range initialData {
		variables[k] = v
	}

	e.mutex.Unlock()

	// Find entry points
	entryPoints := bp.FindEntryPoints()
	if len(entryPoints) == 0 {
		// No entry points, let's find all nodes with no inputs
		// This helps with handling simple data transformation functions
		for _, nodeConfig := range bp.Nodes {
			if len(bp.GetNodeInputConnections(nodeConfig.ID)) == 0 {
				entryPoints = append(entryPoints, nodeConfig.ID)
			}
		}

		if len(entryPoints) == 0 {
			// Still no entry points, try to execute all nodes
			for _, nodeConfig := range bp.Nodes {
				entryPoints = append(entryPoints, nodeConfig.ID)
			}
		}
	}

	// Filter out any nodes with types that aren't registered (to prevent recursion)
	safeEntryPoints := make([]string, 0, len(entryPoints))
	for _, nodeID := range entryPoints {
		nodeConfig := bp.FindNode(nodeID)
		if nodeConfig == nil {
			continue
		}

		_, exists := registry.GetInstance().GetNodeFactory(nodeConfig.Type)

		if exists {
			safeEntryPoints = append(safeEntryPoints, nodeID)
		} else {
			e.logger.Debug("Skipping unregistered node type", map[string]interface{}{
				"nodeType": nodeConfig.Type,
				"nodeId":   nodeID,
			})
		}
	}

	entryPoints = safeEntryPoints

	// Define hooks
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			e.logger.Debug("Mini engine: Node started", map[string]interface{}{
				"nodeId":   nodeID,
				"nodeType": nodeType,
			})
		},
		OnNodeComplete: func(nodeID, nodeType string) {
			e.logger.Debug("Mini engine: Node completed", map[string]interface{}{
				"nodeId":   nodeID,
				"nodeType": nodeType,
			})
		},
		OnNodeError: func(nodeID string, err error) {
			e.logger.Error("Mini engine: Node error", map[string]interface{}{
				"nodeId": nodeID,
				"error":  err.Error(),
			})
		},
		OnPinValue: func(nodeID, pinName string, value interface{}) {
			e.mutex.Lock()
			defer e.mutex.Unlock()

			// Store the output value
			if _, exists := e.outputs[nodeID]; !exists {
				e.outputs[nodeID] = make(map[string]types.Value)
			}

			// Create appropriate type for the value
			var pinType *types.PinType
			switch value.(type) {
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

			e.outputs[nodeID][pinName] = types.NewValue(pinType, value)

			e.logger.Debug("Mini engine: Pin value", map[string]interface{}{
				"nodeId": nodeID,
				"pinId":  pinName,
				"value":  value,
			})
		},
	}

	// Process each entry point in parallel
	wg := sync.WaitGroup{}
	errors := make(chan error, len(entryPoints))

	for _, entryNodeID := range entryPoints {
		wg.Add(1)

		go func(nodeID string) {
			defer wg.Done()

			if err := e.executeNode(nodeID, bp, blueprintID, variables, hooks); err != nil {
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
		return false, lastError
	}

	return true, nil
}

// executeNode executes a single node in the mini engine
func (e *MiniExecutionEngine) executeNode(nodeID string, bp *blueprint.Blueprint, blueprintID string, variables map[string]types.Value, hooks *node.ExecutionHooks) error {
	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		return fmt.Errorf("mini-engine: node not found: %s", nodeID)
	}

	e.mutex.RLock()
	factory, exists := e.nodeRegistry[nodeConfig.Type]
	e.mutex.RUnlock()

	if !exists {
		e.logger.Warn("Node type not registered, skipping execution", map[string]interface{}{
			"nodeType": nodeConfig.Type,
			"nodeId":   nodeID,
		})
		// Don't treat this as an error - just skip this node
		// This is important for handling user functions where we deliberately
		// don't register the function itself to prevent recursion
		return nil
	}

	// Create the node instance
	nodeInstance := factory()

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

			// Check if we have a value for this pin
			e.mutex.RLock()
			if nodeOutputs, exists := e.outputs[sourceNodeID]; exists {
				if value, exists := nodeOutputs[sourcePinID]; exists {
					inputValues[targetPinID] = value
				}
			}
			e.mutex.RUnlock()
		}
	}

	// Add any variables as inputs if matching input pin names exist
	for _, pin := range nodeInstance.GetInputPins() {
		if value, exists := variables[pin.ID]; exists {
			inputValues[pin.ID] = value
		}
	}

	// Create a function to activate output flows
	executionID := fmt.Sprintf("mini-%s-%d", blueprintID, time.Now().UnixNano())
	activateFlowFn := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		// Store outputs
		e.mutex.Lock()
		if _, exists := e.outputs[nodeID]; !exists {
			e.outputs[nodeID] = make(map[string]types.Value)
		}

		for pin, value := range ctx.outputs {
			e.outputs[nodeID][pin] = value
		}
		e.mutex.Unlock()

		// Find connections from this output pin
		outputConnections := bp.GetNodeOutputConnections(nodeID)
		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID

				// Execute the target node
				if err := e.executeNode(targetNodeID, bp, blueprintID, variables, hooks); err != nil {
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

	// Handle errors
	if err != nil {
		if hooks != nil && hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, err)
		}
		return err
	}

	// Store the outputs
	e.mutex.Lock()
	if _, exists := e.outputs[nodeID]; !exists {
		e.outputs[nodeID] = make(map[string]types.Value)
	}

	for pin, value := range ctx.outputs {
		e.outputs[nodeID][pin] = value
	}
	e.mutex.Unlock()

	// Process any execution flows
	for _, outputPin := range ctx.GetActivatedOutputFlows() {
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

// GetOutputs returns all the output values collected during execution
func (e *MiniExecutionEngine) GetOutputs() map[string]types.Value {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// Collect outputs that match expected function outputs
	// This would need to be implemented based on your actual output mapping logic
	// For now, we'll just collect all outputs from all nodes
	results := make(map[string]types.Value)

	for _, nodeOutputs := range e.outputs {
		for pinID, value := range nodeOutputs {
			results[pinID] = value
		}
	}

	return results
}
