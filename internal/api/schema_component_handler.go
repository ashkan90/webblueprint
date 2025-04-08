package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"webblueprint/internal/db" // Import the db package
	"webblueprint/pkg/dto"

	"github.com/go-chi/chi/v5" // Assuming chi router is used, adjust if different
)

// SchemaComponentHandler handles API requests related to schema components.
type SchemaComponentHandler struct {
	Store db.SchemaComponentStore
}

// NewSchemaComponentHandler creates a new handler instance.
func NewSchemaComponentHandler(store db.SchemaComponentStore) *SchemaComponentHandler {
	return &SchemaComponentHandler{Store: store}
}

// CreateSchemaComponent handles POST requests to create a new schema component.
func (h *SchemaComponentHandler) CreateSchemaComponent(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name             string `json:"name"`
		SchemaDefinition string `json:"schema_definition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Name == "" || payload.SchemaDefinition == "" {
		http.Error(w, "Missing required fields: name and schema_definition", http.StatusBadRequest)
		return
	}

	component, err := h.Store.CreateSchemaComponent(payload.Name, payload.SchemaDefinition)
	if err != nil {
		fmt.Printf("Error creating schema component: %v\n", err) // Replace with proper logging
		http.Error(w, "Failed to create schema component", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(component)
}

// ListSchemaComponents handles GET requests to list all schema components.
func (h *SchemaComponentHandler) ListSchemaComponents(w http.ResponseWriter, r *http.Request) {
	components, err := h.Store.ListSchemaComponents()
	if err != nil {
		fmt.Printf("Error listing schema components: %v\n", err) // Replace with proper logging
		http.Error(w, "Failed to list schema components", http.StatusInternalServerError)
		return
	}

	componentsDto := make([]*dto.SchemaDefinition, len(components))
	for i, component := range components {
		componentsDto[i], _ = h.Store.ToPkgSchema(&component)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(componentsDto)
}

// GetSchemaComponent handles GET requests for a specific schema component by ID.
func (h *SchemaComponentHandler) GetSchemaComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // Assuming chi router for URL params
	if id == "" {
		http.Error(w, "Missing schema component ID", http.StatusBadRequest)
		return
	}

	component, err := h.Store.GetSchemaComponent(id)
	if err != nil {
		// Check if it's a 'not found' error
		if err.Error() == fmt.Sprintf("schema component not found with id: %s", id) { // Basic check, improve if needed
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Printf("Error getting schema component: %v\n", err) // Replace with proper logging
			http.Error(w, "Failed to get schema component", http.StatusInternalServerError)
		}
		return
	}

	componentDto, _ := h.Store.ToPkgSchema(component)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(componentDto)
}

// UpdateSchemaComponent handles PUT requests to update a schema component.
func (h *SchemaComponentHandler) UpdateSchemaComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing schema component ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Name             string `json:"name"`
		SchemaDefinition string `json:"schema_definition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Name == "" || payload.SchemaDefinition == "" {
		http.Error(w, "Missing required fields: name and schema_definition", http.StatusBadRequest)
		return
	}

	component, err := h.Store.UpdateSchemaComponent(id, payload.Name, payload.SchemaDefinition)
	if err != nil {
		if err.Error() == fmt.Sprintf("schema component not found with id for update: %s", id) { // Basic check
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Printf("Error updating schema component: %v\n", err) // Replace with proper logging
			http.Error(w, "Failed to update schema component", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(component)
}

// DeleteSchemaComponent handles DELETE requests to remove a schema component.
func (h *SchemaComponentHandler) DeleteSchemaComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing schema component ID", http.StatusBadRequest)
		return
	}

	err := h.Store.DeleteSchemaComponent(id)
	if err != nil {
		if err.Error() == fmt.Sprintf("schema component not found with id for deletion: %s", id) { // Basic check
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			fmt.Printf("Error deleting schema component: %v\n", err) // Replace with proper logging
			http.Error(w, "Failed to delete schema component", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // Success, no content to return
}
