package api

import (
	"encoding/json"
	"net/http"
	"time"
	"webblueprint/internal/event"
	"webblueprint/internal/types"
	"webblueprint/pkg/service"

	"github.com/gorilla/mux"
)

// EventAPIHandler handles API endpoints for event management
type EventAPIHandler struct {
	// Use the concrete type to access all methods
	eventManager *event.EventManager
	eventService *service.EventService
	wsManager    *WebSocketManager
}

// NewEventAPIHandler creates a new event API handler
func NewEventAPIHandler(eventManager *event.EventManager, eventService *service.EventService, wsManager *WebSocketManager) *EventAPIHandler {
	return &EventAPIHandler{
		eventManager: eventManager,
		eventService: eventService,
		wsManager:    wsManager,
	}
}

// RegisterRoutes registers the event API routes with the router
func (h *EventAPIHandler) RegisterEventRoutes(router *mux.Router) {
	// Event definition endpoints
	router.HandleFunc("/events", h.GetEvents).Methods("GET")
	router.HandleFunc("/events", h.CreateEventDispatcher).Methods("POST") // Using CreateEventDispatcher instead of CreateEvent
	router.HandleFunc("/events/{id}", h.GetEvent).Methods("GET")
	router.HandleFunc("/events/{id}", h.UpdateEvent).Methods("PUT")
	router.HandleFunc("/events/{id}", h.DeleteEvent).Methods("DELETE")
	router.HandleFunc("/events/blueprint/{blueprintID}", h.GetBlueprintEvents).Methods("GET")

	// Event binding endpoints
	router.HandleFunc("/events/bindings", h.GetAllBindings).Methods("GET")
	router.HandleFunc("/events/bindings/{id}", h.GetBinding).Methods("GET")
	router.HandleFunc("/events/bindings", h.CreateBinding).Methods("POST")
	router.HandleFunc("/events/bindings/{id}", h.UpdateBinding).Methods("PUT")
	router.HandleFunc("/events/bindings/{id}", h.DeleteBinding).Methods("DELETE")

	// System event endpoints
	router.HandleFunc("/events/system", h.GetSystemEvents).Methods("GET")

	// Event dispatch testing endpoint
	router.HandleFunc("/events/dispatch", h.DispatchEvent).Methods("POST")
}

// CreateEvent creates a new event
func (h *EventAPIHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var event event.EventDefinition
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// If ID is not provided, generate one from the name
	if event.ID == "" && event.Name != "" {
		event.ID = h.generateEventID(event.Name)
	}

	// If category is not provided, default to "Custom Events"
	if event.Category == "" {
		event.Category = "Custom Events"
	}

	// Set creation time
	event.CreatedAt = time.Now()

	// Create the event
	err := h.eventService.CreateEvent(r.Context(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Also register with the in-memory event manager
	// No need for type assertion as h.eventManager is now *event.EventManager
	// We might want to handle the error here, but for now, mirroring the previous logic.
	_ = h.eventManager.RegisterEvent(event)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetEvents returns all registered events
func (h *EventAPIHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	// Get events from service
	events, err := h.eventService.GetAllEvents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get in-memory events to ensure we have system events too
	// No need for type assertion
	inMemoryEvents := h.eventManager.GetAllEvents()

	// Merge the lists, prioritizing DB events
	mergedEvents := make(map[string]event.EventDefinition)

	// Add in-memory events first
	for _, evt := range inMemoryEvents {
		mergedEvents[evt.ID] = evt
	}

	// Override with DB events
	for _, evt := range events {
		mergedEvents[evt.ID] = evt
	}

	// Convert back to slice
	result := make([]event.EventDefinition, 0, len(mergedEvents))
	for _, evt := range mergedEvents {
		result = append(result, evt)
	}

	for _, definition := range events {
		h.eventManager.RegisterEvent(definition)
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetEvent returns a specific event by ID
func (h *EventAPIHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	// Get event ID from URL
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Get event from service
	event, err := h.eventService.GetEventByID(r.Context(), eventID)
	if err != nil {
		// If not found in DB, check in-memory directly
		// No need for type assertion
		if eventDef, exists := h.eventManager.GetEventDefinition(eventID); exists {
			// Found in memory
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(eventDef); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return // Return after handling the in-memory case
		}
		// If not found in DB and not found in memory
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// If found in DB (err == nil)
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateEvent updates an existing event
func (h *EventAPIHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	// Get event ID from URL
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Parse request
	var updatedEvent event.EventDefinition
	if err := json.NewDecoder(r.Body).Decode(&updatedEvent); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Ensure ID matches
	updatedEvent.ID = eventID

	// Update the event
	err := h.eventService.UpdateEvent(r.Context(), updatedEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Also update in the in-memory event manager
	// No need for type assertion
	// Assuming RegisterEvent handles updates (e.g., overwrites if exists)
	_ = h.eventManager.RegisterEvent(updatedEvent)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(updatedEvent); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteEvent deletes an event
func (h *EventAPIHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	// Get event ID from URL
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Delete the event
	err := h.eventService.DeleteEvent(r.Context(), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Also remove the event definition from the in-memory event manager
	// No need for type assertion
	// Assuming UnregisterEventDefinition exists or will be added
	_ = h.eventManager.UnregisterEventDefinition(eventID) // TODO: Add UnregisterEventDefinition to EventManager

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// GetBlueprintEvents returns all events for a specific blueprint
func (h *EventAPIHandler) GetBlueprintEvents(w http.ResponseWriter, r *http.Request) {
	// Get blueprint ID from URL
	vars := mux.Vars(r)
	blueprintID := vars["blueprintID"]

	// Get events from service
	events, err := h.eventService.GetBlueprintEvents(r.Context(), blueprintID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllBindings returns all event bindings
func (h *EventAPIHandler) GetAllBindings(w http.ResponseWriter, r *http.Request) {
	// Get bindings from service
	bindings, err := h.eventService.GetAllBindings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(bindings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetBinding returns a specific binding by ID
func (h *EventAPIHandler) GetBinding(w http.ResponseWriter, r *http.Request) {
	// Get binding ID from URL
	vars := mux.Vars(r)
	bindingID := vars["id"]

	// Get binding from service
	binding, err := h.eventService.GetBindingByID(r.Context(), bindingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(binding); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// BindingRequest is the JSON structure for creating/updating bindings
type BindingRequest struct {
	EventID     string `json:"eventId"`
	HandlerID   string `json:"handlerId"`
	HandlerType string `json:"handlerType"`
	BlueprintID string `json:"blueprintId"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// CreateBinding creates a new event binding
func (h *EventAPIHandler) CreateBinding(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var request BindingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Create binding
	binding := event.EventBinding{
		ID:          request.EventID + "." + request.HandlerID,
		EventID:     request.EventID,
		HandlerID:   request.HandlerID,
		HandlerType: request.HandlerType,
		BlueprintID: request.BlueprintID,
		Priority:    request.Priority,
		CreatedAt:   time.Now(),
		Enabled:     request.Enabled,
	}

	// Create binding in service
	err := h.eventService.CreateBinding(r.Context(), binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register binding with in-memory event manager
	err = h.eventManager.BindEvent(binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(binding); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateBinding updates an existing binding
func (h *EventAPIHandler) UpdateBinding(w http.ResponseWriter, r *http.Request) {
	// Get binding ID from URL
	vars := mux.Vars(r)
	bindingID := vars["id"]

	// Parse request body
	var request BindingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Get existing binding
	existingBinding, err := h.eventService.GetBindingByID(r.Context(), bindingID)
	if err != nil {
		http.Error(w, "Binding not found", http.StatusNotFound)
		return
	}

	// First remove old binding from in-memory manager
	h.eventManager.RemoveBinding(bindingID)

	// Delete old binding from service
	err = h.eventService.DeleteBinding(r.Context(), bindingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create new binding with updated values
	binding := event.EventBinding{
		ID:          bindingID,
		EventID:     request.EventID,
		HandlerID:   request.HandlerID,
		HandlerType: request.HandlerType,
		BlueprintID: request.BlueprintID,
		Priority:    request.Priority,
		CreatedAt:   existingBinding.CreatedAt,
		Enabled:     request.Enabled,
	}

	// Create binding in service
	err = h.eventService.CreateBinding(r.Context(), binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register binding with in-memory event manager
	err = h.eventManager.BindEvent(binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(binding); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteBinding deletes a binding
func (h *EventAPIHandler) DeleteBinding(w http.ResponseWriter, r *http.Request) {
	// Get binding ID from URL
	vars := mux.Vars(r)
	bindingID := vars["id"]

	// Remove binding from in-memory manager
	h.eventManager.RemoveBinding(bindingID)

	// Delete binding from service
	err := h.eventService.DeleteBinding(r.Context(), bindingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Send empty successful response
	w.WriteHeader(http.StatusNoContent)
}

// GetSystemEvents returns all system events
func (h *EventAPIHandler) GetSystemEvents(w http.ResponseWriter, r *http.Request) {
	// Get events directly from the event manager
	// No need for type assertion
	allEvents := h.eventManager.GetAllEvents()
	// Note: If GetAllEvents could potentially fail or not be implemented,
	// we might need error handling here, but the concrete type guarantees the method exists.

	// Filter to system events
	var systemEvents []event.EventDefinition
	for _, evt := range allEvents {
		if evt.Category == "System" {
			systemEvents = append(systemEvents, evt)
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(systemEvents); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DispatchEvent dispatches an event (for testing)
func (h *EventAPIHandler) DispatchEvent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EventID string                 `json:"eventId"`
		Params  map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Convert params to Value types
	params := make(map[string]types.Value)
	for k, v := range request.Params {
		// Create a value with the appropriate type
		params[k] = createValueFromInterface(v)
	}

	// Create dispatch request
	dispatchRequest := event.EventDispatchRequest{
		EventID:    request.EventID,
		Parameters: params,
	}

	// Dispatch the event
	errs := h.eventManager.DispatchEvent(dispatchRequest)
	if len(errs) > 0 {
		http.Error(w, errs[0].Error(), http.StatusBadRequest)
		return
	}

	// Return success response
	response := struct {
		Success bool   `json:"success"`
		EventID string `json:"eventId"`
	}{
		Success: true,
		EventID: request.EventID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// generateEventID creates a standardized event ID from a name
func (h *EventAPIHandler) generateEventID(name string) string {
	// Replace spaces with hyphens
	id := ""
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			id += string(r - 'A' + 'a')
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			id += string(r)
		} else if r == ' ' {
			id += "-"
		}
	}

	// Add custom prefix if needed
	if len(id) > 0 && id[:7] != "custom." {
		id = "custom." + id
	}

	return id
}

// createValueFromInterface creates a types.Value from an interface{} value
func createValueFromInterface(v interface{}) types.Value {
	// Use the appropriate pin type and wrap the value
	switch v.(type) {
	case string:
		return types.NewValue(types.PinTypes.String, v)
	case float64, int, int64:
		return types.NewValue(types.PinTypes.Number, v)
	case bool:
		return types.NewValue(types.PinTypes.Boolean, v)
	case []interface{}:
		return types.NewValue(types.PinTypes.Array, v)
	case map[string]interface{}:
		return types.NewValue(types.PinTypes.Object, v)
	case nil:
		// For nil, return a null value
		return types.NewValue(types.PinTypes.Object, nil)
	default:
		// For unknown types, use Any type
		return types.NewValue(types.PinTypes.Any, v)
	}
}
