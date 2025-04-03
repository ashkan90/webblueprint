package event

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"webblueprint/internal/core" // Now includes EngineController
	"webblueprint/internal/types"
)

// EventManager implements event management capabilities
type EventManager struct {
	definitions      map[string]EventDefinition      // EventID -> EventDefinition
	bindings         map[string][]EventBinding       // EventID -> []EventBinding sorted by priority desc
	handlerFuncs     map[string]EventHandlerFunc     // BindingID -> Generated HandlerFunc that triggers engine
	systemEvents     map[core.SystemEventType]string // SystemEventType -> EventID
	blueprintEvents  map[string][]string             // BlueprintID -> []EventID
	engineController core.EngineController           // Interface to trigger node execution
	mutex            sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager(engineController core.EngineController) *EventManager {
	if engineController == nil {
		// Or handle this more gracefully, maybe a default no-op controller?
		panic("EventManager requires a non-nil EngineController")
	}
	manager := &EventManager{
		definitions:      make(map[string]EventDefinition),
		bindings:         make(map[string][]EventBinding),
		handlerFuncs:     make(map[string]EventHandlerFunc),
		systemEvents:     make(map[core.SystemEventType]string),
		blueprintEvents:  make(map[string][]string),
		engineController: engineController, // Store engine controller reference
	}
	// Register built-in system events
	manager.registerSystemEvents()
	return manager
}

// registerSystemEvents registers the built-in system events
func (em *EventManager) registerSystemEvents() {
	// Define system events (using constants from event_types.go)
	initEvent := EventDefinition{
		ID:          "system.initialize",
		Name:        string(EventTypeInitialize),
		Description: "Triggered when a blueprint starts execution",
		Parameters: []EventParameter{
			{Name: "blueprintID", Type: types.PinTypes.String, Description: "ID of the blueprint being initialized", Optional: false},
			{Name: "executionID", Type: types.PinTypes.String, Description: "ID of the execution instance", Optional: false},
		},
		Category:  "System",
		CreatedAt: time.Now(),
	}

	shutdownEvent := EventDefinition{
		ID:          "system.shutdown",
		Name:        string(EventTypeShutdown),
		Description: "Triggered when a blueprint execution ends",
		Parameters: []EventParameter{
			{Name: "blueprintID", Type: types.PinTypes.String, Description: "ID of the blueprint being shut down", Optional: false},
			{Name: "executionID", Type: types.PinTypes.String, Description: "ID of the execution instance", Optional: false},
			{Name: "success", Type: types.PinTypes.Boolean, Description: "Whether the execution completed successfully", Optional: false},
			{Name: "errorMessage", Type: types.PinTypes.String, Description: "Error message if execution failed", Optional: true},
		},
		Category:  "System",
		CreatedAt: time.Now(),
	}

	// Register system events
	em.definitions[initEvent.ID] = initEvent
	em.bindings[initEvent.ID] = make([]EventBinding, 0)
	em.definitions[shutdownEvent.ID] = shutdownEvent
	em.bindings[shutdownEvent.ID] = make([]EventBinding, 0)

	// Map system event types to event IDs
	em.systemEvents[EventTypeInitialize] = initEvent.ID
	em.systemEvents[EventTypeShutdown] = shutdownEvent.ID
	// Add other system events here...
}

// RegisterEvent registers a new event definition
func (em *EventManager) RegisterEvent(event EventDefinition) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.definitions[event.ID]; exists {
		return fmt.Errorf("event with ID %s already exists", event.ID)
	}
	em.definitions[event.ID] = event
	if _, exists := em.bindings[event.ID]; !exists {
		em.bindings[event.ID] = make([]EventBinding, 0)
	}
	if event.BlueprintID != "" {
		em.blueprintEvents[event.BlueprintID] = append(em.blueprintEvents[event.BlueprintID], event.ID)
	}
	return nil
}

// UnregisterEventDefinition removes an event definition and all its bindings/handlers
func (em *EventManager) UnregisterEventDefinition(eventID string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// 1. Check if event definition exists
	eventDef, exists := em.definitions[eventID]
	if !exists {
		return fmt.Errorf("event definition with ID %s not found", eventID)
	}

	// 2. Remove event definition
	delete(em.definitions, eventID)

	// 3. Remove associated bindings and handlers
	if bindings, ok := em.bindings[eventID]; ok {
		for _, binding := range bindings {
			// Remove handler function
			delete(em.handlerFuncs, binding.ID)
		}
		// Remove all bindings for this event
		delete(em.bindings, eventID)
	}

	// 4. Remove from blueprintEvents map if it was a custom event
	if eventDef.BlueprintID != "" {
		if bpEvents, ok := em.blueprintEvents[eventDef.BlueprintID]; ok {
			updatedBpEvents := make([]string, 0, len(bpEvents)-1)
			for _, id := range bpEvents {
				if id != eventID {
					updatedBpEvents = append(updatedBpEvents, id)
				}
			}
			if len(updatedBpEvents) == 0 {
				delete(em.blueprintEvents, eventDef.BlueprintID)
			} else {
				em.blueprintEvents[eventDef.BlueprintID] = updatedBpEvents
			}
		}
	}

	// 5. Remove from systemEvents map if it was a system event (less likely to be unregistered, but for completeness)
	for sysType, id := range em.systemEvents {
		if id == eventID {
			delete(em.systemEvents, sysType)
			// Assuming an event ID maps to only one system type
			break
		}
	}

	return nil
}

// RegisterHandler generates and registers the actual handler function (which triggers the engine) for a binding.
// This is intended to be called internally by BindEvent.
func (em *EventManager) RegisterHandler(binding EventBinding) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Check if handler already registered for this binding
	if _, exists := em.handlerFuncs[binding.ID]; exists {
		// This might happen if BindEvent is called multiple times for the same binding ID
		// Or if there's a race condition. Log a warning but maybe don't return error?
		fmt.Printf("Warning: Handler already registered for binding %s. Overwriting.\n", binding.ID)
		// return fmt.Errorf("handler already registered for binding %s", binding.ID)
	}

	// --- Generate the handler function ---
	handlerFunc := func(ctx EventHandlerContext) error {
		// Convert local event context to core event context
		coreCtx := core.EventHandlerContext{
			EventID:     ctx.EventID,
			Parameters:  ctx.Parameters,
			SourceID:    ctx.SourceID,
			BlueprintID: ctx.BlueprintID, // Source blueprint ID
			ExecutionID: ctx.ExecutionID,
			HandlerID:   ctx.HandlerID, // Target node ID
			BindingID:   ctx.BindingID,
			Timestamp:   ctx.Timestamp,
		}

		// Use the captured engineController reference to trigger the node execution
		// Pass the core.EventHandlerContext
		err := em.engineController.TriggerNodeExecution(binding.BlueprintID, binding.HandlerID, coreCtx)
		if err != nil {
			// Log or handle the error from triggering the node
			// TODO: Use a proper logger passed to EventManager or context
			fmt.Printf("Error triggering event handler node %s (blueprint %s) for binding %s: %v\n",
				binding.HandlerID, binding.BlueprintID, binding.ID, err)
			return err // Propagate the error
		}
		return nil
	}
	// --- Store the generated function ---
	em.handlerFuncs[binding.ID] = handlerFunc
	return nil
}

// BindEvent creates a binding and registers its handler function.
func (em *EventManager) BindEvent(binding EventBinding) error {
	em.mutex.Lock() // Lock for modifying bindings list

	// Check if event exists
	if _, exists := em.definitions[binding.EventID]; !exists {
		em.mutex.Unlock()
		return fmt.Errorf("event with ID %s does not exist", binding.EventID)
	}

	// Check if this exact binding already exists to prevent duplicates
	if bindings, ok := em.bindings[binding.EventID]; ok {
		for _, existingBinding := range bindings {
			if existingBinding.ID == binding.ID {
				em.mutex.Unlock()
				// Decide if this should be an error or just a no-op
				fmt.Printf("Warning: Binding with ID %s already exists for event %s. Skipping.\n", binding.ID, binding.EventID)
				return nil // Or return an error fmt.Errorf("binding with ID %s already exists", binding.ID)
			}
		}
	}

	// Add binding to the event's bindings list
	em.bindings[binding.EventID] = append(em.bindings[binding.EventID], binding)

	// Sort bindings by priority (higher priority first)
	sort.Slice(em.bindings[binding.EventID], func(i, j int) bool {
		return em.bindings[binding.EventID][i].Priority > em.bindings[binding.EventID][j].Priority
	})

	em.mutex.Unlock() // Unlock before calling RegisterHandler

	// Register the actual handler function (which triggers the engine)
	// RegisterHandler handles its own locking
	err := em.RegisterHandler(binding)
	if err != nil {
		// If handler registration fails, we should ideally roll back the binding addition.
		fmt.Printf("CRITICAL: Failed to register handler for binding %s after adding binding: %v. Binding is orphaned.\n", binding.ID, err)
		// Attempt rollback (careful with locks)
		em.mutex.Lock()
		if bindings, ok := em.bindings[binding.EventID]; ok {
			for i, b := range bindings {
				if b.ID == binding.ID {
					em.bindings[binding.EventID] = append(bindings[:i], bindings[i+1:]...)
					break
				}
			}
		}
		em.mutex.Unlock()
		return fmt.Errorf("failed to register handler after adding binding: %w", err)
	}

	return nil
}

// DispatchEvent dispatches an event to all matching handlers (which trigger node execution)
func (em *EventManager) DispatchEvent(request EventDispatchRequest) []error {
	em.mutex.RLock()

	event, exists := em.definitions[request.EventID]
	if !exists {
		em.mutex.RUnlock()
		return []error{fmt.Errorf("event with ID %s does not exist", request.EventID)}
	}

	// Get a snapshot of bindings and handlers under read lock
	eventBindings := make([]EventBinding, 0)
	if bindings, ok := em.bindings[request.EventID]; ok {
		eventBindings = append(eventBindings, bindings...) // Copy slice
	}

	handlersSnapshot := make(map[string]EventHandlerFunc)
	for id, handler := range em.handlerFuncs {
		handlersSnapshot[id] = handler
	}

	em.mutex.RUnlock() // Release lock before validation and execution

	// Validate parameters
	validationErrors := validateParameters(event, request.Parameters)
	if len(validationErrors) > 0 {
		return validationErrors // Don't dispatch if params are invalid
	}

	var dispatchErrors []error

	// Execute handlers based on (copied and sorted) bindings
	for _, binding := range eventBindings {
		if !binding.Enabled {
			continue
		}

		handlerFunc, exists := handlersSnapshot[binding.ID]
		if !exists {
			// This indicates an inconsistency, likely RegisterHandler failed after BindEvent
			dispatchErrors = append(dispatchErrors, fmt.Errorf("handler function for binding %s not found during dispatch", binding.ID))
			continue
		}

		// Create handler context for this specific binding
		// Ensure parameters are copied if they might be modified by handlers concurrently
		paramsCopy := make(map[string]types.Value, len(request.Parameters))
		for k, v := range request.Parameters {
			paramsCopy[k] = v // Shallow copy is usually fine for Value struct
		}

		handlerCtx := EventHandlerContext{
			EventID:     request.EventID,
			Parameters:  paramsCopy, // Use copy
			SourceID:    request.SourceID,
			BlueprintID: request.BlueprintID, // Source blueprint ID
			BindingID:   binding.ID,
			ExecutionID: request.ExecutionID,
			HandlerID:   binding.HandlerID, // Target node ID
			Timestamp:   request.Timestamp,
		}

		// Execute the handler function (which triggers the engine) synchronously
		err := handlerFunc(handlerCtx)
		if err != nil {
			// Collect errors from triggering the node execution
			dispatchErrors = append(dispatchErrors, fmt.Errorf("error executing handler for binding %s: %w", binding.ID, err))
			// Decide if one handler error should stop others. For now, continue.
		}
	}

	return dispatchErrors
}

// validateParameters validates event parameters against the event definition
func validateParameters(event EventDefinition, params map[string]types.Value) []error {
	var errors []error
	// Implementation remains the same as previous version
	providedParams := make(map[string]bool)
	for name := range params {
		providedParams[name] = true
	}
	for _, paramDef := range event.Parameters {
		value, exists := params[paramDef.Name]
		providedParams[paramDef.Name] = true
		if !exists {
			if !paramDef.Optional {
				errors = append(errors, fmt.Errorf("required parameter '%s' missing for event '%s'", paramDef.Name, event.ID))
			}
		} else {
			// Allow Any type to match anything, otherwise check type ID
			if paramDef.Type != nil && paramDef.Type.ID != types.PinTypes.Any.ID && value.Type.ID != paramDef.Type.ID {
				errors = append(errors, fmt.Errorf("parameter '%s' has incorrect type for event '%s': expected %s, got %s",
					paramDef.Name, event.ID, paramDef.Type.ID, value.Type.ID))
			}
		}
	}
	// Optional: Check for extra parameters not defined in the event
	// for name := range params { ... }
	return errors
}

// GetSystemEventID returns the event ID for a system event type
func (em *EventManager) GetSystemEventID(eventType core.SystemEventType) (string, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	id, exists := em.systemEvents[eventType]
	return id, exists
}

// GetEventDefinition returns the definition for an event
func (em *EventManager) GetEventDefinition(eventID string) (EventDefinition, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	def, exists := em.definitions[eventID]
	return def, exists
}

// GetBlueprintEvents returns all events defined in a blueprint
func (em *EventManager) GetBlueprintEvents(blueprintID string) []EventDefinition {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	eventIDs, exists := em.blueprintEvents[blueprintID]
	if !exists {
		return nil
	}
	events := make([]EventDefinition, 0, len(eventIDs))
	for _, id := range eventIDs {
		if def, exists := em.definitions[id]; exists {
			events = append(events, def)
		}
	}
	return events
}

// RemoveBinding removes an event binding and its associated handler function
func (em *EventManager) RemoveBinding(bindingID string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	found := false
	for eventID, bindings := range em.bindings {
		for i, binding := range bindings {
			if binding.ID == bindingID {
				em.bindings[eventID] = append(bindings[:i], bindings[i+1:]...)
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if found {
		delete(em.handlerFuncs, bindingID) // Remove the generated handler func
	}
}

// ClearBindings removes all bindings for a given blueprint and their handlers
func (em *EventManager) ClearBindings(blueprintID string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	bindingsToRemove := make(map[string]struct{})
	for eventID, bindings := range em.bindings {
		updatedBindings := make([]EventBinding, 0, len(bindings))
		for _, binding := range bindings {
			if binding.BlueprintID != blueprintID {
				updatedBindings = append(updatedBindings, binding)
			} else {
				bindingsToRemove[binding.ID] = struct{}{}
			}
		}
		if len(updatedBindings) < len(bindings) {
			em.bindings[eventID] = updatedBindings
		}
	}
	for bindingID := range bindingsToRemove {
		delete(em.handlerFuncs, bindingID) // Remove the generated handler funcs
	}
}

// GetAllEvents returns all registered events
func (em *EventManager) GetAllEvents() []EventDefinition {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	events := make([]EventDefinition, 0, len(em.definitions))
	for _, def := range em.definitions {
		events = append(events, def)
	}
	return events
}

// GetEventBindings returns all bindings for a given event
func (em *EventManager) GetEventBindings(eventID string) ([]EventBinding, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	bindings, exists := em.bindings[eventID]
	if !exists || len(bindings) == 0 {
		return nil, false
	}
	result := make([]EventBinding, len(bindings))
	copy(result, bindings)
	return result, true
}

// --- Adapter for core.EventManagerInterface ---

// AsEventManagerInterface converts the EventManager to the core interface
func (em *EventManager) AsEventManagerInterface() core.EventManagerInterface {
	return &eventManagerAdapter{manager: em}
}

// eventManagerAdapter adapts EventManager to core.EventManagerInterface
type eventManagerAdapter struct {
	manager *EventManager
}

// DispatchEvent adapts the core interface call to the local DispatchEvent
// The interface now expects core.EventDispatchRequest directly.
func (a *eventManagerAdapter) DispatchEvent(request core.EventDispatchRequest) []error {
	// Convert core type to local event type
	localReq := EventDispatchRequest{
		EventID:     request.EventID,
		Parameters:  request.Parameters,
		SourceID:    request.SourceID,
		BlueprintID: request.BlueprintID,
		ExecutionID: request.ExecutionID,
		Timestamp:   request.Timestamp,
	}
	// Call the manager's DispatchEvent with the local type
	return a.manager.DispatchEvent(localReq)
}

// RegisterEventHandler - Adapter method. No longer supported directly.
func (a *eventManagerAdapter) RegisterEventHandler(eventID string, handler interface{}) error {
	return fmt.Errorf("RegisterEventHandler via core interface is not supported; use BindEvent")
}

// UnregisterEventHandler - Adapter method. No longer supported directly.
func (a *eventManagerAdapter) UnregisterEventHandler(eventID string, handlerID string) error {
	return fmt.Errorf("UnregisterEventHandler via core interface is not supported; use RemoveBinding")
}

// GetEventHandlers - Adapter method. No longer supported directly.
func (a *eventManagerAdapter) GetEventHandlers(eventID string) []interface{} {
	return []interface{}{} // Returns empty slice as EventHandler interfaces are not stored
}
