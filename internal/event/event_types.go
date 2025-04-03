package event

import (
	"time"
	"webblueprint/internal/core"
	"webblueprint/internal/types"
)

const (
	EventTypeInitialize  core.SystemEventType = "OnInitialize"  // Blueprint initialization
	EventTypeShutdown    core.SystemEventType = "OnShutdown"    // Blueprint shutdown
	EventTypeTimer       core.SystemEventType = "OnTimer"       // Periodic timer event
	EventTypeWebhook     core.SystemEventType = "OnWebhook"     // External webhook received
	EventTypeAPIResponse core.SystemEventType = "OnAPIResponse" // API response received
	EventTypeDataChanged core.SystemEventType = "OnDataChanged" // Data variable changed
	EventTypeNodeCreated core.SystemEventType = "OnNodeCreated" // New node created
	EventTypeNodeDeleted core.SystemEventType = "OnNodeDeleted" // Node deleted
	EventTypeError       core.SystemEventType = "OnError"       // Error occurred
)

// EventHandler is the interface that must be implemented by all event handlers
type EventHandler interface {
	HandleEvent(event EventDispatchRequest) error
	GetHandlerID() string
}

// EventDefinition represents a blueprint event
type EventDefinition struct {
	ID          string           `json:"id"`          // Unique identifier for the event
	Name        string           `json:"name"`        // Human-readable name
	Description string           `json:"description"` // Description of what the event does
	Parameters  []EventParameter `json:"parameters"`  // Parameters that can be passed with the event
	Category    string           `json:"category"`    // Category for organization (System, UI, Custom, etc.)
	BlueprintID string           `json:"blueprintId"` // ID of the blueprint that defined this event (empty for system events)
	CreatedAt   time.Time        `json:"createdAt"`   // When the event was defined
}

// EventParameter represents a parameter that can be passed with an event
type EventParameter struct {
	Name        string         // Parameter name
	Type        *types.PinType // Parameter data type
	Description string         // Description of the parameter
	Optional    bool           // Whether the parameter is optional
	Default     interface{}    // Default value for optional parameters
}

// EventBinding represents a connection between an event source and a handler
type EventBinding struct {
	ID          string    // Unique identifier for this binding
	EventID     string    // ID of the event being bound to
	HandlerID   string    // ID of the node that handles the event
	HandlerType string    // Type of the handler node
	BlueprintID string    // ID of the blueprint containing the handler
	Priority    int       // Priority (higher numbers execute first)
	CreatedAt   time.Time // When the binding was created
	Enabled     bool      // Whether the binding is active
}

// EventDispatchRequest represents a request to dispatch an event
type EventDispatchRequest struct {
	EventID     string                 // ID of the event to dispatch
	Parameters  map[string]types.Value // Parameter values to pass
	SourceID    string                 // ID of the node or entity that triggered the event
	BlueprintID string                 // ID of the blueprint dispatching the event
	ExecutionID string                 // Current execution ID
	Timestamp   time.Time              // When the event was dispatched
}

// EventHandlerContext provides context for an event handler
type EventHandlerContext struct {
	EventID     string                 // ID of the event being handled
	Parameters  map[string]types.Value // Parameter values passed with the event
	SourceID    string                 // ID of the node or entity that triggered the event
	BlueprintID string                 // ID of the source blueprint
	BindingID   string                 // ID of the binding being executed
	ExecutionID string                 // Current execution ID
	HandlerID   string                 // ID of the handler node
	Timestamp   time.Time              // When the event was dispatched
}

// EventHandlerFunc is a function that handles an event
type EventHandlerFunc func(context EventHandlerContext) error

// EventManagerInterface defines the interface for the event manager within the event package.
// It reflects the methods available on the concrete *EventManager.
type EventManagerInterface interface {
	// Core dispatching
	DispatchEvent(request EventDispatchRequest) []error

	// Event Definition Management
	RegisterEvent(event EventDefinition) error
	UnregisterEventDefinition(eventID string) error // Added this based on API handler needs
	GetEventDefinition(eventID string) (EventDefinition, bool)
	GetAllEvents() []EventDefinition
	GetBlueprintEvents(blueprintID string) []EventDefinition
	GetSystemEventID(eventType core.SystemEventType) (string, bool)

	// Binding and Handler Management (using EventHandlerFunc)
	BindEvent(binding EventBinding) error
	RegisterHandler(binding EventBinding) error // Changed signature to match implementation
	RemoveBinding(bindingID string)
	ClearBindings(blueprintID string)
	GetEventBindings(eventID string) ([]EventBinding, bool)

	// Adapter to the core interface (for external use)
	AsEventManagerInterface() core.EventManagerInterface
}
