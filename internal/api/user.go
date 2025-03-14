package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"webblueprint/pkg/service"

	"github.com/gorilla/mux"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/users", h.handleGetUsers).Methods("GET")
	router.HandleFunc("/api/users", h.handleCreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", h.handleUpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/me", h.handleGetCurrentUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", h.handleGetUser).Methods("GET")
	router.HandleFunc("/api/auth/login", h.handleLogin).Methods("POST")
}

// handleGetUsers gets all users (admin only)
func (h *UserHandler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get all users with proper pagination
	// This would require a GetAll method on the UserService

	respondWithError(w, http.StatusNotImplemented, "Not implemented yet")
}

// handleGetUser gets a specific user by ID
func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get the user using the service
	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// handleCreateUser creates a new user
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the user from request body
	var request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user format")
		return
	}

	// Create the user using the service
	userID, err := h.userService.CreateUser(
		r.Context(),
		request.Username,
		request.Email,
		request.Password,
		request.FullName,
	)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error creating user: %v", err))
		return
	}

	// Get the created user
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		// User was created but can't be retrieved
		respondWithJSON(w, http.StatusCreated, map[string]string{
			"id":      userID,
			"message": "User created successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

// handleUpdateUser updates an existing user
func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the user update from request body
	var request struct {
		Email    string `json:"email"`
		FullName string `json:"fullName"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user format")
		return
	}

	// TODO: Check if the current user has permission to update this user

	// Update the user using the service
	err := h.userService.UpdateUser(
		r.Context(),
		id,
		request.Email,
		request.FullName,
		request.Password,
	)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error updating user: %v", err))
		return
	}

	// Get the updated user
	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		// User was updated but can't be retrieved
		respondWithJSON(w, http.StatusOK, map[string]string{
			"id":      id,
			"message": "User updated successfully but could not be retrieved",
		})
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// handleGetCurrentUser gets the currently authenticated user
func (h *UserHandler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request
	userID := getUserIDFromRequest(r)
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get the user using the service
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// handleLogin authenticates a user and returns a token
func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse login request
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid login format")
		return
	}

	// Verify credentials using the service
	user, err := h.userService.VerifyCredentials(r.Context(), request.Username, request.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// TODO: Generate JWT token for authentication
	// For now, we'll just return the user ID

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"userId":   user.ID,
		"username": user.Username,
		"role":     user.Role,
		// In a real implementation, you would return a token here
		"token": "dummy-token",
	})
}
