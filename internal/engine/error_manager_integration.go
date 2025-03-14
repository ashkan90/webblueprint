package engine

import (
	errors "webblueprint/internal/bperrors"
	"webblueprint/internal/common"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// ErrorAwareExecutionEngine adds error handling capabilities to the execution engine
type ErrorAwareExecutionEngine struct {
	*ExecutionEngine
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
	validator       *errors.BlueprintValidator
}

// NewErrorAwareExecutionEngine creates a new error-aware execution engine
func NewErrorAwareExecutionEngine(engine *ExecutionEngine) *ErrorAwareExecutionEngine {
	// Create error handling components
	errorManager := errors.NewErrorManager()
	recoveryManager := errors.NewRecoveryManager(errorManager)
	validator := errors.NewBlueprintValidator(errorManager)

	// Register error handlers
	errorManager.RegisterErrorHandler(errors.ErrorTypeExecution, func(err *errors.BlueprintError) error {
		// Log execution errors to the debug manager
		engine.debugManager.StoreNodeDebugData(err.ExecutionID, err.NodeID, map[string]interface{}{
			"error": map[string]interface{}{
				"type":    err.Type,
				"code":    err.Code,
				"message": err.Message,
				"details": err.Details,
			},
		})
		return nil
	})

	return &ErrorAwareExecutionEngine{
		ExecutionEngine: engine,
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
		validator:       validator,
	}
}

// GetErrorManager returns the error manager
func (e *ErrorAwareExecutionEngine) GetErrorManager() *errors.ErrorManager {
	return e.errorManager
}

// GetRecoveryManager returns the recovery manager
func (e *ErrorAwareExecutionEngine) GetRecoveryManager() *errors.RecoveryManager {
	return e.recoveryManager
}

// Execute overrides the base Execute method to add error handling
func (e *ErrorAwareExecutionEngine) Execute(bp *blueprint.Blueprint, executionID string, initialData map[string]types.Value) (common.ExecutionResult, error) {
	// Validate blueprint before execution
	validationResult := e.validator.ValidateBlueprint(bp)
	if !validationResult.Valid {
		// Create a comprehensive error message
		errorMsg := "Blueprint validation failed: "
		if len(validationResult.Errors) > 0 {
			errorMsg += validationResult.Errors[0].Error()
		}

		// Create execution result with error
		result := common.ExecutionResult{
			Success:     false,
			ExecutionID: executionID,
			Error:       validationResult.Errors[0],
		}

		return result, validationResult.Errors[0]
	}

	// Clear previous errors for this execution
	e.errorManager.ClearErrors(executionID)
	e.recoveryManager.ClearRecoveryAttempts(executionID)

	// Execute the blueprint using the base engine
	result, err := e.ExecutionEngine.Execute(bp, executionID, initialData)

	// If there was an error, check if we can recover
	if err != nil {
		// Wrap the error if it's not already a BlueprintError
		var bpErr *errors.BlueprintError
		var ok bool

		if bpErr, ok = err.(*errors.BlueprintError); !ok {
			bpErr = errors.Wrap(
				err,
				errors.ErrorTypeExecution,
				errors.ErrNodeExecutionFailed,
				"Error during blueprint execution",
				errors.SeverityHigh,
			)
			bpErr.WithBlueprintInfo(bp.ID, executionID)
		}

		// Try to recover
		if success, details := e.recoveryManager.RecoverFromError(executionID, bpErr); success {
			// Log recovery attempt
			e.debugManager.StoreExecutionDebugData(executionID, map[string]interface{}{
				"recovery": map[string]interface{}{
					"successful": true,
					"details":    details,
					"error":      bpErr,
				},
			})

			// If recovery was successful, update result
			result.Success = true
			result.Error = nil
			return result, nil
		}

		// If recovery failed, update result with error analysis
		errorAnalysis := e.errorManager.AnalyzeErrors(executionID)
		result.Error = bpErr
		result.Success = false
		result.ErrorAnalysis = errorAnalysis

		return result, bpErr
	}

	// Add error analysis to successful results too
	if len(e.errorManager.GetErrors(executionID)) > 0 {
		result.ErrorAnalysis = e.errorManager.AnalyzeErrors(executionID)
		result.PartialSuccess = true
	}

	return result, nil
}

// executeNode overrides the base executeNode method to add error handling
func (e *ErrorAwareExecutionEngine) executeNode(nodeID string, bp *blueprint.Blueprint, blueprintID, executionID string, variables map[string]types.Value, hooks *node.ExecutionHooks) error {
	// Get the basic execution context from base implementation
	baseCtx := NewExecutionContext(
		nodeID,
		bp.FindNode(nodeID).Type,
		blueprintID,
		executionID,
		make(map[string]types.Value),
		variables,
		e.logger,
		hooks,
		nil, // We'll set this after creating the error-aware context
	)

	// Create error-aware context
	ctx := NewErrorAwareExecutionContext(
		baseCtx,
		e.errorManager,
		e.recoveryManager,
	)

	// Set activateFlow function (needs to reference the error-aware context)
	activateFlowFn := func(ctx *DefaultExecutionContext, nodeID, pinID string) error {
		// Find connections from this output pin
		outputConnections := bp.GetNodeOutputConnections(nodeID)
		for _, conn := range outputConnections {
			if conn.ConnectionType == "execution" && conn.SourcePinID == pinID {
				targetNodeID := conn.TargetNodeID

				// Execute the target node with error handling
				if err := e.executeNode(targetNodeID, bp, blueprintID, executionID, variables, hooks); err != nil {
					// Try to handle the error
					if bpErr, ok := err.(*errors.BlueprintError); ok {
						if success, _ := e.recoveryManager.RecoverFromError(executionID, bpErr); success {
							// If recovery was successful, continue to the next connection
							continue
						}
					}
					return err
				}
			}
		}
		return nil
	}

	baseCtx.activateFlow = activateFlowFn

	// Find the node in the blueprint
	nodeConfig := bp.FindNode(nodeID)
	if nodeConfig == nil {
		err := errors.New(
			errors.ErrorTypeExecution,
			errors.ErrNodeNotFound,
			"Node not found in blueprint",
			errors.SeverityHigh,
		).WithNodeInfo(nodeID, "").WithBlueprintInfo(blueprintID, executionID)

		e.errorManager.RecordError(executionID, err)
		return err
	}

	// Get the node factory
	factory, exists := e.nodeRegistry[nodeConfig.Type]
	if !exists {
		err := errors.New(
			errors.ErrorTypeExecution,
			errors.ErrNodeTypeNotRegistered,
			"Node type not registered",
			errors.SeverityHigh,
		).WithNodeInfo(nodeID, "").WithBlueprintInfo(blueprintID, executionID)

		e.errorManager.RecordError(executionID, err)
		return err
	}

	// Create the node instance
	nodeInstance := factory()

	// Prepare input values with error handling
	inputConnections := bp.GetNodeInputConnections(nodeID)
	for _, conn := range inputConnections {
		if conn.ConnectionType == "data" {
			sourceNodeID := conn.SourceNodeID
			sourcePinID := conn.SourcePinID
			targetPinID := conn.TargetPinID

			// Try to get value with error recovery
			if outputs, ok := e.debugManager.GetNodeOutputValue(executionID, sourceNodeID, sourcePinID); ok {
				value := types.NewValue(types.PinTypes.Any, outputs)
				ctx.DefaultExecutionContext.inputs[targetPinID] = value
			} else {
				// Missing input value - ctx.GetInputValue will handle recovery
			}
		}
	}

	// Notify node start
	if hooks != nil && hooks.OnNodeStart != nil {
		hooks.OnNodeStart(nodeID, nodeConfig.Type)
	}

	// Execute the node with error handling
	err := nodeInstance.Execute(ctx)
	if err != nil {
		// Create a BlueprintError from the error
		bpErr := errors.Wrap(
			err,
			errors.ErrorTypeExecution,
			errors.ErrNodeExecutionFailed,
			"Node execution failed",
			errors.SeverityHigh,
		).WithNodeInfo(nodeID, "").WithBlueprintInfo(blueprintID, executionID)

		// Record the error
		e.errorManager.RecordError(executionID, bpErr)

		// Call error hook
		if hooks != nil && hooks.OnNodeError != nil {
			hooks.OnNodeError(nodeID, bpErr)
		}

		// Try to recover
		if success, _ := e.recoveryManager.RecoverFromError(executionID, bpErr); success {
			// If recovery was successful, continue execution
			// Store debug data
			e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

			// Notify node completion
			if hooks != nil && hooks.OnNodeComplete != nil {
				hooks.OnNodeComplete(nodeID, nodeConfig.Type)
			}

			return nil
		}

		return bpErr
	}

	// Store debug data
	e.debugManager.StoreNodeDebugData(executionID, nodeID, ctx.GetDebugData())

	// Store outputs in debug manager
	for pinID, outValue := range ctx.outputs {
		e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinID, outValue.RawValue)
	}

	// Follow execution flows
	activatedFlows := ctx.GetActivatedOutputFlows()
	for _, outputPin := range activatedFlows {
		if err := activateFlowFn(ctx.DefaultExecutionContext, nodeID, outputPin); err != nil {
			// Try to recover from flow activation error
			if bpErr, ok := err.(*errors.BlueprintError); ok {
				if success, _ := e.recoveryManager.RecoverFromError(executionID, bpErr); success {
					// If recovery was successful, continue to the next flow
					continue
				}
			}
			return err
		}
	}

	// Notify node completion
	if hooks != nil && hooks.OnNodeComplete != nil {
		hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	return nil
}

// UpdateExecutionResult extends ExecutionResult with error information
func (e *ErrorAwareExecutionEngine) UpdateExecutionResult(result *common.ExecutionResult, executionID string) {
	// Add error analysis to the result
	errs := e.errorManager.GetErrors(executionID)
	if len(errs) > 0 {
		result.ErrorAnalysis = e.errorManager.AnalyzeErrors(executionID)

		// Check if any errors were critical
		hasCritical := false
		for _, err := range errs {
			if err.Severity == errors.SeverityCritical || err.Severity == errors.SeverityHigh {
				hasCritical = true
				break
			}
		}

		if hasCritical {
			result.Success = false
			result.PartialSuccess = true
		}
	}
}
