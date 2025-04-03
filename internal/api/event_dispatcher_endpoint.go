package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"webblueprint/internal/event"
)

// CreateEventDispatcherRequest is the JSON structure for creating a new event dispatcher
type CreateEventDispatcherRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
}

// CreateEventDispatcher handles the creation of a new event dispatcher
func (h *EventAPIHandler) CreateEventDispatcher(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var request CreateEventDispatcherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate the request
	if request.Name == "" {
		http.Error(w, "Event name is required", http.StatusBadRequest)
		return
	}

	// Normalize to create the event ID
	eventID := "custom." + strings.ToLower(strings.ReplaceAll(request.Name, " ", "-"))

	// Set a default category if not provided
	category := request.Category
	if category == "" {
		category = "Custom Events"
	}

	// Extract blueprint ID from query parameters if provided
	blueprintID := r.URL.Query().Get("blueprintId")

	// Create the event definition
	eventDef := event.EventDefinition{
		ID:          eventID,
		Name:        request.Name,
		Description: request.Description,
		Category:    category,
		Parameters:  []event.EventParameter{},
		BlueprintID: blueprintID,
		CreatedAt:   time.Now(),
	}

	// First, save to the database
	err := h.eventService.CreateEvent(r.Context(), eventDef)
	if err != nil {
		http.Error(w, "Failed to save event to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Register the event
	err = h.eventManager.RegisterEvent(eventDef)
	if err != nil {
		// Log the error but don't fail the request since we already saved to the database
		fmt.Printf("Warning: Failed to register event in memory: %s\n", err.Error())
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Send the created event back in the response
	json.NewEncoder(w).Encode(eventDef)
}

// Note: We do not need to add this route manually anymore.
// It's now properly registered in the RegisterEventRoutes method in event_api_handler.go
