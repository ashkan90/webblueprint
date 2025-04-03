package engineext

import (
	"webblueprint/internal/bperrors"
	"webblueprint/internal/core"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// ErrorAwareContext extends DefaultExecutionContext with error handling capabilities
type ErrorAwareContext struct {
	node.ExecutionContext
	errorManager    *bperrors.ErrorManager
	recoveryManager *bperrors.RecoveryManager
}

// NewErrorAwareContext creates a new error-aware execution context
func NewErrorAwareContext(
	baseCtx node.ExecutionContext,
	errorManager *bperrors.ErrorManager,
	recoveryManager *bperrors.RecoveryManager,
) *ErrorAwareContext {
	return &ErrorAwareContext{
		ExecutionContext: baseCtx,
		errorManager:     errorManager,
		recoveryManager:  recoveryManager,
	}
}

// GetInputValue retrieves an input value with error handling
func (ctx *ErrorAwareContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.ExecutionContext.GetInputValue(pinID)

	// If the value doesn't exist, try to recover with a default value
	if !exists {
		// Create a BlueprintError
		err := bperrors.New(
			bperrors.ErrorTypeExecution,
			bperrors.ErrMissingRequiredInput,
			"Required input value missing",
			bperrors.SeverityMedium,
		).WithNodeInfo(ctx.GetNodeID(), pinID).WithBlueprintInfo(ctx.GetBlueprintID(), ctx.GetExecutionID())

		// Record the error
		ctx.errorManager.RecordError(ctx.GetExecutionID(), err)

		// Attempt recovery using default value strategy
		if success, _ := ctx.recoveryManager.RecoverFromError(ctx.GetExecutionID(), err); success {
			// Find the pin type for this pin (fallback to Any if not found)
			pinType := types.PinTypes.Any

			// Get a default value for this type
			if defaultValue, err := ctx.recoveryManager.GetDefaultValue(pinType); err == nil {
				// Log the recovery
				ctx.Logger().Info("Recovered from missing input by using default value", map[string]interface{}{
					"nodeId": ctx.GetNodeID(),
					"pinId":  pinID,
					"value":  defaultValue.RawValue,
				})

				// Add to debug data
				ctx.RecordDebugInfo(types.DebugInfo{
					NodeID:      ctx.GetNodeID(),
					PinID:       pinID,
					Description: "Recovered from missing input using default value",
					Value: map[string]interface{}{
						"default":  defaultValue.RawValue,
						"recovery": "default_value",
					},
				})

				return defaultValue, true
			}
		}
	}

	return value, exists
}

// ReportError reports an error during node execution
func (ctx *ErrorAwareContext) ReportError(
	errType bperrors.ErrorType,
	code bperrors.BlueprintErrorCode,
	message string,
	originalErr error,
) *bperrors.BlueprintError {
	// Create a BlueprintError
	severity := bperrors.SeverityMedium

	// Set severity based on error type
	switch errType {
	case bperrors.ErrorTypeExecution:
		severity = bperrors.SeverityHigh
	case bperrors.ErrorTypeConnection:
		severity = bperrors.SeverityMedium
	case bperrors.ErrorTypeValidation:
		severity = bperrors.SeverityMedium
	case bperrors.ErrorTypePermission:
		severity = bperrors.SeverityHigh
	case bperrors.ErrorTypeDatabase:
		severity = bperrors.SeverityHigh
	default:
		severity = bperrors.SeverityMedium
	}

	// Create the error
	var err *bperrors.BlueprintError
	if originalErr != nil {
		err = bperrors.Wrap(
			originalErr,
			errType,
			code,
			message,
			severity,
		)
	} else {
		err = bperrors.New(
			errType,
			code,
			message,
			severity,
		)
	}

	// Add context
	err.WithNodeInfo(ctx.GetNodeID(), "").WithBlueprintInfo(ctx.GetBlueprintID(), ctx.GetExecutionID())

	// Record the error
	ctx.errorManager.RecordError(ctx.GetExecutionID(), err)

	// Log the error
	ctx.Logger().Error(message, map[string]interface{}{
		"nodeId":      ctx.GetNodeID(),
		"errorType":   string(errType),
		"errorCode":   string(code),
		"originalErr": originalErr,
	})

	return err
}

// AttemptRecovery tries to recover from an error
func (ctx *ErrorAwareContext) AttemptRecovery(err *bperrors.BlueprintError) (bool, map[string]interface{}) {
	return ctx.recoveryManager.RecoverFromError(ctx.GetExecutionID(), err)
}

// GetErrorSummary gets a summary of errors for this node
func (ctx *ErrorAwareContext) GetErrorSummary() map[string]interface{} {
	nodeErrors := ctx.errorManager.GetNodeErrors(ctx.GetExecutionID(), ctx.GetNodeID())

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
func (ctx *ErrorAwareContext) GetDefaultValue(pinType *types.PinType) (types.Value, bool) {
	value, err := ctx.recoveryManager.GetDefaultValue(pinType)
	if err != nil {
		return types.Value{}, false
	}
	return value, true
}

// Ensure ErrorAwareContext implements core.ErrorAwareContext
var _ core.ErrorAwareContext = (*ErrorAwareContext)(nil)

// Unwrap returns the underlying ExecutionContext that this context decorates.
func (ctx *ErrorAwareContext) Unwrap() node.ExecutionContext {
	return ctx.ExecutionContext
}
