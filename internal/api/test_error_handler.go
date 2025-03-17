package api

import (
	"encoding/json"
	"net/http"
	errors "webblueprint/internal/bperrors"

	"github.com/gorilla/mux"
)

// TestErrorHandler provides endpoints for testing error handling
type TestErrorHandler struct {
	ErrorManager    *errors.ErrorManager
	RecoveryManager *errors.RecoveryManager
	WSManager       *WebSocketManager
}

// NewTestErrorHandler creates a new test error handler
func NewTestErrorHandler(errorManager *errors.ErrorManager, recoveryManager *errors.RecoveryManager, wsManager *WebSocketManager) *TestErrorHandler {
	return &TestErrorHandler{
		ErrorManager:    errorManager,
		RecoveryManager: recoveryManager,
		WSManager:       wsManager,
	}
}

// HandleGenerateTestError handles requests to generate test errors
func (h *TestErrorHandler) HandleGenerateTestError(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		ExecutionID string `json:"executionId"`
		ErrorType   string `json:"errorType"`
		ErrorCode   string `json:"errorCode"`
		Message     string `json:"message"`
		Severity    string `json:"severity"`
		NodeID      string `json:"nodeId"`
		Recoverable bool   `json:"recoverable"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate minimum required fields
	if req.ExecutionID == "" {
		http.Error(w, "Missing executionId", http.StatusBadRequest)
		return
	}

	// Set defaults if not provided
	if req.ErrorType == "" {
		req.ErrorType = string(errors.ErrorTypeExecution)
	}

	if req.ErrorCode == "" {
		req.ErrorCode = string(errors.ErrNodeExecutionFailed)
	}

	if req.Message == "" {
		req.Message = "Test error message"
	}

	if req.Severity == "" {
		req.Severity = string(errors.SeverityMedium)
	}

	// Generate error
	generator := errors.NewTestErrorGenerator()
	err := generator.GenerateTestError(
		errors.ErrorType(req.ErrorType),
		errors.BlueprintErrorCode(req.ErrorCode),
		req.Message,
		errors.ErrorSeverity(req.Severity),
	)

	// Set node info if provided
	if req.NodeID != "" {
		err.WithNodeInfo(req.NodeID, "")
	}

	// Set blueprint info
	err.WithBlueprintInfo("test-blueprint", req.ExecutionID)

	// Make recoverable if requested
	if req.Recoverable {
		err.WithRecoveryOptions(errors.RecoveryRetry, errors.RecoverySkipNode)
	}

	// Record error
	// h.ErrorManager.RecordError(req.ExecutionID, err)

	// Send WebSocket notifications
	if h.WSManager != nil {
		// Send error notification
		h.WSManager.BroadcastMessage(MsgTypeNodeError, ErrorNotification{
			Type:        "error",
			Error:       err,
			ExecutionID: req.ExecutionID,
		})

		// Send error analysis
		analysis := h.ErrorManager.AnalyzeErrors(req.ExecutionID)
		h.WSManager.BroadcastMessage(MsgTypeDebugData, ErrorAnalysisNotification{
			Type:        "error_analysis",
			Analysis:    analysis,
			ExecutionID: req.ExecutionID,
		})
	}

	// Return response
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"error":   err,
	})
}

// HandleGenerateErrorScenario handles requests to generate test error scenarios
func (h *TestErrorHandler) HandleGenerateErrorScenario(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		ExecutionID  string `json:"executionId"`
		ScenarioType string `json:"scenarioType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ExecutionID == "" {
		http.Error(w, "Missing executionId", http.StatusBadRequest)
		return
	}

	if req.ScenarioType == "" {
		http.Error(w, "Missing scenarioType", http.StatusBadRequest)
		return
	}

	// Generate scenario
	generator := errors.NewTestErrorGenerator()
	analysis, err := generator.SimulateErrorScenario(req.ScenarioType, req.ExecutionID)

	// Send WebSocket notifications for all errors
	if h.WSManager != nil {
		// Get all errors
		allErrors := h.ErrorManager.GetErrors(req.ExecutionID)

		// Send notifications for each error
		for _, e := range allErrors {
			h.WSManager.BroadcastMessage(MsgTypeNodeError, ErrorNotification{
				Type:        "error",
				Error:       e,
				ExecutionID: req.ExecutionID,
			})
		}

		// Send error analysis
		h.WSManager.BroadcastMessage(MsgTypeDebugData, ErrorAnalysisNotification{
			Type:        "error_analysis",
			Analysis:    analysis,
			ExecutionID: req.ExecutionID,
		})
	}

	// If we got an error, include it in the response
	var errorData interface{} = nil
	if err != nil {
		if bpErr, ok := err.(*errors.BlueprintError); ok {
			errorData = bpErr
		} else {
			errorData = map[string]interface{}{
				"message": err.Error(),
			}
		}
	}

	// Return response
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"analysis": analysis,
		"error":    errorData,
	})
}

// RegisterRoutes registers the API routes
func (h *TestErrorHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/test/generate-error", h.HandleGenerateTestError).Methods("POST")
	router.HandleFunc("/api/test/generate-scenario", h.HandleGenerateErrorScenario).Methods("POST")
}
