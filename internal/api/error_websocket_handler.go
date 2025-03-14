package api

import (
	"webblueprint/internal/bperrors"
)

// ErrorWebSocketHandler handles error-related WebSocket messages
type ErrorWebSocketHandler struct {
	WSHandler *WebSocketManager
}

// NewErrorWebSocketHandler creates a new error WebSocket handler
func NewErrorWebSocketHandler(wsHandler *WebSocketManager) *ErrorWebSocketHandler {
	return &ErrorWebSocketHandler{
		WSHandler: wsHandler,
	}
}

// SendErrorNotification sends an error notification to clients
func (h *ErrorWebSocketHandler) SendErrorNotification(executionID string, err *bperrors.BlueprintError) {
	notification := ErrorNotification{
		Type:        "error",
		Error:       err,
		ExecutionID: executionID,
	}

	h.WSHandler.BroadcastMessage(MsgTypeNodeError, notification)
}

// SendErrorAnalysisNotification sends an error analysis notification to clients
func (h *ErrorWebSocketHandler) SendErrorAnalysisNotification(executionID string, analysis map[string]interface{}) {
	notification := ErrorAnalysisNotification{
		Type:        "error_analysis",
		Analysis:    analysis,
		ExecutionID: executionID,
	}

	h.WSHandler.BroadcastMessage(MsgTypeDebugData, notification)
}

// SendRecoveryNotification sends a recovery notification to clients
func (h *ErrorWebSocketHandler) SendRecoveryNotification(executionID, nodeID, errorCode, strategy string, successful bool, details map[string]interface{}) {
	notification := RecoveryNotification{
		Type:        "recovery_attempt",
		Successful:  successful,
		Strategy:    strategy,
		NodeID:      nodeID,
		ErrorCode:   errorCode,
		Details:     details,
		ExecutionID: executionID,
	}

	h.WSHandler.BroadcastMessage(MsgTypeExecStatus, notification)
}

// RegisterWithErrorManager registers handlers with the error manager
func (h *ErrorWebSocketHandler) RegisterWithErrorManager(errorManager *bperrors.ErrorManager) {
	// Register a handler for all error types to send WebSocket notifications
	errorManager.RegisterErrorHandler(bperrors.ErrorTypeExecution, func(err *bperrors.BlueprintError) error {
		h.SendErrorNotification(err.ExecutionID, err)
		return nil
	})

	// Register handlers for other error types
	for _, errType := range []bperrors.ErrorType{
		bperrors.ErrorTypeConnection,
		bperrors.ErrorTypeValidation,
		bperrors.ErrorTypePermission,
		bperrors.ErrorTypeDatabase,
		bperrors.ErrorTypeNetwork,
		bperrors.ErrorTypePlugin,
		bperrors.ErrorTypeSystem,
		bperrors.ErrorTypeUnknown,
	} {
		// Use local variable to avoid closure issues
		errorType := errType

		errorManager.RegisterErrorHandler(errorType, func(err *bperrors.BlueprintError) error {
			h.SendErrorNotification(err.ExecutionID, err)
			return nil
		})
	}
}
