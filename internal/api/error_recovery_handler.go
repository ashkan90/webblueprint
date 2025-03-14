package api

import (
	"encoding/json"
	"net/http"
	errors "webblueprint/internal/bperrors"
)

// RecoveryRequest represents a request to recover from an error
type RecoveryRequest struct {
	ExecutionID string `json:"executionId"`
	NodeID      string `json:"nodeId"`
	ErrorCode   string `json:"errorCode"`
	Strategy    string `json:"strategy"`
}

// RecoveryResponse represents the response to an error recovery attempt
type RecoveryResponse struct {
	Success  bool                   `json:"success"`
	Strategy string                 `json:"strategy"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// ErrorRecoveryHandler handles error recovery requests
type ErrorRecoveryHandler struct {
	wsManager       *WebSocketManager
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
}

// NewErrorRecoveryHandler creates a new error recovery handler
func NewErrorRecoveryHandler(
	wsManager *WebSocketManager,
	errorManager *errors.ErrorManager,
	recoveryManager *errors.RecoveryManager,
) *ErrorRecoveryHandler {
	return &ErrorRecoveryHandler{
		wsManager:       wsManager,
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
	}
}

// HandleErrorRecovery handles error recovery requests
func (h *ErrorRecoveryHandler) HandleErrorRecovery(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req RecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ExecutionID == "" || req.NodeID == "" || req.ErrorCode == "" || req.Strategy == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Find matching errors
	nodeErrors := h.errorManager.GetNodeErrors(req.ExecutionID, req.NodeID)
	var targetError *errors.BlueprintError

	for _, err := range nodeErrors {
		if string(err.Code) == req.ErrorCode {
			targetError = err
			break
		}
	}

	if targetError == nil {
		http.Error(w, "Error not found", http.StatusNotFound)
		return
	}

	// Check if the strategy is valid
	strategy := errors.RecoveryStrategy(req.Strategy)
	validStrategy := false

	for _, s := range targetError.RecoveryOptions {
		if s == strategy {
			validStrategy = true
			break
		}
	}

	if !validStrategy {
		http.Error(w, "Invalid recovery strategy", http.StatusBadRequest)
		return
	}

	// Attempt recovery
	success, details := h.recoveryManager.RecoverFromError(req.ExecutionID, targetError)

	// Send recovery notification via WebSocket
	h.sendRecoveryNotification(
		req.ExecutionID,
		req.NodeID,
		req.ErrorCode,
		req.Strategy,
		success,
		details,
	)

	// Return response
	response := RecoveryResponse{
		Success:  success,
		Strategy: req.Strategy,
		Details:  details,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendRecoveryNotification sends a recovery notification to clients
func (h *ErrorRecoveryHandler) sendRecoveryNotification(
	executionID, nodeID, errorCode, strategy string,
	successful bool,
	details map[string]interface{},
) {
	notification := RecoveryNotification{
		Type:        "recovery_attempt",
		Successful:  successful,
		Strategy:    strategy,
		NodeID:      nodeID,
		ErrorCode:   errorCode,
		Details:     details,
		ExecutionID: executionID,
	}

	h.wsManager.BroadcastMessage(MsgTypeExecStatus, notification)
}

// RegisterRoutes registers the API routes for error recovery
func (h *ErrorRecoveryHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/api/errors/recover", h.HandleErrorRecovery)
}
