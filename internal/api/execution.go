package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"webblueprint/pkg/service"

	"github.com/gorilla/mux"
)

// ExecutionHandler handles execution-related API requests
type ExecutionHandler struct {
	executionService *service.ExecutionService
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(executionService *service.ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{
		executionService: executionService,
	}
}

// RegisterRoutes registers all execution-related routes
func (h *ExecutionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/executions", h.handleGetExecutions).Methods("GET")
	router.HandleFunc("/api/executions/{id}", h.handleGetExecution).Methods("GET")
	router.HandleFunc("/api/executions/{id}/logs", h.handleGetExecutionLogs).Methods("GET")
	router.HandleFunc("/api/executions/{id}/cancel", h.handleCancelExecution).Methods("POST")

	//
	router.HandleFunc("/api/executions/{id}/nodes/{nodeId}", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotImplemented)
		writer.Write([]byte("Not Implemented"))
	}).Methods("GET")

	// Blueprint execution endpoint (could also be in BlueprintHandler)
	router.HandleFunc("/api/blueprints/{id}/execute", h.handleExecuteBlueprint).Methods("POST")
}

// handleGetExecutions gets executions, optionally filtered by blueprint
func (h *ExecutionHandler) handleGetExecutions(w http.ResponseWriter, r *http.Request) {
	// Get blueprint ID from query parameters (required)
	blueprintID := r.URL.Query().Get("blueprint")
	if blueprintID == "" {
		respondWithError(w, http.StatusBadRequest, "Blueprint ID is required")
		return
	}

	// Get executions using the service
	executions, err := h.executionService.GetExecutionsByBlueprint(r.Context(), blueprintID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving executions: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, executions)
}

// handleGetExecution gets a specific execution by ID
func (h *ExecutionHandler) handleGetExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the execution using the service
	execution, err := h.executionService.GetExecution(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Execution not found: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, execution)
}

// handleGetExecutionLogs gets logs for an execution
func (h *ExecutionHandler) handleGetExecutionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the logs using the service
	logs, err := h.executionService.GetExecutionLogs(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving logs: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, logs)
}

// handleCancelExecution cancels a running execution
func (h *ExecutionHandler) handleCancelExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Cancel the execution using the service
	err := h.executionService.CancelExecution(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error canceling execution: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Execution cancelled successfully",
	})
}

// handleExecuteBlueprint executes a blueprint
func (h *ExecutionHandler) handleExecuteBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body for execution parameters
	var request struct {
		Variables map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If body can't be parsed, use empty variables
		request.Variables = make(map[string]interface{})
	}

	// Get the user ID
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Execute the blueprint using the service
	executionID, err := h.executionService.StartExecution(r.Context(), id, request.Variables, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error executing blueprint: %v", err))
		return
	}

	respondWithJSON(w, http.StatusAccepted, map[string]string{
		"executionId": executionID,
		"status":      "running",
	})
}
