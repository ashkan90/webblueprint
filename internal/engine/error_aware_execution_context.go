package engine

import (
	errors "webblueprint/internal/bperrors"
	"webblueprint/internal/types"
)

// ErrorAwareExecutionContext extends DefaultExecutionContext with error handling capabilities
type ErrorAwareExecutionContext struct {
	*DefaultExecutionContext
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
}

// NewErrorAwareExecutionContext creates a new error-aware execution context
func NewErrorAwareExecutionContext(
	ctx *DefaultExecutionContext,
	errorManager *errors.ErrorManager,
	recoveryManager *errors.RecoveryManager,
) *ErrorAwareExecutionContext {
	return &ErrorAwareExecutionContext{
		DefaultExecutionContext: ctx,
		errorManager:            errorManager,
		recoveryManager:         recoveryManager,
	}
}

// GetInputValue retrieves an input value with error handling
func (ctx *ErrorAwareExecutionContext) GetInputValue(pinID string) (types.Value, bool) {
	value, exists := ctx.DefaultExecutionContext.GetInputValue(pinID)

	// If the value doesn't exist, try to recover with a default value
	if !exists {
		// Create a BlueprintError
		err := errors.New(
			errors.ErrorTypeExecution,
			errors.ErrMissingRequiredInput,
			"Required input value missing",
			errors.SeverityMedium,
		).WithNodeInfo(ctx.GetNodeID(), pinID).WithBlueprintInfo(ctx.GetBlueprintID(), ctx.GetExecutionID())

		// Record the error
		ctx.errorManager.RecordError(ctx.GetExecutionID(), err)

		// Attempt recovery using default value strategy
		if success, _ := ctx.recoveryManager.RecoverFromError(ctx.GetExecutionID(), err); success {
			// Find the pin type for this pin
			// In a real implementation, we would get this from node registry or blueprint definition
			pinType := types.PinTypes.Any // Fallback type

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
func (ctx *ErrorAwareExecutionContext) ReportError(
	errType errors.ErrorType,
	code errors.BlueprintErrorCode,
	message string,
	originalErr error,
) *errors.BlueprintError {
	// Create a BlueprintError
	severity := errors.SeverityMedium // Default severity

	// Set severity based on error type
	switch errType {
	case errors.ErrorTypeExecution:
		severity = errors.SeverityHigh
	case errors.ErrorTypeConnection:
		severity = errors.SeverityMedium
	case errors.ErrorTypeValidation:
		severity = errors.SeverityMedium
	case errors.ErrorTypePermission:
		severity = errors.SeverityHigh
	case errors.ErrorTypeDatabase:
		severity = errors.SeverityHigh
	default:
		severity = errors.SeverityMedium
	}

	// Create the error
	var err *errors.BlueprintError
	if originalErr != nil {
		err = errors.Wrap(
			originalErr,
			errType,
			code,
			message,
			severity,
		)
	} else {
		err = errors.New(
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
func (ctx *ErrorAwareExecutionContext) AttemptRecovery(err *errors.BlueprintError) (bool, map[string]interface{}) {
	return ctx.recoveryManager.RecoverFromError(ctx.GetExecutionID(), err)
}

// GetErrorSummary gets a summary of errors for this node
func (ctx *ErrorAwareExecutionContext) GetErrorSummary() map[string]interface{} {
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
