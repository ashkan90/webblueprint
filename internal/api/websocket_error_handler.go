package api

import (
	errors "webblueprint/internal/bperrors"
)

// ErrorNotification represents a WebSocket error notification
type ErrorNotification struct {
	Type        string                 `json:"type"`
	Error       *errors.BlueprintError `json:"error"`
	ExecutionID string                 `json:"executionId"`
}

// ErrorAnalysisNotification represents a WebSocket error analysis notification
type ErrorAnalysisNotification struct {
	Type        string                 `json:"type"`
	Analysis    map[string]interface{} `json:"analysis"`
	ExecutionID string                 `json:"executionId"`
}

// RecoveryNotification represents a WebSocket recovery attempt notification
type RecoveryNotification struct {
	Type        string                 `json:"type"`
	Successful  bool                   `json:"successful"`
	Strategy    string                 `json:"strategy"`
	NodeID      string                 `json:"nodeId"`
	ErrorCode   string                 `json:"errorCode"`
	Details     map[string]interface{} `json:"details,omitempty"`
	ExecutionID string                 `json:"executionId"`
}

// Logger interface for the WebSocketManager
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
}

// sendErrorNotification sends an error notification to the client
func (h *WebSocketManager) sendErrorNotification(executionID string, err *errors.BlueprintError) {
	notification := ErrorNotification{
		Type:        "error",
		Error:       err,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeNodeError, notification)
}

// sendErrorAnalysisNotification sends an error analysis notification to the client
func (h *WebSocketManager) sendErrorAnalysisNotification(executionID string, analysis map[string]interface{}) {
	notification := ErrorAnalysisNotification{
		Type:        "error_analysis",
		Analysis:    analysis,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeDebugData, notification)
}

// sendRecoveryNotification sends a recovery attempt notification to the client
func (h *WebSocketManager) sendRecoveryNotification(executionID, nodeID, errorCode, strategy string, successful bool, details map[string]interface{}) {
	notification := RecoveryNotification{
		Type:        "recovery_attempt",
		Successful:  successful,
		Strategy:    strategy,
		NodeID:      nodeID,
		ErrorCode:   errorCode,
		Details:     details,
		ExecutionID: executionID,
	}

	h.BroadcastMessage(MsgTypeExecStatus, notification)
}

// RegisterErrorHandlers registers error handlers with the error manager
func (h *WebSocketManager) RegisterErrorHandlers(errorManager *errors.ErrorManager, logger Logger) {
	// Set logger for error handlers
	h.Logger = logger

	// Register handler for all error types
	for _, errType := range []errors.ErrorType{
		errors.ErrorTypeExecution,
		errors.ErrorTypeConnection,
		errors.ErrorTypeValidation,
		errors.ErrorTypePermission,
		errors.ErrorTypeDatabase,
		errors.ErrorTypeNetwork,
		errors.ErrorTypePlugin,
		errors.ErrorTypeSystem,
		errors.ErrorTypeUnknown,
	} {
		// Use a separate variable in the closure to avoid issues
		errorType := errType

		errorManager.RegisterErrorHandler(errorType, func(err *errors.BlueprintError) error {
			// Send error notification to clients
			h.sendErrorNotification(err.ExecutionID, err)
			return nil
		})
	}
}
