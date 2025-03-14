package api

import (
	"time"
	errors "webblueprint/internal/bperrors"
)

// ErrorNotificationHandler processes error notifications
type ErrorNotificationHandler struct {
	wsHandler       *WebSocketManager
	wsLogger        *WebSocketLogger
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
}

// NewErrorNotificationHandler creates a new error notification handler
func NewErrorNotificationHandler(wsHandler *WebSocketManager, wsLogger *WebSocketLogger, errorManager *errors.ErrorManager, recoveryManager *errors.RecoveryManager) *ErrorNotificationHandler {
	return &ErrorNotificationHandler{
		wsHandler:       wsHandler,
		wsLogger:        wsLogger,
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
	}
}

// HandleExecutionError handles execution errors
func (h *ErrorNotificationHandler) HandleExecutionError(executionID string, err error) {
	// Ensure we have a BlueprintError
	var bpErr *errors.BlueprintError
	var ok bool
	if bpErr, ok = err.(*errors.BlueprintError); !ok {
		// Wrap the error
		bpErr = errors.Wrap(
			err,
			errors.ErrorTypeExecution,
			errors.ErrUnknown,
			err.Error(),
			errors.SeverityHigh,
		)
	}

	// Set execution ID if not already set
	if bpErr.ExecutionID == "" {
		bpErr.ExecutionID = executionID
	}

	// Record the error
	h.errorManager.RecordError(executionID, bpErr)

	// Send error notification
	h.sendErrorNotification(executionID, bpErr)

	// If it's recoverable, try automatic recovery
	if bpErr.Recoverable && len(bpErr.RecoveryOptions) > 0 {
		// Choose the first recovery strategy
		strategy := bpErr.RecoveryOptions[0]

		// Attempt recovery
		success, details := h.recoveryManager.RecoverFromError(executionID, bpErr)

		// Send recovery notification
		h.sendRecoveryNotification(
			executionID,
			bpErr.NodeID,
			string(bpErr.Code),
			string(strategy),
			success,
			details,
		)
	}

	// Send error analysis after a short delay (to collect any related errors)
	time.AfterFunc(500*time.Millisecond, func() {
		analysis := h.errorManager.AnalyzeErrors(executionID)
		h.sendErrorAnalysisNotification(executionID, analysis)
	})
}

// sendErrorNotification sends an error notification to clients
func (h *ErrorNotificationHandler) sendErrorNotification(executionID string, err *errors.BlueprintError) {
	notification := ErrorNotification{
		Type:        "error",
		Error:       err,
		ExecutionID: executionID,
	}

	h.wsHandler.BroadcastMessage(MsgTypeNodeError, notification)
}

// sendErrorAnalysisNotification sends an error analysis notification to clients
func (h *ErrorNotificationHandler) sendErrorAnalysisNotification(executionID string, analysis map[string]interface{}) {
	notification := ErrorAnalysisNotification{
		Type:        "error_analysis",
		Analysis:    analysis,
		ExecutionID: executionID,
	}

	h.wsHandler.BroadcastMessage(MsgTypeDebugData, notification)
}

// sendRecoveryNotification sends a recovery notification to clients
func (h *ErrorNotificationHandler) sendRecoveryNotification(executionID, nodeID, errorCode, strategy string, successful bool, details map[string]interface{}) {
	notification := RecoveryNotification{
		Type:        "recovery_attempt",
		Successful:  successful,
		Strategy:    strategy,
		NodeID:      nodeID,
		ErrorCode:   errorCode,
		Details:     details,
		ExecutionID: executionID,
	}

	h.wsHandler.BroadcastMessage(MsgTypeExecStatus, notification)
}

// TestErrorScenario triggers a test error scenario
func (h *ErrorNotificationHandler) TestErrorScenario(scenarioType, executionID string) {
	// Create a test error generator
	generator := errors.NewTestErrorGenerator()

	// Generate test scenario
	analysis, err := generator.SimulateErrorScenario(scenarioType, executionID)
	if err != nil {
		// Send the generated error
		if bpErr, ok := err.(*errors.BlueprintError); ok {
			h.sendErrorNotification(executionID, bpErr)
		}
	}

	// Send the analysis
	if analysis != nil {
		h.sendErrorAnalysisNotification(executionID, analysis)
	}
}
