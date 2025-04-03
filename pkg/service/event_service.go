package service

import (
	"context"
	"fmt"
	"time"
	"webblueprint/internal/event"
	"webblueprint/pkg/repository"
)

// EventService provides business logic for event operations
type EventService struct {
	eventRepo repository.EventRepository
}

// NewEventService creates a new event service
func NewEventService(eventRepo repository.EventRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
	}
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, event event.EventDefinition) error {
	// Set creation time if not already set
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Validate event fields
	if err := s.validateEvent(event); err != nil {
		return err
	}

	// Create event in the repository
	return s.eventRepo.Create(ctx, event)
}

// GetEventByID retrieves an event by its ID
func (s *EventService) GetEventByID(ctx context.Context, id string) (event.EventDefinition, error) {
	return s.eventRepo.GetByID(ctx, id)
}

// GetAllEvents retrieves all events
func (s *EventService) GetAllEvents(ctx context.Context) ([]event.EventDefinition, error) {
	return s.eventRepo.GetAll(ctx)
}

// GetBlueprintEvents retrieves all events for a blueprint
func (s *EventService) GetBlueprintEvents(ctx context.Context, blueprintID string) ([]event.EventDefinition, error) {
	return s.eventRepo.GetByBlueprintID(ctx, blueprintID)
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, event event.EventDefinition) error {
	// Validate event fields
	if err := s.validateEvent(event); err != nil {
		return err
	}

	// Update event in the repository
	return s.eventRepo.Update(ctx, event)
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, id string) error {
	return s.eventRepo.Delete(ctx, id)
}

// CreateBinding creates a new event binding
func (s *EventService) CreateBinding(ctx context.Context, binding event.EventBinding) error {
	// Set creation time if not already set
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = time.Now()
	}

	// Create binding in the repository
	return s.eventRepo.CreateBinding(ctx, binding)
}

// GetBindingByID retrieves a binding by its ID
func (s *EventService) GetBindingByID(ctx context.Context, id string) (event.EventBinding, error) {
	return s.eventRepo.GetBindingByID(ctx, id)
}

// GetBindingsByEventID retrieves all bindings for an event
func (s *EventService) GetBindingsByEventID(ctx context.Context, eventID string) ([]event.EventBinding, error) {
	return s.eventRepo.GetBindingsByEventID(ctx, eventID)
}

// GetAllBindings retrieves all bindings
func (s *EventService) GetAllBindings(ctx context.Context) ([]event.EventBinding, error) {
	return s.eventRepo.GetAllBindings(ctx)
}

// DeleteBinding deletes a binding
func (s *EventService) DeleteBinding(ctx context.Context, id string) error {
	return s.eventRepo.DeleteBinding(ctx, id)
}

// DeleteBindingsByEventID deletes all bindings for an event
func (s *EventService) DeleteBindingsByEventID(ctx context.Context, eventID string) error {
	return s.eventRepo.DeleteBindingsByEventID(ctx, eventID)
}

// DeleteBindingsByBlueprintID deletes all bindings for a blueprint
func (s *EventService) DeleteBindingsByBlueprintID(ctx context.Context, blueprintID string) error {
	return s.eventRepo.DeleteBindingsByBlueprintID(ctx, blueprintID)
}

// validateEvent validates an event's fields
func (s *EventService) validateEvent(event event.EventDefinition) error {
	if event.ID == "" {
		return fmt.Errorf("event ID is required")
	}

	if event.Name == "" {
		return fmt.Errorf("event name is required")
	}

	// Check that category is valid
	validCategories := map[string]bool{
		"System":        true,
		"Blueprint":     true,
		"Custom Events": true,
		"User":          true,
		"Input":         true,
	}

	if event.Category == "" {
		return fmt.Errorf("event category is required")
	}

	if !validCategories[event.Category] {
		// If not in our predefined list, check that it starts with "Custom" or accept it anyway
		if len(event.Category) < 6 || event.Category[:6] != "Custom" {
			// We'll accept it but log a warning (which we can't do here, so we'll just accept it)
		}
	}

	return nil
}
