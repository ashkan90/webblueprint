package bperrors

import (
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ErrorAwareContext is an interface that extends the node.ExecutionContext
// with error handling capabilities
type ErrorAwareContext interface {
	// Embed the standard ExecutionContext interface to ensure compatibility
	node.ExecutionContext

	// Error handling methods
	ReportError(errType ErrorType, code BlueprintErrorCode, message string, originalErr error) *BlueprintError
	AttemptRecovery(err *BlueprintError) (bool, map[string]interface{})
	GetErrorSummary() map[string]interface{}

	// Default value management
	GetDefaultValue(pinType *types.PinType) (types.Value, bool)
}

// ErrorContextWrapper wraps an existing ExecutionContext to add error handling
// This allows us to adapt any existing execution context implementation
type ErrorContextWrapper struct {
	// The wrapped execution context
	Context node.ExecutionContext

	// Error handling components
	ErrorManager    *ErrorManager
	RecoveryManager *RecoveryManager
}

// NewErrorContextWrapper creates a new wrapper around an existing execution context
func NewErrorContextWrapper(ctx node.ExecutionContext, em *ErrorManager, rm *RecoveryManager) *ErrorContextWrapper {
	return &ErrorContextWrapper{
		Context:         ctx,
		ErrorManager:    em,
		RecoveryManager: rm,
	}
}

// Forward all standard ExecutionContext methods to the wrapped context

// IsInputPinActive checks if an input pin triggered execution
func (w *ErrorContextWrapper) IsInputPinActive(pinID string) bool {
	return w.Context.IsInputPinActive(pinID)
}

// GetInputValue retrieves an input value by pin ID, with error recovery if needed
func (w *ErrorContextWrapper) GetInputValue(pinID string) (types.Value, bool) {
	// First try to get the value from the original context
	value, exists := w.Context.GetInputValue(pinID)

	// If the value doesn't exist, try to recover with a default value
	if !exists {
		// Get the pin type by looking at the node definition
		// In a real implementation, you would need to get this from the node metadata
		pinType := types.PinTypes.Any // Fallback type

		// Create a BlueprintError
		err := New(
			ErrorTypeExecution,
			ErrMissingRequiredInput,
			"Required input value missing",
			SeverityMedium,
		).WithNodeInfo(w.Context.GetNodeID(), pinID).WithBlueprintInfo(w.Context.GetBlueprintID(), w.Context.GetExecutionID())

		// Record the error
		w.ErrorManager.RecordError(w.Context.GetExecutionID(), err)

		// Attempt recovery
		if success, _ := w.RecoveryManager.RecoverFromError(w.Context.GetExecutionID(), err); success {
			// Get a default value for this type
			if defaultValue, err := w.RecoveryManager.GetDefaultValue(pinType); err == nil {
				// Log the recovery
				w.Context.Logger().Info("Recovered from missing input by using default value", map[string]interface{}{
					"nodeId": w.Context.GetNodeID(),
					"pinId":  pinID,
					"value":  defaultValue.RawValue,
				})

				return defaultValue, true
			}
		}
	}

	return value, exists
}

// SetOutputValue sets an output value by pin ID
func (w *ErrorContextWrapper) SetOutputValue(pinID string, value types.Value) {
	w.Context.SetOutputValue(pinID, value)
}

// ActivateOutputFlow activates an output execution flow
func (w *ErrorContextWrapper) ActivateOutputFlow(pinID string) error {
	return w.Context.ActivateOutputFlow(pinID)
}

// ExecuteConnectedNodes executes all nodes connected to the given output pin
func (w *ErrorContextWrapper) ExecuteConnectedNodes(pinID string) error {
	return w.Context.ExecuteConnectedNodes(pinID)
}

// GetVariable retrieves a variable by name
func (w *ErrorContextWrapper) GetVariable(name string) (types.Value, bool) {
	return w.Context.GetVariable(name)
}

// SetVariable sets a variable by name
func (w *ErrorContextWrapper) SetVariable(name string, value types.Value) {
	w.Context.SetVariable(name, value)
}

// Logger returns the execution logger
func (w *ErrorContextWrapper) Logger() node.Logger {
	return w.Context.Logger()
}

// RecordDebugInfo stores debug information
func (w *ErrorContextWrapper) RecordDebugInfo(info types.DebugInfo) {
	w.Context.RecordDebugInfo(info)
}

// GetDebugData returns all debug data
func (w *ErrorContextWrapper) GetDebugData() map[string]interface{} {
	return w.Context.GetDebugData()
}

// GetNodeID returns the ID of the executing node
func (w *ErrorContextWrapper) GetNodeID() string {
	return w.Context.GetNodeID()
}

// GetNodeType returns the type of the executing node
func (w *ErrorContextWrapper) GetNodeType() string {
	return w.Context.GetNodeType()
}

// GetBlueprintID returns the ID of the executing blueprint
func (w *ErrorContextWrapper) GetBlueprintID() string {
	return w.Context.GetBlueprintID()
}

// GetExecutionID returns the current execution ID
func (w *ErrorContextWrapper) GetExecutionID() string {
	return w.Context.GetExecutionID()
}

// Error handling methods

// ReportError reports an error during node execution
func (w *ErrorContextWrapper) ReportError(errType ErrorType, code BlueprintErrorCode, message string, originalErr error) *BlueprintError {
	// Create a BlueprintError
	severity := SeverityMedium // Default severity

	// Set severity based on error type
	switch errType {
	case ErrorTypeExecution:
		severity = SeverityHigh
	case ErrorTypeConnection:
		severity = SeverityMedium
	case ErrorTypeValidation:
		severity = SeverityMedium
	case ErrorTypePermission:
		severity = SeverityHigh
	case ErrorTypeDatabase:
		severity = SeverityHigh
	default:
		severity = SeverityMedium
	}

	// Create the error
	var err *BlueprintError
	if originalErr != nil {
		err = Wrap(
			originalErr,
			errType,
			code,
			message,
			severity,
		)
	} else {
		err = New(
			errType,
			code,
			message,
			severity,
		)
	}

	// Add context
	err.WithNodeInfo(w.Context.GetNodeID(), "").WithBlueprintInfo(w.Context.GetBlueprintID(), w.Context.GetExecutionID())

	// Record the error
	w.ErrorManager.RecordError(w.Context.GetExecutionID(), err)

	// Log the error
	w.Context.Logger().Error(message, map[string]interface{}{
		"nodeId":      w.Context.GetNodeID(),
		"errorType":   string(errType),
		"errorCode":   string(code),
		"originalErr": originalErr,
	})

	return err
}

// AttemptRecovery tries to recover from an error
func (w *ErrorContextWrapper) AttemptRecovery(err *BlueprintError) (bool, map[string]interface{}) {
	return w.RecoveryManager.RecoverFromError(w.Context.GetExecutionID(), err)
}

// GetErrorSummary gets a summary of errors for this node
func (w *ErrorContextWrapper) GetErrorSummary() map[string]interface{} {
	nodeErrors := w.ErrorManager.GetNodeErrors(w.Context.GetExecutionID(), w.Context.GetNodeID())

	if len(nodeErrors) == 0 {
		return map[string]interface{}{
			"hasErrors": false,
		}
	}

	// Summarize error data
	errorTypes := make(map[string]int)
	errorCodes := make(map[string]int)
	severityCounts := make(map[string]int)

	for _, err := range nodeErrors {
		errorTypes[string(err.Type)]++
		errorCodes[string(err.Code)]++
		severityCounts[string(err.Severity)]++
	}

	return map[string]interface{}{
		"hasErrors":      true,
		"errorCount":     len(nodeErrors),
		"errorTypes":     errorTypes,
		"errorCodes":     errorCodes,
		"severityCounts": severityCounts,
		"latestError":    nodeErrors[len(nodeErrors)-1],
	}
}

// GetDefaultValue gets a default value for a pin type
func (w *ErrorContextWrapper) GetDefaultValue(pinType *types.PinType) (types.Value, bool) {
	value, err := w.RecoveryManager.GetDefaultValue(pinType)
	if err != nil {
		return types.Value{}, false
	}
	return value, true
}
