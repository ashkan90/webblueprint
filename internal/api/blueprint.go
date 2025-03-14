package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"webblueprint/pkg/blueprint"
	"webblueprint/pkg/service"

	"github.com/gorilla/mux"
)

// BlueprintHandler handles blueprint-related API requests
type BlueprintHandler struct {
	blueprintService *service.BlueprintService
}

// NewBlueprintHandler creates a new blueprint handler
func NewBlueprintHandler(blueprintService *service.BlueprintService) *BlueprintHandler {
	return &BlueprintHandler{
		blueprintService: blueprintService,
	}
}

// RegisterRoutes registers all blueprint-related routes
func (h *BlueprintHandler) RegisterRoutes(router *mux.Router) {
	// Blueprint operations
	router.HandleFunc("/api/blueprints", h.handleGetBlueprints).Methods("GET")
	router.HandleFunc("/api/blueprints", h.handleCreateBlueprint).Methods("POST")
	router.HandleFunc("/api/blueprints/{id}", h.handleGetBlueprint).Methods("GET")
	router.HandleFunc("/api/blueprints/{id}", h.handleUpdateBlueprint).Methods("PUT")
	router.HandleFunc("/api/blueprints/{id}", h.handleDeleteBlueprint).Methods("DELETE")

	// Blueprint versions
	router.HandleFunc("/api/blueprints/{id}/versions", h.handleGetVersions).Methods("GET")
	router.HandleFunc("/api/blueprints/{id}/versions", h.handleCreateVersion).Methods("POST")
	router.HandleFunc("/api/blueprints/{id}/versions/{version}", h.handleGetVersion).Methods("GET")

	// Blueprint execution
	//router.HandleFunc("/api/blueprints/{id}/execute", h.handleExecuteBlueprint).Methods("POST")
}

// handleGetBlueprints gets all blueprints, optionally filtered by workspace
func (h *BlueprintHandler) handleGetBlueprints(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from query parameters (optional)
	workspaceID := r.URL.Query().Get("workspace")
	limit := 100
	offset := 0

	var blueprints []*blueprint.Blueprint
	var err error

	// Get blueprints by workspace if provided
	if workspaceID != "" {
		blueprints, err = h.blueprintService.GetByWorkspace(r.Context(), workspaceID)
	} else {
		// Without a workspace ID, get all accessible blueprints
		blueprints, err = h.blueprintService.GetAll(r.Context(), limit, offset)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving blueprints: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, blueprints)
}

// handleGetBlueprint gets a specific blueprint by ID
func (h *BlueprintHandler) handleGetBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	bp, err := h.blueprintService.GetBlueprint(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Blueprint not found: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, bp)
}

// handleCreateBlueprint creates a new blueprint
func (h *BlueprintHandler) handleCreateBlueprint(w http.ResponseWriter, r *http.Request) {
	// Parse the blueprint from request body
	var bp blueprint.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid blueprint format")
		return
	}

	// Get the workspace ID from query parameters
	workspaceID := r.URL.Query().Get("workspace")
	if workspaceID == "" {
		respondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Get the user ID from context or request
	// In a real implementation, this would come from authentication middleware
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Create the blueprint
	blueprintID, err := h.blueprintService.CreateBlueprint(r.Context(), &bp, workspaceID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating blueprint: %v", err))
		return
	}

	// Get the created blueprint to return
	createdBP, err := h.blueprintService.GetBlueprint(r.Context(), blueprintID)
	if err != nil {
		// Blueprint was created but can't be retrieved
		respondWithJSON(w, http.StatusCreated, map[string]string{
			"id":      blueprintID,
			"message": "Blueprint created successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusCreated, createdBP)
}

// handleUpdateBlueprint updates an existing blueprint
func (h *BlueprintHandler) handleUpdateBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the blueprint from request body
	var bp blueprint.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid blueprint format")
		return
	}

	// Ensure IDs match
	if bp.ID != id {
		respondWithError(w, http.StatusBadRequest, "Blueprint ID mismatch")
		return
	}

	// Get the user ID from context or request
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get the workspace ID from query parameters
	workspaceID := r.URL.Query().Get("workspace")
	if workspaceID == "" {
		respondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Get the existing blueprint
	_, err := h.blueprintService.GetBlueprintOrCreate(
		r.Context(),
		&bp,
		workspaceID,
		userID,
	)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Blueprint not found: %v", err))
		return
	}

	// Create a new version with the updated blueprint
	versionNumber, err := h.blueprintService.SaveVersion(r.Context(), id, &bp, "Updated via API", userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating blueprint: %v", err))
		return
	}

	// Get the updated blueprint
	updatedBP, err := h.blueprintService.GetBlueprint(r.Context(), id)
	if err != nil {
		// Update succeeded but retrieval failed
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"id":            id,
			"versionNumber": versionNumber,
			"message":       "Blueprint updated successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusOK, updatedBP)
}

// handleDeleteBlueprint deletes a blueprint
func (h *BlueprintHandler) handleDeleteBlueprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete the blueprint
	err := h.blueprintService.DeleteBlueprint(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting blueprint: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Blueprint deleted successfully",
	})
}

// handleGetVersions gets all versions of a blueprint
func (h *BlueprintHandler) handleGetVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get versions
	versions, err := h.blueprintService.GetVersions(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving versions: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, versions)
}

// handleCreateVersion creates a new version of a blueprint
func (h *BlueprintHandler) handleCreateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body for comment
	var request struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If body can't be parsed, continue with empty comment
		request.Comment = ""
	}

	// Get the user ID
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get the current blueprint
	bp, err := h.blueprintService.GetBlueprint(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Blueprint not found: %v", err))
		return
	}

	// Create a new version
	versionNumber, err := h.blueprintService.SaveVersion(r.Context(), id, bp, request.Comment, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating version: %v", err))
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":            id,
		"versionNumber": versionNumber,
		"message":       "Version created successfully",
	})
}

// handleGetVersion gets a specific version of a blueprint
func (h *BlueprintHandler) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	versionStr := vars["version"]

	// Parse version number
	versionNumber, err := strconv.Atoi(versionStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	// Get the specific version
	bp, err := h.blueprintService.GetVersion(r.Context(), id, versionNumber)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Version not found: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, bp)
}

// handleExecuteBlueprint executes a blueprint
func (h *BlueprintHandler) handleExecuteBlueprint(w http.ResponseWriter, r *http.Request) {
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

	// Execute the blueprint
	executionID, err := h.blueprintService.ExecuteBlueprint(r.Context(), id, request.Variables, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error executing blueprint: %v", err))
		return
	}

	respondWithJSON(w, http.StatusAccepted, map[string]string{
		"executionId": executionID,
		"status":      "running",
	})
}

// Helper function to get user ID from request
// In a real implementation, this would come from authentication middleware
func getUserIDFromRequest(r *http.Request) string {
	// For demo purposes, get a default user ID or from a header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Try to get from the database if possible
		// For now, return a placeholder
		return "00000000-0000-0000-0000-000000000001"
	}
	return userID
}
