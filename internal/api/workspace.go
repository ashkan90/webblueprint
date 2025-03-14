package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"webblueprint/pkg/service"
)

// WorkspaceHandler handles workspace-related API requests
type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
	}
}

// RegisterRoutes registers all workspace-related routes
func (h *WorkspaceHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/workspaces", h.handleGetWorkspaces).Methods("GET")
	router.HandleFunc("/api/workspaces", h.handleCreateWorkspace).Methods("POST")
	router.HandleFunc("/api/workspaces/{id}", h.handleGetWorkspace).Methods("GET")
	router.HandleFunc("/api/workspaces/{id}", h.handleUpdateWorkspace).Methods("PUT")
	router.HandleFunc("/api/workspaces/{id}", h.handleDeleteWorkspace).Methods("DELETE")
	router.HandleFunc("/api/workspaces/{id}/members", h.handleGetWorkspaceMembers).Methods("GET")
	router.HandleFunc("/api/workspaces/{id}/members", h.handleAddWorkspaceMember).Methods("POST")
	router.HandleFunc("/api/workspaces/{id}/members/{userId}", h.handleRemoveWorkspaceMember).Methods("DELETE")
	router.HandleFunc("/api/workspaces/{id}/blueprints", h.handleGetWorkspaceBlueprints).Methods("GET")
}

// handleGetWorkspaces gets all workspaces for the current user
func (h *WorkspaceHandler) handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get user's workspaces
	workspaces, err := h.workspaceService.GetUserWorkspaces(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving workspaces: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, workspaces)
}

// handleGetWorkspace gets a specific workspace by ID
func (h *WorkspaceHandler) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the workspace
	workspace, err := h.workspaceService.GetWorkspace(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Workspace not found: %v", err))
		return
	}

	// TODO: Check if user has access to this workspace

	respondWithJSON(w, http.StatusOK, workspace)
}

// handleCreateWorkspace creates a new workspace
func (h *WorkspaceHandler) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	// Parse the workspace from request body
	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"isPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid workspace format")
		return
	}

	// Get user ID from request
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Create the workspace using the service
	workspaceID, err := h.workspaceService.CreateWorkspace(
		r.Context(),
		request.Name,
		request.Description,
		request.IsPublic,
		"user", // Owner type
		userID, // Owner ID
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating workspace: %v", err))
		return
	}

	// Get the created workspace
	workspace, err := h.workspaceService.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		// Workspace was created but can't be retrieved
		respondWithJSON(w, http.StatusCreated, map[string]string{
			"id":      workspaceID,
			"message": "Workspace created successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusCreated, workspace)
}

// handleUpdateWorkspace updates an existing workspace
func (h *WorkspaceHandler) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the workspace update from request body
	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"isPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid workspace format")
		return
	}

	// TODO: Check if user has permission to update this workspace

	// Update the workspace using the service
	err := h.workspaceService.UpdateWorkspace(
		r.Context(),
		id,
		request.Name,
		request.Description,
		request.IsPublic,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating workspace: %v", err))
		return
	}

	// Get the updated workspace
	workspace, err := h.workspaceService.GetWorkspace(r.Context(), id)
	if err != nil {
		// Workspace was updated but can't be retrieved
		respondWithJSON(w, http.StatusOK, map[string]string{
			"id":      id,
			"message": "Workspace updated successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusOK, workspace)
}

// handleDeleteWorkspace deletes a workspace
func (h *WorkspaceHandler) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Check if user has permission to delete this workspace

	// Delete the workspace using the service
	err := h.workspaceService.DeleteWorkspace(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting workspace: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Workspace deleted successfully",
	})
}

// handleGetWorkspaceMembers gets all members of a workspace
func (h *WorkspaceHandler) handleGetWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the workspace members using the service
	members, err := h.workspaceService.GetWorkspaceMembers(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving workspace members: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, members)
}

// handleAddWorkspaceMember adds a user to a workspace
func (h *WorkspaceHandler) handleAddWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body
	var request struct {
		UserID string `json:"userId"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Add member to workspace using the service
	err := h.workspaceService.AddWorkspaceMember(r.Context(), id, request.UserID, request.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error adding member: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Member added successfully",
	})
}

// handleRemoveWorkspaceMember removes a user from a workspace
func (h *WorkspaceHandler) handleRemoveWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := vars["userId"]

	// Remove member from workspace using the service
	err := h.workspaceService.RemoveWorkspaceMember(r.Context(), id, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error removing member: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Member removed successfully",
	})
}

// handleGetWorkspaceBlueprints gets all blueprints in a workspace
func (h *WorkspaceHandler) handleGetWorkspaceBlueprints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get workspace blueprints using the service
	blueprints, err := h.workspaceService.GetWorkspaceBlueprints(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving workspace blueprints: %v", err))
		return
	}

	// Return simplified blueprints with only necessary fields
	simplifiedBlueprints := make([]map[string]interface{}, len(blueprints))
	for i, bp := range blueprints {
		simplifiedBlueprints[i] = map[string]interface{}{
			"id":           bp.ID,
			"name":         bp.Name,
			"description":  bp.Description.String,
			"createdAt":    bp.CreatedAt,
			"updatedAt":    bp.UpdatedAt,
			"createdBy":    bp.CreatedBy,
			"updatedBy":    bp.UpdatedBy,
			"isPublic":     bp.IsPublic,
			"tags":         bp.Tags,
			"thumbnailURL": bp.ThumbnailURL,
			"nodeCount":    bp.NodeCount,
			"category":     bp.Category,
		}
	}

	respondWithJSON(w, http.StatusOK, simplifiedBlueprints)
}
