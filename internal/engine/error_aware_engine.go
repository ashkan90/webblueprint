package engine

import (
	"errors"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/common"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// ErrorAwareEngine adds error handling capabilities to the execution engine
// It works with both standard and actor-based execution by wrapping the execution hooks
type ErrorAwareEngine struct {
	BaseEngine      *ExecutionEngine
	ErrorManager    *bperrors.ErrorManager
	RecoveryManager *bperrors.RecoveryManager
	InfoStore       *bperrors.ExecutionInfoStore
}

// NewErrorAwareEngine creates a new error-aware engine wrapper
func NewErrorAwareEngine(baseEngine *ExecutionEngine) *ErrorAwareEngine {
	errorManager := bperrors.NewErrorManager()
	recoveryManager := bperrors.NewRecoveryManager(errorManager)
	infoStore := bperrors.NewExecutionInfoStore()

	return &ErrorAwareEngine{
		BaseEngine:      baseEngine,
		ErrorManager:    errorManager,
		RecoveryManager: recoveryManager,
		InfoStore:       infoStore,
	}
}

// Execute runs a blueprint with enhanced error handling
func (e *ErrorAwareEngine) Execute(bp *blueprint.Blueprint, executionID string, initialData map[string]types.Value) (map[string]interface{}, error) {
	// Validate blueprint before execution
	validator := bperrors.NewBlueprintValidator(e.ErrorManager)
	validationResult := validator.ValidateBlueprint(bp)

	// Create extended execution info
	var baseResult common.ExecutionResult
	extendedInfo := bperrors.NewExtendedExecutionInfo(&baseResult)
	extendedInfo.AddValidationResults(&validationResult)

	// If validation failed, return without executing
	if !validationResult.Valid {
		var validationError error
		if len(validationResult.Errors) > 0 {
			validationError = validationResult.Errors[0]
			baseResult = common.ExecutionResult{
				Success:     false,
				ExecutionID: executionID,
				Error:       validationError,
			}
		} else {
			validationError = bperrors.New(
				bperrors.ErrorTypeValidation,
				bperrors.ErrInvalidBlueprintStructure,
				"Blueprint validation failed",
				bperrors.SeverityHigh,
			)
			baseResult = common.ExecutionResult{
				Success:     false,
				ExecutionID: executionID,
				Error:       validationError,
			}
		}

		// Store extended info
		e.InfoStore.StoreExecutionInfo(executionID, extendedInfo)

		// Return the result as a map for JSON serialization
		return extendedInfo.ToMap(), validationError
	}

	// Create error handling hooks to wrap the base hooks
	baseHooks := &node.ExecutionHooks{}
	_ = e.createErrorAwareHooks(executionID, baseHooks, extendedInfo)

	// Clear previous errors for this execution
	e.ErrorManager.ClearErrors(executionID)
	e.RecoveryManager.ClearRecoveryAttempts(executionID)

	// Execute the blueprint using the base engine
	result, err := e.BaseEngine.Execute(bp, executionID, initialData)

	// Generate error analysis
	if err != nil {
		// Wrap the error if it's not already a BlueprintError
		var blueprintError *bperrors.BlueprintError
		if !errors.As(err, &blueprintError) {
			err = bperrors.Wrap(
				err,
				bperrors.ErrorTypeExecution,
				bperrors.ErrNodeExecutionFailed,
				"Error during blueprint execution",
				bperrors.SeverityHigh,
			)
			err.(*bperrors.BlueprintError).WithBlueprintInfo(bp.ID, executionID)
		}

		// Add error analysis
		analysis := e.ErrorManager.AnalyzeErrors(executionID)
		extendedInfo.AddErrorAnalysis(analysis)

		// Store extended info
		e.InfoStore.StoreExecutionInfo(executionID, extendedInfo)

		// Return the result as a map for JSON serialization
		return extendedInfo.ToMap(), err
	}

	// Convert execution engine result to common result
	baseResult = common.ExecutionResult{
		Success:     result.Success,
		ExecutionID: result.ExecutionID,
		StartTime:   result.StartTime,
		EndTime:     result.EndTime,
		Error:       result.Error,
		NodeResults: result.NodeResults,
	}

	// If successful, still add error analysis if errors occurred
	if errors := e.ErrorManager.GetErrors(executionID); len(errors) > 0 {
		analysis := e.ErrorManager.AnalyzeErrors(executionID)
		extendedInfo.AddErrorAnalysis(analysis)
		extendedInfo.PartialSuccess = true
	}

	// Store extended info
	e.InfoStore.StoreExecutionInfo(executionID, extendedInfo)

	// Return the result as a map for JSON serialization
	return extendedInfo.ToMap(), nil
}

// GetErrorManager returns the error manager
func (e *ErrorAwareEngine) GetErrorManager() *bperrors.ErrorManager {
	return e.ErrorManager
}

// GetRecoveryManager returns the recovery manager
func (e *ErrorAwareEngine) GetRecoveryManager() *bperrors.RecoveryManager {
	return e.RecoveryManager
}

// GetExecutionInfo gets extended info for an execution
func (e *ErrorAwareEngine) GetExecutionInfo(executionID string) (*bperrors.ExtendedExecutionInfo, bool) {
	return e.InfoStore.GetExecutionInfo(executionID)
}

// WrapExecutionContext wraps an execution context with error handling capabilities
func (e *ErrorAwareEngine) WrapExecutionContext(ctx node.ExecutionContext) bperrors.ErrorAwareContext {
	return bperrors.NewErrorContextWrapper(ctx, e.ErrorManager, e.RecoveryManager)
}

// createErrorAwareHooks creates execution hooks with error handling
func (e *ErrorAwareEngine) createErrorAwareHooks(
	executionID string,
	baseHooks *node.ExecutionHooks,
	info *bperrors.ExtendedExecutionInfo,
) *node.ExecutionHooks {
	errorHooks := &node.ExecutionHooks{
		OnNodeStart: func(nodeID, nodeType string) {
			// Call base hook if provided
			if baseHooks.OnNodeStart != nil {
				baseHooks.OnNodeStart(nodeID, nodeType)
			}
		},

		OnNodeComplete: func(nodeID, nodeType string) {
			// Record successful node
			info.AddNodeStatus(nodeID, true)

			// Call base hook if provided
			if baseHooks.OnNodeComplete != nil {
				baseHooks.OnNodeComplete(nodeID, nodeType)
			}
		},

		OnNodeError: func(nodeID string, err error) {
			// Record failed node
			info.AddNodeStatus(nodeID, false)

			// Convert to BlueprintError if needed
			var bpErr *bperrors.BlueprintError
			if existingErr, ok := err.(*bperrors.BlueprintError); ok {
				bpErr = existingErr
			} else {
				bpErr = bperrors.Wrap(
					err,
					bperrors.ErrorTypeExecution,
					bperrors.ErrNodeExecutionFailed,
					err.Error(),
					bperrors.SeverityHigh,
				)
				bpErr.WithNodeInfo(nodeID, "")
				bpErr.WithBlueprintInfo("", executionID)
			}

			// Record the error
			e.ErrorManager.RecordError(executionID, bpErr)

			// Call base hook if provided
			if baseHooks.OnNodeError != nil {
				baseHooks.OnNodeError(nodeID, bpErr)
			}
		},

		OnPinValue: func(nodeID, pinName string, value interface{}) {
			// Call base hook if provided
			if baseHooks.OnPinValue != nil {
				baseHooks.OnPinValue(nodeID, pinName, value)
			}
		},

		OnLog: func(nodeID, message string) {
			// Call base hook if provided
			if baseHooks.OnLog != nil {
				baseHooks.OnLog(nodeID, message)
			}
		},
	}

	return errorHooks
}
