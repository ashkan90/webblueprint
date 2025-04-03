package event

import (
	"fmt"
	"webblueprint/internal/core"
)

// EventListener interface for subscribing to event notifications
type EventListener interface {
	OnEventDispatched(eventID string, request EventDispatchRequest)
	OnEventBound(binding EventBinding)
	OnEventUnbound(bindingID string)
}

// defaultLogger is a simple default logger implementation
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, fields map[string]interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *defaultLogger) Info(msg string, fields map[string]interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *defaultLogger) Warn(msg string, fields map[string]interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, fields)
}

func (l *defaultLogger) Error(msg string, fields map[string]interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *defaultLogger) Opts(options map[string]interface{}) {
	// No-op for default logger
}

func ExtractEventManager(e core.EventManagerInterface) *EventManager {
	m, ok := e.(*eventManagerAdapter)
	if !ok {
		return nil
	}

	return m.manager
}

// EventManager manages event registration, binding, and dispatching
//type EventManager struct {
//	definitions     map[string]EventDefinition  // EventID -> EventDefinition
//	bindings        map[string][]EventBinding   // EventID -> []EventBinding
//	handlers        map[string]EventHandlerFunc // BindingID -> HandlerFunc
//	systemEvents    map[SystemEventType]string  // SystemEventType -> EventID
//	blueprintEvents map[string][]string         // BlueprintID -> []EventID
//	listeners       []EventListener             // Event listeners for notifications
//	logger          Logger                      // Logger for event operations
//	mutex           sync.RWMutex
//}
//
//// NewEventManager creates a new event manager
//func NewEventManager() *EventManager {
//	manager := &EventManager{
//		definitions:     make(map[string]EventDefinition),
//		bindings:        make(map[string][]EventBinding),
//		handlers:        make(map[string]EventHandlerFunc),
//		systemEvents:    make(map[SystemEventType]string),
//		blueprintEvents: make(map[string][]string),
//		listeners:       make([]EventListener, 0),
//		logger:          &defaultLogger{},
//	}
//
//	// Register built-in system events
//	manager.registerSystemEvents()
//
//	return manager
//}
//
//// SetLogger sets the logger for the event manager
//func (em *EventManager) SetLogger(logger Logger) {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//	em.logger = logger
//}
//
//// AddEventListener adds a listener for event notifications
//func (em *EventManager) AddEventListener(listener EventListener) {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//	em.listeners = append(em.listeners, listener)
//}
//
//// RemoveEventListener removes a listener
//func (em *EventManager) RemoveEventListener(listener EventListener) {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//
//	// Find and remove the listener
//	for i, l := range em.listeners {
//		if l == listener {
//			em.listeners = append(em.listeners[:i], em.listeners[i+1:]...)
//			return
//		}
//	}
//}
//
//// notifyListeners notifies all listeners about an event
//func (em *EventManager) notifyListeners(eventID string, request EventDispatchRequest) {
//	for _, listener := range em.listeners {
//		listener.OnEventDispatched(eventID, request)
//	}
//}
//
//// notifyBindingListeners notifies all listeners about a binding event
//func (em *EventManager) notifyBindingListeners(binding EventBinding) {
//	for _, listener := range em.listeners {
//		listener.OnEventBound(binding)
//	}
//}
//
//// notifyUnbindListeners notifies all listeners about an unbind event
//func (em *EventManager) notifyUnbindListeners(bindingID string) {
//	for _, listener := range em.listeners {
//		listener.OnEventUnbound(bindingID)
//	}
//}
//
//// RegisterSystemEvent registers a system event type with a specific event ID
//func (em *EventManager) RegisterSystemEvent(eventType SystemEventType, eventID string) {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//	em.systemEvents[eventType] = eventID
//}
//
//// GetAllBindings returns all event bindings
//func (em *EventManager) GetAllBindings() []EventBinding {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	var allBindings []EventBinding
//	for _, bindings := range em.bindings {
//		allBindings = append(allBindings, bindings...)
//	}
//
//	return allBindings
//}
//
//// registerSystemEvents registers the built-in system events
//func (em *EventManager) registerSystemEvents() {
//	// Define system events
//	initEvent := EventDefinition{
//		ID:          "system.initialize",
//		Name:        string(EventTypeInitialize),
//		Description: "Triggered when a blueprint starts execution",
//		Parameters: []EventParameter{
//			{
//				Name:        "blueprintID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the blueprint being initialized",
//				Optional:    false,
//			},
//			{
//				Name:        "executionID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the execution instance",
//				Optional:    false,
//			},
//		},
//		Category:  "System",
//		CreatedAt: time.Now(),
//	}
//
//	shutdownEvent := EventDefinition{
//		ID:          "system.shutdown",
//		Name:        string(EventTypeShutdown),
//		Description: "Triggered when a blueprint execution ends",
//		Parameters: []EventParameter{
//			{
//				Name:        "blueprintID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the blueprint being shut down",
//				Optional:    false,
//			},
//			{
//				Name:        "executionID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the execution instance",
//				Optional:    false,
//			},
//			{
//				Name:        "success",
//				Type:        types.PinTypes.Boolean,
//				Description: "Whether the execution completed successfully",
//				Optional:    false,
//			},
//			{
//				Name:        "errorMessage",
//				Type:        types.PinTypes.String,
//				Description: "Error message if execution failed",
//				Optional:    true,
//			},
//		},
//		Category:  "System",
//		CreatedAt: time.Now(),
//	}
//
//	timerEvent := EventDefinition{
//		ID:          "system.timer",
//		Name:        string(EventTypeTimer),
//		Description: "Triggered periodically based on a timer",
//		Parameters: []EventParameter{
//			{
//				Name:        "blueprintID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the blueprint",
//				Optional:    false,
//			},
//			{
//				Name:        "executionID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the execution instance",
//				Optional:    false,
//			},
//			{
//				Name:        "interval",
//				Type:        types.PinTypes.Number,
//				Description: "Timer interval in milliseconds",
//				Optional:    false,
//			},
//			{
//				Name:        "count",
//				Type:        types.PinTypes.Number,
//				Description: "Number of times the timer has fired",
//				Optional:    false,
//			},
//			{
//				Name:        "timerID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the timer",
//				Optional:    false,
//			},
//		},
//		Category:  "System",
//		CreatedAt: time.Now(),
//	}
//
//	webhookEvent := EventDefinition{
//		ID:          "system.webhook",
//		Name:        string(EventTypeWebhook),
//		Description: "Triggered when a webhook is received",
//		Parameters: []EventParameter{
//			{
//				Name:        "blueprintID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the blueprint",
//				Optional:    false,
//			},
//			{
//				Name:        "executionID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the execution instance",
//				Optional:    false,
//			},
//			{
//				Name:        "path",
//				Type:        types.PinTypes.String,
//				Description: "Webhook path",
//				Optional:    false,
//			},
//			{
//				Name:        "method",
//				Type:        types.PinTypes.String,
//				Description: "HTTP method",
//				Optional:    false,
//			},
//			{
//				Name:        "data",
//				Type:        types.PinTypes.Object,
//				Description: "Webhook payload data",
//				Optional:    true,
//			},
//			{
//				Name:        "headers",
//				Type:        types.PinTypes.Object,
//				Description: "HTTP headers",
//				Optional:    true,
//			},
//		},
//		Category:  "System",
//		CreatedAt: time.Now(),
//	}
//
//	// Error event
//	errorEvent := EventDefinition{
//		ID:          "system.error",
//		Name:        string(EventTypeError),
//		Description: "Triggered when an error occurs during blueprint execution",
//		Parameters: []EventParameter{
//			{
//				Name:        "blueprintID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the blueprint",
//				Optional:    false,
//			},
//			{
//				Name:        "executionID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the execution instance",
//				Optional:    false,
//			},
//			{
//				Name:        "nodeID",
//				Type:        types.PinTypes.String,
//				Description: "ID of the node where the error occurred",
//				Optional:    true,
//			},
//			{
//				Name:        "errorMessage",
//				Type:        types.PinTypes.String,
//				Description: "Error message",
//				Optional:    false,
//			},
//			{
//				Name:        "errorDetails",
//				Type:        types.PinTypes.Object,
//				Description: "Additional error details",
//				Optional:    true,
//			},
//		},
//		Category:  "System",
//		CreatedAt: time.Now(),
//	}
//
//	// Register system events
//	em.RegisterEvent(initEvent)
//	em.RegisterEvent(shutdownEvent)
//	em.RegisterEvent(timerEvent)
//	em.RegisterEvent(webhookEvent)
//	em.RegisterEvent(errorEvent)
//
//	// Map system event types to event IDs
//	em.systemEvents[EventTypeInitialize] = initEvent.ID
//	em.systemEvents[EventTypeShutdown] = shutdownEvent.ID
//	em.systemEvents[EventTypeTimer] = timerEvent.ID
//	em.systemEvents[EventTypeWebhook] = webhookEvent.ID
//	em.systemEvents[EventTypeError] = errorEvent.ID
//}
//
//// RegisterEvent registers a new event definition
//func (em *EventManager) RegisterEvent(event EventDefinition) error {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//
//	// Check if event already exists
//	if _, exists := em.definitions[event.ID]; exists {
//		return fmt.Errorf("event with ID %s already exists", event.ID)
//	}
//
//	// Store the event definition
//	em.definitions[event.ID] = event
//
//	// Initialize empty bindings list
//	em.bindings[event.ID] = make([]EventBinding, 0)
//
//	// Add to blueprint events if not a system event
//	if event.BlueprintID != "" {
//		em.blueprintEvents[event.BlueprintID] = append(
//			em.blueprintEvents[event.BlueprintID],
//			event.ID,
//		)
//	}
//
//	// Log successful registration
//	em.logger.Info(fmt.Sprintf("Event registered: %s", event.ID), map[string]interface{}{
//		"eventName":   event.Name,
//		"category":    event.Category,
//		"blueprintID": event.BlueprintID,
//		"paramCount":  len(event.Parameters),
//	})
//
//	return nil
//}
//
//// BindEvent creates a binding between an event and a handler
//func (em *EventManager) BindEvent(binding EventBinding) error {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//
//	// Check if event exists
//	if _, exists := em.definitions[binding.EventID]; !exists {
//		return fmt.Errorf("event with ID %s does not exist", binding.EventID)
//	}
//
//	// Add binding to the event's bindings
//	em.bindings[binding.EventID] = append(em.bindings[binding.EventID], binding)
//
//	// Sort bindings by priority (higher priority first)
//	sort.Slice(em.bindings[binding.EventID], func(i, j int) bool {
//		return em.bindings[binding.EventID][i].Priority > em.bindings[binding.EventID][j].Priority
//	})
//
//	// Log successful binding
//	em.logger.Info(fmt.Sprintf("Event bound: %s -> %s", binding.EventID, binding.HandlerID), map[string]interface{}{
//		"bindingID":   binding.ID,
//		"blueprintID": binding.BlueprintID,
//		"priority":    binding.Priority,
//	})
//
//	// Notify listeners
//	for _, listener := range em.listeners {
//		listener.OnEventBound(binding)
//	}
//
//	return nil
//}
//
//// RegisterHandler registers a handler function for a binding
//func (em *EventManager) RegisterHandler(bindingID string, handler EventHandlerFunc) {
//	em.mutex.Lock()
//	defer em.mutex.Unlock()
//
//	em.handlers[bindingID] = handler
//
//	// Log handler registration
//	em.logger.Debug(fmt.Sprintf("Handler registered for binding: %s", bindingID), nil)
//}
//
//// DispatchEvent dispatches an event to all matching handlers
//func (em *EventManager) DispatchEvent(request EventDispatchRequest) []error {
//	em.mutex.RLock()
//
//	// Check if event exists
//	event, exists := em.definitions[request.EventID]
//	if !exists {
//		em.mutex.RUnlock()
//		em.logger.Error(fmt.Sprintf("Failed to dispatch event: %s does not exist", request.EventID), nil)
//		return []error{fmt.Errorf("event with ID %s does not exist", request.EventID)}
//	}
//
//	// Get bindings for this event
//	eventBindings := make([]EventBinding, len(em.bindings[request.EventID]))
//	copy(eventBindings, em.bindings[request.EventID])
//
//	// Copy handlers to avoid holding the lock during execution
//	handlers := make(map[string]EventHandlerFunc, len(em.handlers))
//	for id, handler := range em.handlers {
//		handlers[id] = handler
//	}
//
//	// Copy listeners to avoid holding the lock when notifying
//	listeners := make([]EventListener, len(em.listeners))
//	copy(listeners, em.listeners)
//
//	// Get logger for consistent logging
//	logger := em.logger
//
//	em.mutex.RUnlock()
//
//	// Validate parameters against event definition
//	errors := em.validateParameters(event, request.Parameters)
//	if len(errors) > 0 {
//		logger.Error("Parameter validation failed for event", map[string]interface{}{
//			"eventID": request.EventID,
//			"errors":  errors,
//		})
//		return errors
//	}
//
//	// Log event dispatch
//	logger.Info(fmt.Sprintf("Dispatching event: %s", request.EventID), map[string]interface{}{
//		"sourceID":    request.SourceID,
//		"blueprintID": request.BlueprintID,
//		"executionID": request.ExecutionID,
//		"bindings":    len(eventBindings),
//	})
//
//	// Execute handlers
//	for _, binding := range eventBindings {
//		if !binding.Enabled {
//			logger.Debug(fmt.Sprintf("Skipping disabled binding: %s", binding.ID), nil)
//			continue
//		}
//
//		handler, exists := handlers[binding.ID]
//		if !exists {
//			errors = append(errors, fmt.Errorf("handler for binding %s not found", binding.ID))
//			logger.Error(fmt.Sprintf("Handler not found for binding: %s", binding.ID), nil)
//			continue
//		}
//
//		// Create handler context
//		ctx := EventHandlerContext{
//			EventID:     request.EventID,
//			Parameters:  request.Parameters,
//			SourceID:    request.SourceID,
//			BlueprintID: request.BlueprintID,
//			BindingID:   binding.ID,
//			ExecutionID: request.ExecutionID,
//			HandlerID:   binding.HandlerID,
//		}
//
//		// Execute handler (in real implementation, this might be done asynchronously)
//		go func(handler EventHandlerFunc, ctx EventHandlerContext, binding EventBinding) {
//			logger.Debug(fmt.Sprintf("Executing handler for binding: %s", binding.ID), map[string]interface{}{
//				"handlerID":   binding.HandlerID,
//				"handlerType": binding.HandlerType,
//			})
//
//			if err := handler(ctx); err != nil {
//				logger.Error(fmt.Sprintf("Error executing event handler: %v", err), map[string]interface{}{
//					"bindingID":   binding.ID,
//					"handlerID":   binding.HandlerID,
//					"eventID":     ctx.EventID,
//					"blueprintID": binding.BlueprintID,
//				})
//			}
//		}(handler, ctx, binding)
//	}
//
//	// Notify listeners
//	for _, listener := range listeners {
//		go listener.OnEventDispatched(request.EventID, request)
//	}
//
//	return errors
//}
//
//// validateParameters validates event parameters against the event definition
//func (em *EventManager) validateParameters(event EventDefinition, params map[string]types.Value) []error {
//	var errors []error
//
//	// Check for required parameters
//	for _, param := range event.Parameters {
//		if !param.Optional {
//			value, exists := params[param.Name]
//			if !exists {
//				errors = append(errors, fmt.Errorf("required parameter %s missing", param.Name))
//				continue
//			}
//
//			// Validate parameter type
//			if value.Type.ID != param.Type.ID {
//				errors = append(errors, fmt.Errorf("parameter %s has incorrect type: expected %s, got %s",
//					param.Name, param.Type.ID, value.Type.ID))
//			}
//		}
//	}
//
//	return errors
//}
//
//// GetSystemEventID returns the event ID for a system event type
//func (em *EventManager) GetSystemEventID(eventType SystemEventType) (string, bool) {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	id, exists := em.systemEvents[eventType]
//	return id, exists
//}
//
//// GetEventDefinition returns the definition for an event
//func (em *EventManager) GetEventDefinition(eventID string) (EventDefinition, bool) {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	def, exists := em.definitions[eventID]
//	return def, exists
//}
//
//// GetBlueprintEvents returns all events defined in a blueprint
//func (em *EventManager) GetBlueprintEvents(blueprintID string) []EventDefinition {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	eventIDs, exists := em.blueprintEvents[blueprintID]
//	if !exists {
//		return nil
//	}
//
//	events := make([]EventDefinition, 0, len(eventIDs))
//	for _, id := range eventIDs {
//		if def, exists := em.definitions[id]; exists {
//			events = append(events, def)
//		}
//	}
//
//	return events
//}
//
//// RemoveBinding removes an event binding
//func (em *EventManager) RemoveBinding(bindingID string) {
//	em.mutex.Lock()
//
//	// Find and remove the binding
//	var removedBinding *EventBinding
//	for eventID, bindings := range em.bindings {
//		for i, binding := range bindings {
//			if binding.ID == bindingID {
//				// Make a copy of the binding before removing it
//				bindingCopy := binding
//				removedBinding = &bindingCopy
//
//				// Remove binding from slice
//				em.bindings[eventID] = append(bindings[:i], bindings[i+1:]...)
//
//				// Remove handler
//				delete(em.handlers, bindingID)
//				break
//			}
//		}
//		if removedBinding != nil {
//			break
//		}
//	}
//
//	// Get listeners to notify after releasing the lock
//	var listeners []EventListener
//	if removedBinding != nil {
//		listeners = make([]EventListener, len(em.listeners))
//		copy(listeners, em.listeners)
//
//		em.logger.Info(fmt.Sprintf("Binding removed: %s", bindingID), map[string]interface{}{
//			"eventID":     removedBinding.EventID,
//			"handlerID":   removedBinding.HandlerID,
//			"blueprintID": removedBinding.BlueprintID,
//		})
//	}
//
//	em.mutex.Unlock()
//
//	// Notify listeners about unbinding
//	if removedBinding != nil && len(listeners) > 0 {
//		for _, listener := range listeners {
//			listener.OnEventUnbound(bindingID)
//		}
//	}
//}
//
//// ClearBindings removes all bindings for a given blueprint
//func (em *EventManager) ClearBindings(blueprintID string) {
//	em.mutex.Lock()
//
//	// Find all bindings for this blueprint
//	removedBindings := make([]string, 0)
//	for eventID, bindings := range em.bindings {
//		updatedBindings := make([]EventBinding, 0)
//		for _, binding := range bindings {
//			if binding.BlueprintID != blueprintID {
//				updatedBindings = append(updatedBindings, binding)
//			} else {
//				// Add to removed list
//				removedBindings = append(removedBindings, binding.ID)
//
//				// Remove handler for this binding
//				delete(em.handlers, binding.ID)
//			}
//		}
//		em.bindings[eventID] = updatedBindings
//	}
//
//	// Get listeners to notify after releasing the lock
//	var listeners []EventListener
//	if len(removedBindings) > 0 {
//		listeners = make([]EventListener, len(em.listeners))
//		copy(listeners, em.listeners)
//
//		em.logger.Info(fmt.Sprintf("Cleared %d bindings for blueprint: %s", len(removedBindings), blueprintID), nil)
//	}
//
//	em.mutex.Unlock()
//
//	// Notify listeners about unbinding
//	if len(removedBindings) > 0 && len(listeners) > 0 {
//		for _, bindingID := range removedBindings {
//			for _, listener := range listeners {
//				listener.OnEventUnbound(bindingID)
//			}
//		}
//	}
//}
//
//// GetAllEvents returns all registered events
//func (em *EventManager) GetAllEvents() []EventDefinition {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	events := make([]EventDefinition, 0, len(em.definitions))
//	for _, def := range em.definitions {
//		events = append(events, def)
//	}
//
//	return events
//}
//
//// GetEventBindings returns all bindings for a given event
//func (em *EventManager) GetEventBindings(eventID string) ([]EventBinding, bool) {
//	em.mutex.RLock()
//	defer em.mutex.RUnlock()
//
//	bindings, exists := em.bindings[eventID]
//	if !exists {
//		return nil, false
//	}
//
//	// Make a copy to avoid concurrent modification
//	result := make([]EventBinding, len(bindings))
//	copy(result, bindings)
//
//	return result, true
//}
