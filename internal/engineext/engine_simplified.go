package engineext

//import (
//	"fmt"
//	"sync"
//	"time"
//	"webblueprint/internal/bperrors"
//	"webblueprint/internal/core"
//	"webblueprint/internal/node"
//	"webblueprint/internal/types"
//)
//
//// DefaultExecutionContext is a simplified implementation of node.ExecutionContext
//type DefaultExecutionContext struct {
//	nodeID             string
//	nodeType           string
//	blueprintID        string
//	executionID        string
//	inputs             map[string]types.Value
//	outputs            map[string]types.Value
//	variables          map[string]types.Value
//	debugData          map[string]interface{}
//	logger             node.Logger
//	hooks              *node.ExecutionHooks
//	activateFlow       func(ctx *DefaultExecutionContext, nodeID, pinID string) error
//	activatedFlows     []string
//	activatedFlowMutex sync.Mutex
//	activePins         map[string]bool
//}
//
//// NewExecutionContext creates a new execution context
//func NewExecutionContext(
//	nodeID string,
//	nodeType string,
//	blueprintID string,
//	executionID string,
//	inputs map[string]types.Value,
//	variables map[string]types.Value,
//	logger node.Logger,
//	hooks *node.ExecutionHooks,
//	activateFlow func(ctx *DefaultExecutionContext, nodeID, pinID string) error,
//) *DefaultExecutionContext {
//	if logger != nil {
//		logger.Opts(map[string]interface{}{"nodeId": nodeID})
//	}
//	return &DefaultExecutionContext{
//		nodeID:         nodeID,
//		nodeType:       nodeType,
//		blueprintID:    blueprintID,
//		executionID:    executionID,
//		inputs:         inputs,
//		outputs:        make(map[string]types.Value),
//		variables:      variables,
//		debugData:      make(map[string]interface{}),
//		logger:         logger,
//		hooks:          hooks,
//		activateFlow:   activateFlow,
//		activatedFlows: make([]string, 0),
//		activePins:     make(map[string]bool),
//	}
//}
//
//// SetActiveInputPin marks an input pin as the one that triggered execution
//func (ctx *DefaultExecutionContext) SetActiveInputPin(pinID string) {
//	ctx.activePins[pinID] = true
//}
//
//// IsInputPinActive checks if the input pin triggered execution
//func (ctx *DefaultExecutionContext) IsInputPinActive(pinID string) bool {
//	if len(ctx.activePins) == 0 && pinID == "execute" {
//		return true
//	}
//
//	if active, exists := ctx.activePins[pinID]; exists && active {
//		return true
//	}
//
//	return false
//}
//
//// GetInputValue retrieves an input value by pin ID
//func (ctx *DefaultExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
//	value, exists := ctx.inputs[pinID]
//	if exists {
//		if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
//			ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
//		}
//		return value, true
//	}
//	return types.Value{}, false
//}
//
//// SetOutputValue sets an output value by pin ID
//func (ctx *DefaultExecutionContext) SetOutputValue(pinID string, value types.Value) {
//	ctx.outputs[pinID] = value
//
//	if ctx.hooks != nil && ctx.hooks.OnPinValue != nil {
//		ctx.hooks.OnPinValue(ctx.nodeID, pinID, value.RawValue)
//	}
//}
//
//// GetOutputValue retrieves an output value by pin ID
//func (ctx *DefaultExecutionContext) GetOutputValue(pinID string) (types.Value, bool) {
//	value, exists := ctx.outputs[pinID]
//	return value, exists
//}
//
//// ActivateOutputFlow activates an output execution flow
//func (ctx *DefaultExecutionContext) ActivateOutputFlow(pinID string) error {
//	ctx.activatedFlowMutex.Lock()
//	defer ctx.activatedFlowMutex.Unlock()
//
//	ctx.activatedFlows = append(ctx.activatedFlows, pinID)
//	return nil
//}
//
//// ExecuteConnectedNodes executes all nodes connected to the given output pin
//func (ctx *DefaultExecutionContext) ExecuteConnectedNodes(pinID string) error {
//	if ctx.logger != nil {
//		ctx.logger.Debug("Executing connected nodes", map[string]interface{}{
//			"pin": pinID,
//		})
//	}
//
//	if ctx.activateFlow != nil {
//		return ctx.activateFlow(ctx, ctx.nodeID, pinID)
//	}
//	return nil
//}
//
//// GetActivatedOutputFlows returns the list of output pins that were activated
//func (ctx *DefaultExecutionContext) GetActivatedOutputFlows() []string {
//	ctx.activatedFlowMutex.Lock()
//	defer ctx.activatedFlowMutex.Unlock()
//
//	result := make([]string, len(ctx.activatedFlows))
//	copy(result, ctx.activatedFlows)
//	return result
//}
//
//// GetVariable retrieves a variable by name
//func (ctx *DefaultExecutionContext) GetVariable(name string) (types.Value, bool) {
//	value, exists := ctx.variables[name]
//	return value, exists
//}
//
//// SetVariable sets a variable by name
//func (ctx *DefaultExecutionContext) SetVariable(name string, value types.Value) {
//	ctx.variables[name] = value
//}
//
//// Logger returns the execution logger
//func (ctx *DefaultExecutionContext) Logger() node.Logger {
//	return ctx.logger
//}
//
//// RecordDebugInfo stores debug information
//func (ctx *DefaultExecutionContext) RecordDebugInfo(info types.DebugInfo) {
//	key := fmt.Sprintf("debug_%d", time.Now().UnixNano())
//	ctx.debugData[key] = info
//}
//
//// GetDebugData returns all debug data
//func (ctx *DefaultExecutionContext) GetDebugData() map[string]interface{} {
//	return ctx.debugData
//}
//
//// GetNodeID returns the ID of the executing node
//func (ctx *DefaultExecutionContext) GetNodeID() string {
//	return ctx.nodeID
//}
//
//// GetNodeType returns the type of the executing node
//func (ctx *DefaultExecutionContext) GetNodeType() string {
//	return ctx.nodeType
//}
//
//// GetBlueprintID returns the ID of the executing blueprint
//func (ctx *DefaultExecutionContext) GetBlueprintID() string {
//	return ctx.blueprintID
//}
//
//// GetExecutionID returns the current execution ID
//func (ctx *DefaultExecutionContext) GetExecutionID() string {
//	return ctx.executionID
//}
//
//// GetOutputs returns all outputs from this execution context
//func (ctx *DefaultExecutionContext) GetOutputs() map[string]types.Value {
//	outputsCopy := make(map[string]types.Value)
//	for k, v := range ctx.outputs {
//		outputsCopy[k] = v
//	}
//	return outputsCopy
//}
