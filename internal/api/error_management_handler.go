package api

import (
	"encoding/json"
	"net/http"
	"time"
	errors "webblueprint/internal/bperrors"
)

// ErrorManagementHandler handles error management API endpoints
type ErrorManagementHandler struct {
	errorManager    *errors.ErrorManager
	recoveryManager *errors.RecoveryManager
}

// NewErrorManagementHandler creates a new error management handler
func NewErrorManagementHandler(errorManager *errors.ErrorManager, recoveryManager *errors.RecoveryManager) *ErrorManagementHandler {
	return &ErrorManagementHandler{
		errorManager:    errorManager,
		recoveryManager: recoveryManager,
	}
}

// ListErrors returns all errors for an execution
func (h *ErrorManagementHandler) ListErrors(w http.ResponseWriter, r *http.Request) error {
	// Get execution ID from query params
	executionID := r.URL.Query().Get("executionId")
	if executionID == "" {
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrMissingRequiredInput,
			"Missing executionId parameter",
			errors.SeverityMedium,
		)
	}

	// Get errors
	errorList := h.errorManager.GetErrors(executionID)

	// Return errors
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"errors":  errorList,
		"count":   len(errorList),
	})
}

// GetErrorAnalysis returns error analysis for an execution
func (h *ErrorManagementHandler) GetErrorAnalysis(w http.ResponseWriter, r *http.Request) error {
	// Get execution ID from query params
	executionID := r.URL.Query().Get("executionId")
	if executionID == "" {
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrMissingRequiredInput,
			"Missing executionId parameter",
			errors.SeverityMedium,
		)
	}

	// Get analysis
	analysis := h.errorManager.AnalyzeErrors(executionID)

	// Return analysis
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"analysis": analysis,
	})
}

// RecoverFromError attempts to recover from an error
func (h *ErrorManagementHandler) RecoverFromError(w http.ResponseWriter, r *http.Request) error {
	// Parse request
	var req struct {
		ExecutionID string `json:"executionId"`
		NodeID      string `json:"nodeId"`
		ErrorCode   string `json:"errorCode"`
		Strategy    string `json:"strategy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.Wrap(
			err,
			errors.ErrorTypeValidation,
			errors.ErrInvalidNodeConfiguration,
			"Invalid request format",
			errors.SeverityMedium,
		)
	}

	// Validate request
	if req.ExecutionID == "" || req.ErrorCode == "" || req.Strategy == "" {
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrMissingRequiredInput,
			"Missing required fields",
			errors.SeverityMedium,
		)
	}

	// Find matching errors
	var targetError *errors.BlueprintError

	if req.NodeID != "" {
		// Look for specific node error
		nodeErrors := h.errorManager.GetNodeErrors(req.ExecutionID, req.NodeID)
		for _, err := range nodeErrors {
			if string(err.Code) == req.ErrorCode {
				targetError = err
				break
			}
		}
	} else {
		// Look through all errors
		allErrors := h.errorManager.GetErrors(req.ExecutionID)
		for _, err := range allErrors {
			if string(err.Code) == req.ErrorCode {
				targetError = err
				break
			}
		}
	}

	if targetError == nil {
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrNodeNotFound,
			"Error not found",
			errors.SeverityMedium,
		)
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
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrInvalidPropertyValue,
			"Invalid recovery strategy",
			errors.SeverityMedium,
		)
	}

	// Attempt recovery
	success, details := h.recoveryManager.RecoverFromError(req.ExecutionID, targetError)

	// Add timestamp to details
	if details == nil {
		details = make(map[string]interface{})
	}
	details["timestamp"] = time.Now()

	// Return result
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  success,
		"strategy": req.Strategy,
		"details":  details,
		"error":    targetError,
	})
}

// ClearErrors clears all errors for an execution
func (h *ErrorManagementHandler) ClearErrors(w http.ResponseWriter, r *http.Request) error {
	// Parse request
	var req struct {
		ExecutionID string `json:"executionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.Wrap(
			err,
			errors.ErrorTypeValidation,
			errors.ErrInvalidNodeConfiguration,
			"Invalid request format",
			errors.SeverityMedium,
		)
	}

	// Validate request
	if req.ExecutionID == "" {
		return errors.New(
			errors.ErrorTypeValidation,
			errors.ErrMissingRequiredInput,
			"Missing executionId",
			errors.SeverityMedium,
		)
	}

	// Clear errors
	h.errorManager.ClearErrors(req.ExecutionID)

	// Also clear recovery attempts
	h.recoveryManager.ClearRecoveryAttempts(req.ExecutionID)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// RegisterRoutes registers the API routes
func (h *ErrorManagementHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/api/errors/list", ErrorHandler(h.ListErrors))
	router.HandleFunc("/api/errors/analysis", ErrorHandler(h.GetErrorAnalysis))
	router.HandleFunc("/api/errors/recover", ErrorHandler(h.RecoverFromError))
	router.HandleFunc("/api/errors/clear", ErrorHandler(h.ClearErrors))
}
