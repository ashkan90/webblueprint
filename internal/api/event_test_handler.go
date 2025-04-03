package api

import (
	"encoding/json"
	"net/http"
	"time"
	"webblueprint/internal/event"
	"webblueprint/internal/types"

	"github.com/gorilla/mux"
)

// EventTestHandler handles API endpoints for testing events
type EventTestHandler struct {
	eventManager event.EventManagerInterface
	wsManager    *WebSocketManager
}

// NewEventTestHandler creates a new event test handler
func NewEventTestHandler(eventManager event.EventManagerInterface, wsManager *WebSocketManager) *EventTestHandler {
	return &EventTestHandler{
		eventManager: eventManager,
		wsManager:    wsManager,
	}
}

// RegisterRoutes registers the event test routes with the router
func (h *EventTestHandler) RegisterRoutes(router *mux.Router) {
	// List all available events
	router.HandleFunc("/events/list", h.ListEvents).Methods("GET")

	// Dispatch an event
	router.HandleFunc("/events/dispatch", h.DispatchEvent).Methods("POST")

	// Get event bindings
	router.HandleFunc("/events/bindings", h.ListBindings).Methods("GET")
}

// ListEvents returns a list of all registered events
func (h *EventTestHandler) ListEvents(w http.ResponseWriter, r *http.Request) {

	// Get all events from the event manager
	events := h.eventManager.GetAllEvents()

	// Convert to a response format
	type EventResponse struct {
		ID          string                 `json:"id"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Category    string                 `json:"category"`
		Parameters  []event.EventParameter `json:"parameters"`
		BlueprintID string                 `json:"blueprintId,omitempty"`
	}

	response := make([]EventResponse, 0, len(events))
	for _, evt := range events {
		response = append(response, EventResponse{
			ID:          evt.ID,
			Name:        evt.Name,
			Description: evt.Description,
			Category:    evt.Category,
			Parameters:  evt.Parameters,
			BlueprintID: evt.BlueprintID,
		})
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DispatchEventRequest represents a request to dispatch an event
type DispatchEventRequest struct {
	EventID     string                 `json:"eventId"`
	Parameters  map[string]interface{} `json:"parameters"`
	SourceID    string                 `json:"sourceId"`
	BlueprintID string                 `json:"blueprintId"`
	ExecutionID string                 `json:"executionId"`
}

// DispatchEvent handles dispatching an event
func (h *EventTestHandler) DispatchEvent(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var request DispatchEventRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Get the event definition
	eventDef, exists := h.eventManager.GetEventDefinition(request.EventID)
	if !exists {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Convert parameters to types.Value
	params := make(map[string]types.Value)
	for name, value := range request.Parameters {
		// Find parameter definition to get the correct type
		var paramDef *event.EventParameter
		for _, p := range eventDef.Parameters {
			if p.Name == name {
				paramDef = &p
				break
			}
		}

		if paramDef == nil {
			// Parameter not defined in event
			continue
		}

		// Convert the value based on parameter type
		var typedValue types.Value
		switch paramDef.Type.ID {
		case types.PinTypes.String.ID:
			if strVal, ok := value.(string); ok {
				typedValue = types.NewValue(types.PinTypes.String, strVal)
			}
		case types.PinTypes.Number.ID:
			if numVal, ok := value.(float64); ok {
				typedValue = types.NewValue(types.PinTypes.Number, numVal)
			}
		case types.PinTypes.Boolean.ID:
			if boolVal, ok := value.(bool); ok {
				typedValue = types.NewValue(types.PinTypes.Boolean, boolVal)
			}
		case types.PinTypes.Object.ID:
			typedValue = types.NewValue(types.PinTypes.Object, value)
		default:
			// Default to Any type
			typedValue = types.NewValue(types.PinTypes.Any, value)
		}

		params[name] = typedValue
	}

	// Create the dispatch request
	dispatchRequest := event.EventDispatchRequest{
		EventID:     request.EventID,
		Parameters:  params,
		SourceID:    request.SourceID,
		BlueprintID: request.BlueprintID,
		ExecutionID: request.ExecutionID,
		Timestamp:   time.Now(),
	}

	// Dispatch the event
	errors := h.eventManager.DispatchEvent(dispatchRequest)

	// Prepare the response
	response := struct {
		Success bool     `json:"success"`
		Errors  []string `json:"errors,omitempty"`
	}{
		Success: len(errors) == 0,
	}

	if len(errors) > 0 {
		response.Errors = make([]string, len(errors))
		for i, err := range errors {
			response.Errors[i] = err.Error()
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListBindings returns a list of all event bindings
func (h *EventTestHandler) ListBindings(w http.ResponseWriter, r *http.Request) {
	// Get all bindings from the event manager
	//bindings := h.eventManager.GetAllBindings()
	bindings := []interface{}{}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(bindings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
