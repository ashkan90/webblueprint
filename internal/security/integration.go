package security

import (
	"context"
	"fmt"
	"webblueprint/internal/common"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// IntegrateWithEngine integrates security features with the execution engine
func IntegrateWithEngine(engine interface{}) *SecureEngineAdapter {
	adapter := NewSecureEngineAdapter()

	// Hook into engine via existing hooks

	// This is where we would add hooks to the engine to integrate security
	// In a real implementation, we would need to modify the engine to support these hooks
	// Since we can't modify the engine directly in this example, this is a placeholder

	return adapter
}

// SecureExecutionWrapper wraps the Execute method to add security checks
func SecureExecutionWrapper(
	adapter *SecureEngineAdapter,
	execEngine interface{},
) func(bp *blueprint.Blueprint, executionID string, userID string, initialData map[string]types.Value) (common.ExecutionResult, error) {
	return func(bp *blueprint.Blueprint, executionID string, userID string, initialData map[string]types.Value) (common.ExecutionResult, error) {
		// Perform pre-execution security checks
		if err := adapter.CheckSecurityPreExecution(bp, userID); err != nil {
			return common.ExecutionResult{
				ExecutionID: executionID,
				Success:     false,
				Error:       err,
			}, err
		}

		// Check variables for security issues
		if err := adapter.CheckBlueprintVariables(bp, initialData); err != nil {
			return common.ExecutionResult{
				ExecutionID: executionID,
				Success:     false,
				Error:       err,
			}, err
		}

		// Create custom hooks for tracking
		hooks := &node.ExecutionHooks{
			OnNodeStart: func(nodeID, nodeType string) {
				// Track node execution with security manager
				// In a real implementation, we would use this to enforce resource limits
			},
			OnNodeComplete: func(nodeID, nodeType string) {
				// Track node completion
			},
			OnNodeError: func(nodeID string, err error) {
				// Track node errors
			},
		}

		// We would use these hooks in a real implementation
		_ = hooks

		// Here we would wrap all execution contexts with sandboxed contexts
		// In a real implementation, we would replace the engine's NewExecutionContext method

		// Execute the blueprint
		// Execute the blueprint using type assertion
		result := common.ExecutionResult{}
		var err error

		// Try to call Execute method using type assertion
		if engine, ok := execEngine.(interface {
			Execute(bp *blueprint.Blueprint, executionID string, initialData map[string]types.Value) (common.ExecutionResult, error)
		}); ok {
			result, err = engine.Execute(bp, executionID, initialData)
		} else {
			err = fmt.Errorf("engine does not implement Execute method")
		}

		// Clean up resources
		adapter.RemoveCachedContext(userID, bp.ID, executionID)

		return result, err
	}
}

// CreateSecureExecutionContext creates a secure execution context
func CreateSecureExecutionContext(
	adapter *SecureEngineAdapter,
	nodeID string,
	nodeType string,
	blueprintID string,
	executionID string,
	userID string,
	inputs map[string]types.Value,
	variables map[string]types.Value,
	logger node.Logger,
	hooks *node.ExecutionHooks,
	activateFlow func(ctx context.Context, nodeID, pinID string) error,
) node.ExecutionContext {
	// Create a base execution context - this is normally a standard execution context
	// For this example, we're creating a mock implementation
	baseCtx := &mockExecutionContext{
		nodeID:      nodeID,
		nodeType:    nodeType,
		blueprintID: blueprintID,
		executionID: executionID,
		inputs:      inputs,
		outputs:     make(map[string]types.Value),
		variables:   variables,
		logger:      logger,
		hooks:       hooks,
	}

	// Wrap with sandboxed context
	return adapter.WrapExecutionContext(baseCtx, userID, blueprintID, executionID)
}

// A simple mock implementation of execution context for demonstration
type mockExecutionContext struct {
	nodeID      string
	nodeType    string
	blueprintID string
	executionID string
	inputs      map[string]types.Value
	outputs     map[string]types.Value
	variables   map[string]types.Value
	logger      node.Logger
	hooks       *node.ExecutionHooks
}

// IsInputPinActive checks if an input pin is active
func (m *mockExecutionContext) IsInputPinActive(pinID string) bool {
	// For the mock, assume the execute pin is always active
	if pinID == "execute" {
		return true
	}
	return false
}

func (m *mockExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := m.inputs[pinID]
	return value, exists
}

func (m *mockExecutionContext) SetOutputValue(pinID string, value types.Value) {
	m.outputs[pinID] = value
}

func (m *mockExecutionContext) GetOutputValue(pinID string) (types.Value, bool) {
	value, exists := m.outputs[pinID]
	return value, exists
}

func (m *mockExecutionContext) ActivateOutputFlow(pinID string) error {
	return nil
}

func (m *mockExecutionContext) GetVariable(name string) (types.Value, bool) {
	value, exists := m.variables[name]
	return value, exists
}

func (m *mockExecutionContext) SetVariable(name string, value types.Value) {
	m.variables[name] = value
}

func (m *mockExecutionContext) Logger() node.Logger {
	return m.logger
}

func (m *mockExecutionContext) ExecuteConnectedNodes(pinID string) error {
	return nil
}

func (m *mockExecutionContext) GetActivatedOutputFlows() []string {
	return []string{}
}

func (m *mockExecutionContext) GetNodeID() string {
	return m.nodeID
}

func (m *mockExecutionContext) GetNodeType() string {
	return m.nodeType
}

func (m *mockExecutionContext) GetBlueprintID() string {
	return m.blueprintID
}

func (m *mockExecutionContext) GetExecutionID() string {
	return m.executionID
}

func (m *mockExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	// No-op in mock
}

func (m *mockExecutionContext) GetDebugData() map[string]interface{} {
	return make(map[string]interface{})
}
