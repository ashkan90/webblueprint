package core

import (
	"time"
	"webblueprint/internal/bperrors"
	"webblueprint/internal/node"
	"webblueprint/internal/types"
)

// EventManagerInterface defines the interface for event management
// This is in the core package to avoid import cycles
type EventManagerInterface interface {
	// DispatchEvent triggers handlers bound to the specified event.
	// It uses the specific core.EventDispatchRequest type.
	DispatchEvent(request EventDispatchRequest) []error

	// NOTE: Methods related to the old EventHandler interface are removed.
	// Registration and unregistration are now handled implicitly via
	// BindEvent and RemoveBinding in the concrete implementation, which
	// manage the associated handler functions.
	// RegisterEventHandler(eventID string, handler interface{}) error
	// UnregisterEventHandler(eventID string, handlerID string) error
	// GetEventHandlers(eventID string) []interface{}
}

// EventHandlerContext contains context information for event handlers
type EventHandlerContext struct {
	// The ID of the event being handled
	EventID string

	// Parameters passed with the event
	Parameters map[string]types.Value

	// The source node/component that triggered the event
	SourceID string

	// Blueprint ID associated with this event
	BlueprintID string

	// Execution ID for tracking the execution flow
	ExecutionID string

	// Handler ID for the node handling this event
	HandlerID string

	// Binding ID to identify the event binding
	BindingID string

	// The timestamp when the event was triggered
	Timestamp time.Time
}

// EventDispatchRequest contains information for dispatching an event
type EventDispatchRequest struct {
	EventID     string
	Parameters  map[string]types.Value
	SourceID    string
	BlueprintID string
	ExecutionID string
	Timestamp   time.Time
}

// EventHandler defines a handler for events
type EventHandler interface {
	HandleEvent(event EventDispatchRequest) error
	GetHandlerID() string
}

// EventAwareContext extends ExecutionContext with event capabilities
type EventAwareContext interface {
	node.ExecutionContext

	// Event management
	GetEventManager() EventManagerInterface
	DispatchEvent(eventID string, params map[string]types.Value) error
	IsEventHandlerActive() bool
	GetEventHandlerContext() *EventHandlerContext
}

// ErrorAwareContext extends ExecutionContext with error handling capabilities
type ErrorAwareContext interface {
	node.ExecutionContext

	// Error handling methods
	ReportError(errType bperrors.ErrorType, code bperrors.BlueprintErrorCode, message string, originalErr error) *bperrors.BlueprintError
	AttemptRecovery(err *bperrors.BlueprintError) (bool, map[string]interface{})
	GetErrorSummary() map[string]interface{}

	// Default value management
	GetDefaultValue(pinType *types.PinType) (types.Value, bool)
}

type ActorAwareContext interface {
	node.ExecutionContext

	SetInputPinActive(pinID string)
}

// ContextProvider defines an interface for providing execution contexts with additional capabilities
type ContextProvider interface {
	// CreateEventAwareContext creates a context with event capabilities
	CreateEventAwareContext(
		baseCtx node.ExecutionContext,
		isEventHandler bool,
		eventHandlerContext *EventHandlerContext,
	) node.ExecutionContext

	CreateErrorAwareContext(
		baseCtx node.ExecutionContext,
	) node.ExecutionContext

	CreateActorContext(
		baseCtx node.ExecutionContext,
	) node.ExecutionContext

	CreateFunctionContext(
		baseCtx node.ExecutionContext,
		functionID string,
	) node.ExecutionContext
}

// EngineController defines an interface for controlling engine execution,
// specifically for triggering event handler nodes.
type EngineController interface {
	TriggerNodeExecution(blueprintID string, nodeID string, triggerContext EventHandlerContext) error
}
