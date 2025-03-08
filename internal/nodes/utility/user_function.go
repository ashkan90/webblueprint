package utility

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"webblueprint/internal/db"
	"webblueprint/internal/engine"
	"webblueprint/internal/node"
	"webblueprint/internal/registry"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// UserFunctionNode implements a node that executes a user-defined function
type UserFunctionNode struct {
	node.BaseNode
	FunctionDef blueprint.Function // The function definition
	mutex       sync.RWMutex
}

// NewUserFunctionNode creates a new user function node factory
func NewUserFunctionNode(function blueprint.Function) func() node.Node {
	return func() node.Node {
		var fnNode = &UserFunctionNode{
			BaseNode: node.BaseNode{
				Metadata: node.NodeMetadata{
					TypeID:      strings.ToLower(function.Name),
					Name:        function.Name,
					Description: function.Description,
					Category:    "Function",
					Version:     "1.0.0",
				},
			},
			FunctionDef: function, // Store the entire function definition
		}

		// Set up the input pins based on the function's node type
		for _, input := range function.NodeType.Inputs {
			pinType := mapPinType(input.Type)
			fnNode.Inputs = append(fnNode.Inputs, types.Pin{
				ID:          input.ID,
				Name:        input.Name,
				Description: input.Description,
				Type:        pinType,
				Optional:    input.Optional,
				Default:     input.Default,
			})
		}

		// Ensure we have at least one execution input pin
		hasExecutionInput := false
		for _, input := range fnNode.Inputs {
			if input.Type == types.PinTypes.Execution {
				hasExecutionInput = true
				break
			}
		}

		if !hasExecutionInput {
			// Add default execution input
			fnNode.Inputs = append(fnNode.Inputs, types.Pin{
				ID:          "execute",
				Name:        "Execute",
				Description: "Execution input",
				Type:        types.PinTypes.Execution,
				Optional:    false,
			})
		}

		// Set up the output pins based on the function's node type
		for _, output := range function.NodeType.Outputs {
			pinType := mapPinType(output.Type)
			fnNode.Outputs = append(fnNode.Outputs, types.Pin{
				ID:          output.ID,
				Name:        output.Name,
				Description: output.Description,
				Type:        pinType,
				Optional:    output.Optional,
				Default:     output.Default,
			})
		}

		// Ensure we have at least one execution output pin
		hasExecutionOutput := false
		for _, output := range fnNode.Outputs {
			if output.Type == types.PinTypes.Execution {
				hasExecutionOutput = true
				break
			}
		}

		if !hasExecutionOutput {
			// Add default execution output
			fnNode.Outputs = append(fnNode.Outputs, types.Pin{
				ID:          "then",
				Name:        "Then",
				Description: "Execution output",
				Type:        types.PinTypes.Execution,
				Optional:    false,
			})
		}

		// Set up properties
		for _, property := range function.NodeType.Properties {
			fnNode.Properties = append(fnNode.Properties, types.Property{
				Name:  property.Name,
				Value: property.Value,
			})
		}

		return fnNode
	}
}

// Execute runs the node logic
func (n *UserFunctionNode) Execute(ctx node.ExecutionContext) error {
	logger := ctx.Logger()
	logger.Debug("Executing Function node", map[string]interface{}{
		"functionName": n.Metadata.Name,
	})

	// Collect debug data for inputs
	inputDebugData := make(map[string]interface{})
	inputValues := make(map[string]types.Value)

	// Collect all input values
	for _, input := range n.GetInputPins() {
		inputValue, exists := ctx.GetInputValue(input.ID)
		if exists {
			inputValues[input.ID] = inputValue
			inputDebugData[input.ID] = inputValue.RawValue
		}
	}

	// Log the input values for debugging
	logger.Debug("Function input values", inputDebugData)

	// Create a unique ID for this function execution
	executionID := fmt.Sprintf("func-%s-%d", ctx.GetExecutionID(), time.Now().UnixNano())

	// Create a mini-blueprint from the function definition
	functionBlueprint := &blueprint.Blueprint{
		ID:          fmt.Sprintf("%s-func-%d", n.FunctionDef.ID, time.Now().UnixNano()),
		Name:        n.FunctionDef.Name,
		Description: n.FunctionDef.Description,
		Version:     "1.0.0",
		Nodes:       n.FunctionDef.Nodes,
		Connections: n.FunctionDef.Connections,
	}

	// Store the blueprint temporarily
	db.Blueprints.AddBlueprint(functionBlueprint)
	defer func() {
		// Remove the temporary blueprint
		delete(db.Blueprints, functionBlueprint.ID)
	}()

	// Find the entry point nodes in the function
	entryPoints := functionBlueprint.FindEntryPoints()
	if len(entryPoints) == 0 {
		logger.Warn("No entry points found in function", nil)
		// Even if there are no entry points, we might have simple data transformations
	}

	// Create a function execution context using the engine's implementation
	functionContext := engine.NewFunctionExecutionContext(
		ctx,                          // parent context
		ctx.GetNodeID(),              // node ID
		n.Metadata.TypeID,            // node type
		functionBlueprint.ID,         // function/blueprint ID
		executionID,                  // execution ID
		inputValues,                  // input values
		make(map[string]types.Value), // variables
		logger,                       // logger
	)

	// Execute the function nodes
	err := n.executeFunction(functionContext, entryPoints)
	if err != nil {
		logger.Error("Function execution error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Map function outputs to the node outputs
	outputDebugData := make(map[string]interface{})
	for _, output := range n.GetOutputPins() {
		// Skip execution pins as they're handled by the flow activation
		if output.Type == types.PinTypes.Execution {
			continue
		}

		// Get the output value from the function context if available
		if value, exists := functionContext.GetAllOutputs()[output.ID]; exists {
			ctx.SetOutputValue(output.ID, value)
			outputDebugData[output.ID] = value.RawValue
		} else {
			// For pins without explicit outputs, check if we have a matching input
			// (simple pass-through for inputs to outputs with same name)
			if inputValue, exists := inputValues[output.ID]; exists {
				ctx.SetOutputValue(output.ID, inputValue)
				outputDebugData[output.ID] = inputValue.RawValue
			}
		}
	}

	// Log the output values for debugging
	logger.Debug("Function output values", outputDebugData)

	// Record debug info
	ctx.RecordDebugInfo(types.DebugInfo{
		NodeID:      ctx.GetNodeID(),
		Description: "Function Execution",
		Value: map[string]interface{}{
			"inputs":          inputDebugData,
			"outputs":         outputDebugData,
			"functionName":    n.Metadata.Name,
			"nodeCount":       len(n.FunctionDef.Nodes),
			"connectionCount": len(n.FunctionDef.Connections),
		},
		Timestamp: time.Now(),
	})

	// Continue execution flow - use "then" by default if it exists
	for _, output := range n.GetOutputPins() {
		if output.ID == "then" && output.Type == types.PinTypes.Execution {
			return ctx.ActivateOutputFlow("then")
		}
	}

	// If "then" doesn't exist, activate the first execution output (if any)
	for _, output := range n.GetOutputPins() {
		if output.Type == types.PinTypes.Execution {
			return ctx.ActivateOutputFlow(output.ID)
		}
	}

	// No execution outputs found, just return nil
	return nil
}

// executeFunction is a helper method to execute the internal nodes of the function
func (n *UserFunctionNode) executeFunction(functionContext *engine.FunctionExecutionContext, entryPoints []string) error {
	logger := functionContext.Logger()

	// We no longer filter out recursive entry points - instead, we'll handle them properly
	// in executeNode by simulating their execution

	// If there are no entry points, try to use the mini execution engine
	if len(entryPoints) == 0 {
		logger.Debug("No entry points found in function, using mini execution engine", nil)

		// Create a mini execution engine for this function
		miniEngine := engine.NewMiniExecutionEngine(logger, nil)

		// Register all node types EXCEPT this function type to prevent recursion
		nodeRegistry := getNodeRegistry()
		safeFunctionName := strings.ToLower(n.FunctionDef.Name)
		for typeID, factory := range nodeRegistry {
			// Skip registering this function type to prevent recursion
			if strings.ToLower(typeID) != safeFunctionName {
				miniEngine.RegisterNodeType(typeID, factory)
			}
		}

		// Execute the mini-blueprint with the function context's inputs
		success, err := miniEngine.Execute(functionContext.GetBlueprintID(), functionContext.GetAllOutputs())
		if err != nil {
			logger.Error("Mini execution engine error", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}

		if !success {
			logger.Warn("Mini execution engine completed with warnings", nil)
		}

		// Get outputs from the mini engine's results and map them to function outputs
		outputs := miniEngine.GetOutputs()
		for pinID, value := range outputs {
			// Check if this is a valid output pin of the function
			for _, output := range n.GetOutputPins() {
				if output.ID == pinID && output.Type != types.PinTypes.Execution {
					functionContext.SetOutputValue(pinID, value)
				}
			}
		}

		logger.Debug("Mini execution engine completed", map[string]interface{}{
			"outputCount": len(outputs),
		})

		return nil
	}

	logger.Debug("Executing function entry points", map[string]interface{}{
		"entryPointCount": len(entryPoints),
	})

	// With entry points, we execute each entry point node
	for _, entryPointID := range entryPoints {
		// We no longer need to filter here - we'll handle recursive nodes in executeNode
		err := n.executeNode(functionContext, entryPointID)
		if err != nil {
			logger.Error("Error executing function entry point", map[string]interface{}{
				"entryPointId": entryPointID,
				"error":        err.Error(),
			})
			return err
		}
	}

	logger.Debug("Function entry points executed successfully", nil)
	return nil
}

// executeNode executes a single node in the function
func (n *UserFunctionNode) executeNode(functionContext *engine.FunctionExecutionContext, nodeID string) error {
	// Get the blueprint that contains this function
	bp, err := db.Blueprints.GetBlueprint(functionContext.GetBlueprintID())
	if err != nil {
		return fmt.Errorf("function blueprint not found: %s", functionContext.GetBlueprintID())
	}

	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		return fmt.Errorf("node not found in function: %s", nodeID)
	}

	// CRITICAL: Handle recursive function calls
	// If this node is of the same type as our function, we need to stop the recursion
	if strings.ToLower(nodeConfig.Type) == strings.ToLower(n.FunctionDef.Name) {
		functionContext.Logger().Warn("Detected recursive function call - simulating execution", map[string]interface{}{
			"functionName": n.FunctionDef.Name,
			"nodeId":       nodeID,
		})

		// Instead of simply skipping, we need to simulate the function's behavior
		// by activating its output execution flows as if it had executed

		// Find all execution connections from this node and follow them
		outputConnections := bp.GetNodeOutputConnections(nodeID)
		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" {
				// Execute the target node - this simulates the function activating its outputs
				functionContext.Logger().Debug("Following connection from simulated function call", map[string]interface{}{
					"sourceNodeId": nodeID,
					"sourcePinId":  conn.SourcePinID,
					"targetNodeId": conn.TargetNodeID,
				})

				if err := n.executeNode(functionContext, conn.TargetNodeID); err != nil {
					return err
				}
			}
		}

		// We've simulated the function's behavior, so return without actually executing it
		return nil
	}

	// Get the node registry
	registry := getNodeRegistry()

	// Get the node factory
	factory, exists := registry[nodeConfig.Type]
	if !exists {
		return fmt.Errorf("node type not registered: %s", nodeConfig.Type)
	}

	// Create the node instance
	nodeInstance := factory()

	// Collect input values from connected nodes and function inputs
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

			// Check if we have a result for this pin in the function context
			if value, exists := functionContext.GetInternalOutput(sourceNodeID, sourcePinID); exists {
				inputValues[targetPinID] = value
			}
		}
	}

	// Check if any input pins match function inputs
	for _, pin := range nodeInstance.GetInputPins() {
		if value, exists := functionContext.GetInputValue(pin.ID); exists && pin.Type != types.PinTypes.Execution {
			inputValues[pin.ID] = value
		}
	}

	// Create a function to activate output flows
	executionID := fmt.Sprintf("func-%s-%d", functionContext.GetBlueprintID(), time.Now().UnixNano())
	activateFlowFn := func(ctx *engine.DefaultExecutionContext, nodeID, pinID string) error {
		// Store all outputs in the function context
		for outPinID, outValue := range ctx.GetOutputs() {
			functionContext.StoreInternalOutput(nodeID, outPinID, outValue)

			// If this is an output pin of the function, set it in the function's outputs
			// This is for data flow pins of the function node itself
			for _, output := range n.GetOutputPins() {
				if output.ID == outPinID && output.Type != types.PinTypes.Execution {
					functionContext.SetOutputValue(outPinID, outValue)
				}
			}
		}

		// Find connections from this output pin
		outputConnections := bp.GetNodeOutputConnections(nodeID)
		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID

				// Execute the target node
				if err := n.executeNode(functionContext, targetNodeID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Create logger for this node
	nodeLogger := functionContext.Logger()

	// Create execution hooks
	hooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			nodeLogger.Debug("Function internal node started", map[string]interface{}{
				"nodeId":   nodeID,
				"nodeType": nodeType,
			})
		},
		OnNodeComplete: func(nodeID, nodeType string) {
			nodeLogger.Debug("Function internal node completed", map[string]interface{}{
				"nodeId":   nodeID,
				"nodeType": nodeType,
			})
		},
		OnNodeError: func(nodeID string, err error) {
			nodeLogger.Error("Function internal node error", map[string]interface{}{
				"nodeId": nodeID,
				"error":  err.Error(),
			})
		},
		OnPinValue: func(nodeID, pinName string, value interface{}) {
			// Store the output value in the function context
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

			functionContext.StoreInternalOutput(nodeID, pinName, types.NewValue(pinType, value))

			// Check if this pin maps to a function output
			for _, output := range n.GetOutputPins() {
				if output.ID == pinName && output.Type != types.PinTypes.Execution {
					functionContext.SetOutputValue(pinName, types.NewValue(pinType, value))
				}
			}
		},
	}

	// Create execution context
	ctx := engine.NewExecutionContext(
		nodeID,
		nodeConfig.Type,
		functionContext.GetBlueprintID(),
		executionID,
		inputValues,
		functionContext.GetAllVariables(),
		nodeLogger,
		hooks,
		activateFlowFn,
	)

	// Execute the node
	err = nodeInstance.Execute(ctx)
	if err != nil {
		return err
	}

	// Store all outputs in the function context
	for pinID, value := range ctx.GetOutputs() {
		functionContext.StoreInternalOutput(nodeID, pinID, value)

		// If this is an output pin of the function, set it in the function's outputs
		for _, output := range n.GetOutputPins() {
			if output.ID == pinID && output.Type != types.PinTypes.Execution {
				functionContext.SetOutputValue(pinID, value)
			}
		}
	}

	// Get activated flows
	activatedFlows := ctx.GetActivatedOutputFlows()

	// Process function output execution flows
	for _, outputPin := range activatedFlows {
		// Check if this flow maps to a function output flow
		for _, output := range n.GetOutputPins() {
			if output.ID == outputPin && output.Type == types.PinTypes.Execution {
				functionContext.ActivateOutputFlow(outputPin)
			}
		}

		// Process internal connections
		if err := activateFlowFn(ctx, nodeID, outputPin); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to map blueprint pin types to internal pin types
func mapPinType(pinType *blueprint.NodePinType) *types.PinType {
	if pinType == nil {
		return types.PinTypes.Any
	}

	switch pinType.ID {
	case "execution":
		return types.PinTypes.Execution
	case "string":
		return types.PinTypes.String
	case "number":
		return types.PinTypes.Number
	case "boolean":
		return types.PinTypes.Boolean
	case "object":
		return types.PinTypes.Object
	case "array":
		return types.PinTypes.Array
	default:
		return types.PinTypes.Any
	}
}

// Helper function to get the node registry
func getNodeRegistry() map[string]node.NodeFactory {
	// Access the global node registry
	return registry.GetInstance().GetAllNodeFactories()
}
