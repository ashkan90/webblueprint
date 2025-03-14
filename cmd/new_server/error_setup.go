package main

import (
	"log"
	"webblueprint/internal/api"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/engine"
	"webblueprint/internal/types"
	"webblueprint/pkg/blueprint"
	loggerPkg "webblueprint/pkg/logger"
)

// SetupErrorHandling initializes the error handling and recovery system
func SetupErrorHandling(baseEngine *engine.ExecutionEngine, wsManager *api.WebSocketManager) *engine.ErrorAwareExecutionEngine {
	// Create the error-aware execution engine wrapper
	errorAwareEngine := engine.NewErrorAwareExecutionEngine(baseEngine)

	// Create a WebSocket error handler and register it
	errorHandler := api.NewErrorWebSocketHandler(wsManager)
	errorHandler.RegisterWithErrorManager(errorAwareEngine.GetErrorManager())

	// Register default recovery handlers
	registerDefaultRecoveryHandlers(errorAwareEngine.GetRecoveryManager())

	// Register error listeners for logging
	registerErrorLoggers(errorAwareEngine.GetErrorManager(), baseEngine.GetLogger())

	return errorAwareEngine
}

// registerDefaultRecoveryHandlers adds default recovery handlers for common scenarios
func registerDefaultRecoveryHandlers(recoveryManager *bperrors.RecoveryManager) {
	// Register default value providers for different types
	recoveryManager.RegisterDefaultValueProvider(types.PinTypes.String.Name, func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.String, "")
	})

	recoveryManager.RegisterDefaultValueProvider(types.PinTypes.Number.Name, func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Number, 0)
	})

	recoveryManager.RegisterDefaultValueProvider(types.PinTypes.Boolean.Name, func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Boolean, false)
	})

	recoveryManager.RegisterDefaultValueProvider(types.PinTypes.Array.Name, func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Array, []interface{}{})
	})

	recoveryManager.RegisterDefaultValueProvider(types.PinTypes.Object.Name, func(pt *types.PinType) types.Value {
		return types.NewValue(types.PinTypes.Object, map[string]interface{}{})
	})

	// For custom types, you could add specific handlers here
}

// registerErrorLoggers registers handlers that log errors
func registerErrorLoggers(errorManager *bperrors.ErrorManager, logger loggerPkg.Logger) {
	// Log critical and high severity errors
	errorManager.RegisterErrorHandler(bperrors.ErrorTypeExecution, func(err *bperrors.BlueprintError) error {
		if err.Severity == bperrors.SeverityCritical || err.Severity == bperrors.SeverityHigh {
			logger.Error(err.Error(), map[string]interface{}{
				"errorType":   string(err.Type),
				"errorCode":   string(err.Code),
				"nodeId":      err.NodeID,
				"blueprintId": err.BlueprintID,
				"details":     err.Details,
			})
		} else {
			logger.Warn(err.Error(), map[string]interface{}{
				"errorType":   string(err.Type),
				"errorCode":   string(err.Code),
				"nodeId":      err.NodeID,
				"blueprintId": err.BlueprintID,
			})
		}
		return nil
	})

	// Similar handlers for other error types...
	errorManager.RegisterErrorHandler(bperrors.ErrorTypeDatabase, func(err *bperrors.BlueprintError) error {
		logger.Error("Database error: "+err.Error(), map[string]interface{}{
			"errorCode": string(err.Code),
			"details":   err.Details,
		})
		return nil
	})
}

// ValidateBlueprintWithErrorHandling runs enhanced validation with better error reporting
func ValidateBlueprintWithErrorHandling(bp *blueprint.Blueprint, validator *bperrors.BlueprintValidator) (bool, []*bperrors.BlueprintError) {
	result := validator.ValidateBlueprint(bp)

	// Log validation results
	if !result.Valid {
		log.Printf("Blueprint validation failed for blueprint %s: %d errors", bp.ID, len(result.Errors))
		for _, err := range result.Errors {
			log.Printf("- Error [%s-%s]: %s",
				err.(*bperrors.BlueprintError).Type,
				err.(*bperrors.BlueprintError).Code,
				err.(*bperrors.BlueprintError).Message,
			)
		}
	}

	if len(result.Warnings) > 0 {
		log.Printf("Blueprint validation warnings for blueprint %s: %d warnings", bp.ID, len(result.Warnings))
		for _, warn := range result.Warnings {
			log.Printf("- Warning [%s-%s]: %s",
				warn.(*bperrors.BlueprintError).Type,
				warn.(*bperrors.BlueprintError).Code,
				warn.(*bperrors.BlueprintError).Message,
			)
		}
	}

	var resultErrors = make([]*bperrors.BlueprintError, 0, len(result.Errors))
	for _, err := range result.Errors {
		resultErrors = append(resultErrors, err.(*bperrors.BlueprintError))
	}

	return result.Valid, resultErrors
}

// GetEngineWithErrorHandling creates a new execution engine with error handling
func GetEngineWithErrorHandling(wsManager *api.WebSocketManager, logger *api.WebSocketLogger) *engine.ErrorAwareExecutionEngine {
	// Create base engine
	debugManager := engine.NewDebugManager()
	baseEngine := engine.NewExecutionEngine(logger, debugManager)

	// Wrap with error handling
	return SetupErrorHandling(baseEngine, wsManager)
}
