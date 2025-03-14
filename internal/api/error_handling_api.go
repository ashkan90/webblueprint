package api

import (
	"encoding/json"
	"net/http"
	"webblueprint/internal/bperrors"
)

// ErrorHandlingAPI provides endpoints for error management
type ErrorHandlingAPI struct {
	ErrorManager    *bperrors.ErrorManager
	RecoveryManager *bperrors.RecoveryManager
	InfoStore       *bperrors.ExecutionInfoStore
}

// NewErrorHandlingAPI creates a new error handling API
func NewErrorHandlingAPI(errorManager *bperrors.ErrorManager, recoveryManager *bperrors.RecoveryManager, infoStore *bperrors.ExecutionInfoStore) *ErrorHandlingAPI {
	return &ErrorHandlingAPI{
		ErrorManager:    errorManager,
		RecoveryManager: recoveryManager,
		InfoStore:       infoStore,
	}
}

// HandleListErrors handles the request to list all errors for an execution
func (api *ErrorHandlingAPI) HandleListErrors(w http.ResponseWriter, r *http.Request) {
	// Get execution ID from query params
	executionID := r.URL.Query().Get("executionId")
	if executionID == "" {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Missing executionId parameter",
			ErrorCode: string(bperrors.ErrMissingRequiredInput),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Get errors
	errors := api.ErrorManager.GetErrors(executionID)

	// Return response
	response := map[string]interface{}{
		"success": true,
		"errors":  errors,
		"count":   len(errors),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleGetErrorAnalysis handles the request to get error analysis for an execution
func (api *ErrorHandlingAPI) HandleGetErrorAnalysis(w http.ResponseWriter, r *http.Request) {
	// Get execution ID from query params
	executionID := r.URL.Query().Get("executionId")
	if executionID == "" {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Missing executionId parameter",
			ErrorCode: string(bperrors.ErrMissingRequiredInput),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Get analysis
	analysis := api.ErrorManager.AnalyzeErrors(executionID)

	// Return response
	response := map[string]interface{}{
		"success":  true,
		"analysis": analysis,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleGetExecutionInfo handles the request to get extended execution info
func (api *ErrorHandlingAPI) HandleGetExecutionInfo(w http.ResponseWriter, r *http.Request) {
	// Get execution ID from query params
	executionID := r.URL.Query().Get("executionId")
	if executionID == "" {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Missing executionId parameter",
			ErrorCode: string(bperrors.ErrMissingRequiredInput),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// If InfoStore is nil, respond with an error
	if api.InfoStore == nil {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Execution info store not available",
			ErrorCode: string(bperrors.ErrInternalServerError),
		}
		respondWithJSON(w, http.StatusInternalServerError, errorResponse)
		return
	}

	// Get execution info
	info, exists := api.InfoStore.GetExecutionInfo(executionID)
	if !exists {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Execution info not found",
			ErrorCode: string(bperrors.ErrBlueprintNotFound),
		}
		respondWithJSON(w, http.StatusNotFound, errorResponse)
		return
	}

	// Return response
	response := map[string]interface{}{
		"success": true,
		"info":    info.ToMap(),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleAttemptRecovery handles the request to attempt recovery from an error
func (api *ErrorHandlingAPI) HandleAttemptRecovery(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		ExecutionID string `json:"executionId"`
		NodeID      string `json:"nodeId"`
		ErrorCode   string `json:"errorCode"`
		Strategy    string `json:"strategy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Invalid request format",
			ErrorCode: string(bperrors.ErrInvalidNodeConfiguration),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Validate request
	if req.ExecutionID == "" || req.NodeID == "" || req.ErrorCode == "" || req.Strategy == "" {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Missing required fields",
			ErrorCode: string(bperrors.ErrMissingRequiredInput),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Find matching errors
	nodeErrors := api.ErrorManager.GetNodeErrors(req.ExecutionID, req.NodeID)
	var targetError *bperrors.BlueprintError

	for _, err := range nodeErrors {
		if string(err.Code) == req.ErrorCode {
			targetError = err
			break
		}
	}

	if targetError == nil {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Error not found",
			ErrorCode: string(bperrors.ErrNodeNotFound),
		}
		respondWithJSON(w, http.StatusNotFound, errorResponse)
		return
	}

	// Attempt recovery
	success, details := api.RecoveryManager.RecoverFromError(req.ExecutionID, targetError)

	// Return response
	response := map[string]interface{}{
		"success":  success,
		"strategy": req.Strategy,
		"details":  details,
		"error":    targetError,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleClearErrors handles the request to clear all errors for an execution
func (api *ErrorHandlingAPI) HandleClearErrors(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		ExecutionID string `json:"executionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Invalid request format",
			ErrorCode: string(bperrors.ErrInvalidNodeConfiguration),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Validate request
	if req.ExecutionID == "" {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Missing executionId",
			ErrorCode: string(bperrors.ErrMissingRequiredInput),
		}
		respondWithJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	// Clear errors
	api.ErrorManager.ClearErrors(req.ExecutionID)

	// Clear recovery attempts
	api.RecoveryManager.ClearRecoveryAttempts(req.ExecutionID)

	// Return response
	response := map[string]interface{}{
		"success": true,
	}

	respondWithJSON(w, http.StatusOK, response)
}
