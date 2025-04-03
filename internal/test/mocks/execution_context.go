package mocks

import (
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// MockExecutionContext implements the ExecutionContext interface for testing
type MockExecutionContext struct {
	nodeID        string
	nodeType      string
	blueprintID   string
	executionID   string
	inputValues   map[string]types.Value
	outputValues  map[string]types.Value
	variables     map[string]types.Value
	activatedFlow string
	logger        node.Logger
	debugInfo     []types.DebugInfo
	debugData     map[string]interface{}
	executedPins  map[string]bool
	activePins    map[string]bool
}

// NewMockExecutionContext creates a new mock execution context for testing
func NewMockExecutionContext(nodeID string, nodeType string, logger node.Logger) *MockExecutionContext {
	return &MockExecutionContext{
		nodeID:       nodeID,
		nodeType:     nodeType,
		blueprintID:  "test-blueprint",
		executionID:  "test-execution",
		inputValues:  make(map[string]types.Value),
		outputValues: make(map[string]types.Value),
		variables:    make(map[string]types.Value),
		logger:       logger,
		debugData:    make(map[string]interface{}),
		executedPins: make(map[string]bool),
		activePins:   make(map[string]bool),
	}
}

// IsInputPinActive checks if an input pin is active
func (m *MockExecutionContext) IsInputPinActive(pinID string) bool {
	// By default, assume the "execute" pin is active when no specific active pin is set
	if len(m.activePins) == 0 && pinID == "execute" {
		return true
	}

	active, exists := m.activePins[pinID]
	return exists && active
}

// SetActiveInputPin sets an input pin as active
func (m *MockExecutionContext) SetActiveInputPin(pinID string, active bool) {
	m.activePins[pinID] = active
}

// GetInputValue returns an input value
func (m *MockExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	val, exists := m.inputValues[pinID]
	return val, exists
}

// SetInputValue sets an input value for testing
func (m *MockExecutionContext) SetInputValue(pinID string, value types.Value) {
	m.inputValues[pinID] = value
}

// SetOutputValue sets an output value
func (m *MockExecutionContext) SetOutputValue(pinID string, value types.Value) {
	m.outputValues[pinID] = value
}

// GetOutputValue gets an output value for assertions
func (m *MockExecutionContext) GetOutputValue(pinID string) (types.Value, bool) {
	val, exists := m.outputValues[pinID]
	return val, exists
}

// ActivateOutputFlow activates an output flow
func (m *MockExecutionContext) ActivateOutputFlow(pinID string) error {
	m.activatedFlow = pinID
	return nil
}

// GetActivatedFlow returns which flow was activated
func (m *MockExecutionContext) GetActivatedFlow() string {
	return m.activatedFlow
}

// ExecuteConnectedNodes executes nodes connected to an output pin
func (m *MockExecutionContext) ExecuteConnectedNodes(pinID string) error {
	m.executedPins[pinID] = true
	return nil
}

// WasExecuted returns whether a pin was executed
func (m *MockExecutionContext) WasExecuted(pinID string) bool {
	return m.executedPins[pinID]
}

// GetVariable gets a variable
func (m *MockExecutionContext) GetVariable(name string) (types.Value, bool) {
	val, exists := m.variables[name]
	return val, exists
}

// SetVariable sets a variable
func (m *MockExecutionContext) SetVariable(name string, value types.Value) {
	m.variables[name] = value
}

// Logger returns the logger
func (m *MockExecutionContext) Logger() node.Logger {
	return m.logger
}

// RecordDebugInfo records debug info
func (m *MockExecutionContext) RecordDebugInfo(info types.DebugInfo) {
	m.debugInfo = append(m.debugInfo, info)
}

// GetDebugInfos returns all recorded debug info
func (m *MockExecutionContext) GetDebugInfos() []types.DebugInfo {
	return m.debugInfo
}

// GetDebugData returns debug data
func (m *MockExecutionContext) GetDebugData() map[string]interface{} {
	return m.debugData
}

// GetNodeID returns the node ID
func (m *MockExecutionContext) GetNodeID() string {
	return m.nodeID
}

// GetNodeType returns the node type
func (m *MockExecutionContext) GetNodeType() string {
	return m.nodeType
}

// GetBlueprintID returns the blueprint ID
func (m *MockExecutionContext) GetBlueprintID() string {
	return m.blueprintID
}

// GetExecutionID returns the execution ID
func (m *MockExecutionContext) GetExecutionID() string {
	return m.executionID
}
