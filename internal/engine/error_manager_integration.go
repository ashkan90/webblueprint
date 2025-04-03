package engine

import (
	"reflect"
	errors "webblueprint/internal/bperrors"
	"webblueprint/internal/common"
	"webblueprint/internal/engineext"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
)

// SetInputValue uses reflection to set an input value
// This is a temporary workaround until proper methods are added
func SetInputValue(ctx interface{}, pinID string, value types.Value) {
	// Try to convert to DefaultExecutionContext
	if defaultCtx, ok := ctx.(*engineext.DefaultExecutionContext); ok {
		val := reflect.ValueOf(defaultCtx).Elem()
		inputsField := val.FieldByName("inputs")

		if inputsField.IsValid() && inputsField.Kind() == reflect.Map {
			// Create a reflect.Value for the key
			keyValue := reflect.ValueOf(pinID)
			// Create a reflect.Value for the value
			valueValue := reflect.ValueOf(value)
			// Set the map entry
			inputsField.SetMapIndex(keyValue, valueValue)
		}
	}
}

// SetActivateFlow uses reflection to set the activateFlow field
// This is a temporary workaround until proper access methods are added
func SetActivateFlow(ctx *engineext.DefaultExecutionContext, fn interface{}) {
	// Skip if either is nil
	if ctx == nil || fn == nil {
		return
	}

	// Get the value of the context
	val := reflect.ValueOf(ctx).Elem()

	// Find the activateFlow field
	field := val.FieldByName("activateFlow")
	if !field.IsValid() || !field.CanSet() {
		return // Field not found or can't be set
	}

	// Set the field value
	fnVal := reflect.ValueOf(fn)
	if fnVal.Type().AssignableTo(field.Type()) {
		field.Set(fnVal)
	}
}

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

/*
// executeNode overrides the base executeNode method to add error handling
func (e *ErrorAwareExecutionEngine) executeNode(nodeID string, bp *blueprint.Blueprint, blueprintID, executionID string, variables map[string]types.Value, hooks *node.ExecutionHooks) error {
	// Get the basic execution context from base implementation
	baseCtx := engineext.NewExecutionContext(
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
	ctx := engineext.NewErrorAwareExecutionContext(
		baseCtx,
		e.errorManager,
		e.recoveryManager,
	)

	// Set activateFlow function (needs to reference the error-aware context)
	activateFlowFn := func(ctx *engineext.DefaultExecutionContext, nodeID, pinID string) error {
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

	// Use reflection to set activateFlow
	// This is a temporary workaround for the unexported field
	SetActivateFlow(baseCtx, activateFlowFn)

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
				// Define a helper method to set the input value
				// This is a temporary workaround until proper methods are added
				SetInputValue(ctx, targetPinID, value)
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
	// Check if the underlying context implements ExtendedExecutionContext
	if underlyingCtx := ctx.ExecutionContext; underlyingCtx != nil { // Assuming ErrorAwareContext has ExecutionContext field
		if extCtx, ok := underlyingCtx.(node.ExtendedExecutionContext); ok {
			outputs := extCtx.GetAllOutputs()
			for pinID, outValue := range outputs {
			e.debugManager.StoreNodeOutputValue(executionID, nodeID, pinID, outValue.RawValue)
		}
	}

	// Follow execution flows
	activatedFlows := ctx.GetActivatedOutputFlows() // Assuming ErrorAwareContext implements this via embedding
	for _, outputPin := range activatedFlows {
		// Check if the underlying context is DefaultExecutionContext
		if underlyingCtx := ctx.ExecutionContext; underlyingCtx != nil { // Assuming ErrorAwareContext has ExecutionContext field
			if defaultCtx, ok := underlyingCtx.(*engineext.DefaultExecutionContext); ok {
				if err := activateFlowFn(defaultCtx, nodeID, outputPin); err != nil {
					// Try to recover from flow activation error
					// Add missing closing brace for the inner if err != nil
					if bpErr, ok := err.(*errors.BlueprintError); ok {
						if success, _ := e.recoveryManager.RecoverFromError(executionID, bpErr); success {
							// If recovery was successful, continue to the next flow
						continue
					}
				} // End if bpErr
				return err
			} // End if err != nil <-- Add this closing brace
		} // End if defaultCtx
	} // End if underlyingCtx

	// Notify node completion
	if hooks != nil && hooks.OnNodeComplete != nil {
		hooks.OnNodeComplete(nodeID, nodeConfig.Type)
	}

	return nil
}
*/ // End of commented out executeNode override

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
