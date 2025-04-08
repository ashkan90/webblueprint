package engineext

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"webblueprint/pkg/blueprint"

	// "webblueprint/internal/bperrors" // Not used directly in this basic context
	// "webblueprint/internal/core" // Not used directly in this basic context
	"webblueprint/internal/db" // Added import for db package
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/repository" // Added import
)

// DefaultExecutionContext is a simplified implementation of node.ExecutionContext
// It also implements node.ExtendedExecutionContext
type DefaultExecutionContext struct {
	storeCtx           context.Context
	nodeID             string
	nodeType           string
	blueprintID        string
	executionID        string
	inputs             map[string]types.Value
	outputs            map[string]types.Value // Unexported, accessed via methods
	variables          map[string]types.Value
	debugData          map[string]interface{}
	logger             node.Logger
	hooks              *node.ExecutionHooks
	activateFlow       func(ctx *DefaultExecutionContext, nodeID, pinID string) error
	activatedFlows     []string
	activatedFlowMutex sync.Mutex
	activePins         map[string]bool // Tracks which input execution pin was activated
	mutex              sync.RWMutex
	repoFactory        repository.RepositoryFactory // Added field
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
	storeContext context.Context,
	repoFactory repository.RepositoryFactory, // Added parameter
) *DefaultExecutionContext {
	if logger != nil {
		logger.Opts(map[string]interface{}{"nodeId": nodeID})
	}
	return &DefaultExecutionContext{
		storeCtx:       storeContext,
		nodeID:         nodeID,
		nodeType:       nodeType,
		blueprintID:    blueprintID,
		executionID:    executionID,
		inputs:         inputs,
		outputs:        make(map[string]types.Value),
		variables:      variables,
		debugData:      make(map[string]interface{}),
		logger:         logger,
		hooks:          hooks,
		activateFlow:   activateFlow,
		activatedFlows: make([]string, 0),
		activePins:     make(map[string]bool),
		repoFactory:    repoFactory, // Added assignment
	}
}

// SetActiveInputPin marks an input pin as the one that triggered execution
func (ctx *DefaultExecutionContext) SetActiveInputPin(pinID string) {
	ctx.activePins[pinID] = true
}

// IsInputPinActive checks if the input pin triggered execution
func (ctx *DefaultExecutionContext) IsInputPinActive(pinID string) bool {
	// Default execution pin if no specific pin was activated
	if len(ctx.activePins) == 0 && pinID == "execute" {
		return true
	}
	// Check if the specific pin was activated
	if active, exists := ctx.activePins[pinID]; exists && active {
		return true
	}
	return false
}

// GetInputValue retrieves an input value by pin ID
func (ctx *DefaultExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
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
	// Get input connections for this node
	bp := ctx.storeCtx.Value("bp").(*blueprint.Blueprint)
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

			// TODO: some confusion going here ?
			//if sourceNode != nil {
			//	for _, property := range sourceNode.Properties {
			//		var propType = ctx.resolveConstantType(ctx.nodeType)
			//		return types.NewValue(propType, property.Value), true
			//	}
			//}
		}
	}

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
	inputPins, ok := ctx.storeCtx.Value("node.inputPins").([]types.Pin)
	if !ok {
		ctx.Logger().Error("somehow DefaultExecutionContext has no inputPins", nil)
	}

	defNode := bp.FindNode(ctx.GetNodeID())
	for _, pin := range inputPins {
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
		} else if pin.ID == pinID && defNode != nil {
			if defaults, ok := defNode.Data["defaults"].(map[string]interface{}); ok {
				defaultValue := types.NewValue(pin.Type, defaults[pin.ID])

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
	}

	// No value or default found
	return types.Value{}, false
}

// SetOutputValue sets an output value by pin ID
func (ctx *DefaultExecutionContext) SetOutputValue(pinID string, value types.Value) {
	ctx.outputs[pinID] = value
	// Trigger hook when value is set
	if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
		ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
	}
}

// GetOutputValue retrieves an output value by pin ID (implements ExtendedExecutionContext)
func (ctx *DefaultExecutionContext) GetOutputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.outputs[pinID]
	return value, exists
}

func (ctx *DefaultExecutionContext) SetInput(pingID string, value types.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.inputs[pingID] = value
}

// ActivateOutputFlow activates an output execution flow
func (ctx *DefaultExecutionContext) ActivateOutputFlow(pinID string) error {
	ctx.activatedFlowMutex.Lock()
	// Check if already activated to prevent duplicates if called multiple times
	alreadyActivated := false
	for _, flow := range ctx.activatedFlows {
		if flow == pinID {
			alreadyActivated = true
			break
		}
	}
	if !alreadyActivated {
		ctx.activatedFlows = append(ctx.activatedFlows, pinID)
	}
	ctx.activatedFlowMutex.Unlock()

	// Call the engine's activateFlow function immediately
	if ctx.activateFlow != nil {
		return ctx.activateFlow(ctx, ctx.nodeID, pinID)
	}
	return nil
}

// ExecuteConnectedNodes executes all nodes connected to the given output pin
// This might be redundant if ActivateOutputFlow handles it.
func (ctx *DefaultExecutionContext) ExecuteConnectedNodes(pinID string) error {
	if ctx.logger != nil {
		ctx.logger.Debug("Executing connected nodes (via ExecuteConnectedNodes)", map[string]interface{}{
			"pin": pinID,
		})
	}
	if ctx.activateFlow != nil {
		return ctx.activateFlow(ctx, ctx.nodeID, pinID)
	}
	return nil
}

// GetActivatedOutputFlows returns the list of output pins that were activated (implements ExtendedExecutionContext)
func (ctx *DefaultExecutionContext) GetActivatedOutputFlows() []string {
	ctx.activatedFlowMutex.Lock()
	defer ctx.activatedFlowMutex.Unlock()
	// Return a copy to prevent external modification
	result := make([]string, len(ctx.activatedFlows))
	copy(result, ctx.activatedFlows)
	return result
}

// GetVariable retrieves a variable by name
func (ctx *DefaultExecutionContext) GetVariable(name string) (types.Value, bool) {
	value, exists := ctx.variables[name]
	return value, exists
}

// SetVariable sets a variable by name
func (ctx *DefaultExecutionContext) SetVariable(name string, value types.Value) {
	// TODO: Consider adding hooks for variable changes if needed
	ctx.variables[name] = value
}

// Logger returns the execution logger
func (ctx *DefaultExecutionContext) Logger() node.Logger {
	return ctx.logger
}

// RecordDebugInfo stores debug information
func (ctx *DefaultExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// Use a more structured key if needed, e.g., based on timestamp or sequence
	key := fmt.Sprintf("debug_%d_%s", time.Now().UnixNano(), info.PinID)
	if ctx.mutex.TryLock() {
		defer ctx.mutex.Unlock()
	}
	ctx.debugData[key] = info

}

// GetDebugData returns all debug data
func (ctx *DefaultExecutionContext) GetDebugData() map[string]interface{} {
	// Return a copy to prevent external modification
	dataCopy := make(map[string]interface{}, len(ctx.debugData))
	for k, v := range ctx.debugData {
		dataCopy[k] = v
	}
	return dataCopy
}

// GetNodeID returns the ID of the executing node
func (ctx *DefaultExecutionContext) GetNodeID() string {
	return ctx.nodeID
}

// GetNodeType returns the type of the executing node
func (ctx *DefaultExecutionContext) GetNodeType() string {
	return ctx.nodeType
}

// GetBlueprintID returns the ID of the executing blueprint
func (ctx *DefaultExecutionContext) GetBlueprintID() string {
	return ctx.blueprintID
}

// GetExecutionID returns the current execution ID
func (ctx *DefaultExecutionContext) GetExecutionID() string {
	return ctx.executionID
}

// GetAllOutputs returns all outputs from this execution context (implements ExtendedExecutionContext)
func (ctx *DefaultExecutionContext) GetAllOutputs() map[string]types.Value {
	// Return a copy to prevent external modification
	outputsCopy := make(map[string]types.Value)
	for k, v := range ctx.outputs {
		outputsCopy[k] = v
	}
	return outputsCopy
}

func (ctx *DefaultExecutionContext) SaveData(key string, value interface{}) {
	ctx.storeCtx = context.WithValue(ctx.storeCtx, key, value)
}

func (ctx *DefaultExecutionContext) GetProperty(name string) (interface{}, bool) {
	properties, ok := ctx.storeCtx.Value("node.properties").([]types.Property)
	if !ok {
		return nil, false
	}

	for _, prop := range properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}

	return nil, false
}

func (ctx *DefaultExecutionContext) resolveConstantType(nodeType string) *types.PinType {
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
} // End of resolveConstantType

// GetSchemaComponentStore returns the SchemaComponentStore from the repository factory
// This method makes DefaultExecutionContext implement node.SchemaAccessContext
func (ctx *DefaultExecutionContext) GetSchemaComponentStore() db.SchemaComponentStore {
	if ctx.repoFactory == nil {
		// This shouldn't happen if injection is set up correctly, but return nil defensively.
		// Consider logging an error here.
		if ctx.logger != nil {
			ctx.logger.Error("RepositoryFactory is nil in execution context", nil)
		}
		return nil
	}
	return ctx.repoFactory.GetSchemaComponentStore()
}

// getPropertyValue retrieves a property value from the actor's properties
func (ctx *DefaultExecutionContext) getPropertyValue(name string) (interface{}, bool) {
	return ctx.GetProperty(name)
}

// getPinTypeForInput determines the appropriate type for an input pin
func (ctx *DefaultExecutionContext) getPinTypeForInput(pinID string) *types.PinType {
	// Try to get the type from pin definitions
	inputPins, ok := ctx.storeCtx.Value("node.inputPins").([]types.Pin)
	if !ok {
		return nil
	}

	for _, pin := range inputPins {
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
